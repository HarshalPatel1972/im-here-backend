package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/guardian/im-here/internal/config"
	"github.com/guardian/im-here/internal/detector"
	"github.com/guardian/im-here/internal/models"
	"github.com/guardian/im-here/internal/notifier"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Poller struct {
	cfg        *config.Config
	db         *pgxpool.Pool
	notifier   *notifier.Notifier
	client     *http.Client
	lastETag   string
	pollInt    time.Duration
}

func New(cfg *config.Config, db *pgxpool.Pool, notif *notifier.Notifier) *Poller {
	return &Poller{
		cfg:      cfg,
		db:       db,
		notifier: notif,
		client:  &http.Client{Timeout: 10 * time.Second},
		pollInt: time.Duration(cfg.PollIntervalSeconds) * time.Second,
	}
}

func (p *Poller) Start(ctx context.Context) {
	log.Println("Starting GitHub events poller...")
	
	// Create a ticker for polling based on dynamic poll interval
	for {
		select {
		case <-ctx.Done():
			return
		default:
			start := time.Now()
			p.poll(ctx)
			
			// Wait for the remainder of the interval, enforcing minimum from GitHub
			elapsed := time.Since(start)
			if elapsed < p.pollInt {
				select {
				case <-ctx.Done():
					return
				case <-time.After(p.pollInt - elapsed):
				}
			}
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/events", nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+p.cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	if p.lastETag != "" {
		req.Header.Set("If-None-Match", p.lastETag)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		log.Printf("Error polling events: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return
	}

	p.handleRateLimit(resp)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code from GitHub: %d", resp.StatusCode)
		return
	}

	p.lastETag = resp.Header.Get("ETag")

	// Update poll interval if GitHub requests it
	if pi := resp.Header.Get("X-Poll-Interval"); pi != "" {
		if sec, err := strconv.Atoi(pi); err == nil {
			p.pollInt = time.Duration(sec) * time.Second
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	var events []map[string]interface{}
	if err := json.Unmarshal(body, &events); err != nil {
		log.Printf("Error parsing events: %v", err)
		return
	}

	for _, ev := range events {
		if evType, ok := ev["type"].(string); ok && evType == "PushEvent" {
			p.processPushEvent(ctx, ev)
		}
	}
}

func (p *Poller) handleRateLimit(resp *http.Response) {
	if rem := resp.Header.Get("X-RateLimit-Remaining"); rem != "" {
		if r, err := strconv.Atoi(rem); err == nil && r < 100 {
			if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
				if rs, err := strconv.ParseInt(reset, 10, 64); err == nil {
					resetTime := time.Unix(rs, 0)
					wait := time.Until(resetTime)
					if wait > 0 {
						log.Printf("Rate limit nearly exhausted (%d remaining). Sleeping for %v", r, wait)
						time.Sleep(wait)
					}
				}
			}
		}
	}
}

func (p *Poller) processPushEvent(ctx context.Context, ev map[string]interface{}) {
	repo, ok := ev["repo"].(map[string]interface{})
	if !ok {
		return
	}
	repoName, _ := repo["name"].(string)

	payload, ok := ev["payload"].(map[string]interface{})
	if !ok {
		return
	}

	commits, ok := payload["commits"].([]interface{})
	if !ok {
		return
	}

	for _, c := range commits {
		cMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		sha, _ := cMap["sha"].(string)
		
		author, _ := cMap["author"].(map[string]interface{})
		email, _ := author["email"].(string)
		name, _ := author["name"].(string)

		p.fetchAndScanCommit(ctx, repoName, sha, email, name)
	}
}

func (p *Poller) fetchAndScanCommit(ctx context.Context, repo string, sha string, authorEmail string, authorName string) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", repo, sha)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", "Bearer "+p.cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github.v3.diff")

	resp, err := p.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	diffBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	diffContent := string(diffBytes)

	// Since the detector requires diff parsing per file, we'll pass the whole diff 
	// and extract a dummy file path, but a robust implementation parses diff hunks for file names.
	// For this task's scope, we parse "diff --git a/file b/file"
	
	files := strings.Split(diffContent, "diff --git ")
	for _, fDiff := range files {
		if fDiff == "" {
			continue
		}
		
		// extract file path
		firstLine := strings.SplitN(fDiff, "\n", 2)[0]
		parts := strings.Split(firstLine, " ")
		filePath := ""
		if len(parts) >= 2 {
			filePath = strings.TrimPrefix(parts[1], "b/")
		}

		findings := detector.Detect("diff --git "+fDiff, filePath)
		for _, f := range findings {
			f.RepoFullName = repo
			f.CommitSHA = sha
			f.CommitterEmail = authorEmail
			f.CommitterName = authorName
			
			p.persistFinding(ctx, f)
		}
	}
}

func (p *Poller) persistFinding(ctx context.Context, f models.Finding) {
	query := `
		INSERT INTO findings (repo_full_name, file_path, line_number, secret_type, commit_sha, committer_email, committer_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (repo_full_name, commit_sha, file_path, line_number) DO NOTHING
		RETURNING id
	`
	var id int64
	err := p.db.QueryRow(ctx, query,
		f.RepoFullName, f.FilePath, f.LineNumber, f.SecretType,
		f.CommitSHA, f.CommitterEmail, f.CommitterName,
	).Scan(&id)

	shaShort := f.CommitSHA
	if len(shaShort) > 7 {
		shaShort = shaShort[:7]
	}

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			log.Printf("[SKIP] already notified: %s/%s:%d", f.RepoFullName, f.FilePath, f.LineNumber)
		} else {
			log.Printf("Error saving finding: %v", err)
		}
		return
	}

	log.Printf("[FOUND] %s in %s/%s:%d commit:%s", f.SecretType, f.RepoFullName, f.FilePath, f.LineNumber, shaShort)

	f.ID = id
	go p.notifier.Notify(context.Background(), f)
}

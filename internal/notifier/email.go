package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/guardian/im-here/internal/models"
)

const emailTemplate = `Hey %s,

We noticed an exposed %s in a recent push to %s.

File: %s · Line: %d · Commit: %s

Don't panic. Nobody else has seen this. We found it first and
this message is only going to you.

HERE'S WHAT TO DO RIGHT NOW:

1. Revoke the key at your provider's dashboard
2. Generate a new one
3. Store it in a .env file (never commit .env)
4. Add .env to your .gitignore

Even free API keys are yours alone. The fix takes less than 5 minutes.

We've got you.

— I'm Here
A free, open-source guardian for developers.
We never display your keys to anyone.`

func (n *Notifier) sendEmail(ctx context.Context, f models.Finding) error {
	repoName := f.RepoFullName
	// repoName could be "owner/repo", email says "we found something in repo" 
	// The prompt uses {repo_name} for subject and {repo_full_name} for body.

	shaShort := f.CommitSHA
	if len(shaShort) > 7 {
		shaShort = shaShort[:7]
	}

	subject := fmt.Sprintf("I'm Here — we found something in %s", repoName)
	text := fmt.Sprintf(emailTemplate,
		f.CommitterName,
		f.SecretType,
		f.RepoFullName,
		f.FilePath,
		f.LineNumber,
		shaShort,
	)

	payload := map[string]interface{}{
		"from":    n.cfg.ResendFrom,
		"to":      []string{f.CommitterEmail},
		"subject": subject,
		"text":    text,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal resend payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create resend request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+n.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("resend api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

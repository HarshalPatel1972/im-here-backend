package notifier

import (
	"context"
	"log"
	"time"

	"github.com/guardian/im-here/internal/config"
	"github.com/guardian/im-here/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notifier struct {
	cfg *config.Config
	db  *pgxpool.Pool
}

func New(cfg *config.Config, db *pgxpool.Pool) *Notifier {
	return &Notifier{
		cfg: cfg,
		db:  db,
	}
}

func (n *Notifier) Notify(ctx context.Context, f models.Finding) error {
	var issueURL string
	var errIssue error
	var errEmail error

	// 1. Create GitHub Private Issue
	issueURL, errIssue = n.createPrivateIssue(ctx, f)
	if errIssue != nil {
		log.Printf("[ERROR] failed to create GitHub issue for %d: %v", f.ID, errIssue)
	} else {
		f.GithubIssueURL = issueURL
	}

	// 2. Send email via Resend
	errEmail = n.sendEmail(ctx, f)
	if errEmail != nil {
		log.Printf("[ERROR] failed to send email to %s for %d: %v", f.CommitterEmail, f.ID, errEmail)
	} else {
		f.EmailSent = true
	}

	if errIssue == nil || errEmail == nil {
		f.NotifiedAt = time.Now()
	}

	// 3. Persist to DB
	if f.NotifiedAt.IsZero() {
		// Both failed, do not mark as notified
		return errIssue
	}

	query := `
		UPDATE findings
		SET notified_at = $1, github_issue_url = $2, email_sent = $3
		WHERE id = $4
	`
	_, err := n.db.Exec(ctx, query, f.NotifiedAt,
		NullString(f.GithubIssueURL), f.EmailSent, f.ID)

	if err != nil {
		log.Printf("[ERROR] failed to update db notification status for %d: %v", f.ID, err)
		return err
	}

	log.Printf("[NOTIFY] Successfully notified %s (Issue: %v, Email: %v)", f.RepoFullName, errIssue == nil, errEmail == nil)
	return nil
}

func NullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

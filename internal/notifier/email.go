package notifier

import (
	"context"
	"fmt"
	"net/smtp"
        "strings"
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

        // Sanitize headers to prevent CRLF injection
        safeEmail := strings.ReplaceAll(strings.ReplaceAll(f.CommitterEmail, "\r", ""), "\n", "")
        safeName := strings.ReplaceAll(strings.ReplaceAll(f.CommitterName, "\r", ""), "\n", "")
        safeRepo := strings.ReplaceAll(strings.ReplaceAll(repoName, "\r", ""), "\n", "")

        subject := fmt.Sprintf("I'm Here — we found something in %s", safeRepo)
        text := fmt.Sprintf(emailTemplate,
                safeName,
                f.SecretType,
                f.RepoFullName,
                f.FilePath,
                f.LineNumber,
                shaShort,
        )

        // Combine headers and body
        msg := []byte(fmt.Sprintf("To: %s\r\n"+
                "From: I'm Here <%s>\r\n"+
                "Subject: %s\r\n"+
                "\r\n"+
                "%s\r\n", safeEmail, n.cfg.SMTPUser, subject, text))

        auth := smtp.PlainAuth("", n.cfg.SMTPUser, n.cfg.SMTPPassword, n.cfg.SMTPHost)
        addr := fmt.Sprintf("%s:%s", n.cfg.SMTPHost, n.cfg.SMTPPort)

        err := smtp.SendMail(addr, auth, n.cfg.SMTPUser, []string{safeEmail}, msg)
}

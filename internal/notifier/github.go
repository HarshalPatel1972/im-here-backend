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

const githubIssueTemplate = `Hey there,

We noticed something important in a recent commit to this repository.

**What we found:** An exposed %s
**File:** ` + "`" + `%s` + "`" + ` (line %d)
**Commit:** %s

---

**Don't panic.** We found this first. This notification is private — only you can see it.

**Do this right now:**

1. **Revoke the key** at your provider's dashboard immediately
2. **Generate a new key**
3. **Store it safely** in a ` + "`" + `.env` + "`" + ` file:
   ` + "```" + `
   YOUR_KEY=your_new_key_here
   ` + "```" + `
   Then add ` + "`" + `.env` + "`" + ` to your ` + "`" + `.gitignore` + "`" + `:
   ` + "```" + `
   echo ".env" >> .gitignore
   ` + "```" + `
4. **Update your code** to read from environment variables

---

**Why this matters:** Even free API keys are yours alone. Anyone with access
can consume your quota, rack up charges, or impersonate you with that provider.

You've got this. The fix takes less than 5 minutes.

— I'm Here
*A free, open-source guardian for developers. We never display your keys publicly.*`

func (n *Notifier) createPrivateIssue(ctx context.Context, f models.Finding) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues", f.RepoFullName)

	shaShort := f.CommitSHA
	if len(shaShort) > 7 {
		shaShort = shaShort[:7]
	}

	title := fmt.Sprintf("[I'm Here] Security Notice — Exposed %s detected", f.SecretType)
	body := fmt.Sprintf(githubIssueTemplate, f.SecretType, f.FilePath, f.LineNumber, shaShort)

	payload := map[string]string{
		"title": title,
		"body":  body,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal github payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create github request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+n.cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode github response: %w", err)
	}

	return result.HTMLURL, nil
}

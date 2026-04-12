package models

import "time"

type Finding struct {
	ID             int64     `json:"id"`
	RepoFullName   string    `json:"repo_full_name"`
	FilePath       string    `json:"file_path"`
	LineNumber     int       `json:"line_number"`
	SecretType     string    `json:"secret_type"`
	CommitSHA      string    `json:"commit_sha"`
	CommitterEmail string    `json:"committer_email"`
	CommitterName  string    `json:"committer_name"`
	DetectedAt     time.Time `json:"detected_at"`
	NotifiedAt     time.Time `json:"notified_at,omitempty"`
	GithubIssueURL string    `json:"github_issue_url,omitempty"`
	EmailSent      bool      `json:"email_sent"`
}

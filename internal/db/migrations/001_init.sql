-- 001_init.sql

CREATE TABLE IF NOT EXISTS findings (
  id              BIGSERIAL PRIMARY KEY,
  repo_full_name  TEXT NOT NULL,
  file_path       TEXT NOT NULL,
  line_number     INTEGER NOT NULL,
  secret_type     TEXT NOT NULL,
  commit_sha      TEXT NOT NULL,
  committer_email TEXT NOT NULL,
  committer_name  TEXT NOT NULL,
  detected_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  notified_at     TIMESTAMPTZ,
  github_issue_url TEXT,
  email_sent      BOOLEAN NOT NULL DEFAULT FALSE,
  UNIQUE(repo_full_name, commit_sha, file_path, line_number)
);

CREATE INDEX IF NOT EXISTS idx_findings_repo ON findings(repo_full_name);
CREATE INDEX IF NOT EXISTS idx_findings_detected ON findings(detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_findings_notified ON findings(notified_at);

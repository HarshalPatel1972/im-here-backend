package detector

import "regexp"

type SecretPattern struct {
	Type  string
	Regex *regexp.Regexp
}

var secretPatterns = []SecretPattern{
	{Type: "openai_api_key", Regex: regexp.MustCompile(`sk-proj-[A-Za-z0-9_-]{40,}|sk-[A-Za-z0-9]{48}`)},
	{Type: "anthropic_api_key", Regex: regexp.MustCompile(`sk-ant-[A-Za-z0-9_-]{40,}`)},
	{Type: "aws_access_key", Regex: regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{Type: "aws_secret_key", Regex: regexp.MustCompile(`(?i)aws.{0,20}secret.{0,20}['\"][0-9a-zA-Z/+]{40}['\"]`)},
	{Type: "github_token", Regex: regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`)},
	{Type: "huggingface_token", Regex: regexp.MustCompile(`hf_[A-Za-z0-9]{34,}`)},
	{Type: "groq_api_key", Regex: regexp.MustCompile(`gsk_[A-Za-z0-9]{52}`)},
	{Type: "stripe_secret_key", Regex: regexp.MustCompile(`sk_live_[A-Za-z0-9]{24,}`)},
	{Type: "stripe_restricted", Regex: regexp.MustCompile(`rk_live_[A-Za-z0-9]{24,}`)},
	{Type: "sendgrid_api_key", Regex: regexp.MustCompile(`SG\.[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43}`)},
	{Type: "twilio_api_key", Regex: regexp.MustCompile(`SK[0-9a-fA-F]{32}`)},
	{Type: "replicate_api_token", Regex: regexp.MustCompile(`r8_[A-Za-z0-9]{40}`)},
	{Type: "deepseek_api_key", Regex: regexp.MustCompile(`sk-[a-f0-9]{32}`)},
	{Type: "google_api_key", Regex: regexp.MustCompile(`AIza[0-9A-Za-z_-]{35}`)},
	{Type: "private_key_block", Regex: regexp.MustCompile(`-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----`)},
}

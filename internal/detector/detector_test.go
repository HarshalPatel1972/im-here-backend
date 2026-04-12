package detector

import (
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name        string
		diffLine    string
		expectedLen int
		secretType  string
	}{
		{"openai_api_key", "+sk-proj-1234567890123456789012345678901234567890", 1, "openai_api_key"},
		{"anthropic_api_key", "+sk-ant-1234567890123456789012345678901234567890", 1, "anthropic_api_key"},
		{"aws_access_key", "+AKIA1234567890123456", 1, "aws_access_key"},
		{"aws_secret_key", "+aws_secret_access_key = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMN'", 1, "aws_secret_key"},
		{"github_token", "+ghp_123456789012345678901234567890123456", 1, "github_token"},
		{"huggingface_token", "+hf_1234567890123456789012345678901234", 1, "huggingface_token"},
		{"groq_api_key", "+gsk_1234567890123456789012345678901234567890123456789012", 1, "groq_api_key"},
		{"stripe_secret_key", "+sk_live_123456789012345678901234", 1, "stripe_secret_key"},
		{"stripe_restricted", "+rk_live_123456789012345678901234", 1, "stripe_restricted"},
		{"sendgrid_api_key", "+SG.1234567890123456789012.1234567890123456789012345678901234567890123", 1, "sendgrid_api_key"},
		{"twilio_api_key", "+SK1234567890abcdef1234567890abcdef", 1, "twilio_api_key"},
		{"replicate_api_token", "+r8_1234567890123456789012345678901234567890", 1, "replicate_api_token"},
		{"deepseek_api_key", "+sk-1234567890abcdef1234567890abcdef", 1, "deepseek_api_key"},
		{"google_api_key", "+AIza1234567890ABCDEF1234567890ABCDEF123", 1, "google_api_key"},
		{"private_key_block", "+-----BEGIN RSA PRIVATE KEY-----", 1, "private_key_block"},
		{"removed_line", "-sk-proj-1234567890123456789012345678901234567890", 0, ""},
		{"false_positive", "+const s = 'just_a_normal_string_over_20_chars'", 0, ""},
		{"high_entropy_secret", "+export const KEY = 'aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789'", 1, "high_entropy_secret"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			findings := Detect(tc.diffLine, "test.go")
			if len(findings) != tc.expectedLen {
				t.Fatalf("expected %d findings, got %d for %s", tc.expectedLen, len(findings), tc.diffLine)
			}
			if tc.expectedLen > 0 {
				if findings[0].SecretType != tc.secretType {
					t.Fatalf("expected secret type %s, got %s for %s", tc.secretType, findings[0].SecretType, tc.diffLine)
				}
			}
		})
	}
}

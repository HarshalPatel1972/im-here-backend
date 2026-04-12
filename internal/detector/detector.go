package detector

import (
	"strings"

	"github.com/guardian/im-here/internal/models"
)

func Detect(diffContent string, filePath string) []models.Finding {
	var findings []models.Finding
	lines := strings.Split(diffContent, "\n")
	
	// Track line numbers based on hunk headers if needed, but since we are just 
	// scanning simple diff content for this tool, we assume consecutive numbering 
	// for simplistic scanning or just use the index for now.
	// For actual implementation, parsing accurate diff line numbers from @@ -x,y +a,b @@ is better.
	// To keep it simple per docs, we'll just track the added line index.
	
	lineNumber := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "@@ ") {
			// Basic hunk tracking could go here
			continue
		}
		
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			if !strings.HasPrefix(line, "-") {
				lineNumber++
			}
			continue
		}
		
		lineNumber++
		actualLine := line[1:] // strip the '+'

		detected := false
		for _, pattern := range secretPatterns {
			if pattern.Regex.MatchString(actualLine) {
				findings = append(findings, models.Finding{
					FilePath:   filePath,
					LineNumber: lineNumber,
					SecretType: pattern.Type,
				})
				detected = true
				// For atomic tracking, we might only log the first finding per line
				break
			}
		}

		if !detected && len(actualLine) > 20 && DetectHighEntropySecret(actualLine) {
			findings = append(findings, models.Finding{
				FilePath:   filePath,
				LineNumber: lineNumber,
				SecretType: "high_entropy_secret",
			})
		}
	}

	return findings
}

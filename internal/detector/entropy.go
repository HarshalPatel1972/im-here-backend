package detector

import (
	"math"
	"regexp"
)

var assignmentRegex = regexp.MustCompile(`(?i)(KEY|TOKEN|SECRET|PASSWORD)\s*=\s*['"]?([a-zA-Z0-9_\-\+/=]{21,})['"]?`)

func CalculateEntropy(s string) float64 {
	length := len(s)
	if length == 0 {
		return 0
	}

	charCounts := make(map[rune]int)
	for _, char := range s {
		charCounts[char]++
	}

	entropy := 0.0
	for _, count := range charCounts {
		probability := float64(count) / float64(length)
		entropy -= probability * math.Log2(probability)
	}

	return entropy
}

func DetectHighEntropySecret(line string) bool {
	matches := assignmentRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) == 3 {
			value := match[2]
			if CalculateEntropy(value) > 4.5 {
				return true
			}
		}
	}
	return false
}

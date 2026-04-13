package dlp

import (
	"fmt"
	"math"
	"regexp"
)

// ShannonEntropy calculates the Shannon entropy of a string.
// Used to detect high-entropy (likely secret) tokens in command output.
func ShannonEntropy(data string) float64 {
	if data == "" {
		return 0
	}
	frequencies := make(map[rune]float64)
	for _, char := range data {
		frequencies[char]++
	}

	entropy := 0.0
	length := float64(len(data))

	for _, freq := range frequencies {
		p := freq / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}

// StaticFilter applies all regex and entropy-based redaction rules to rawOutput
// and returns the sanitized string (Layer 4 DLP).
func StaticFilter(rawOutput string) string {
	cleanOutput := rawOutput

	reURI := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+-.]*:\/\/[^\s:@\/]+:)([^\s:@\/]+)(@[^\s\/]+)`)
	cleanOutput = reURI.ReplaceAllString(cleanOutput, "$1[URI_PASSWORD_REDACTED]$3")

	reSecret := regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token|api[_-]?key|private[_-]?key|salt|bearer|client[_-]?secret)\s*[:=]\s*([^\s\n\r"']+)`)
	cleanOutput = reSecret.ReplaceAllString(cleanOutput, "$1 = [🔒 DATA_REDACTED]")

	reAPIKeys := regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16}|sk_live_[0-9a-zA-Z]{24}|ghp_[0-9a-zA-Z]{36}|xox[bap]-[0-9a-zA-Z_-]+|eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+)`)
	cleanOutput = reAPIKeys.ReplaceAllString(cleanOutput, "[🔑 API_TOKEN_REDACTED]")

	reSSH := regexp.MustCompile(`(?s)-----BEGIN.*?PRIVATE KEY.*?-----END.*?PRIVATE KEY-----`)
	cleanOutput = reSSH.ReplaceAllString(cleanOutput, "[🚫 PRIVATE_KEY_REDACTED]")

	reWeirdAssignments := regexp.MustCompile(`(?i)([a-zA-Z0-9_-]+)\s*[:=]\s*([^\s\n\r]*?[!@#$%^&*][^\s\n\r]*)`)
	cleanOutput = reWeirdAssignments.ReplaceAllString(cleanOutput, "$1 = [🔒 SPECIAL_CHAR_PASSWORD_REDACTED]")

	reGenericSecret := regexp.MustCompile(`(?i)([a-z0-9_-]+)[\s:=]+\s*([^\s]{8,})`)
	cleanOutput = reGenericSecret.ReplaceAllString(cleanOutput, "$1 = [🔒 BLOCKED_PASSWORD]")

	rePotentialSecrets := regexp.MustCompile(`\S{14,}`)
	cleanOutput = rePotentialSecrets.ReplaceAllStringFunc(cleanOutput, func(match string) string {
		ent := ShannonEntropy(match)
		if ent > 3.8 {
			return fmt.Sprintf("[🔒 ANOMALOUS_STRING_REDACTED | Entropy: %.2f]", ent)
		}
		return match
	})

	if len(cleanOutput) > 10000 {
		cleanOutput = cleanOutput[:10000] + "\n\n...[SYSTEM TRUNCATED OUTPUT DUE TO LENGTH]..."
	}

	return cleanOutput
}

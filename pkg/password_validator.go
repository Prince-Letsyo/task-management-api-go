package pkg

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/trustelem/zxcvbn"
)

type PasswordValidator struct {
	MinScore          int     // zxcvbn score (0-4, default 2 = good)
	MaxSimilarity     float64 // 80% similarity threshold
	CommonPasswords   map[string]struct{}
	scoreDescriptions map[int]string
}

func (pv *PasswordValidator) normalizeString(s string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(s) {
		if !unicode.IsSpace(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func (pv *PasswordValidator) calculateSimilarity(password, otherField string) float64 {
	normPassword := pv.normalizeString(password)
	normField := pv.normalizeString(otherField)

	if len(normPassword) == 0 || len(normField) == 0 {
		return 0.0
	}

	minLen := math.Min(float64(len(normPassword)), float64(len(normField)))
	maxLen := math.Max(float64(len(normPassword)), float64(len(normField)))

	// Substring check
	if strings.Contains(normField, normPassword) || strings.Contains(normPassword, normField) {
		return minLen / maxLen
	}

	// Sequential common characters
	commonChars := 0
	for i := 0; i < int(minLen); i++ {
		if normPassword[i] == normField[i] {
			commonChars++
		}
	}

	return float64(commonChars) / maxLen
}

func (pv *PasswordValidator) isTooSimilar(password, username, email string) (bool, string) {
	simUsername := pv.calculateSimilarity(password, username)

	emailLocal := strings.Split(email, "@")[0]
	simEmail := pv.calculateSimilarity(password, emailLocal)

	if simUsername > pv.MaxSimilarity {
		return true, fmt.Sprintf("Password is too similar to username (similarity: %.1f%%)", simUsername*100)
	}
	if simEmail > pv.MaxSimilarity {
		return true, fmt.Sprintf("Password is too similar to email (similarity: %.1f%%)", simEmail*100)
	}
	return false, ""
}

type ValidationResult struct {
	IsValid            bool     `json:"is_valid"`
	Errors             []string `json:"errors"`
	Score              int      `json:"score"`
	FeedbackWarning    string   `json:"feedback_warning,omitempty"`    // Note: this port has no feedback
	FeedbackSuggestion string   `json:"feedback_suggestion,omitempty"` // Note: this port has no feedback
	SimilarityUsername float64  `json:"similarity_username"`
	SimilarityEmail    float64  `json:"similarity_email"`
	Guesses            float64  `json:"guesses,omitempty"`
	CalcTime           float64  `json:"calc_time,omitempty"`
}

func (pv *PasswordValidator) ValidatePassword(password, username, email string) ValidationResult {
	var errors []string

	// 1. Strength validation (zxcvbn)
	userInputs := []string{username, email} // Helps penalize passwords containing personal info
	result := zxcvbn.PasswordStrength(password, userInputs)

	if result.Score < pv.MinScore {
		desc := pv.scoreDescriptions[result.Score]
		// This port does not provide detailed feedback/suggestions like the Python/JS versions
		errorMsg := fmt.Sprintf("Password is too weak (score: %s). Use a stronger password.", desc)
		errors = append(errors, errorMsg)
	}

	// 2. Similarity validation
	isSimilar, similarityMsg := pv.isTooSimilar(password, username, email)
	if isSimilar {
		errors = append(errors, similarityMsg)
	}

	// 3. Common passwords
	if _, found := pv.CommonPasswords[strings.ToLower(password)]; found {
		errors = append(errors, "Password is too common")
	}

	emailLocal := strings.Split(email, "@")[0]

	return ValidationResult{
		IsValid:            len(errors) == 0,
		Errors:             errors,
		Score:              result.Score,
		FeedbackWarning:    "",                         // Not available in trustelem/zxcvbn
		FeedbackSuggestion: "Use a stronger password.", // Fallback suggestion
		SimilarityUsername: pv.calculateSimilarity(password, username),
		SimilarityEmail:    pv.calculateSimilarity(password, emailLocal),
		Guesses:            result.Guesses,
		CalcTime:           result.CalcTime,
	}
}

func newPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		MinScore:      2,
		MaxSimilarity: 0.8,
		CommonPasswords: map[string]struct{}{
			"password": {},
			"123456":   {},
			"qwerty":   {},
			"admin":    {},
			"letmein":  {},
		},
		scoreDescriptions: map[int]string{
			0: "very weak",
			1: "weak",
			2: "good",
			3: "strong",
			4: "very strong",
		},
	}
}

var DefaultValidator = newPasswordValidator()

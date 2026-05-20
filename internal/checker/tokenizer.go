package checker

import (
	"strings"
	"unicode"
)

// Tokenize lowers text and splits on non-letter/non-digit boundaries (matches HOW_TO_CHECK: punctuation stripped including /).
func Tokenize(s string) []string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.Fields(b.String())
}

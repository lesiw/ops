package cmd

import (
	"regexp"
	"strings"
)

var shUnsafe = regexp.MustCompile(`[^\w@%+=:,./-]`)

func shQuote(s string) string {
	if s == "" {
		return `''`
	}
	if !shUnsafe.MatchString(s) {
		return s
	}
	return `'` + strings.ReplaceAll(s, `'`, `\'`) + `'`
}

func shJoin(parts []string) string {
	quotedParts := make([]string, len(parts))
	for i, part := range parts {
		quotedParts[i] = shQuote(part)
	}
	return strings.Join(quotedParts, " ")
}

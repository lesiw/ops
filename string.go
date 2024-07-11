package ci

import (
	"fmt"
	"strings"
	"unicode"
)

func stringify(in []any) []string {
	out := make([]string, len(in))
	for i := range in {
		out[i] = fmt.Sprint(in[i])
	}
	return out
}

func snakecase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

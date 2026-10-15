package grammar

import (
	"sort"
	"strings"
)

// Show returns s in double quotes, with each code point outside '!'..'~'
// replaced by one '?'.
func Show(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		if r >= '!' && r <= '~' {
			b.WriteRune(r)
		} else {
			b.WriteByte('?')
		}
	}
	b.WriteByte('"')
	return b.String()
}

func isLetterOrDigit(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
}

func nameFrom(s string, first func(byte) bool) bool {
	if s == "" || !first(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isLetterOrDigit(s[i]) {
			return false
		}
	}
	return true
}

// IsNonterminal reports whether s is a nonterminal name such as Expr2.
func IsNonterminal(s string) bool {
	return nameFrom(s, func(c byte) bool { return c >= 'A' && c <= 'Z' })
}

// IsTerminal reports whether s is a terminal name (id) or a quoted terminal ('+').
func IsTerminal(s string) bool {
	if nameFrom(s, func(c byte) bool { return c >= 'a' && c <= 'z' }) {
		return true
	}
	n := len(s)
	if n < 3 || n > 5 || s[0] != '\'' || s[n-1] != '\'' {
		return false
	}
	for i := 1; i < n-1; i++ {
		if s[i] < '!' || s[i] > '~' || s[i] == '\'' {
			return false
		}
	}
	return true
}

// SortTerminals returns the members of set in code point order, "$" last.
func SortTerminals(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i] == "$" || out[j] == "$" {
			return out[j] == "$" && out[i] != "$"
		}
		return out[i] < out[j]
	})
	return out
}

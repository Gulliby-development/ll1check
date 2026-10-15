package grammar

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseInlineCommentsValid(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"space then hash token", "A -> a # x", "A@1 [a]@1"},
		{"hash-prefixed token", "A -> a #x", "A@1 [a]@1"},
		{"tab then hash-prefixed token", "A -> a\t#x", "A@1 [a]@1"},
		{"quoted hash terminal", "A -> '#' | b # c", "A@1 ['#']@1 [b]@1"},
		{"continuation comment", "A -> a\n| b # c", "A@1 [a]@1 [b]@2"},
		{"empty production comment", "A -> %empty # c", "A@1 []@1"},
		{"comment hides invalid symbol", "A -> a # $", "A@1 [a]@1"},
		{"comment hides following rule", "A -> a # B -> b", "A@1 [a]@1"},
		{"later hash remains part of comment", "A -> a # first # second", "A@1 [a]@1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const display = "g.g"
			g, err := Parse(display, tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := dumpGrammar(g); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
			if g.Display != display {
				t.Errorf("display %q, want %q", g.Display, display)
			}
			if len(g.Rules) != 1 || len(g.Index) != 1 {
				t.Fatalf("got %d rules and %d index entries, want one each", len(g.Rules), len(g.Index))
			}
			if got, ok := g.Index["A"]; !ok || got != 0 {
				t.Errorf("index of A is %d (present %t), want 0", got, ok)
			}
		})
	}
}

func TestParseInlineCommentsErrors(t *testing.T) {
	tests := []struct {
		name string
		text string
		line int
		msg  string
	}{
		{"attached hash is not a comment", "A -> a#b", 1, `invalid symbol "a#b"`},
		{"comment leaves too few rule tokens", "A -> # c", 1, `expected "NAME -> ALTERNATIVES"`},
		{"name with comment leaves too few tokens", "A # c", 1, `expected "NAME -> ALTERNATIVES"`},
		{"comment after continuation bar leaves empty alternative", "A -> a\n| # c", 2, "empty alternative (write %empty)"},
		{"comment after alternative bar leaves empty alternative", "A -> a | # c", 1, "empty alternative (write %empty)"},
		{"comment does not hide preceding invalid symbol", "A -> $ # c", 1, `invalid symbol "$"`},
		{"orphan continuation with symbols", "| a # c", 1, `"|" without a rule`},
		{"orphan continuation with comment only", "| # c", 1, `"|" without a rule`},
		{"prefixed attached hash is not a comment", "\n# prefix\nA -> a#b", 3, `invalid symbol "a#b"`},
		{"prefixed comment leaves too few rule tokens", "\n# prefix\nA -> # c", 3, `expected "NAME -> ALTERNATIVES"`},
		{"prefixed name with comment leaves too few tokens", "\n# prefix\nA # c", 3, `expected "NAME -> ALTERNATIVES"`},
		{"prefixed continuation comment leaves empty alternative", "\n# prefix\nA -> a\n| # c", 4, "empty alternative (write %empty)"},
		{"prefixed alternative comment leaves empty alternative", "\n# prefix\nA -> a | # c", 3, "empty alternative (write %empty)"},
		{"prefixed comment does not hide preceding invalid symbol", "\n# prefix\nA -> $ # c", 3, `invalid symbol "$"`},
		{"prefixed orphan continuation with symbols", "\n# prefix\n| a # c", 3, `"|" without a rule`},
		{"prefixed orphan continuation with comment only", "\n# prefix\n| # c", 3, `"|" without a rule`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, err := Parse("g.g", tc.text)
			if err == nil {
				t.Fatalf("expected error, got grammar %s", dumpGrammar(g))
			}
			expected := fmt.Sprintf("g.g:%d: %s", tc.line, tc.msg)
			if got := err.Error(); got != expected {
				t.Errorf("got %q, want %q", got, expected)
			}
			if g != nil {
				t.Errorf("grammar should be nil on error")
			}
		})
	}
}

func TestParseInlineCommentsPreserveBOMBlankLinesAndCRLF(t *testing.T) {
	lf := "\uFEFF\n# whole-line comment\nA -> a # c\n| b # c\n"
	plain, err := Parse("g.g", lf)
	if err != nil {
		t.Fatal(err)
	}

	crlf := strings.ReplaceAll(lf, "\n", "\r\n")
	other, err := Parse("g.g", crlf)
	if err != nil {
		t.Fatal(err)
	}

	const want = "A@3 [a]@3 [b]@4"
	if got := dumpGrammar(plain); got != want {
		t.Errorf("LF grammar %q, want %q", got, want)
	}
	if got := dumpGrammar(other); got != want {
		t.Errorf("CRLF grammar %q, want %q", got, want)
	}
	if dumpGrammar(plain) != dumpGrammar(other) {
		t.Errorf("LF grammar %q differs from CRLF grammar %q", dumpGrammar(plain), dumpGrammar(other))
	}

	errorLF := "\uFEFF\n# whole-line comment\nA -> a # c\n| $ # later error\n"
	errorCRLF := strings.ReplaceAll(errorLF, "\n", "\r\n")
	for _, tc := range []struct {
		name string
		text string
	}{
		{"LF later error", errorLF},
		{"CRLF later error", errorCRLF},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, err := Parse("g.g", tc.text)
			if err == nil {
				t.Fatalf("expected error, got grammar %s", dumpGrammar(g))
			}
			const wantError = `g.g:4: invalid symbol "$"`
			if err.Error() != wantError {
				t.Errorf("got %q, want %q", err.Error(), wantError)
			}
			if g != nil {
				t.Errorf("grammar should be nil on error")
			}
		})
	}
}

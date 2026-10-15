package grammar

import (
	"fmt"
	"strings"
	"testing"
)

func dumpGrammar(g *Grammar) string {
	var parts []string
	for _, r := range g.Rules {
		var alts []string
		for _, a := range r.Alts {
			alts = append(alts, fmt.Sprintf("[%s]@%d", strings.Join(a.Symbols, " "), a.Line))
		}
		parts = append(parts, fmt.Sprintf("%s@%d %s", r.Name, r.Line, strings.Join(alts, " ")))
	}
	return strings.Join(parts, "; ")
}

func TestParseValid(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"single rule", "A -> a\n", "A@1 [a]@1"},
		{"no trailing newline", "A -> a", "A@1 [a]@1"},
		{"comments and blank lines", "# c\n\n  # indented\nA -> a\n", "A@4 [a]@4"},
		{"tabs and runs of spaces", "\tA\t->\ta   B \t\nB -> b", "A@1 [a B]@1; B@2 [b]@2"},
		{"continuation lines", "A -> a\n  | b\n  | c d | e\nB -> b", "A@1 [a]@1 [b]@2 [c d]@3 [e]@3; B@4 [b]@4"},
		{"continuation after comment", "A -> a\n# x\n| b\n", "A@1 [a]@1 [b]@3"},
		{"empty alternative", "A -> a | %empty", "A@1 [a]@1 []@1"},
		{"only empty", "A -> %empty", "A@1 []@1"},
		{"quoted terminals", "A -> '+' | '<=' | '|'", "A@1 ['+']@1 ['<=']@1 ['|']@1"},
		{"names with digits", "Expr2 -> Expr2_x id1\nExpr2_x -> x", "Expr2@1 [Expr2_x id1]@1; Expr2_x@2 [x]@2"},
		{"forward reference", "S -> A\nA -> a", "S@1 [A]@1; A@2 [a]@2"},
		{"self reference", "A -> A a | a", "A@1 [A a]@1 [a]@1"},
		{"BOM", "\uFEFFA -> a\nB -> b\n", "A@1 [a]@1; B@2 [b]@2"},
		{"CRLF", "A -> a\r\n| b\r\n\r\nB -> b\r\n", "A@1 [a]@1 [b]@2; B@4 [b]@4"},
		{"BOM and CRLF", "\uFEFF# c\r\nA -> a\r\n  | b\r\nB -> b", "A@2 [a]@2 [b]@3; B@4 [b]@4"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, err := Parse("g.g", tc.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := dumpGrammar(g); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
			if g.Display != "g.g" {
				t.Errorf("display %q", g.Display)
			}
			for i, r := range g.Rules {
				if g.Index[r.Name] != i {
					t.Errorf("index of %s is %d, want %d", r.Name, g.Index[r.Name], i)
				}
			}
		})
	}
}

func TestParseBOMAndCRLFSameAsLF(t *testing.T) {
	lf := "# c\nS -> A | %empty\n  | 'x'\nA -> a\n"
	plain, err := Parse("g.g", lf)
	if err != nil {
		t.Fatal(err)
	}
	crlf := "\uFEFF" + strings.ReplaceAll(lf, "\n", "\r\n")
	other, err := Parse("g.g", crlf)
	if err != nil {
		t.Fatal(err)
	}
	if dumpGrammar(plain) != dumpGrammar(other) {
		t.Errorf("%q != %q", dumpGrammar(plain), dumpGrammar(other))
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"invalid name lowercase", "a -> b", `g.g:1: invalid name "a"`},
		{"invalid name dash", "E-1 -> b", `g.g:1: invalid name "E-1"`},
		{"invalid name non-ASCII", "é -> a", `g.g:1: invalid name "?"`},
		{"empty alternative trailing bar", "A -> b |", "g.g:1: empty alternative (write %empty)"},
		{"empty alternative leading bar", "A -> | b", "g.g:1: empty alternative (write %empty)"},
		{"empty alternative double bar", "A -> a | | b", "g.g:1: empty alternative (write %empty)"},
		{"empty alternative continuation", "A -> a\n|", "g.g:2: empty alternative (write %empty)"},
		{"empty must stand alone", "A -> b %empty", "g.g:1: %empty must stand alone"},
		{"empty twice", "A -> %empty %empty", "g.g:1: %empty must stand alone"},
		{"invalid symbol dollar", "A -> $", `g.g:1: invalid symbol "$"`},
		{"invalid symbol split quote", "A -> 'ab c'", `g.g:1: invalid symbol "'ab"`},
		{"invalid symbol too long", "A -> 'abcd'", `g.g:1: invalid symbol "'abcd'"`},
		{"invalid symbol digit start", "A -> 1a", `g.g:1: invalid symbol "1a"`},
		{"undefined rule", "A -> B", `g.g:1: undefined rule "B"`},
		{"undefined rule first use", "S -> a\n  | B\nT -> C B", `g.g:2: undefined rule "B"`},
		{"duplicate rule", "A -> a\nA -> a", `g.g:2: duplicate rule "A" (first on line 1)`},
		{"duplicate rule first line", "# c\nA -> a\n\nA -> b", `g.g:4: duplicate rule "A" (first on line 2)`},
		{"expected one token", "x", `g.g:1: expected "NAME -> ALTERNATIVES"`},
		{"expected two tokens", "A ->", `g.g:1: expected "NAME -> ALTERNATIVES"`},
		{"expected wrong arrow", "A => a", `g.g:1: expected "NAME -> ALTERNATIVES"`},
		{"bar without rule", "| a", `g.g:1: "|" without a rule`},
		{"bar without rule after comment", "# c\n| a", `g.g:2: "|" without a rule`},
		{"no rules empty text", "", "g.g: no rules"},
		{"no rules only comments", "# a\n\n   \n# b\n", "g.g: no rules"},
		{"line numbers with CRLF", "\r\n\r\nx\r\n", `g.g:3: expected "NAME -> ALTERNATIVES"`},
		{"first bad line wins", "A -> a\nx\nb -> c", `g.g:2: expected "NAME -> ALTERNATIVES"`},
		{"expected before name", "a ->", `g.g:1: expected "NAME -> ALTERNATIVES"`},
		{"name before symbol", "a -> $", `g.g:1: invalid name "a"`},
		{"duplicate before symbol", "A -> a\nA -> 'abcd'", `g.g:2: duplicate rule "A" (first on line 1)`},
		{"invalid name before empty alt", "a -> |", `g.g:1: invalid name "a"`},
		{"empty alternative before bad symbol", "A -> | $", "g.g:1: empty alternative (write %empty)"},
		{"bad symbol left of empty alternative", "A -> $ |", `g.g:1: invalid symbol "$"`},
		{"must stand alone before bad symbol", "A -> %empty $", "g.g:1: %empty must stand alone"},
		{"line error before undefined", "A -> B\nC -> $", `g.g:2: invalid symbol "$"`},
		{"duplicate before undefined", "A -> B\nA -> b", `g.g:2: duplicate rule "A" (first on line 1)`},
		{"first undefined in file order", "S -> B C\nB -> C", `g.g:1: undefined rule "C"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, err := Parse("g.g", tc.text)
			if err == nil {
				t.Fatalf("expected error %q, got grammar %s", tc.want, dumpGrammar(g))
			}
			if err.Error() != tc.want {
				t.Errorf("got %q, want %q", err.Error(), tc.want)
			}
			if g != nil {
				t.Errorf("grammar should be nil on error")
			}
		})
	}
}

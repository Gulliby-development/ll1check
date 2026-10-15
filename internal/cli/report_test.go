package cli

import (
	"strings"
	"testing"

	"github.com/Gulliby-development/ll1check/internal/grammar"
)

func termSet(items ...string) map[string]bool {
	m := map[string]bool{}
	for _, s := range items {
		m[s] = true
	}
	return m
}

func rulesOf(names ...string) *grammar.Grammar {
	g := &grammar.Grammar{Display: "t.g", Index: map[string]int{}}
	for i, n := range names {
		g.Rules = append(g.Rules, &grammar.Rule{Name: n, Line: i + 1})
		g.Index[n] = i
	}
	return g
}

func exprFixture() (*grammar.Grammar, *grammar.Analysis) {
	g := rulesOf("Expr", "Expr2", "Term", "Term2", "Factor")
	a := &grammar.Analysis{
		Nullable: map[string]bool{"Expr2": true, "Term2": true},
		First: map[string]map[string]bool{
			"Expr":   termSet("id", "'('"),
			"Expr2":  termSet("'+'"),
			"Term":   termSet("id", "'('"),
			"Term2":  termSet("'*'"),
			"Factor": termSet("id", "'('"),
		},
		Follow: map[string]map[string]bool{
			"Expr":   termSet("$", "')'"),
			"Expr2":  termSet("$", "')'"),
			"Term":   termSet("$", "')'", "'+'"),
			"Term2":  termSet("$", "')'", "'+'"),
			"Factor": termSet("$", "')'", "'+'", "'*'"),
		},
	}
	return g, a
}

func TestFormatSetsExpr(t *testing.T) {
	g, a := exprFixture()
	want := "RULE    NULLABLE  FIRST   FOLLOW\n" +
		"Expr    no        '(' id  ')' $\n" +
		"Expr2   yes       '+'     ')' $\n" +
		"Term    no        '(' id  ')' '+' $\n" +
		"Term2   yes       '*'     ')' '+' $\n" +
		"Factor  no        '(' id  ')' '*' '+' $\n"
	if got := formatSets(g, a); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestFormatSetsEmptySetsShowDash(t *testing.T) {
	g := rulesOf("A")
	a := &grammar.Analysis{
		Nullable: map[string]bool{},
		First:    map[string]map[string]bool{},
		Follow:   map[string]map[string]bool{},
	}
	want := "RULE  NULLABLE  FIRST  FOLLOW\n" +
		"A     no        -      -\n"
	if got := formatSets(g, a); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestFormatSetsCodePointWidths(t *testing.T) {
	g := rulesOf("S", "T")
	a := &grammar.Analysis{
		Nullable: map[string]bool{"T": true},
		First: map[string]map[string]bool{
			"S": termSet("éééééé"),
			"T": termSet("a"),
		},
		Follow: map[string]map[string]bool{
			"S": termSet("$"),
		},
	}
	want := "RULE  NULLABLE  FIRST   FOLLOW\n" +
		"S     no        éééééé  $\n" +
		"T     yes       a       -\n"
	if got := formatSets(g, a); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestFormatSetsLineEnds(t *testing.T) {
	g, a := exprFixture()
	got := formatSets(g, a)
	if strings.Contains(got, "\r") {
		t.Errorf("output contains CR: %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("output does not end with LF: %q", got)
	}
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 6 {
		t.Fatalf("got %d lines, want 6", len(lines))
	}
	for _, l := range lines {
		if strings.HasSuffix(l, " ") {
			t.Errorf("line ends with a space: %q", l)
		}
	}
}

func TestFormatFinding(t *testing.T) {
	tests := []struct {
		display string
		f       grammar.Finding
		want    string
	}{
		{"loose.g", grammar.Finding{Line: 2, Code: "LL004", Message: "Stmt: alternatives 1 and 2 both predict id"}, "loose.g:2: LL004 Stmt: alternatives 1 and 2 both predict id"},
		{"<stdin>", grammar.Finding{Line: 10, Code: "LL001", Message: "Dead is not reachable from Stmt"}, "<stdin>:10: LL001 Dead is not reachable from Stmt"},
		{"a/b.g", grammar.Finding{Line: 1, Code: "LL003", Message: "A is left recursive"}, "a/b.g:1: LL003 A is left recursive"},
	}
	for _, tc := range tests {
		if got := formatFinding(tc.display, tc.f); got != tc.want {
			t.Errorf("got %q want %q", got, tc.want)
		}
	}
}

func TestOkLine(t *testing.T) {
	tests := []struct {
		rules int
		want  string
	}{
		{1, "expr.g: ok, 1 rule\n"},
		{2, "expr.g: ok, 2 rules\n"},
		{5, "expr.g: ok, 5 rules\n"},
	}
	for _, tc := range tests {
		if got := okLine("expr.g", tc.rules); got != tc.want {
			t.Errorf("got %q want %q", got, tc.want)
		}
	}
}

func TestAcceptedLine(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "accepted, 0 productions\n"},
		{1, "accepted, 1 production\n"},
		{2, "accepted, 2 productions\n"},
		{9, "accepted, 9 productions\n"},
	}
	for _, tc := range tests {
		if got := acceptedLine(tc.n); got != tc.want {
			t.Errorf("got %q want %q", got, tc.want)
		}
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0 rules"},
		{1, "1 rule"},
		{2, "2 rules"},
	}
	for _, tc := range tests {
		if got := plural(tc.n, "rule"); got != tc.want {
			t.Errorf("got %q want %q", got, tc.want)
		}
	}
}

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuietAcceptedPrintsOnlyTheCount(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	tk := writeTemp(t, "sum.txt", "id '+' id\n")
	want := "accepted, 9 productions\n"
	expect(t, []string{"parse", "--quiet", g, tk}, "", 0, want, "")
	expect(t, []string{"parse", g, "--quiet", tk}, "", 0, want, "")
	expect(t, []string{"parse", g, tk, "--quiet"}, "", 0, want, "")
	expect(t, []string{"parse", "--quiet", "--quiet", g, tk}, "", 0, want, "")
	expect(t, []string{"parse", g, tk}, "", 0, sumProductions+want, "")
}

func TestQuietCountsOneProduction(t *testing.T) {
	want := "accepted, 1 production\n"
	g := writeTemp(t, "a.g", "S -> a\n")
	tk := writeTemp(t, "t.txt", "a")
	expect(t, []string{"parse", "--quiet", g, tk}, "", 0, want, "")
	g2 := writeTemp(t, "b.g", "S -> a | %empty\n")
	tk2 := writeTemp(t, "e.txt", "")
	expect(t, []string{"parse", "--quiet", g2, tk2}, "", 0, want, "")
}

func TestQuietSyntaxErrorPrintsNothingOnStdout(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	short := writeTemp(t, "short.txt", "id '+'")
	expect(t, []string{"parse", "--quiet", g, short}, "", 1, "",
		"ll1check: "+shown(short)+": unexpected end of input (expected '(', id)\n")
	bad := writeTemp(t, "bad.txt", "id id")
	wantErr := "ll1check: " + shown(bad) + ": token 2: unexpected \"id\" (expected ')', '*', '+', end of input)\n"
	expect(t, []string{"parse", "--quiet", g, bad}, "", 1, "", wantErr)
	expect(t, []string{"parse", g, bad}, "", 1,
		"Expr -> Term Expr2\nTerm -> Factor Term2\nFactor -> id\n", wantErr)
}

func TestQuietReadsStandardInput(t *testing.T) {
	g := writeTemp(t, "g.g", "S -> a\n")
	expect(t, []string{"parse", "--quiet", g, "-"}, "a", 0, "accepted, 1 production\n", "")
	expect(t, []string{"parse", "--quiet", "-", "-"}, "", 2, "", "ll1check: standard input given twice\n")
}

func TestQuietKeepsTheGrammarChecks(t *testing.T) {
	g := writeTemp(t, "loose.g", looseGrammar)
	tk := writeTemp(t, "sum.txt", "id '+' id")
	expect(t, []string{"parse", "--quiet", g, tk}, "", 2, "",
		"ll1check: "+shown(g)+": the grammar is not LL(1) (run check)\n")
	ok := writeTemp(t, "g.g", "S -> a\n")
	missing := filepath.Join(t.TempDir(), "none.txt")
	expect(t, []string{"parse", "--quiet", ok, missing}, "", 2, "",
		"ll1check: "+shown(missing)+": cannot read the file\n")
	expect(t, []string{"parse", "--quiet", ok, "TOKENS"}, "", 2, "",
		"ll1check: TOKENS: cannot read the file\n")
}

func TestQuietIsAnOptionOfParseOnly(t *testing.T) {
	g := writeTemp(t, "g.g", "S -> a\n")
	expect(t, []string{"check", "--quiet", g}, "", 2, "", "ll1check: unknown option \"--quiet\"\n")
	expect(t, []string{"sets", g, "--quiet"}, "", 2, "", "ll1check: unknown option \"--quiet\"\n")
	expect(t, strings.Fields("parse --Quiet g g"), "", 2, "", "ll1check: unknown option \"--Quiet\"\n")
	expect(t, strings.Fields("parse --quiet --loud g g"), "", 2, "", "ll1check: unknown option \"--loud\"\n")
	expect(t, []string{"--quiet"}, "", 2, "", "ll1check: unknown command \"--quiet\"\n")
}

func TestQuietCountsTheOperands(t *testing.T) {
	g := writeTemp(t, "g.g", "S -> a\n")
	expect(t, []string{"parse", "--quiet"}, "", 2, "", "ll1check: missing argument\n")
	expect(t, []string{"parse", "--quiet", g}, "", 2, "", "ll1check: missing argument\n")
	expect(t, []string{"parse", "--quiet", g, g, g}, "", 2, "", "ll1check: too many arguments\n")
}

func TestReadmeDocumentsQuiet(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	if !strings.Contains(s, "\n       ll1check parse [--quiet] GRAMMAR TOKENS\n") {
		t.Error("README lacks the new usage line")
	}
	if !strings.Contains(s, "With `parse --quiet` the productions are not printed") {
		t.Error("README lacks the --quiet paragraph")
	}
}

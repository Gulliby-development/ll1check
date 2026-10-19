package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const exprGrammar = "# arithmetic expressions\n" +
	"Expr -> Term Expr2\n" +
	"Expr2 -> '+' Term Expr2\n" +
	"      | %empty\n" +
	"Term -> Factor Term2\n" +
	"Term2 -> '*' Factor Term2 | %empty\n" +
	"Factor -> '(' Expr ')' | id\n"

const looseGrammar = "# a statement grammar with problems\n" +
	"Stmt -> id Tail | id '(' ')' | Block\n" +
	"Tail -> '=' num | %empty\n" +
	"Block -> '{' Stmt '}'\n" +
	"Dead -> Dead x | y\n" +
	"Loop -> Loop Loop\n"

const exprSets = "RULE    NULLABLE  FIRST   FOLLOW\n" +
	"Expr    no        '(' id  ')' $\n" +
	"Expr2   yes       '+'     ')' $\n" +
	"Term    no        '(' id  ')' '+' $\n" +
	"Term2   yes       '*'     ')' '+' $\n" +
	"Factor  no        '(' id  ')' '*' '+' $\n"

const sumProductions = "Expr -> Term Expr2\n" +
	"Term -> Factor Term2\n" +
	"Factor -> id\n" +
	"Term2 -> %empty\n" +
	"Expr2 -> '+' Term Expr2\n" +
	"Term -> Factor Term2\n" +
	"Factor -> id\n" +
	"Term2 -> %empty\n" +
	"Expr2 -> %empty\n"

func execRun(args []string, stdin string) (int, string, string) {
	var out, errb bytes.Buffer
	code := Run(args, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func shown(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}

func expect(t *testing.T, args []string, stdin string, code int, out, errText string) {
	t.Helper()
	c, o, e := execRun(args, stdin)
	if c != code || o != out || e != errText {
		t.Fatalf("args %q: got code %d stdout %q stderr %q; want code %d stdout %q stderr %q",
			args, c, o, e, code, out, errText)
	}
}

func TestHelp(t *testing.T) {
	u := "usage: ll1check check GRAMMAR\n       ll1check sets GRAMMAR\n       ll1check parse GRAMMAR TOKENS\n"
	expect(t, []string{"--help"}, "", 0, u, "")
	expect(t, []string{"check", "--help"}, "", 2, "", "ll1check: unknown option \"--help\"\n")
	expect(t, []string{"--help", "x"}, "", 2, "", "ll1check: unknown command \"--help\"\n")
}

func TestVersion(t *testing.T) {
	expect(t, []string{"--version"}, "", 0, "ll1check devel\n", "")
	expect(t, []string{"--version", "check"}, "", 2, "", "ll1check: unknown command \"--version\"\n")
	expect(t, []string{"check", "--version"}, "", 2, "", "ll1check: unknown option \"--version\"\n")
	expect(t, []string{"--version", "--help"}, "", 2, "", "ll1check: unknown command \"--version\"\n")
	expect(t, []string{"--version=1"}, "", 2, "", "ll1check: unknown command \"--version=1\"\n")
}

func TestCheckOK(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	expect(t, []string{"check", g}, "", 0, shown(g)+": ok, 5 rules\n", "")
}

func TestCheckOneRule(t *testing.T) {
	g := writeTemp(t, "one.g", "A -> a\n")
	expect(t, []string{"check", g}, "", 0, shown(g)+": ok, 1 rule\n", "")
}

func TestCheckLoose(t *testing.T) {
	g := writeTemp(t, "loose.g", looseGrammar)
	f := shown(g)
	want := f + ":2: LL004 Stmt: alternatives 1 and 2 both predict id\n" +
		f + ":5: LL001 Dead is not reachable from Stmt\n" +
		f + ":5: LL003 Dead is left recursive\n" +
		f + ":5: LL004 Dead: alternatives 1 and 2 both predict y\n" +
		f + ":6: LL001 Loop is not reachable from Stmt\n" +
		f + ":6: LL002 Loop derives no terminal string\n" +
		f + ":6: LL003 Loop is left recursive\n"
	expect(t, []string{"check", g}, "", 1, want, "")
}

func TestCheckStdin(t *testing.T) {
	expect(t, []string{"check", "-"}, exprGrammar, 0, "<stdin>: ok, 5 rules\n", "")
	expect(t, []string{"check", "-"}, "A -> A a | b\n", 1,
		"<stdin>:1: LL003 A is left recursive\n<stdin>:1: LL004 A: alternatives 1 and 2 both predict b\n", "")
}

func TestSetsExpr(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	expect(t, []string{"sets", g}, "", 0, exprSets, "")
	expect(t, []string{"sets", "-"}, exprGrammar, 0, exprSets, "")
}

func TestSetsEmptySet(t *testing.T) {
	expect(t, []string{"sets", "-"}, "S -> S a\n", 0,
		"RULE  NULLABLE  FIRST  FOLLOW\nS     no        -      a $\n", "")
}

func TestParseAccepted(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	tk := writeTemp(t, "sum.txt", "id '+' id\n")
	expect(t, []string{"parse", g, tk}, "", 0, sumProductions+"accepted, 9 productions\n", "")
}

func TestParseSyntaxError(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	tk := writeTemp(t, "bad.txt", "id '+' '*' id")
	first5 := strings.Join(strings.SplitAfter(sumProductions, "\n")[:5], "")
	expect(t, []string{"parse", g, tk}, "", 1, first5,
		"ll1check: "+shown(tk)+": token 3: unexpected \"'*'\" (expected '(', id)\n")
}

func TestParseUnexpectedEnd(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	tk := writeTemp(t, "short.txt", "id '+'")
	first6 := strings.Join(strings.SplitAfter(sumProductions, "\n")[:5], "")
	expect(t, []string{"parse", g, tk}, "", 1, first6,
		"ll1check: "+shown(tk)+": unexpected end of input (expected '(', id)\n")
}

func TestParseLeftoverTokens(t *testing.T) {
	g := writeTemp(t, "a.g", "S -> a\n")
	tk := writeTemp(t, "t.txt", "a a")
	expect(t, []string{"parse", g, tk}, "", 1, "S -> a\n",
		"ll1check: "+shown(tk)+": token 2: unexpected \"a\" (expected end of input)\n")
}

func TestParseNotLL1(t *testing.T) {
	g := writeTemp(t, "loose.g", looseGrammar)
	tk := writeTemp(t, "sum.txt", "id '+' id")
	want := "ll1check: " + shown(g) + ": the grammar is not LL(1) (run check)\n"
	expect(t, []string{"parse", g, tk}, "", 2, "", want)
	missing := filepath.Join(t.TempDir(), "none.txt")
	expect(t, []string{"parse", g, missing}, "", 2, "", want)
}

func TestParseOnlyLL001NotBlocked(t *testing.T) {
	g := writeTemp(t, "u.g", "S -> a\nB -> b\n")
	tk := writeTemp(t, "t.txt", "a")
	expect(t, []string{"check", g}, "", 1, shown(g)+":2: LL001 B is not reachable from S\n", "")
	expect(t, []string{"parse", g, tk}, "", 0, "S -> a\naccepted, 1 production\n", "")
}

func TestMissingFiles(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "none.g")
	expect(t, []string{"check", missing}, "", 2, "", "ll1check: "+shown(missing)+": cannot read the file\n")
	g := writeTemp(t, "g.g", "S -> a\n")
	expect(t, []string{"parse", g, missing}, "", 2, "", "ll1check: "+shown(missing)+": cannot read the file\n")
}

func TestInvalidUTF8(t *testing.T) {
	bad := writeTemp(t, "bad.g", "S -> a\xff\n")
	expect(t, []string{"sets", bad}, "", 2, "", "ll1check: "+shown(bad)+": input is not valid UTF-8\n")
	g := writeTemp(t, "g.g", "S -> a\n")
	tk := writeTemp(t, "t.txt", "a\xc3")
	expect(t, []string{"parse", g, tk}, "", 2, "", "ll1check: "+shown(tk)+": input is not valid UTF-8\n")
}

func TestGrammarErrorBeatsTokenError(t *testing.T) {
	g := writeTemp(t, "g.g", "a -> b\n")
	missing := filepath.Join(t.TempDir(), "none.txt")
	expect(t, []string{"parse", g, missing}, "", 2, "", "ll1check: "+shown(g)+":1: invalid name \"a\"\n")
}

func TestGrammarErrors(t *testing.T) {
	tests := []struct{ text, msg string }{
		{"a -> b\n", "1: invalid name \"a\""},
		{"A -> b |\n", "1: empty alternative (write %empty)"},
		{"A -> b %empty\n", "1: %empty must stand alone"},
		{"A -> $\n", "1: invalid symbol \"$\""},
		{"A -> 'ab c'\n", "1: invalid symbol \"'ab\""},
		{"A -> B\n", "1: undefined rule \"B\""},
		{"A -> a\nA -> a\n", "2: duplicate rule \"A\" (first on line 1)"},
		{"x\n", "1: expected \"NAME -> ALTERNATIVES\""},
		{"| a\n", "1: \"|\" without a rule"},
	}
	for _, tc := range tests {
		g := writeTemp(t, "g.g", tc.text)
		expect(t, []string{"check", g}, "", 2, "", "ll1check: "+shown(g)+":"+tc.msg+"\n")
	}
	g := writeTemp(t, "c.g", "# only a comment\n")
	expect(t, []string{"sets", g}, "", 2, "", "ll1check: "+shown(g)+": no rules\n")
}

func TestBackslashPathShown(t *testing.T) {
	p := filepath.Join(t.TempDir(), "no\\such.g")
	expect(t, []string{"check", p}, "", 2, "", "ll1check: "+strings.ReplaceAll(p, "\\", "/")+": cannot read the file\n")
}

func TestBOMAndCRLF(t *testing.T) {
	plain := writeTemp(t, "plain.g", exprGrammar)
	crlf := "\uFEFF" + strings.ReplaceAll(exprGrammar, "\n", "\r\n")
	bom := writeTemp(t, "bom.g", crlf)
	_, want, _ := execRun([]string{"sets", plain}, "")
	expect(t, []string{"sets", bom}, "", 0, want, "")
	tk := writeTemp(t, "t.txt", "\uFEFFid\r\n'+'\tid\r\n")
	expect(t, []string{"parse", bom, tk}, "", 0, sumProductions+"accepted, 9 productions\n", "")
}

func TestParseStdinTokens(t *testing.T) {
	g := writeTemp(t, "g.g", "S -> a\n")
	expect(t, []string{"parse", g, "-"}, "a", 0, "S -> a\naccepted, 1 production\n", "")
	expect(t, []string{"parse", "-", "-"}, "", 2, "", "ll1check: standard input given twice\n")
}

func TestParseEmptyTokensNullableStart(t *testing.T) {
	g := writeTemp(t, "g.g", "S -> a | %empty\n")
	tk := writeTemp(t, "t.txt", "")
	expect(t, []string{"parse", g, tk}, "", 0, "S -> %empty\naccepted, 1 production\n", "")
}

func TestParseEmptyTokensNonNullable(t *testing.T) {
	g := writeTemp(t, "g.g", "S -> a\n")
	tk := writeTemp(t, "t.txt", "")
	expect(t, []string{"parse", g, tk}, "", 1, "",
		"ll1check: "+shown(tk)+": unexpected end of input (expected a)\n")
}

func TestUsageErrorsEmptyStdout(t *testing.T) {
	expect(t, nil, "", 2, "", "ll1check: missing command (check, sets or parse)\n")
	expect(t, []string{"foo"}, "", 2, "", "ll1check: unknown command \"foo\"\n")
	expect(t, []string{"check"}, "", 2, "", "ll1check: missing argument\n")
	expect(t, []string{"check", "a", "b"}, "", 2, "", "ll1check: too many arguments\n")
}

package cli

import (
	"path/filepath"
	"testing"
)

func TestOnlyKeepsTheFindingsOfOneCode(t *testing.T) {
	g := writeTemp(t, "loose.g", looseGrammar)
	f := shown(g)
	expect(t, []string{"check", "--only", "LL003", g}, "", 1,
		f+":5: LL003 Dead is left recursive\n"+f+":6: LL003 Loop is left recursive\n", "")
	expect(t, []string{"check", "--only", "LL001", g}, "", 1,
		f+":5: LL001 Dead is not reachable from Stmt\n"+f+":6: LL001 Loop is not reachable from Stmt\n", "")
	expect(t, []string{"check", "--only", "LL002", g}, "", 1,
		f+":6: LL002 Loop derives no terminal string\n", "")
	expect(t, []string{"check", "--only", "LL004", g}, "", 1,
		f+":2: LL004 Stmt: alternatives 1 and 2 both predict id\n"+f+":5: LL004 Dead: alternatives 1 and 2 both predict y\n", "")
}

func TestOnlyAcceptsTheInlineForm(t *testing.T) {
	g := writeTemp(t, "loose.g", looseGrammar)
	f := shown(g)
	expect(t, []string{"check", "--only=LL002", g}, "", 1,
		f+":6: LL002 Loop derives no terminal string\n", "")
}

func TestOnlyMayComeBeforeOrAfterTheGrammar(t *testing.T) {
	g := writeTemp(t, "loose.g", looseGrammar)
	want := shown(g) + ":6: LL002 Loop derives no terminal string\n"
	expect(t, []string{"check", g, "--only", "LL002"}, "", 1, want, "")
	expect(t, []string{"check", g, "--only=LL002"}, "", 1, want, "")
}

func TestOnlyReadsStandardInput(t *testing.T) {
	expect(t, []string{"check", "--only", "LL003", "-"}, "A -> A a | b\n", 1,
		"<stdin>:1: LL003 A is left recursive\n", "")
}

func TestOnlyPrintsANoneLineWhenTheCodeHasNoFinding(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	expect(t, []string{"check", "--only", "LL002", g}, "", 0, shown(g)+": no LL002 findings, 5 rules\n", "")
	one := writeTemp(t, "one.g", "A -> a\n")
	expect(t, []string{"check", "--only", "LL004", one}, "", 0, shown(one)+": no LL004 findings, 1 rule\n", "")
}

func TestOnlyExitsZeroWhenOtherCodesHaveFindings(t *testing.T) {
	expect(t, []string{"check", "--only", "LL001", "-"}, "A -> A a | b\n", 0,
		"<stdin>: no LL001 findings, 1 rule\n", "")
	expect(t, []string{"check", "-"}, "A -> A a | b\n", 1,
		"<stdin>:1: LL003 A is left recursive\n<stdin>:1: LL004 A: alternatives 1 and 2 both predict b\n", "")
}

func TestOnlyDoesNotChangeTheOkLine(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	expect(t, []string{"check", g}, "", 0, shown(g)+": ok, 5 rules\n", "")
}

func TestOnlyRejectsAnUnknownCode(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	expect(t, []string{"check", "--only", "LL005", g}, "", 2, "", "ll1check: unknown finding code \"LL005\"\n")
	expect(t, []string{"check", "--only", "ll003", g}, "", 2, "", "ll1check: unknown finding code \"ll003\"\n")
	expect(t, []string{"check", "--only=", g}, "", 2, "", "ll1check: unknown finding code \"\"\n")
	expect(t, []string{"check", "--only", "LL003 ", g}, "", 2, "", "ll1check: unknown finding code \"LL003?\"\n")
}

func TestOnlyTakesTheNextWordAsItsValue(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	expect(t, []string{"check", "--only", "expr.g"}, "", 2, "", "ll1check: unknown finding code \"expr.g\"\n")
	expect(t, []string{"check", "--only", "--only", g}, "", 2, "", "ll1check: unknown finding code \"--only\"\n")
}

func TestOnlyAsTheLastWordNeedsAValue(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	expect(t, []string{"check", g, "--only"}, "", 2, "", "ll1check: --only needs a value\n")
}

func TestOnlyGivenTwiceIsAnError(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	expect(t, []string{"check", "--only", "LL001", "--only", "LL002", g}, "", 2, "", "ll1check: --only given twice\n")
	expect(t, []string{"check", "--only=LL001", "--only=LL001", g}, "", 2, "", "ll1check: --only given twice\n")
	expect(t, []string{"check", "--only", "LL001", "--only"}, "", 2, "", "ll1check: --only given twice\n")
}

func TestOnlyErrorsComeInTheOrderOfReading(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	expect(t, []string{"check", "--only", "LL009", "--bogus", g}, "", 2, "", "ll1check: unknown finding code \"LL009\"\n")
	expect(t, []string{"check", "--bogus", "--only", "LL009", g}, "", 2, "", "ll1check: unknown option \"--bogus\"\n")
}

func TestOnlyBelongsToCheckOnly(t *testing.T) {
	g := writeTemp(t, "expr.g", exprGrammar)
	tk := writeTemp(t, "sum.txt", "id")
	expect(t, []string{"sets", "--only", "LL001", g}, "", 2, "", "ll1check: unknown option \"--only\"\n")
	expect(t, []string{"parse", "--only", "LL001", g, tk}, "", 2, "", "ll1check: unknown option \"--only\"\n")
	expect(t, []string{"parse", "--only=LL001", g, tk}, "", 2, "", "ll1check: unknown option \"--only=LL001\"\n")
}

func TestOnlyStillNeedsAGrammar(t *testing.T) {
	expect(t, []string{"check", "--only", "LL001"}, "", 2, "", "ll1check: missing argument\n")
	g := writeTemp(t, "expr.g", exprGrammar)
	expect(t, []string{"check", "--only", "LL001", g, g}, "", 2, "", "ll1check: too many arguments\n")
}

func TestOnlyKeepsTheGrammarAndFileErrors(t *testing.T) {
	bad := writeTemp(t, "bad.g", "a -> b\n")
	expect(t, []string{"check", "--only", "LL001", bad}, "", 2, "", "ll1check: "+shown(bad)+":1: invalid name \"a\"\n")
	missing := filepath.Join(t.TempDir(), "none.g")
	expect(t, []string{"check", "--only", "LL001", missing}, "", 2, "", "ll1check: "+shown(missing)+": cannot read the file\n")
}

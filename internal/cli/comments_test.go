package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestInlineCommentsCalculatorCommands(t *testing.T) {
	grammar := writeTemp(t, "calc.g", "# calculator\n"+
		"Expr -> Term Expr2          # a sum\n"+
		"Expr2 -> '+' Term Expr2     # more terms\n"+
		"       | %empty             # or none\n"+
		"Term -> id | '#'            # a quoted '#' is a terminal\n")
	tokens := writeTemp(t, "t.txt", "id '+' '#'")

	expect(t, []string{"check", grammar}, "", 0, shown(grammar)+": ok, 3 rules\n", "")
	expect(t, []string{"sets", grammar}, "", 0,
		"RULE   NULLABLE  FIRST   FOLLOW\n"+
			"Expr   no        '#' id  $\n"+
			"Expr2  yes       '+'     $\n"+
			"Term   no        '#' id  '+' $\n", "")
	expect(t, []string{"parse", grammar, tokens}, "", 0,
		"Expr -> Term Expr2\n"+
			"Term -> id\n"+
			"Expr2 -> '+' Term Expr2\n"+
			"Term -> '#'\n"+
			"Expr2 -> %empty\n"+
			"accepted, 5 productions\n", "")
}

func TestInlineCommentsFindings(t *testing.T) {
	grammar := writeTemp(t, "f.g", "S -> A # start\n"+
		"A -> a | a # twice\n"+
		"B -> b # unused\n")
	expect(t, []string{"check", grammar}, "", 1,
		shown(grammar)+":2: LL004 A: alternatives 1 and 2 both predict a\n"+
			shown(grammar)+":3: LL001 B is not reachable from S\n", "")
}

func TestInlineCommentErrors(t *testing.T) {
	tests := []struct {
		name string
		text string
		line int
		msg  string
	}{
		{"joined token is not a comment", "A -> a#b\n", 1, "invalid symbol \"a#b\""},
		{"rule only has two tokens after comment", "A -> # c\n", 1, "expected \"NAME -> ALTERNATIVES\""},
		{"name only has one token after comment", "A # c\n", 1, "expected \"NAME -> ALTERNATIVES\""},
		{"empty continuation alternative", "A -> a\n| # c\n", 2, "empty alternative (write %empty)"},
		{"empty rule alternative", "A -> a | # c\n", 1, "empty alternative (write %empty)"},
		{"invalid symbol before comment", "A -> $ # c\n", 1, "invalid symbol \"$\""},
		{"orphan continuation with alternative", "| a # c\n", 1, "\"|\" without a rule"},
		{"orphan continuation with comment only", "| # c\n", 1, "\"|\" without a rule"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixtures := []struct {
				name   string
				prefix string
				line   int
			}{
				{"unprefixed", "", tc.line},
				{"prefixed", "\n# prefix comment\n", tc.line + 2},
			}
			for _, fixture := range fixtures {
				t.Run(fixture.name, func(t *testing.T) {
					grammar := writeTemp(t, "bad.g", fixture.prefix+tc.text)
					expect(t, []string{"check", grammar}, "", 2, "",
						"ll1check: "+shown(grammar)+":"+strconv.Itoa(fixture.line)+": "+tc.msg+"\n")
				})
			}
		})
	}
}

func TestInlineCommentsFromStdinCRLF(t *testing.T) {
	expect(t, []string{"check", "-"}, "S -> a # first\r\n | b # second\r\n",
		0, "<stdin>: ok, 1 rule\n", "")
}

func TestInlineCommentsReadme(t *testing.T) {
	raw, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("cannot read README.md: %v", err)
	}
	readme := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if strings.Contains(readme, "a comment cannot follow other text") {
		t.Fatal("README.md still says a comment cannot follow other text")
	}

	paragraph := "The file is UTF-8. One leading BOM is skipped. Lines are split at LF and one CR before the LF is removed. Line numbers count every physical line from 1. A line is trimmed of spaces and tabs; an empty line or one starting with `#` is skipped. Tokens are separated by runs of spaces and tabs. On a rule line or a `|` line the first token that starts with `#`, and everything after it, is a comment (`'#'` in quotes is an ordinary terminal; `a#b` is an invalid symbol, not a comment):"
	example := "Expr2 -> '+' Term Expr2     # more terms\n       | %empty             # or none\n"
	want := paragraph + "\n\n" + fenced(example) + "\n- A rule line is `NAME -> ALT | ALT ...`."
	if !strings.Contains(readme, want) {
		t.Errorf("README.md does not contain the exact Grammar file paragraph and example immediately before the bullets")
	}
}

func TestBuiltProgramInlineCommentsOnBOMCRLFStdin(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "ll1check")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}

	buildCtx, cancelBuild := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", executable, ".")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	input := []byte("\xef\xbb\xbf\r\n# whole-line comment\r\nS -> a # rule comment\r\n | b # continuation comment\r\n")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "check", "-")
	cmd.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		t.Fatalf("built program failed: %v; stdout %q; stderr %q", err, stdout.String(), stderr.String())
	}
	if got, want := stdout.String(), "<stdin>: ok, 1 rule\n"; got != want {
		t.Errorf("stdout %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr %q, want empty", got)
	}
}

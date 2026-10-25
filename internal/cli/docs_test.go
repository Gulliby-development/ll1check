package cli

import (
	"os"
	"strings"
	"testing"
)

const docUsage = `usage: ll1check check GRAMMAR
       ll1check sets GRAMMAR
       ll1check parse [--quiet] GRAMMAR TOKENS
`

const docExprG = `# arithmetic expressions
Expr -> Term Expr2
Expr2 -> '+' Term Expr2
      | %empty
Term -> Factor Term2
Term2 -> '*' Factor Term2 | %empty
Factor -> '(' Expr ')' | id
`

const docSets = `RULE    NULLABLE  FIRST   FOLLOW
Expr    no        '(' id  ')' $
Expr2   yes       '+'     ')' $
Term    no        '(' id  ')' '+' $
Term2   yes       '*'     ')' '+' $
Factor  no        '(' id  ')' '*' '+' $
`

const docParse = `Expr -> Term Expr2
Term -> Factor Term2
Factor -> id
Term2 -> %empty
Expr2 -> '+' Term Expr2
Term -> Factor Term2
Factor -> id
Term2 -> %empty
Expr2 -> %empty
accepted, 9 productions
`

func fenced(body string) string {
	return "```\n" + body + "```\n"
}

func TestReadmeContents(t *testing.T) {
	raw, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("cannot read README.md: %v", err)
	}
	readme := strings.ReplaceAll(string(raw), "\r\n", "\n")

	cases := []struct {
		name  string
		wants []string
	}{
		{"headings", []string{"\n## Install\n", "\n## Usage\n", "\n## Grammar file\n", "\n## Findings\n", "\n## Exit codes\n", "\n## Development\n"}},
		{"usage lines", []string{fenced(docUsage)}},
		{"expr.g block", []string{fenced(docExprG)}},
		{"sets output", []string{fenced(docSets)}},
		{"parse output", []string{fenced(docParse)}},
		{"install command", []string{"go build -o ll1check ."}},
		{"development commands", []string{"go vet ./...", "gofmt -l .", "go test ./..."}},
		{"finding codes", []string{"LL001", "LL002", "LL003", "LL004"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, want := range c.wants {
				if !strings.Contains(readme, want) {
					t.Errorf("README.md does not contain %q", want)
				}
			}
		})
	}
}

func TestGoDirectiveMatchesReadme(t *testing.T) {
	goModRaw, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatalf("cannot read go.mod: %v", err)
	}
	readmeRaw, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("cannot read README.md: %v", err)
	}

	goDirective := ""
	for _, line := range strings.Split(string(goModRaw), "\n") {
		if strings.HasPrefix(line, "go ") {
			goDirective = strings.TrimSpace(strings.TrimPrefix(line, "go "))
		}
	}
	if goDirective != "1.24" {
		t.Errorf("go.mod go directive = %q, want 1.24", goDirective)
	}

	if !strings.Contains(string(readmeRaw), "Go 1.24 or newer") {
		t.Error("README.md does not contain Go 1.24 or newer")
	}
}

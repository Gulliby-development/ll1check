package cli

import (
	"os"
	"strings"
	"testing"
)

const docUsage = `usage: ll1check check GRAMMAR
       ll1check sets GRAMMAR
       ll1check parse GRAMMAR TOKENS
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

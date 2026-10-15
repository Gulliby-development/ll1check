package grammar

import (
	"strings"
	"testing"
)

const ptExprGrammar = `# arithmetic expressions
Expr -> Term Expr2
Expr2 -> '+' Term Expr2
      | %empty
Term -> Factor Term2
Term2 -> '*' Factor Term2 | %empty
Factor -> '(' Expr ')' | id
`

func ptSetup(t *testing.T, text string) (*Grammar, *Analysis) {
	t.Helper()
	g, err := Parse("g", text)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return g, Analyze(g)
}

func TestPredictSets(t *testing.T) {
	g, a := ptSetup(t, ptExprGrammar)
	preds := Predict(g, a)
	tests := []struct {
		name string
		rule int
		alt  int
		want string
	}{
		{"Expr", 0, 0, "'(' id"},
		{"Expr2 plus", 1, 0, "'+'"},
		{"Expr2 empty uses FOLLOW", 1, 1, "')' $"},
		{"Term", 2, 0, "'(' id"},
		{"Term2 star", 3, 0, "'*'"},
		{"Term2 empty uses FOLLOW", 3, 1, "')' '+' $"},
		{"Factor paren", 4, 0, "'('"},
		{"Factor id", 4, 1, "id"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.Join(SortTerminals(preds[tc.rule][tc.alt]), " ")
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"spaces", "id '+' id", "id|'+'|id"},
		{"tabs", "a\tb\t\tc", "a|b|c"},
		{"crlf", "a\r\nb\r\n", "a|b"},
		{"bom", "\uFEFFa b", "a|b"},
		{"empty", "", ""},
		{"only whitespace", " \t\r\n ", ""},
		{"only bom", "\uFEFF", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.Join(Tokenize(tc.in), "|")
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseTokensAccepted(t *testing.T) {
	g, a := ptSetup(t, ptExprGrammar)
	lines, serr := ParseTokens(g, a, Tokenize("id '+' id\n"))
	if serr != nil {
		t.Fatalf("unexpected error %+v", serr)
	}
	want := strings.Join([]string{
		"Expr -> Term Expr2",
		"Term -> Factor Term2",
		"Factor -> id",
		"Term2 -> %empty",
		"Expr2 -> '+' Term Expr2",
		"Term -> Factor Term2",
		"Factor -> id",
		"Term2 -> %empty",
		"Expr2 -> %empty",
	}, "\n")
	if strings.Join(lines, "\n") != want || len(lines) != 9 {
		t.Errorf("got %q", lines)
	}
}

func TestParseTokensRejected(t *testing.T) {
	g, a := ptSetup(t, ptExprGrammar)
	lines, serr := ParseTokens(g, a, Tokenize("id '+' '*' id"))
	if serr == nil {
		t.Fatal("expected a syntax error")
	}
	if len(lines) != 5 || lines[4] != "Expr2 -> '+' Term Expr2" {
		t.Errorf("lines = %q", lines)
	}
	if serr.Index != 3 || serr.Token != "'*'" || serr.AtEnd {
		t.Errorf("error = %+v", serr)
	}
	if strings.Join(serr.Expected, ",") != "'(',id" {
		t.Errorf("expected = %q", serr.Expected)
	}
}

func TestParseTokensLeftover(t *testing.T) {
	g, a := ptSetup(t, "S -> a\n")
	lines, serr := ParseTokens(g, a, []string{"a", "a"})
	if serr == nil {
		t.Fatal("expected a syntax error")
	}
	if len(lines) != 1 {
		t.Errorf("lines = %q", lines)
	}
	if serr.Index != 2 || serr.Token != "a" || serr.AtEnd || strings.Join(serr.Expected, ",") != "$" {
		t.Errorf("error = %+v", serr)
	}
}

func TestParseTokensPrematureEnd(t *testing.T) {
	g, a := ptSetup(t, ptExprGrammar)
	_, serr := ParseTokens(g, a, []string{"id", "'+'"})
	if serr == nil {
		t.Fatal("expected a syntax error")
	}
	if !serr.AtEnd || serr.Index != 0 || strings.Join(serr.Expected, ",") != "'(',id" {
		t.Errorf("error = %+v", serr)
	}
}

func TestParseTokensTerminalOnTop(t *testing.T) {
	g, a := ptSetup(t, "S -> a b\n")
	_, serr := ParseTokens(g, a, []string{"a"})
	if serr == nil || !serr.AtEnd || strings.Join(serr.Expected, ",") != "b" {
		t.Errorf("end: error = %+v", serr)
	}
	_, serr = ParseTokens(g, a, []string{"a", "c"})
	if serr == nil || serr.Index != 2 || serr.Token != "c" || strings.Join(serr.Expected, ",") != "b" {
		t.Errorf("mismatch: error = %+v", serr)
	}
}

func TestParseTokensEmptyList(t *testing.T) {
	g, a := ptSetup(t, "S -> a | %empty\n")
	lines, serr := ParseTokens(g, a, nil)
	if serr != nil {
		t.Fatalf("unexpected error %+v", serr)
	}
	if len(lines) != 1 || lines[0] != "S -> %empty" {
		t.Errorf("lines = %q", lines)
	}
}

func TestParseTokensEmptyListNonNullable(t *testing.T) {
	g, a := ptSetup(t, "S -> a | b\n")
	lines, serr := ParseTokens(g, a, nil)
	if serr == nil || !serr.AtEnd || len(lines) != 0 || strings.Join(serr.Expected, ",") != "a,b" {
		t.Errorf("lines = %q, error = %+v", lines, serr)
	}
}

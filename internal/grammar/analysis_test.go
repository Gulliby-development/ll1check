package grammar

import (
	"strings"
	"testing"
)

func analyzeText(t *testing.T, text string) (*Grammar, *Analysis) {
	t.Helper()
	g, err := Parse("g.g", text)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	return g, Analyze(g)
}

func setText(set map[string]bool) string {
	return strings.Join(SortTerminals(set), " ")
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

type setRow struct {
	rule, nullable, first, follow string
}

func checkSets(t *testing.T, text string, rows []setRow) {
	t.Helper()
	_, a := analyzeText(t, text)
	for _, r := range rows {
		if got := yesNo(a.Nullable[r.rule]); got != r.nullable {
			t.Errorf("%s nullable = %s, want %s", r.rule, got, r.nullable)
		}
		if got := setText(a.First[r.rule]); got != r.first {
			t.Errorf("%s FIRST = %q, want %q", r.rule, got, r.first)
		}
		if got := setText(a.Follow[r.rule]); got != r.follow {
			t.Errorf("%s FOLLOW = %q, want %q", r.rule, got, r.follow)
		}
	}
}

const exprGrammar = `# arithmetic expressions
Expr -> Term Expr2
Expr2 -> '+' Term Expr2
      | %empty
Term -> Factor Term2
Term2 -> '*' Factor Term2 | %empty
Factor -> '(' Expr ')' | id
`

func TestAnalyzeSets(t *testing.T) {
	tests := []struct {
		name string
		text string
		rows []setRow
	}{
		{"expr.g", exprGrammar, []setRow{
			{"Expr", "no", "'(' id", "')' $"},
			{"Expr2", "yes", "'+'", "')' $"},
			{"Term", "no", "'(' id", "')' '+' $"},
			{"Term2", "yes", "'*'", "')' '+' $"},
			{"Factor", "no", "'(' id", "')' '*' '+' $"},
		}},
		{"nullable chain", "A -> B\nB -> C\nC -> %empty\n", []setRow{
			{"A", "yes", "", "$"},
			{"B", "yes", "", "$"},
			{"C", "yes", "", "$"},
		}},
		{"nullable chain with terminal", "A -> B x\nB -> C\nC -> %empty | c\n", []setRow{
			{"A", "no", "c x", "$"},
			{"B", "yes", "c", "x"},
			{"C", "yes", "c", "x"},
		}},
		{"mutually recursive nullable", "S -> A\nA -> B | %empty\nB -> A | b\n", []setRow{
			{"S", "yes", "b", "$"},
			{"A", "yes", "b", "$"},
			{"B", "yes", "b", "$"},
		}},
		{"mutually recursive not nullable", "S -> A\nA -> B a\nB -> A b | c\n", []setRow{
			{"S", "no", "c", "$"},
			{"A", "no", "c", "b $"},
			{"B", "no", "c", "a"},
		}},
		{"start follow holds end marker", "S -> a\n", []setRow{
			{"S", "no", "a", "$"},
		}},
		{"start rule used inside", "S -> a S b | %empty\n", []setRow{
			{"S", "yes", "a", "b $"},
		}},
		{"unreachable and unproductive rules", "S -> a\nDead -> Dead x | y\nLoop -> Loop Loop\n", []setRow{
			{"Dead", "no", "y", "x"},
			{"Loop", "no", "", ""},
		}},
		{"a differs from quoted a", "S -> a | 'a'\n", []setRow{
			{"S", "no", "'a' a", "$"},
		}},
		{"code point order with end marker last", "S -> T z\nT -> '~' | 'A' | b | '!'\n", []setRow{
			{"T", "no", "'!' 'A' '~' b", "z"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkSets(t, tt.text, tt.rows)
		})
	}
}

func TestAnalyzeFlags(t *testing.T) {
	tests := []struct {
		name                  string
		text                  string
		rule                  string
		reachable, productive bool
		leftRec               bool
	}{
		{"left recursive productive", "S -> a\nDead -> Dead x | y\n", "Dead", false, true, true},
		{"self loop unproductive", "S -> a\nLoop -> Loop Loop\n", "Loop", false, false, true},
		{"nullable prefix", "A -> B A | a\nB -> %empty\n", "A", true, true, true},
		{"nullable prefix helper", "A -> B A | a\nB -> %empty\n", "B", true, true, false},
		{"two rules A", "S -> A\nA -> B x\nB -> A y\n", "A", true, false, true},
		{"two rules B", "S -> A\nA -> B x\nB -> A y\n", "B", true, false, true},
		{"right recursion", "A -> a A | a\n", "A", true, true, false},
		{"non-nullable prefix blocks", "A -> B A | a\nB -> b\n", "A", true, true, false},
		{"used only by unreachable", "S -> a\nU -> V\nV -> v\n", "V", false, true, false},
		{"unreachable user", "S -> a\nU -> V\nV -> v\n", "U", false, true, false},
		{"start rule", "S -> a\n", "S", true, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, a := analyzeText(t, tt.text)
			if a.Reachable[tt.rule] != tt.reachable {
				t.Errorf("reachable = %v, want %v", a.Reachable[tt.rule], tt.reachable)
			}
			if a.Productive[tt.rule] != tt.productive {
				t.Errorf("productive = %v, want %v", a.Productive[tt.rule], tt.productive)
			}
			if a.LeftRecursive[tt.rule] != tt.leftRec {
				t.Errorf("left recursive = %v, want %v", a.LeftRecursive[tt.rule], tt.leftRec)
			}
		})
	}
}

package grammar

import (
	"reflect"
	"testing"
)

func TestShow(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", `""`},
		{"printable", "id", `"id"`},
		{"all punctuation kept", "'+'|\\~!", `"'+'|\~!"`},
		{"space", "a b", `"a?b"`},
		{"tab", "a\tb", `"a?b"`},
		{"non-ascii one code point", "é", `"?"`},
		{"multi-byte emoji one mark", "a😀b", `"a?b"`},
		{"newline", "\n", `"?"`},
		{"del", "\x7f", `"?"`},
		{"quote not escaped", `"`, `"""`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Show(tc.in); got != tc.want {
				t.Errorf("Show(%q) = %s, want %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsNonterminal(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"Expr", true},
		{"Expr2", true},
		{"E", true},
		{"A_b", true},
		{"expr", false},
		{"E-1", false},
		{"%empty", false},
		{"", false},
		{"1A", false},
		{"_A", false},
		{"Éa", false},
	}
	for _, tc := range tests {
		if got := IsNonterminal(tc.in); got != tc.want {
			t.Errorf("IsNonterminal(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestIsTerminal(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"id", true},
		{"num", true},
		{"a", true},
		{"a_1", true},
		{"'+'", true},
		{"'<='", true},
		{"'|'", true},
		{"'abc'", true},
		{"$", false},
		{"1a", false},
		{"'ab", false},
		{"'abcd'", false},
		{"''", false},
		{"'''", false},
		{"'a'b'", false},
		{"Expr", false},
		{"", false},
		{"%empty", false},
		{"'é'", false},
		{"' '", false},
	}
	for _, tc := range tests {
		if got := IsTerminal(tc.in); got != tc.want {
			t.Errorf("IsTerminal(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestSortTerminals(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"empty", nil, []string{}},
		{"single", []string{"a"}, []string{"a"}},
		{"code point order", []string{"id", "'+'", "'('"}, []string{"'('", "'+'", "id"}},
		{"a differs from quoted a", []string{"a", "'a'"}, []string{"'a'", "a"}},
		{"dollar last", []string{"$", "id", "'+'", "z"}, []string{"'+'", "id", "z", "$"}},
		{"dollar alone", []string{"$"}, []string{"$"}},
		{"non-ascii after ascii", []string{"é", "z", "$"}, []string{"z", "é", "$"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			set := map[string]bool{}
			for _, s := range tc.in {
				set[s] = true
			}
			if got := SortTerminals(set); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("SortTerminals(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

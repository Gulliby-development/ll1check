package grammar

import (
	"fmt"
	"reflect"
	"testing"
)

func findingLines(t *testing.T, text string) []string {
	t.Helper()
	g, err := Parse("g", text)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	var out []string
	for _, f := range Findings(g, Analyze(g)) {
		out = append(out, fmt.Sprintf("%d %s %s", f.Line, f.Code, f.Message))
	}
	return out
}

func TestFindingsTable(t *testing.T) {
	cases := []struct {
		name string
		text string
		want []string
	}{
		{
			"loose.g",
			"# a statement grammar with problems\n" +
				"Stmt -> id Tail | id '(' ')' | Block\n" +
				"Tail -> '=' num | %empty\n" +
				"Block -> '{' Stmt '}'\n" +
				"Dead -> Dead x | y\n" +
				"Loop -> Loop Loop\n",
			[]string{
				"2 LL004 Stmt: alternatives 1 and 2 both predict id",
				"5 LL001 Dead is not reachable from Stmt",
				"5 LL003 Dead is left recursive",
				"5 LL004 Dead: alternatives 1 and 2 both predict y",
				"6 LL001 Loop is not reachable from Stmt",
				"6 LL002 Loop derives no terminal string",
				"6 LL003 Loop is left recursive",
			},
		},
		{
			"empty alternatives predict end marker",
			"S -> A | B\nA -> %empty\nB -> %empty\n",
			[]string{"1 LL004 S: alternatives 1 and 2 both predict $"},
		},
		{
			"left recursion through nullable prefix",
			"A -> B A | a\nB -> %empty\n",
			[]string{
				"1 LL003 A is left recursive",
				"1 LL004 A: alternatives 1 and 2 both predict a",
			},
		},
		{
			"three alternatives give pairs in order",
			"S -> a | a | a\n",
			[]string{
				"1 LL004 S: alternatives 1 and 2 both predict a",
				"1 LL004 S: alternatives 1 and 3 both predict a",
				"1 LL004 S: alternatives 2 and 3 both predict a",
			},
		},
		{
			"continuation lines are numbered",
			"S -> a\n  | b\n  | a\n",
			[]string{"1 LL004 S: alternatives 1 and 3 both predict a"},
		},
		{
			"single alternative has no LL004",
			"S -> a S | a\nT -> b\n",
			[]string{
				"1 LL004 S: alternatives 1 and 2 both predict a",
				"2 LL001 T is not reachable from S",
			},
		},
		{
			"single alternative rule",
			"S -> a a a\n",
			nil,
		},
		{
			"shared terminals sorted and joined",
			"S -> A | B\nA -> b | a\nB -> a | b\n",
			[]string{"1 LL004 S: alternatives 1 and 2 both predict a, b"},
		},
		{
			"unproductive without left recursion",
			"S -> a | B\nB -> x B\n",
			[]string{"2 LL002 B derives no terminal string"},
		},
		{
			"disjoint alternatives",
			"S -> a | b | c\n",
			nil,
		},
		{
			"expr.g",
			"# arithmetic expressions\nExpr -> Term Expr2\nExpr2 -> '+' Term Expr2\n      | %empty\n" +
				"Term -> Factor Term2\nTerm2 -> '*' Factor Term2 | %empty\nFactor -> '(' Expr ')' | id\n",
			nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := findingLines(t, c.text)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestNotLL1(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{"clean", "S -> a\n", false},
		{"LL001 only", "S -> a\nD -> x\n", false},
		{"LL002 only", "S -> a | B\nB -> x B\n", false},
		{"LL003", "S -> S a | b\n", true},
		{"LL004", "S -> a | a\n", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g, err := Parse("g", c.text)
			if err != nil {
				t.Fatal(err)
			}
			if got := NotLL1(Findings(g, Analyze(g))); got != c.want {
				t.Errorf("NotLL1 = %v, want %v", got, c.want)
			}
		})
	}
	if NotLL1(nil) {
		t.Error("NotLL1(nil) = true")
	}
	if !NotLL1([]Finding{{1, "LL001", "x"}, {2, "LL004", "y"}}) {
		t.Error("mixed findings with LL004 should be not LL(1)")
	}
}

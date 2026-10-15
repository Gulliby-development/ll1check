package grammar

import "strings"

// SyntaxError describes a rejected token list. Index is the 1-based token
// number, 0 at the end of input.
type SyntaxError struct {
	Index    int
	Token    string
	AtEnd    bool
	Expected []string
}

// Predict returns, per rule and per alternative, the prediction set.
func Predict(g *Grammar, a *Analysis) [][]map[string]bool {
	out := make([][]map[string]bool, len(g.Rules))
	for ri, r := range g.Rules {
		out[ri] = make([]map[string]bool, len(r.Alts))
		for i, alt := range r.Alts {
			set := a.FirstOf(alt.Symbols)
			if a.AltNullable(alt) {
				addAll(set, a.Follow[r.Name])
			}
			out[ri][i] = set
		}
	}
	return out
}

// Tokenize strips one leading BOM and splits on ASCII space, tab, CR and LF.
func Tokenize(text string) []string {
	text = strings.TrimPrefix(text, "\uFEFF")
	return strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\r' || r == '\n'
	})
}

// ParseTokens runs a predictive parser and returns the applied productions.
func ParseTokens(g *Grammar, a *Analysis, tokens []string) ([]string, *SyntaxError) {
	var lines []string
	if len(g.Rules) == 0 {
		return lines, nil
	}
	preds := Predict(g, a)
	stack := []string{"$", g.Rules[0].Name}
	pos := 0
	fail := func(expected map[string]bool) *SyntaxError {
		e := &SyntaxError{Expected: SortTerminals(expected)}
		if pos < len(tokens) {
			e.Index = pos + 1
			e.Token = tokens[pos]
		} else {
			e.AtEnd = true
		}
		return e
	}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top == "$" {
			if pos < len(tokens) {
				return lines, fail(map[string]bool{"$": true})
			}
			break
		}
		if !IsNonterminal(top) {
			if pos < len(tokens) && tokens[pos] == top {
				pos++
				continue
			}
			return lines, fail(map[string]bool{top: true})
		}
		ri := g.Index[top]
		r := g.Rules[ri]
		look := "$"
		if pos < len(tokens) {
			look = tokens[pos]
		}
		chosen := -1
		if look != "$" || pos >= len(tokens) {
			for i := range r.Alts {
				if preds[ri][i][look] {
					chosen = i
					break
				}
			}
		}
		if chosen < 0 {
			exp := map[string]bool{}
			for i := range r.Alts {
				addAll(exp, preds[ri][i])
			}
			return lines, fail(exp)
		}
		syms := r.Alts[chosen].Symbols
		rhs := "%empty"
		if len(syms) > 0 {
			rhs = strings.Join(syms, " ")
		}
		lines = append(lines, r.Name+" -> "+rhs)
		for i := len(syms) - 1; i >= 0; i-- {
			stack = append(stack, syms[i])
		}
	}
	return lines, nil
}

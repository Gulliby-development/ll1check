package grammar

import (
	"fmt"
	"strings"
)

func splitTokens(line string) []string {
	tokens := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' })
	for i, token := range tokens {
		if strings.HasPrefix(token, "#") {
			return tokens[:i]
		}
	}
	return tokens
}

// Parse reads a grammar file. The error text has no "ll1check: " prefix.
func Parse(display, text string) (*Grammar, error) {
	text = strings.TrimPrefix(text, "\uFEFF")
	g := &Grammar{Display: display, Index: map[string]int{}}
	var cur *Rule
	for i, raw := range strings.Split(text, "\n") {
		lineNo := i + 1
		raw = strings.TrimSuffix(raw, "\r")
		line := strings.Trim(raw, " \t")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fail := func(format string, args ...any) error {
			return fmt.Errorf("%s:%d: %s", display, lineNo, fmt.Sprintf(format, args...))
		}
		toks := splitTokens(line)
		var rest []string
		if toks[0] == "|" {
			if cur == nil {
				return nil, fail("\"|\" without a rule")
			}
			rest = toks[1:]
		} else {
			if len(toks) < 3 || toks[1] != "->" {
				return nil, fail("expected \"NAME -> ALTERNATIVES\"")
			}
			if !IsNonterminal(toks[0]) {
				return nil, fail("invalid name %s", Show(toks[0]))
			}
			if first, ok := g.Index[toks[0]]; ok {
				return nil, fail("duplicate rule %s (first on line %d)", Show(toks[0]), g.Rules[first].Line)
			}
			cur = &Rule{Name: toks[0], Line: lineNo}
			rest = toks[2:]
		}
		var alts []Alt
		var cuts [][]string
		var curAlt []string
		for _, t := range rest {
			if t == "|" {
				cuts = append(cuts, curAlt)
				curAlt = nil
				continue
			}
			curAlt = append(curAlt, t)
		}
		cuts = append(cuts, curAlt)
		alts = alts[:0]
		for _, a := range cuts {
			if len(a) == 0 {
				return nil, fail("empty alternative (write %%empty)")
			}
			for _, s := range a {
				if s == "%empty" {
					if len(a) > 1 {
						return nil, fail("%%empty must stand alone")
					}
					continue
				}
				if !IsNonterminal(s) && !IsTerminal(s) {
					return nil, fail("invalid symbol %s", Show(s))
				}
			}
			syms := a
			if len(a) == 1 && a[0] == "%empty" {
				syms = []string{}
			}
			alts = append(alts, Alt{Line: lineNo, Symbols: syms})
		}
		cur.Alts = append(cur.Alts, alts...)
		if toks[0] != "|" {
			g.Index[cur.Name] = len(g.Rules)
			g.Rules = append(g.Rules, cur)
		}
	}
	if len(g.Rules) == 0 {
		return nil, fmt.Errorf("%s: no rules", display)
	}
	for _, r := range g.Rules {
		for _, a := range r.Alts {
			for _, s := range a.Symbols {
				if IsNonterminal(s) {
					if _, ok := g.Index[s]; !ok {
						return nil, fmt.Errorf("%s:%d: undefined rule %s", display, a.Line, Show(s))
					}
				}
			}
		}
	}
	return g, nil
}

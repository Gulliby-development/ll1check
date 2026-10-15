package grammar

// Analysis holds the results of Analyze, keyed by rule name.
type Analysis struct {
	Nullable      map[string]bool
	Reachable     map[string]bool
	Productive    map[string]bool
	LeftRecursive map[string]bool
	First         map[string]map[string]bool
	Follow        map[string]map[string]bool
}

func addAll(dst, src map[string]bool) bool {
	changed := false
	for k := range src {
		if !dst[k] {
			dst[k] = true
			changed = true
		}
	}
	return changed
}

// Analyze computes nullable, FIRST, FOLLOW, reachable, productive and
// left-recursive rules of g.
func Analyze(g *Grammar) *Analysis {
	a := &Analysis{
		Nullable:      map[string]bool{},
		Reachable:     map[string]bool{},
		Productive:    map[string]bool{},
		LeftRecursive: map[string]bool{},
		First:         map[string]map[string]bool{},
		Follow:        map[string]map[string]bool{},
	}
	for _, r := range g.Rules {
		a.First[r.Name] = map[string]bool{}
		a.Follow[r.Name] = map[string]bool{}
	}
	if len(g.Rules) == 0 {
		return a
	}
	for changed := true; changed; {
		changed = false
		for _, r := range g.Rules {
			for _, alt := range r.Alts {
				if !a.Nullable[r.Name] && a.AltNullable(alt) {
					a.Nullable[r.Name] = true
					changed = true
				}
				if !a.Productive[r.Name] && a.altProductive(alt) {
					a.Productive[r.Name] = true
					changed = true
				}
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for _, r := range g.Rules {
			for _, alt := range r.Alts {
				if addAll(a.First[r.Name], a.FirstOf(alt.Symbols)) {
					changed = true
				}
			}
		}
	}
	a.Follow[g.Rules[0].Name]["$"] = true
	for changed := true; changed; {
		changed = false
		for _, r := range g.Rules {
			for _, alt := range r.Alts {
				for i, sym := range alt.Symbols {
					if !IsNonterminal(sym) {
						continue
					}
					rest := alt.Symbols[i+1:]
					if addAll(a.Follow[sym], a.FirstOf(rest)) {
						changed = true
					}
					if a.seqNullable(rest) && addAll(a.Follow[sym], a.Follow[r.Name]) {
						changed = true
					}
				}
			}
		}
	}
	a.Reachable[g.Rules[0].Name] = true
	for changed := true; changed; {
		changed = false
		for _, r := range g.Rules {
			if !a.Reachable[r.Name] {
				continue
			}
			for _, alt := range r.Alts {
				for _, sym := range alt.Symbols {
					if IsNonterminal(sym) && !a.Reachable[sym] {
						a.Reachable[sym] = true
						changed = true
					}
				}
			}
		}
	}
	edges := map[string][]string{}
	for _, r := range g.Rules {
		for _, alt := range r.Alts {
			for _, sym := range alt.Symbols {
				if !IsNonterminal(sym) {
					break
				}
				edges[r.Name] = append(edges[r.Name], sym)
				if !a.Nullable[sym] {
					break
				}
			}
		}
	}
	for _, r := range g.Rules {
		seen := map[string]bool{}
		stack := append([]string(nil), edges[r.Name]...)
		for len(stack) > 0 {
			n := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if seen[n] {
				continue
			}
			seen[n] = true
			stack = append(stack, edges[n]...)
		}
		if seen[r.Name] {
			a.LeftRecursive[r.Name] = true
		}
	}
	return a
}

func (a *Analysis) seqNullable(syms []string) bool {
	for _, s := range syms {
		if !IsNonterminal(s) || !a.Nullable[s] {
			return false
		}
	}
	return true
}

// AltNullable reports whether the alternative can derive the empty string.
func (a *Analysis) AltNullable(alt Alt) bool {
	return a.seqNullable(alt.Symbols)
}

func (a *Analysis) altProductive(alt Alt) bool {
	for _, s := range alt.Symbols {
		if IsNonterminal(s) && !a.Productive[s] {
			return false
		}
	}
	return true
}

// FirstOf returns the terminals that can start a derivation of syms.
func (a *Analysis) FirstOf(syms []string) map[string]bool {
	out := map[string]bool{}
	for _, s := range syms {
		if !IsNonterminal(s) {
			out[s] = true
			break
		}
		addAll(out, a.First[s])
		if !a.Nullable[s] {
			break
		}
	}
	return out
}

package grammar

import "strings"

// Findings returns the check findings of g in file order per rule.
func Findings(g *Grammar, a *Analysis) []Finding {
	var out []Finding
	if len(g.Rules) == 0 {
		return out
	}
	start := g.Rules[0].Name
	preds := Predict(g, a)
	for ri, r := range g.Rules {
		if !a.Reachable[r.Name] {
			out = append(out, Finding{r.Line, "LL001", r.Name + " is not reachable from " + start})
		}
		if !a.Productive[r.Name] {
			out = append(out, Finding{r.Line, "LL002", r.Name + " derives no terminal string"})
		}
		if a.LeftRecursive[r.Name] {
			out = append(out, Finding{r.Line, "LL003", r.Name + " is left recursive"})
		}
		ps := preds[ri]
		for i := 0; i < len(ps); i++ {
			for j := i + 1; j < len(ps); j++ {
				shared := map[string]bool{}
				for t := range ps[i] {
					if ps[j][t] {
						shared[t] = true
					}
				}
				if len(shared) == 0 {
					continue
				}
				msg := r.Name + ": alternatives " + itoa(i+1) + " and " + itoa(j+1) +
					" both predict " + strings.Join(SortTerminals(shared), ", ")
				out = append(out, Finding{r.Line, "LL004", msg})
			}
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// NotLL1 reports whether any finding is LL003 or LL004.
func NotLL1(findings []Finding) bool {
	for _, f := range findings {
		if f.Code == "LL003" || f.Code == "LL004" {
			return true
		}
	}
	return false
}

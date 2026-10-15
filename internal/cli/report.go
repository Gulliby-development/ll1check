package cli

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Gulliby-development/ll1check/internal/grammar"
)

func plural(n int, word string) string {
	if n == 1 {
		return strconv.Itoa(n) + " " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}

func setText(set map[string]bool) string {
	if len(set) == 0 {
		return "-"
	}
	return strings.Join(grammar.SortTerminals(set), " ")
}

// formatSets renders the sets table; every line ends with LF.
func formatSets(g *grammar.Grammar, a *grammar.Analysis) string {
	rows := [][]string{{"RULE", "NULLABLE", "FIRST", "FOLLOW"}}
	for _, r := range g.Rules {
		nullable := "no"
		if a.Nullable[r.Name] {
			nullable = "yes"
		}
		rows = append(rows, []string{r.Name, nullable, setText(a.First[r.Name]), setText(a.Follow[r.Name])})
	}
	widths := make([]int, 4)
	for _, row := range rows {
		for i, c := range row {
			if w := utf8.RuneCountInString(c); w > widths[i] {
				widths[i] = w
			}
		}
	}
	var b strings.Builder
	for _, row := range rows {
		for i, c := range row {
			b.WriteString(c)
			if i < 3 {
				b.WriteString(strings.Repeat(" ", widths[i]-utf8.RuneCountInString(c)+2))
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// formatFinding renders one finding line without the line end.
func formatFinding(display string, f grammar.Finding) string {
	return display + ":" + strconv.Itoa(f.Line) + ": " + f.Code + " " + f.Message
}

func okLine(display string, rules int) string {
	return display + ": ok, " + plural(rules, "rule") + "\n"
}

func acceptedLine(productions int) string {
	return "accepted, " + plural(productions, "production") + "\n"
}

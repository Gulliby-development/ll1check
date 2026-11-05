package cli

import (
	"strings"

	"github.com/Gulliby-development/ll1check/internal/grammar"
)

const usage = "usage: ll1check check [--only CODE] GRAMMAR\n" +
	"       ll1check sets GRAMMAR\n" +
	"       ll1check parse [--quiet] GRAMMAR TOKENS\n"

type command struct {
	name    string
	help    bool
	version bool
	quiet   bool
	only    string
	files   []string
}

// parseArgs validates args and returns the command, or an error message.
func parseArgs(args []string) (command, string) {
	var c command
	if len(args) == 0 {
		return c, "missing command (check, sets or parse)"
	}
	if len(args) == 1 && args[0] == "--help" {
		c.help = true
		return c, ""
	}
	if len(args) == 1 && args[0] == "--version" {
		c.version = true
		return c, ""
	}
	need := 1
	switch args[0] {
	case "check", "sets":
	case "parse":
		need = 2
	default:
		return c, "unknown command " + grammar.Show(args[0])
	}
	c.name = args[0]
	var rest []string
	onlySeen := false
	for i := 1; i < len(args); i++ {
		a := args[i]
		if c.name == "check" && (a == "--only" || strings.HasPrefix(a, "--only=")) {
			if onlySeen {
				return c, "--only given twice"
			}
			onlySeen = true
			var v string
			if a == "--only" {
				if i+1 >= len(args) {
					return c, "--only needs a value"
				}
				i++
				v = args[i]
			} else {
				v = a[len("--only="):]
			}
			if !validCode(v) {
				return c, "unknown finding code " + grammar.Show(v)
			}
			c.only = v
			continue
		}
		if a == "--quiet" && c.name == "parse" {
			c.quiet = true
			continue
		}
		if len(a) > 0 && a[0] == '-' && a != "-" {
			return c, "unknown option " + grammar.Show(a)
		}
		rest = append(rest, a)
	}
	if len(rest) < need {
		return c, "missing argument"
	}
	if len(rest) > need {
		return c, "too many arguments"
	}
	if need == 2 && rest[0] == "-" && rest[1] == "-" {
		return c, "standard input given twice"
	}
	c.files = rest
	return c, ""
}

// validCode reports whether s is one of the four finding codes.
func validCode(s string) bool {
	return s == "LL001" || s == "LL002" || s == "LL003" || s == "LL004"
}

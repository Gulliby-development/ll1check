package cli

import "github.com/Gulliby-development/ll1check/internal/grammar"

const usage = "usage: ll1check check GRAMMAR\n" +
	"       ll1check sets GRAMMAR\n" +
	"       ll1check parse GRAMMAR TOKENS\n"

type command struct {
	name    string
	help    bool
	version bool
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
	rest := args[1:]
	for _, a := range rest {
		if len(a) > 0 && a[0] == '-' && a != "-" {
			return c, "unknown option " + grammar.Show(a)
		}
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

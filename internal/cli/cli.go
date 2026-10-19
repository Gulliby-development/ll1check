package cli

import (
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Gulliby-development/ll1check/internal/grammar"
)

func displayName(path string) string {
	if path == "-" {
		return "<stdin>"
	}
	return strings.ReplaceAll(path, "\\", "/")
}

// readInput reads a file or standard input; msg is empty on success.
func readInput(path string, stdin io.Reader) (string, string) {
	display := displayName(path)
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return "", display + ": cannot read the file"
	}
	if !utf8.Valid(data) {
		return "", display + ": input is not valid UTF-8"
	}
	return string(data), ""
}

func fail(stderr io.Writer, msg string, code int) int {
	_, _ = io.WriteString(stderr, "ll1check: "+msg+"\n")
	return code
}

// Run executes the command line and returns the exit code.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	c, msg := parseArgs(args)
	if msg != "" {
		return fail(stderr, msg, 2)
	}
	if c.help {
		_, _ = io.WriteString(stdout, usage)
		return 0
	}
	if c.version {
		version := "devel"
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
		_, _ = io.WriteString(stdout, "ll1check "+version+"\n")
		return 0
	}
	display := displayName(c.files[0])
	text, msg := readInput(c.files[0], stdin)
	if msg != "" {
		return fail(stderr, msg, 2)
	}
	g, err := grammar.Parse(display, text)
	if err != nil {
		return fail(stderr, err.Error(), 2)
	}
	a := grammar.Analyze(g)
	switch c.name {
	case "sets":
		_, _ = io.WriteString(stdout, formatSets(g, a))
		return 0
	case "check":
		findings := grammar.Findings(g, a)
		if len(findings) == 0 {
			_, _ = io.WriteString(stdout, okLine(display, len(g.Rules)))
			return 0
		}
		var b strings.Builder
		for _, f := range findings {
			b.WriteString(formatFinding(display, f))
			b.WriteString("\n")
		}
		_, _ = io.WriteString(stdout, b.String())
		return 1
	}
	if grammar.NotLL1(grammar.Findings(g, a)) {
		return fail(stderr, display+": the grammar is not LL(1) (run check)", 2)
	}
	tokDisplay := displayName(c.files[1])
	tokText, msg := readInput(c.files[1], stdin)
	if msg != "" {
		return fail(stderr, msg, 2)
	}
	lines, serr := grammar.ParseTokens(g, a, grammar.Tokenize(tokText))
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l)
		b.WriteString("\n")
	}
	if serr == nil {
		b.WriteString(acceptedLine(len(lines)))
		_, _ = io.WriteString(stdout, b.String())
		return 0
	}
	_, _ = io.WriteString(stdout, b.String())
	exp := make([]string, len(serr.Expected))
	for i, e := range serr.Expected {
		if e == "$" {
			e = "end of input"
		}
		exp[i] = e
	}
	m := tokDisplay + ": "
	if serr.AtEnd {
		m += "unexpected end of input"
	} else {
		m += "token " + strconv.Itoa(serr.Index) + ": unexpected " + grammar.Show(serr.Token)
	}
	m += " (expected " + strings.Join(exp, ", ") + ")"
	return fail(stderr, m, 1)
}

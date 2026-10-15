package cli

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

const wantUsage = "usage: ll1check check GRAMMAR\n" +
	"       ll1check sets GRAMMAR\n" +
	"       ll1check parse GRAMMAR TOKENS\n"

var argErrorCases = []struct {
	name string
	args []string
	msg  string
}{
	{"no arguments", []string{}, "missing command (check, sets or parse)"},
	{"unknown command", []string{"foo"}, `unknown command "foo"`},
	{"unknown option", []string{"check", "-x", "f"}, `unknown option "-x"`},
	{"check missing argument", []string{"check"}, "missing argument"},
	{"sets missing argument", []string{"sets"}, "missing argument"},
	{"parse missing argument", []string{"parse", "g"}, "missing argument"},
	{"check too many", []string{"check", "a", "b"}, "too many arguments"},
	{"parse too many", []string{"parse", "a", "b", "c"}, "too many arguments"},
	{"stdin twice", []string{"parse", "-", "-"}, "standard input given twice"},
	{"option before too many", []string{"check", "-x", "a", "b"}, `unknown option "-x"`},
	{"option before missing", []string{"parse", "--help"}, `unknown option "--help"`},
	{"help after command", []string{"check", "--help"}, `unknown option "--help"`},
	{"help as first of two", []string{"--help", "x"}, `unknown command "--help"`},
	{"non-ASCII command", []string{"é\t"}, `unknown command "??"`},
	{"non-ASCII option", []string{"sets", "-ü日"}, `unknown option "-??"`},
	{"empty command", []string{""}, `unknown command ""`},
}

func TestParseArgsErrors(t *testing.T) {
	for _, tc := range argErrorCases {
		t.Run(tc.name, func(t *testing.T) {
			_, msg := parseArgs(tc.args)
			if msg != tc.msg {
				t.Fatalf("parseArgs(%q) message = %q, want %q", tc.args, msg, tc.msg)
			}
		})
	}
}

func TestRunArgErrors(t *testing.T) {
	for _, tc := range argErrorCases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errb bytes.Buffer
			code := Run(tc.args, strings.NewReader(""), &out, &errb)
			if code != 2 {
				t.Fatalf("exit = %d, want 2", code)
			}
			if out.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", out.String())
			}
			if want := "ll1check: " + tc.msg + "\n"; errb.String() != want {
				t.Fatalf("stderr = %q, want %q", errb.String(), want)
			}
		})
	}
}

func TestParseArgsValid(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want command
	}{
		{"help", []string{"--help"}, command{help: true}},
		{"check file", []string{"check", "g"}, command{name: "check", files: []string{"g"}}},
		{"check stdin", []string{"check", "-"}, command{name: "check", files: []string{"-"}}},
		{"sets stdin", []string{"sets", "-"}, command{name: "sets", files: []string{"-"}}},
		{"parse grammar stdin", []string{"parse", "-", "t"}, command{name: "parse", files: []string{"-", "t"}}},
		{"parse tokens stdin", []string{"parse", "g", "-"}, command{name: "parse", files: []string{"g", "-"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, msg := parseArgs(tc.args)
			if msg != "" {
				t.Fatalf("unexpected message %q", msg)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("parseArgs(%q) = %+v, want %+v", tc.args, got, tc.want)
			}
		})
	}
}

func TestRunHelp(t *testing.T) {
	var out, errb bytes.Buffer
	code := Run([]string{"--help"}, strings.NewReader(""), &out, &errb)
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if out.String() != wantUsage {
		t.Fatalf("stdout = %q, want %q", out.String(), wantUsage)
	}
	if errb.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errb.String())
	}
}

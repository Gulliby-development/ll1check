<p align="center"><img src="docs/logo.svg" width="96" alt="" /></p>

# ll1check

[![CI](https://github.com/Gulliby-development/ll1check/actions/workflows/ci.yml/badge.svg)](https://github.com/Gulliby-development/ll1check/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

ll1check is a command-line checker for LL(1) grammars. It is for people who write a parser by hand and want to know, before coding, whether one token of lookahead is enough. It reports rules that cannot be used and decisions one token cannot make, prints the nullable, FIRST and FOLLOW sets of every rule, and runs a predictive parser over a list of tokens. It works offline, uses only the Go standard library and prints the same on Windows, macOS and Linux (every line ends with LF).

## Install

```
go build -o ll1check .
```

ll1check needs Go 1.24 or newer (the `go` line of `go.mod`).

## Usage

```
usage: ll1check check GRAMMAR
       ll1check sets GRAMMAR
       ll1check parse [--quiet] GRAMMAR TOKENS
```

GRAMMAR and TOKENS are paths or `-` for standard input (named `<stdin>`). In messages a backslash in a path is shown as `/`. `--help` is accepted only as the only argument.

- `check` prints one line per finding and exits 1, or an `ok` line with the number of rules and exits 0.
- `sets` prints a table with the nullable flag, FIRST and FOLLOW of every rule.
- `parse` runs a predictive parser over the tokens and prints the productions it applies.

`ll1check --version` (as the only argument) prints `ll1check` and the module version from the build information (a tag or a pseudo-version, or `devel` when the build has none) on one line, and exits 0.

A grammar file `expr.g`:

```
# arithmetic expressions
Expr -> Term Expr2
Expr2 -> '+' Term Expr2
      | %empty
Term -> Factor Term2
Term2 -> '*' Factor Term2 | %empty
Factor -> '(' Expr ')' | id
```

`ll1check sets expr.g` prints

```
RULE    NULLABLE  FIRST   FOLLOW
Expr    no        '(' id  ')' $
Expr2   yes       '+'     ')' $
Term    no        '(' id  ')' '+' $
Term2   yes       '*'     ')' '+' $
Factor  no        '(' id  ')' '*' '+' $
```

With a file `sum.txt` holding `id '+' id`, `ll1check parse expr.g sum.txt` prints (exit 0)

```
Expr -> Term Expr2
Term -> Factor Term2
Factor -> id
Term2 -> %empty
Expr2 -> '+' Term Expr2
Term -> Factor Term2
Factor -> id
Term2 -> %empty
Expr2 -> %empty
accepted, 9 productions
```

## Grammar file

The file is UTF-8. One leading BOM is skipped. Lines are split at LF and one CR before the LF is removed. Line numbers count every physical line from 1. A line is trimmed of spaces and tabs; an empty line or one starting with `#` is skipped. Tokens are separated by runs of spaces and tabs. On a rule line or a `|` line the first token that starts with `#`, and everything after it, is a comment (`'#'` in quotes is an ordinary terminal; `a#b` is an invalid symbol, not a comment):

```
Expr2 -> '+' Term Expr2     # more terms
       | %empty             # or none
```

- A rule line is `NAME -> ALT | ALT ...`.
- A line whose first token is `|` adds alternatives to the latest rule.
- `|` and `->` are separate tokens.
- An ALT is one or more symbols, or the single token `%empty` (the empty alternative).
- NAME is a nonterminal name: an uppercase ASCII letter followed by ASCII letters, digits and `_` (`Expr`, `Expr2`).
- A symbol is a nonterminal name, a terminal name (a lowercase ASCII letter followed by letters, digits and `_`, such as `id`) or a quoted terminal (a quote, 1 to 3 characters from `!` to `~` other than the quote, and a quote, such as `'+'`, `'<='`, `'|'`).
- `$` stands for the end of input in output and is never a symbol of the grammar.

The first problem ends the reading with exit 2. Per line the checks run in this order: the line shape, a `|` without a rule, an invalid name, a duplicate rule, then the alternatives left to right (empty alternative, `%empty` not alone, invalid symbol). After the last line, a nonterminal with no rule is reported as an undefined rule. A file with no rules is an error.

Definitions:

- The start rule is the first rule.
- A rule is nullable when one of its alternatives is `%empty` or has only nullable nonterminals.
- FIRST(X) is the set of terminals that can start a string derived from X.
- FOLLOW(X) is the set of terminals, and `$`, that can come right after X in a derivation from the start rule; `$` is in FOLLOW of the start rule. FIRST and FOLLOW are the least solution over all rules, also the unreachable and unproductive ones.
- The prediction set of an alternative is the FIRST of its symbols in order (a symbol counts for the next one only when it is a nullable nonterminal), plus FOLLOW of its rule when the whole alternative can derive the empty string.
- A rule is reachable when it is the start rule or occurs in an alternative of a reachable rule.
- A rule is productive when it derives some string of terminals (possibly empty).
- A rule is left recursive when it can derive a string that starts with itself after one or more steps, a step going from a rule to the nonterminals in the leading symbols of its alternatives (the first symbol, and the next while the previous is a nullable nonterminal).
- Terminals are compared and sorted by their text as written (`a` and `'a'` differ), by code point, with `$` last.

## Findings

`check` prints one line per finding in the form `FILE:LINE: CODE MESSAGE`, with the line of the rule's name, and exits 1. Per rule, in file order, the codes are:

- `LL001`: the rule is not reachable from the start rule.
- `LL002`: the rule derives no terminal string.
- `LL003`: the rule is left recursive.
- `LL004`: for every pair of alternatives I < J (numbered from 1 within the rule, continuation lines included) whose prediction sets share terminals, the shared terminals are listed.

With no findings it prints `FILE: ok, N rules` (`1 rule`) and exits 0.

### How parse works

The parser has the end marker and the start rule on its stack. A nonterminal on top is replaced by the alternative whose prediction set holds the next token (`$` when the tokens are used up), and one line `X -> SYMBOLS` is printed (`%empty` for none). A terminal on top must equal the next token and consumes it. When the tokens are used up exactly, it prints `accepted, N productions` (`1 production`) and exits 0.

TOKENS holds tokens separated by runs of ASCII space, tab, CR and LF (one leading BOM skipped); it may hold none. A syntax error exits 1 and keeps the lines already printed on stdout. Stderr gets `ll1check: TOKENS: token N: unexpected "X" (expected A, B)` or `ll1check: TOKENS: unexpected end of input (expected A, B)`. N counts tokens from 1. The expected list is the union of the prediction sets of the rule on top (a terminal on top: that terminal), sorted, with `end of input` in place of `$`, last.

With `parse --quiet` the productions are not printed: an accepted input prints only the `accepted, N productions` line, and a syntax error prints nothing on stdout (stderr and the exit code stay as they are). `--quiet` is an option of `parse` only and may stand anywhere among its arguments; `check --quiet` and `sets --quiet` are `unknown option "--quiet"` (exit 2). With the files above, `ll1check parse --quiet expr.g sum.txt` prints (exit 0):

```
accepted, 9 productions
```

`parse` refuses a grammar that is not LL(1): when `check` would print an LL003 or LL004 finding, it exits 2 before TOKENS is read.

## Exit codes

- 0: success.
- 1: findings from `check`, or a syntax error from `parse`.
- 2: usage or input error (bad arguments, a file that cannot be read or is not valid UTF-8, a grammar error, or a grammar that is not LL(1) for `parse`). Errors print `ll1check: MESSAGE` on stderr and nothing on stdout.

## Development

```
go vet ./...
gofmt -l .
go test ./...
```

`gofmt -l .` must print nothing. The code uses only the Go standard library.

## License

MIT, see [LICENSE](LICENSE).

---

Built with BAKED, an open-source coding agent (bench loop round 27).

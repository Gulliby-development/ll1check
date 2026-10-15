package grammar

// Alt is one alternative of a rule. An empty Symbols slice is %empty.
type Alt struct {
	Line    int
	Symbols []string
}

// Rule is a nonterminal with its alternatives, in file order.
type Rule struct {
	Name string
	Line int
	Alts []Alt
}

// Grammar is a parsed grammar; the start rule is Rules[0].
type Grammar struct {
	Display string
	Rules   []*Rule
	Index   map[string]int
}

// Finding is one problem reported by check.
type Finding struct {
	Line    int
	Code    string
	Message string
}

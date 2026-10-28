// Package abnf reads a grammar written in ABNF (RFC 5234, with RFC 7405
// case-sensitive strings) and matches strings against its rules.
//
// Two readability extensions are accepted: "|" as well as "/" for
// alternation, and any Unicode character inside a quoted string. Rule
// names are case-insensitive and may contain "_" as well as "-".
//
// Matching explores alternatives in written order with full backtracking,
// and the first derivation that consumes the whole input wins. Repetition
// is greedy first. Where a grammar is ambiguous the written order decides.
package abnf

import (
	"fmt"
	"sort"
	"strings"
)

// Grammar is a set of named rules.
type Grammar struct {
	rules map[string]*rule  // keyed by the lower-cased name
	order []string          // user-defined names in definition order
	refs  map[string]string // lower-cased referenced name -> spelling first seen
}

type rule struct {
	name string // as first defined
	body element
	core bool
}

// coreRules are RFC 5234 appendix B.1. A grammar may redefine any of them.
const coreRules = `
ALPHA  = %x41-5A / %x61-7A
BIT    = "0" / "1"
CHAR   = %x01-7F
CR     = %x0D
CRLF   = CR LF
CTL    = %x00-1F / %x7F
DIGIT  = %x30-39
DQUOTE = %x22
HEXDIG = DIGIT / "A" / "B" / "C" / "D" / "E" / "F"
HTAB   = %x09
LF     = %x0A
LWSP   = *(WSP / CRLF WSP)
OCTET  = %x00-FF
SP     = %x20
VCHAR  = %x21-7E
WSP    = SP / HTAB
`

// Parse reads a grammar. Core rules (DIGIT, ALPHA, SP, ...) are defined
// unless the grammar defines them itself.
func Parse(src string) (*Grammar, error) {
	g := &Grammar{rules: map[string]*rule{}, refs: map[string]string{}}
	if err := g.define(coreRules, true); err != nil {
		panic("abnf: core rules: " + err.Error())
	}
	if err := g.define(src, false); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *Grammar) define(src string, core bool) error {
	for _, d := range logicalRules(src) {
		name, op, body, err := splitRule(d.text)
		if err != nil {
			return fmt.Errorf("line %d: %w", d.line, err)
		}
		p := &parser{src: []rune(body), g: g}
		e, err := p.alternation()
		if err != nil {
			return fmt.Errorf("line %d: %s: %w", d.line, name, err)
		}
		if !p.atEnd() {
			return fmt.Errorf("line %d: %s: unexpected %q", d.line, name, string(p.src[p.pos:]))
		}
		key := strings.ToLower(name)
		existing, ok := g.rules[key]
		switch {
		case op == "=/":
			if !ok {
				return fmt.Errorf("line %d: %s =/ before %s is defined", d.line, name, name)
			}
			existing.body = alt{existing.body, e}
		case ok && existing.core && !core:
			// A grammar's own definition replaces the core rule.
			existing.body, existing.core = e, false
			g.order = append(g.order, existing.name)
		case ok:
			return fmt.Errorf("line %d: %s defined twice", d.line, name)
		default:
			g.rules[key] = &rule{name: name, body: e, core: core}
			if !core {
				g.order = append(g.order, name)
			}
		}
	}
	return nil
}

type decl struct {
	line int
	text string
}

// logicalRules joins continuation lines (those starting with whitespace)
// onto the rule they continue and strips comments.
func logicalRules(src string) []decl {
	var out []decl
	for i, line := range strings.Split(src, "\n") {
		line = strings.TrimRight(line, "\r")
		if j := commentStart(line); j >= 0 {
			line = line[:j]
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(out) > 0 && (line[0] == ' ' || line[0] == '\t') {
			out[len(out)-1].text += " " + line
			continue
		}
		out = append(out, decl{line: i + 1, text: line})
	}
	return out
}

// commentStart is the index of a ";" outside a quoted string, or -1.
func commentStart(line string) int {
	quoted := false
	for i, r := range line {
		switch {
		case r == '"':
			quoted = !quoted
		case r == ';' && !quoted:
			return i
		}
	}
	return -1
}

func splitRule(text string) (name, op, body string, err error) {
	text = strings.TrimSpace(text)
	i := strings.IndexAny(text, " \t=")
	if i <= 0 {
		return "", "", "", fmt.Errorf("expected a rule name in %q", text)
	}
	name = text[:i]
	if !isRuleName(name) {
		return "", "", "", fmt.Errorf("bad rule name %q", name)
	}
	rest := strings.TrimSpace(text[i:])
	switch {
	case strings.HasPrefix(rest, "=/"):
		op, body = "=/", rest[2:]
	case strings.HasPrefix(rest, "="):
		op, body = "=", rest[1:]
	default:
		return "", "", "", fmt.Errorf("expected = after %s", name)
	}
	if strings.TrimSpace(body) == "" {
		return "", "", "", fmt.Errorf("%s has no definition", name)
	}
	return name, op, body, nil
}

func isRuleName(s string) bool {
	for i, r := range s {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z':
		case i > 0 && (r >= '0' && r <= '9' || r == '-' || r == '_'):
		default:
			return false
		}
	}
	return s != ""
}

// Rules is every rule the grammar defines, in definition order; core rules
// are included only if the grammar redefined them.
func (g *Grammar) Rules() []string { return append([]string(nil), g.order...) }

// Undefined is every rule name referenced but not defined, sorted.
func (g *Grammar) Undefined() []string {
	var out []string
	for key, name := range g.refs {
		if _, ok := g.rules[key]; !ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

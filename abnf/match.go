package abnf

import (
	"fmt"
	"strings"
	"unicode"
)

// Node is one rule's match: the rule name as defined, the span of the
// input it covers as byte offsets, its text, and the rules matched within
// it, in order.
type Node struct {
	Rule     string
	Start    int
	End      int
	Text     string
	Children []*Node
}

// Find is the first direct child matched by the rule, or nil. Names are
// case-insensitive.
func (n *Node) Find(rule string) *Node {
	for _, c := range n.Children {
		if strings.EqualFold(c.Rule, rule) {
			return c
		}
	}
	return nil
}

// All is every direct child matched by the rule, in order.
func (n *Node) All(rule string) []*Node {
	var out []*Node
	for _, c := range n.Children {
		if strings.EqualFold(c.Rule, rule) {
			out = append(out, c)
		}
	}
	return out
}

// Match derives input from the named rule, consuming all of it. The error
// names the furthest byte offset the match reached.
func (g *Grammar) Match(rule string, input string) (*Node, error) {
	return g.MatchWhere(rule, input, nil)
}

// MatchWhere is Match with a constraint the grammar cannot state: accept
// is called as each rule's node completes, with its children in place, and
// a false rejects that derivation so that matching backtracks to the next.
func (g *Grammar) MatchWhere(rule string, input string, accept func(*Node) bool) (*Node, error) {
	r, ok := g.rules[strings.ToLower(rule)]
	if !ok {
		return nil, fmt.Errorf("abnf: no rule %s", rule)
	}
	in := []rune(input)
	offs := make([]int, len(in)+1)
	b := 0
	for i, c := range in {
		offs[i] = b
		b += len(string(c))
	}
	offs[len(in)] = b
	m := &matcher{g: g, in: in, offs: offs, input: input, accept: accept}
	root := &Node{}
	m.cur = root
	matched := m.rule(r, 0, func(end int) bool { return end == len(in) })
	if m.err != nil {
		return nil, m.err
	}
	if !matched {
		return nil, fmt.Errorf("abnf: %s does not match at byte %d", r.name, offs[m.far])
	}
	return root.Children[0], nil
}

type matcher struct {
	g      *Grammar
	in     []rune
	offs   []int
	input  string
	far    int   // furthest rune position a terminal failed at
	cur    *Node // the node whose children are being collected
	err    error // an undefined rule was reached
	accept func(*Node) bool
}

func (m *matcher) fail(pos int) bool {
	if pos > m.far {
		m.far = pos
	}
	return false
}

func (m *matcher) rule(r *rule, pos int, k func(int) bool) bool {
	parent := m.cur
	n := &Node{Rule: r.name, Start: m.offs[pos]}
	m.cur = n
	ok := m.match(r.body, pos, func(end int) bool {
		n.End = m.offs[end]
		n.Text = m.input[n.Start:n.End]
		if m.accept != nil && !m.accept(n) {
			return false
		}
		parent.Children = append(parent.Children, n)
		m.cur = parent
		if k(end) {
			return true
		}
		parent.Children = parent.Children[:len(parent.Children)-1]
		m.cur = n
		return false
	})
	m.cur = parent
	return ok
}

func (m *matcher) match(e element, pos int, k func(int) bool) bool {
	switch e := e.(type) {
	case ref:
		r, ok := m.g.rules[strings.ToLower(string(e))]
		if !ok {
			m.err = fmt.Errorf("abnf: undefined rule %s", string(e))
			return false
		}
		return m.rule(r, pos, k)
	case str:
		for i, want := range e.s {
			if pos+i >= len(m.in) || !equalRune(m.in[pos+i], want, e.fold) {
				return m.fail(pos + i)
			}
		}
		return k(pos + len(e.s))
	case rng:
		if pos >= len(m.in) || m.in[pos] < e.lo || m.in[pos] > e.hi {
			return m.fail(pos)
		}
		return k(pos + 1)
	case seq:
		return m.seq(e, pos, k)
	case alt:
		for _, a := range e {
			if m.match(a, pos, k) {
				return true
			}
		}
		return false
	case rep:
		return m.rep(e, pos, 0, k)
	case prose:
		panic(fmt.Sprintf("abnf: prose rule <%s> cannot be matched", string(e)))
	}
	panic(fmt.Sprintf("abnf: unknown element %T", e))
}

func (m *matcher) seq(items seq, pos int, k func(int) bool) bool {
	if len(items) == 0 {
		return k(pos)
	}
	return m.match(items[0], pos, func(p int) bool { return m.seq(items[1:], p, k) })
}

// rep is greedy: more repetitions are tried before fewer. An empty match
// of the repeated element is not counted, so *[x] cannot loop.
func (m *matcher) rep(e rep, pos, count int, k func(int) bool) bool {
	if e.max < 0 || count < e.max {
		if m.match(e.e, pos, func(p int) bool {
			if p == pos {
				return false
			}
			return m.rep(e, p, count+1, k)
		}) {
			return true
		}
	}
	if count >= e.min {
		return k(pos)
	}
	return false
}

func equalRune(a, b rune, fold bool) bool {
	if a == b {
		return true
	}
	return fold && unicode.ToLower(a) == unicode.ToLower(b)
}

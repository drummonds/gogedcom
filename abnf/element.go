package abnf

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// element is one node of a rule's definition.
type element interface{ isElement() }

type (
	alt []element // alternation, in written order
	seq []element // concatenation
	rep struct {  // repetition; max < 0 is unbounded
		min, max int
		e        element
	}
	ref string // a rule name as written
	str struct {
		s    []rune
		fold bool // case-insensitive
	}
	rng   struct{ lo, hi rune } // one code point in a range
	prose string
)

func (alt) isElement()   {}
func (seq) isElement()   {}
func (rep) isElement()   {}
func (ref) isElement()   {}
func (str) isElement()   {}
func (rng) isElement()   {}
func (prose) isElement() {}

// parser reads one rule body.
type parser struct {
	src []rune
	pos int
	g   *Grammar
}

func (p *parser) atEnd() bool { p.skip(); return p.pos >= len(p.src) }

func (p *parser) skip() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

func (p *parser) peek() rune {
	p.skip()
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) alternation() (element, error) {
	var out alt
	for {
		c, err := p.concatenation()
		if err != nil {
			return nil, err
		}
		out = append(out, c)
		if r := p.peek(); r == '/' || r == '|' {
			p.pos++
			continue
		}
		break
	}
	if len(out) == 1 {
		return out[0], nil
	}
	return out, nil
}

func (p *parser) concatenation() (element, error) {
	var out seq
	for {
		before := p.pos
		r := p.peek()
		if r == 0 || r == '/' || r == '|' || r == ')' || r == ']' {
			break
		}
		if len(out) > 0 && p.pos == before {
			return nil, fmt.Errorf("expected whitespace before %q", string(r))
		}
		e, err := p.repetition()
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("expected an element")
	}
	if len(out) == 1 {
		return out[0], nil
	}
	return out, nil
}

func (p *parser) repetition() (element, error) {
	p.skip()
	min, max, hasRepeat := 1, 1, false
	if n, ok := p.digits(); ok {
		min, max, hasRepeat = n, n, true
	}
	if p.pos < len(p.src) && p.src[p.pos] == '*' {
		p.pos++
		if !hasRepeat {
			min = 0
		}
		max = -1
		if n, ok := p.digits(); ok {
			max = n
		}
		hasRepeat = true
	}
	if hasRepeat && max >= 0 && min > max {
		return nil, fmt.Errorf("repeat %d*%d has min above max", min, max)
	}
	e, err := p.element()
	if err != nil {
		return nil, err
	}
	if !hasRepeat {
		return e, nil
	}
	return rep{min: min, max: max, e: e}, nil
}

func (p *parser) digits() (int, bool) {
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
		p.pos++
	}
	if p.pos == start {
		return 0, false
	}
	n, _ := strconv.Atoi(string(p.src[start:p.pos]))
	return n, true
}

func (p *parser) element() (element, error) {
	p.skip()
	if p.pos >= len(p.src) {
		return nil, fmt.Errorf("expected an element")
	}
	switch r := p.src[p.pos]; {
	case r == '(':
		p.pos++
		e, err := p.alternation()
		if err != nil {
			return nil, err
		}
		if p.peek() != ')' {
			return nil, fmt.Errorf("expected )")
		}
		p.pos++
		return e, nil
	case r == '[':
		p.pos++
		e, err := p.alternation()
		if err != nil {
			return nil, err
		}
		if p.peek() != ']' {
			return nil, fmt.Errorf("expected ]")
		}
		p.pos++
		return rep{min: 0, max: 1, e: e}, nil
	case r == '"':
		return p.quoted(true)
	case r == '%':
		return p.terminalValue()
	case r == '<':
		end := indexRune(p.src, p.pos, '>')
		if end < 0 {
			return nil, fmt.Errorf("unterminated prose value")
		}
		e := prose(string(p.src[p.pos+1 : end]))
		p.pos = end + 1
		return e, nil
	case unicode.IsLetter(r):
		start := p.pos
		for p.pos < len(p.src) && (unicode.IsLetter(p.src[p.pos]) || unicode.IsDigit(p.src[p.pos]) || p.src[p.pos] == '-' || p.src[p.pos] == '_') {
			p.pos++
		}
		name := string(p.src[start:p.pos])
		key := strings.ToLower(name)
		if _, seen := p.g.refs[key]; !seen {
			p.g.refs[key] = name
		}
		return ref(name), nil
	default:
		return nil, fmt.Errorf("unexpected %q", string(r))
	}
}

func (p *parser) quoted(fold bool) (element, error) {
	end := indexRune(p.src, p.pos+1, '"')
	if end < 0 {
		return nil, fmt.Errorf("unterminated string")
	}
	s := append([]rune(nil), p.src[p.pos+1:end]...)
	p.pos = end + 1
	return str{s: s, fold: fold}, nil
}

// terminalValue reads %x.., %d.., %b.., %s"..", %i"..".
func (p *parser) terminalValue() (element, error) {
	if p.pos+1 >= len(p.src) {
		return nil, fmt.Errorf("bad terminal value")
	}
	kind := unicode.ToLower(p.src[p.pos+1])
	p.pos += 2
	switch kind {
	case 's':
		return p.quoted(false)
	case 'i':
		return p.quoted(true)
	}
	base := map[rune]int{'x': 16, 'd': 10, 'b': 2}[kind]
	if base == 0 {
		return nil, fmt.Errorf("bad terminal value %%%c", kind)
	}
	first, err := p.number(base)
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.src) && p.src[p.pos] == '-' {
		p.pos++
		last, err := p.number(base)
		if err != nil {
			return nil, err
		}
		return rng{lo: first, hi: last}, nil
	}
	s := []rune{first}
	for p.pos < len(p.src) && p.src[p.pos] == '.' {
		p.pos++
		r, err := p.number(base)
		if err != nil {
			return nil, err
		}
		s = append(s, r)
	}
	if len(s) == 1 {
		return rng{lo: first, hi: first}, nil
	}
	return str{s: s, fold: false}, nil
}

func (p *parser) number(base int) (rune, error) {
	start := p.pos
	for p.pos < len(p.src) && isDigitIn(p.src[p.pos], base) {
		p.pos++
	}
	if p.pos == start {
		return 0, fmt.Errorf("expected a number")
	}
	n, err := strconv.ParseInt(string(p.src[start:p.pos]), base, 32)
	if err != nil {
		return 0, err
	}
	return rune(n), nil
}

func isDigitIn(r rune, base int) bool {
	switch base {
	case 2:
		return r == '0' || r == '1'
	case 10:
		return r >= '0' && r <= '9'
	default:
		return r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F'
	}
}

func indexRune(s []rune, from int, r rune) int {
	for i := from; i < len(s); i++ {
		if s[i] == r {
			return i
		}
	}
	return -1
}

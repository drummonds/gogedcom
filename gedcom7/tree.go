package gedcom7

import (
	"fmt"

	"git.bytestone.uk/hum3/gotreesitter"
)

// parse runs the tree-sitter parser over src. The tree is returned even
// when it holds errors; firstError finds them.
func parse(src []byte) (*gotreesitter.Tree, *gotreesitter.Language, error) {
	lang := Gedcom7Language()
	tree, err := gotreesitter.NewParser(lang).Parse(src)
	if err != nil {
		return nil, nil, err
	}
	return tree, lang, nil
}

// firstError is the first error or missing node under n, or nil.
func firstError(n *gotreesitter.Node) *gotreesitter.Node {
	if !n.HasError() && !n.IsMissing() {
		return nil
	}
	if n.IsError() || n.IsMissing() {
		return n
	}
	for _, c := range n.Children() {
		if e := firstError(c); e != nil {
			return e
		}
	}
	return n
}

// firstGap is the first byte position under n that no leaf covers, and
// whether there is one. gotreesitter (v0.7.0) skips a character the lexer
// cannot match where the C parser reports an ERROR node (gotreesitter
// issue #9: a second space after the level, a tab in a name, a lowercase
// z after a time), so the tree is checked for text it does not account
// for.
func firstGap(n *gotreesitter.Node) (gotreesitter.Point, bool) {
	at := n.StartByte()
	point := n.StartPoint()
	for _, c := range n.Children() {
		if c.StartByte() > at {
			return point, true
		}
		if p, ok := firstGap(c); ok {
			return p, true
		}
		at, point = c.EndByte(), c.EndPoint()
	}
	return point, n.ChildCount() > 0 && at < n.EndByte()
}

// parseError is the error the tree holds, or nil: the first error or
// missing node, else the first gap. line is the file line the parsed text
// starts on.
func parseError(root *gotreesitter.Node, lang *gotreesitter.Language, line int) error {
	if e := firstError(root); e != nil {
		return errorAt(e, lang, line)
	}
	if p, ok := firstGap(root); ok {
		return fmt.Errorf("line %d: unexpected text at column %d", int(p.Row)+line, p.Column+1)
	}
	return nil
}

// errorAt describes an error node by its line and column (1-based); line
// is the file line the parsed text starts on.
func errorAt(n *gotreesitter.Node, lang *gotreesitter.Language, line int) error {
	p := n.StartPoint()
	p.Row += uint32(line - 1)
	what := "unexpected text"
	if n.IsMissing() {
		what = "missing " + n.Type(lang)
	}
	return fmt.Errorf("line %d: %s at column %d", p.Row+1, what, p.Column+1)
}

// child is the first named child of the type, or nil.
func child(n *gotreesitter.Node, lang *gotreesitter.Language, typ string) *gotreesitter.Node {
	for _, c := range n.Children() {
		if c.IsNamed() && c.Type(lang) == typ {
			return c
		}
	}
	return nil
}

// children is every named child of the type, in order.
func children(n *gotreesitter.Node, lang *gotreesitter.Language, typ string) []*gotreesitter.Node {
	var out []*gotreesitter.Node
	for _, c := range n.Children() {
		if c.IsNamed() && c.Type(lang) == typ {
			out = append(out, c)
		}
	}
	return out
}

// firstNamed is the first named child, or nil.
func firstNamed(n *gotreesitter.Node) *gotreesitter.Node {
	for _, c := range n.Children() {
		if c.IsNamed() {
			return c
		}
	}
	return nil
}

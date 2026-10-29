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

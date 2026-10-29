package gedcom7

import (
	"fmt"
	"strings"

	"git.bytestone.uk/hum3/gotreesitter"
)

// payload is a data type's text parsed as the payload of a level-0 line,
// "0 TAG <s>": the grammar types a payload by its tag, so that line is how a
// DateValue, Time, Age, PersonalName, Latitude or Longitude is parsed on its
// own. line is the typed line node (date_line, time_line, ...).
type payload struct {
	tree *gotreesitter.Tree
	src  []byte
	lang *gotreesitter.Language
	line *gotreesitter.Node
}

// parsePayload parses s as the payload of tag. what names the data type in
// errors. The caller releases the result.
func parsePayload(what, tag, s string) (*payload, error) {
	if strings.ContainsAny(s, "\r\n") {
		return nil, fmt.Errorf("gedcom7: %s %q: line break", what, s)
	}
	line := "0 " + tag
	if s != "" {
		line += " " + s
	}
	// gotreesitter (v0.7.0) mis-parses input that ends right after a GLR
	// fork; a trailing EOL avoids it.
	src := []byte(line + "\n")
	tree, lang, err := parse(src)
	if err != nil {
		return nil, fmt.Errorf("gedcom7: %s %q: %w", what, s, err)
	}
	root := tree.RootNode()
	if err := parseError(root, lang, 1); err != nil {
		tree.Release()
		return nil, fmt.Errorf("gedcom7: %s %q: %w", what, s, err)
	}
	typed := child(child(root, lang, "line"), lang, strings.ToLower(tag)+"_line")
	if typed == nil {
		tree.Release()
		return nil, fmt.Errorf("gedcom7: %s %q: not a %s line", what, s, tag)
	}
	return &payload{tree: tree, src: src, lang: lang, line: typed}, nil
}

func (p *payload) release() { p.tree.Release() }

// value is the typed line's value node, or nil when the payload is empty.
func (p *payload) value(typ string) *gotreesitter.Node { return child(p.line, p.lang, typ) }

// text is the source text of n.
func (p *payload) text(n *gotreesitter.Node) string { return n.Text(p.src) }

// childText is the text of n's first child of the type, or "".
func (p *payload) childText(n *gotreesitter.Node, typ string) string {
	if c := child(n, p.lang, typ); c != nil {
		return p.text(c)
	}
	return ""
}

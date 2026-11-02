package gedcom7

import (
	"git.bytestone.uk/hum3/gotreesitter"
)

// parse runs the tree-sitter parser over src. The tree is returned even
// when it holds errors; tstree.ParseError finds them.
func parse(src []byte) (*gotreesitter.Tree, *gotreesitter.Language, error) {
	lang := Gedcom7Language()
	tree, err := gotreesitter.NewParser(lang).Parse(src)
	if err != nil {
		return nil, nil, err
	}
	return tree, lang, nil
}

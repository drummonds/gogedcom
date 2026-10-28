// Package gedcom7 reads GEDCOM 7 with the specification's own grammar.
//
// grammar.abnf is FamilySearch's extracted-files/grammar.abnf, vendored at
// the commit in grammar.commit (task grammar:fetch), licensed under
// LICENSE.gedcom. It is read as it is by package abnf; nothing in it is
// transcribed into code.
package gedcom7

import (
	_ "embed"
	"sync"

	"git.bytestone.uk/hum3/gogedcom/abnf"
)

//go:embed grammar.abnf
var grammarSrc string

var (
	grammarOnce sync.Once
	grammar     *abnf.Grammar
)

// Grammar is the specification's ABNF, parsed once. Its one external
// reference is URI-reference (RFC 3986).
func Grammar() *abnf.Grammar {
	grammarOnce.Do(func() {
		g, err := abnf.Parse(grammarSrc)
		if err != nil {
			panic("gedcom7: vendored grammar: " + err.Error())
		}
		grammar = g
	})
	return grammar
}

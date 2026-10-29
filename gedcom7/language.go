package gedcom7

import (
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/gob"
	"sync"

	"git.bytestone.uk/hum3/gotreesitter"
)

// The parse tables come from the tree-sitter chain, task grammar:generate:
// grammar/gedcom7.abnf -> tree-sitter2abnf -> grammar/src/grammar.json ->
// tree-sitter generate -> grammar/src/parser.c -> ts2go -> grammar_gen.go
// and grammar_blobs/gedcom7.bin. The blob is what gotreesitter's own
// registry holds for its languages: a gzip of the gob-encoded Language.

//go:embed grammar_blobs/*.bin
var blobs embed.FS

var (
	languageOnce sync.Once
	language     *gotreesitter.Language
	languageErr  error
)

// loadEmbeddedLanguage is what the ts2go-generated Gedcom7Language calls.
func loadEmbeddedLanguage(name string) *gotreesitter.Language {
	languageOnce.Do(func() {
		data, err := blobs.ReadFile("grammar_blobs/" + name)
		if err != nil {
			languageErr = err
			return
		}
		gz, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			languageErr = err
			return
		}
		defer gz.Close()
		var lang gotreesitter.Language
		if err := gob.NewDecoder(gz).Decode(&lang); err != nil {
			languageErr = err
			return
		}
		language = &lang
	})
	if languageErr != nil {
		panic("gedcom7: grammar blob " + name + ": " + languageErr.Error())
	}
	return language
}

# gogedcom

GEDCOM 7 and GEDCOM 5.5.1 in Go, parsed from the specification's own ABNF
rather than a hand transcription of it, with a 5.5.1 to 7 converter.

The GEDCOM 7 grammar (`grammar/gedcom7.abnf`, YABNF) is written from the
specification's `extracted-files/grammar.abnf`, vendored at a pinned commit
(`gedcom7/grammar.commit`, `task grammar:fetch`). It becomes a parser by
the tree-sitter chain, `task grammar:generate`: tree-sitter2abnf →
tree-sitter generate → ts2go → a gotreesitter blob embedded in `gedcom7`.

## Status

Stories 001 to 003: the grammar, the six GEDCOM 7 data types (`DateValue`,
`Time`, `Age`, `PersonalName`, `Latitude`, `Longitude`) typed, a file read
as a tree of structures, and the registry validating it: which structures
go where, cardinality, enumerations. See [ROADMAP.md](ROADMAP.md).

## Vocabulary

GEDCOM 7's own words, used in code, tests and docs.

| Term | Meaning |
|---|---|
| Line | One line of the file: level, optional cross-reference identifier, tag, optional payload |
| Structure | A line with its substructures (the lines one level deeper that follow it) |
| Record | A level-0 structure other than the header and trailer |
| Tag | The structure type's name on the line (`BIRT`); extension tags start with `_` |
| Payload | The line's value: text, a pointer, or a typed data type |
| Pointer | A payload that refers to a record by its cross-reference identifier, or `@VOID@` |
| Data type | A payload grammar: `DateValue`, `Time`, `Age`, `PersonalName`, … each a rule in the ABNF |
| Grammar | `grammar/gedcom7.abnf`: the rules the parser is generated from; a line is typed by its tag |
| Registry | The specification's `extracted-files/tags`, one YAML per URI-identified term, vendored in `gedcom7/registry/` |
| Structure type | A registry term: a tag's meaning under a given superstructure (`DATE` is `DATE`, `DATE-exact` or `HEAD-DATE`), its payload type and substructures |
| Cardinality | `{0:1}`, `{0:M}`, `{1:1}`, `{1:M}`: whether a substructure is required and whether it may repeat |
| Violation | A structure that breaks the registry, reported with its line by `Validate` |

## Packages

| Package | Holds |
|---|---|
| `gedcom7` | The generated parser, typed data types, the reader, the registry and its validator; the vendored spec ABNF as reference |
| `grammar/` | The YABNF grammar, its generated `src/`, and the tree-sitter corpus tests |

## Toolchain

`task grammar:generate` needs `tree-sitter2abnf` (Go), `npx tree-sitter-cli`
(node) and a C compiler for `task grammar:test`. `task registry:fetch`
refetches the registry. Consumers need none of them: the blob and the
registry are in the module.

## Licence

MIT. The vendored GEDCOM 7 grammar is Apache 2.0, FamilySearch
(`gedcom7/LICENSE.gedcom`).

## Links

| | |
|---|---|
| Documentation | https://gogedcom.docs.bytestone.uk |
| Source | https://git.bytestone.uk/hum3/gogedcom |
| Mirror | https://github.com/drummonds/gogedcom |
| GEDCOM 7 specification | https://gedcom.io/specifications/FamilySearchGEDCOMv7.html |

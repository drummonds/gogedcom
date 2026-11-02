# gogedcom

GEDCOM 7 and GEDCOM 5.5.1 in Go, parsed from the specification's own ABNF
rather than a hand transcription of it, with a 5.5.1 to 7 converter.

The GEDCOM 7 grammar (`grammar/gedcom7/gedcom7.abnf`, YABNF) is written
from the specification's `extracted-files/grammar.abnf`, vendored at a
pinned commit (`gedcom7/grammar.commit`, `task grammar:fetch`). The GEDCOM
5.5.1 grammar (`grammar/gedcom551/gedcom551.abnf`) is written from the
5.5.1 specification, which has no machine-readable grammar. Each becomes a
parser by the tree-sitter chain, `task grammar:generate GRAMMAR=…`:
tree-sitter2abnf → tree-sitter generate → ts2go → a gotreesitter blob
embedded in its package.

## Status

GEDCOM 7 is read as a tree of structures, validated against the
specification's registry, and written back to a file that reads to the
same tree; the six data types are typed. GEDCOM 5.5.1 is read into the
same tree, every payload kept as the file has it, with `DATE` values typed
on demand; nothing converts or writes 5.5.1 yet.

The finish line for 5.5.1 is one file: the author's own tree exported from
Geni.com as GEDCOM 5.5.1 (60,126 lines; 1,976 `INDI`, 722 `FAM`; the
extension tags `_MARNM` and `_PRIM`; `ADDR` placed where 5.5.1 does not
allow it). 5.5.1 is done when that file goes 5.5.1 → 7 → 5.5.1 and comes
back equal to the original: the same trees after `CONC` and `CONT` are
folded, the header's provenance aside. The file names living people and
is not in the repo. See [ROADMAP.md](ROADMAP.md) and
[story 006](docs/stories/006-readme-status-and-the-roadmap-the-geni-round-trip.md)
for the comparison rules.

## Vocabulary

GEDCOM 7's own words, used in code, tests and docs; the 5.5.1 differences
at the end.

| Term | Meaning |
|---|---|
| Line | One line of the file: level, optional cross-reference identifier, tag, optional payload |
| Structure | A line with its substructures (the lines one level deeper that follow it) |
| Record | A level-0 structure other than the header and trailer |
| Tag | The structure type's name on the line (`BIRT`); extension tags start with `_` |
| Payload | The line's value: text, a pointer, or a typed data type |
| Pointer | A payload that refers to a record by its cross-reference identifier, or `@VOID@` |
| Data type | A payload grammar: `DateValue`, `Time`, `Age`, `PersonalName`, … each a rule in the ABNF |
| Grammar | `grammar/gedcom7/gedcom7.abnf`: the rules the parser is generated from; a line is typed by its tag |
| Registry | The specification's `extracted-files/tags`, one YAML per URI-identified term, vendored in `gedcom7/registry/` |
| Structure type | A registry term: a tag's meaning under a given superstructure (`DATE` is `DATE`, `DATE-exact` or `HEAD-DATE`), its payload type and substructures |
| Cardinality | `{0:1}`, `{0:M}`, `{1:1}`, `{1:M}`: whether a substructure is required and whether it may repeat |
| Violation | A structure that breaks the registry, reported with its line by `Validate` |
| Line break in a payload | A `\n` in `Payload`; in the file it is a `CONT` line one level deeper, folded by `Read`, split by `Write` |
| At-sign escape | A text payload whose value begins with `@` is written `@@…`; `Read` strips it, `Write` adds it; a pointer is never escaped |
| Round trip | `Read(Write(doc))` equals `doc` in every field but `Line`; for 5.5.1, the Geni export back from 7 equal to the original (story 006) |
| `CONC` (5.5.1) | A continuation line whose payload joins its parent's without a separator; folded by `gedcom551.Read` like `CONT` |
| Typed value (5.5.1) | A data type parsed on its own from its tag and payload, `DATE 11 MAR 1961`: the 5.5.1 grammar's second entry, since a 5.5.1 line types nothing and the file is the contract (story 007) |
| Calendar escape (5.5.1) | `@#DJULIAN@` before a date, where 7 writes `JULIAN`; `Date.Calendar` holds the name |

## Packages

| Package | Holds |
|---|---|
| `gedcom` | The tree both readers return: `Structure`, `Document` |
| `gedcom7` | The generated parser, typed data types, the reader and writer, the registry and its validator; the vendored spec ABNF as reference; the specification's test files under `testdata/` |
| `gedcom551` | The generated 5.5.1 parser, the reader, `ParseDate` |
| `grammar/gedcom7/`, `grammar/gedcom551/` | Each grammar in YABNF, its generated `src/`, and its tree-sitter corpus tests |

## Toolchain

`task grammar:generate GRAMMAR=gedcom7|gedcom551` needs `tree-sitter2abnf`
(Go) and `npx tree-sitter-cli` (node); `task grammar:test GRAMMAR=…` needs a
C compiler. `task registry:fetch`
refetches the registry, `task testdata:fetch` the specification's test
files (GEDCOM.io `testfiles/gedcom70`, pinned in the Taskfile). Consumers need none of them: the blob and the
registry are in the module.

## Licence

MIT. The vendored GEDCOM 7 grammar, registry and test files are Apache
2.0, FamilySearch (`gedcom7/LICENSE.gedcom`).

## Links

| | |
|---|---|
| Documentation | https://gogedcom.docs.bytestone.uk |
| Source | https://git.bytestone.uk/hum3/gogedcom |
| Mirror | https://github.com/drummonds/gogedcom |
| GEDCOM 7 specification | https://gedcom.io/specifications/FamilySearchGEDCOMv7.html |
| GEDCOM 5.5.1 specification | https://gedcom.io/specifications/ged551.pdf |

# Roadmap

Milestones for gogedcom. Each item is an open issue labelled `roadmap` on
git.bytestone.uk/hum3/gogedcom; `tp roadmap add` files one, `tp release`
retires those whose stories have shipped.

## GEDCOM 7 from the specification

The current standard, read from its own machine-readable files.
- #1 The specification's ABNF as a Go parser: the data types (date, time, age, name, latitude, longitude) typed from the vendored grammar, and the file as a tree of structures
- #2 The structure schema from the specification's YAML registry: which structures go where, cardinality, enumerations; a validator that reports every violation with its line
- #3 A writer: GEDCOM 7 out of the tree, and GEDZIP with the media; reading what was written gives the same tree
- #4 Extensions: our own tags registered through SCHMA with URIs, defined in a YAML of the registry's shape, validated like the standard ones

## GEDCOM 5.5.1 and the way to 7

Where the corpus is: readable natively, converted at the door.
- #5 A 5.5.1 ABNF in house style, written from the 5.5.1 specification, and the reader on the same structure as the 7 reader
- #6 The 5.5.1 to 7 converter: the mapping as a decision table in the repo; every 5.5.1 structure accounted for, the original wording kept in PHRASE
- #7 Completeness measured on real exports: convert, write, read back, diff; the loss list per source in the README

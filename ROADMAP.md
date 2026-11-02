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

## GEDCOM 5.5.1 there and back: the Geni export round-tripped

Where the corpus is. Done when the author's Geni.com 5.5.1 export goes
5.5.1 → 7 → 5.5.1 and comes back equal to the original (story 006 has the
comparison rules).
- #5 A 5.5.1 ABNF in house style, written from the 5.5.1 specification, and the reader on the same structure as the 7 reader
- #6 The 5.5.1 to 7 converter: the mapping as a decision table in the repo; every 5.5.1 structure accounted for, the original wording kept in PHRASE
- #22 The 7 to 5.5.1 converter: the inverse of #6, PHRASE back to the original wording, extension tags and off-schema structures carried through unchanged; a 5.5.1 writer on the same contract as the 7 writer
- #7 The round trip: the Geni export read as 5.5.1, converted to 7, written, read, converted back and written as 5.5.1; the trees equal after `CONC`/`CONT` folding, header provenance aside; any remaining diff listed in the README

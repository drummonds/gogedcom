# gogedcom

GEDCOM 7 and 5.5.1 in Go, parsed from the specification's own ABNF, with
converters between the two.

- [Story 001 — The specification ABNF as a Go parser](stories/001-the-specification-abnf-as-a-go-parser.html)
- [Story 002 — Time, age, name, latitude and longitude as typed values](stories/002-time-age-name-latitude-and-longitude-as-typed.html)
- [Story 003 — The schema from the registry YAML](stories/003-the-schema-from-the-registry-yaml-structure-types.html)
- [Story 004 — The grammar as LR(1)](stories/004-the-grammar-as-lr1.html)
- [Story 005 — A writer: GEDCOM 7 out of the tree, read back to the same tree](stories/005-a-writer-gedcom-7-out-of-the-tree-read-back-to-the.html)
- [Story 006 — README status and the roadmap: the Geni round trip as the goal](stories/006-readme-status-and-the-roadmap-the-geni-round-trip.html)
- [Story 007 — The 5.5.1 grammar from its specification, and the reader on the 7 reader's tree](stories/007-the-5-5-1-grammar-from-its-specification-and-the.html)
- [Story 008 — The 5.5.1 to 7 converter: the mapping as a decision table](stories/008-the-5-5-1-to-7-converter-the-mapping-as-a-decision.html)
- [Story 009 — The 7 to 5.5.1 converter and writer: the inverse of 008](stories/009-the-7-to-5-5-1-converter-and-writer-the-inverse-of.html)
- [Story 010 — The round trip at file level: the Geni export through both writers](stories/010-the-round-trip-at-file-level-the-geni-export.html)
- [Converting 5.5.1 to 7: the mapping](convert-551-to-7.html) — the decision table and every 5.5.1 tag's rule
- [Converting 7 to 5.5.1: the mapping](convert-7-to-551.html) — the inverse: the decision table and every 7 tag's rule
- [Round tripping: where it loses](round-trip.html) — research for the converter stories
- Source: [git.bytestone.uk/hum3/gogedcom](https://git.bytestone.uk/hum3/gogedcom), mirror [github.com/drummonds/gogedcom](https://github.com/drummonds/gogedcom)

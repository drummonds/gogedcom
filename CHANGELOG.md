# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.11.0] - 2026-11-05

- The round trip at file level (roadmap #7): the Geni export read as 5.5.1, converted to 7, written, read, converted back, written as 5.5.1 and read comes back equal to the original by story 006's rules, the header's provenance aside; the comparison lists every record that differs, and none does

### Added

- The round trip at file level (roadmap #7): the Geni export read as 5.5.1, converted to 7, written, read, converted back, written as 5.5.1 and read comes back equal to the original by story 006's rules, the header's provenance aside; the comparison lists every record that differs, and none does (#32)

## [0.10.0] - 2026-11-05

- `convert.To551`: the 7 tree as a GEDCOM 5.5.1 tree by the decision table in `docs/convert-7-to-551.md`, 008's rules read backwards: the original wording back from `PHRASE` where 008 wrote it, `SNOTE` to `NOTE`, `EXID` to `RFN`/`AFN`/`RIN`, `TRAN` to `FONE`/`ROMN`, `ROLE OTHER` to `RELA`, enumerations in 5.5.1's case, media types to words, `SOUR @VOID@` to an inline citation, relocated `_ADDR` back to `ADDR`, an `OBJE` record pointed to once back inline, and 7-native content 5.5.1 has no home for (`CREA`, `UID`, a `PHRASE` under `PEDI`) carried as an extension tag. `gedcom551.Write` writes the tree as 5.5.1 on `gedcom7.Write`'s contract. The Geni export goes 5.5.1 → 7 → 5.5.1 in memory and comes back equal in every field but `Line`; the table is held to `convert/rules551.go` and to the registry's tag list by a test

### Added

- `convert.To551`: the 7 tree as a GEDCOM 5.5.1 tree by the decision table in `docs/convert-7-to-551.md`, 008's rules read backwards: the original wording back from `PHRASE` where 008 wrote it, `SNOTE` to `NOTE`, `EXID` to `RFN`/`AFN`/`RIN`, `TRAN` to `FONE`/`ROMN`, `ROLE OTHER` to `RELA`, enumerations in 5.5.1's case, media types to words, `SOUR @VOID@` to an inline citation, relocated `_ADDR` back to `ADDR`, an `OBJE` record pointed to once back inline, and 7-native content 5.5.1 has no home for (`CREA`, `UID`, a `PHRASE` under `PEDI`) carried as an extension tag. `gedcom551.Write` writes the tree as 5.5.1 on `gedcom7.Write`'s contract. The Geni export goes 5.5.1 → 7 → 5.5.1 in memory and comes back equal in every field but `Line`; the table is held to `convert/rules551.go` and to the registry's tag list by a test (#29)

## [0.9.0] - 2026-11-05

- Story 009 proposed, the 7 to 5.5.1 converter and the 5.5.1 writer: 008's rules read backwards with the wording taken from `PHRASE` when 008 would have written it, an `OBJE` record pointed to once inlined, and 7-native content 5.5.1 has no home for (`CREA`, `UID`, a `PHRASE` under `PEDI`) carried as an extension tag, the one rule for the whole list; `gedcom551.Write` on `gedcom7.Write`'s contract

### Added

- Story 009 proposed, the 7 to 5.5.1 converter and the 5.5.1 writer: 008's rules read backwards with the wording taken from `PHRASE` when 008 would have written it, an `OBJE` record pointed to once inlined, and 7-native content 5.5.1 has no home for (`CREA`, `UID`, a `PHRASE` under `PEDI`) carried as an extension tag, the one rule for the whole list; `gedcom551.Write` on `gedcom7.Write`'s contract (#29)

## [0.8.0] - 2026-11-04

- `convert.To7`: the 5.5.1 tree as a GEDCOM 7 tree by the decision table in `docs/convert-551-to-7.md`, which names every 5.5.1 tag's rule and is held to `convert/rules.go` by a test. 7's grammar decides what must change (a `DATE`, `AGE` or `NAME` it rejects is re-spelt with the original in `PHRASE` or `_PHRASE`), the registry and the migration guide what may change losslessly (`NOTE` records to `SNOTE`, `RFN`/`AFN`/`RIN` to `EXID`, `FONE`/`ROMN` to `TRAN`, `RELA` to `ROLE`, enumerations upper-cased, media types, inline `OBJE` and `SOUR` to records, an empty `ADDR` composed from its pieces), and a standard tag out of its 7 place is relocated as `_ADDR` declared in `SCHMA`. The Geni export converts, validates with no violation and reads back from the written 7 file equal. `gedcom7.Registry.Types` lists a tag's terms (#26)

### Added

- `convert.To7`: the 5.5.1 tree as a GEDCOM 7 tree by the decision table in `docs/convert-551-to-7.md`, which names every 5.5.1 tag's rule and is held to `convert/rules.go` by a test. 7's grammar decides what must change (a `DATE`, `AGE` or `NAME` it rejects is re-spelt with the original in `PHRASE` or `_PHRASE`), the registry and the migration guide what may change losslessly (`NOTE` records to `SNOTE`, `RFN`/`AFN`/`RIN` to `EXID`, `FONE`/`ROMN` to `TRAN`, `RELA` to `ROLE`, enumerations upper-cased, media types, inline `OBJE` and `SOUR` to records, an empty `ADDR` composed from its pieces), and a standard tag out of its 7 place is relocated as `_ADDR` declared in `SCHMA`. The Geni export converts, validates with no violation and reads back from the written 7 file equal. `gedcom7.Registry.Types` lists a tag's terms (#26) (#26)

## [0.7.0] - 2026-11-04

- Story 008 proposed, the 5.5.1 to 7 converter: the mapping rule (7's grammar decides what must change, the registry what may change losslessly, everything else passes through), the decision table to come in `docs/convert-551-to-7.md`. `docs/round-trip.md` is the research behind it: every point on the 5.5.1 → 7 → 5.5.1 trip where information is lost, comes back changed or stops the trip, with a verdict per row and what the code does today; `task docs:build` renders `docs/*.md`

### Added

- Story 008 proposed, the 5.5.1 to 7 converter: the mapping rule (7's grammar decides what must change, the registry what may change losslessly, everything else passes through), the decision table to come in `docs/convert-551-to-7.md`. `docs/round-trip.md` is the research behind it: every point on the 5.5.1 → 7 → 5.5.1 trip where information is lost, comes back changed or stops the trip, with a verdict per row and what the code does today; `task docs:build` renders `docs/*.md` (#26)

### Roadmap

- #1 The specification's ABNF as a Go parser: the data types (date, time, age, name, latitude, longitude) typed from the vendored grammar, and the file as a tree of structures
- #2 The structure schema from the specification's YAML registry: which structures go where, cardinality, enumerations; a validator that reports every violation with its line
- #5 A 5.5.1 ABNF in house style, written from the 5.5.1 specification, and the reader on the same structure as the 7 reader

## [0.6.0] - 2026-11-02

- `gedcom551`: GEDCOM 5.5.1 read into the same tree as GEDCOM 7. The grammar (`grammar/gedcom551/gedcom551.abnf`) is written from the 5.5.1 specification's line syntax and primitive data types; a line types nothing, so a payload its data type rejects is still read, and `CONC`/`CONT` are folded. `ParseDate` reads a `DATE_VALUE` with the calendar escape, `B.C.`, the dual year, `INT … (phrase)` and `(phrase)`. The tree types `Structure` and `Document` move to package `gedcom` (`gedcom7` keeps them as aliases); the grammars live under `grammar/gedcom7/` and `grammar/gedcom551/`, `task grammar:generate GRAMMAR=…`

### Added

- `gedcom551`: GEDCOM 5.5.1 read into the same tree as GEDCOM 7. The grammar (`grammar/gedcom551/gedcom551.abnf`) is written from the 5.5.1 specification's line syntax and primitive data types; a line types nothing, so a payload its data type rejects is still read, and `CONC`/`CONT` are folded. `ParseDate` reads a `DATE_VALUE` with the calendar escape, `B.C.`, the dual year, `INT … (phrase)` and `(phrase)`. The tree types `Structure` and `Document` move to package `gedcom` (`gedcom7` keeps them as aliases); the grammars live under `grammar/gedcom7/` and `grammar/gedcom551/`, `task grammar:generate GRAMMAR=…` (#24)

## [0.5.0] - 2026-11-02

- README status says what is built and what finishes 5.5.1: the Geni export round-tripped 5.5.1 → 7 → 5.5.1 and equal to the original. The roadmap epic and #7 say the same; a new item for the writer back to 5.5.1 (story 006)

### Changed

- README status says what is built and what finishes 5.5.1: the Geni export round-tripped 5.5.1 → 7 → 5.5.1 and equal to the original. The roadmap epic and #7 say the same; a new item for the writer back to 5.5.1 (story 006) (#21)

## [0.4.0] - 2026-11-02

- `gedcom7.Write`: the tree as a GEDCOM 7 file (LF line ends, `CONT` for line breaks, the at-sign escape on text payloads), refusing a payload with a carriage return or a tag or cross-reference identifier the grammar rejects; reading what it wrote gives the same tree. `Read` strips the at-sign escape, so `Payload` holds the value (`@x`, not `@@x`). The specification's test files vendored under `gedcom7/testdata/gedcom70` (`task testdata:fetch`)

### Added

- `gedcom7.Write`: the tree as a GEDCOM 7 file (LF line ends, `CONT` for line breaks, the at-sign escape on text payloads), refusing a payload with a carriage return or a tag or cross-reference identifier the grammar rejects; reading what it wrote gives the same tree. `Read` strips the at-sign escape, so `Payload` holds the value (`@x`, not `@@x`). The specification's test files vendored under `gedcom7/testdata/gedcom70` (`task testdata:fetch`) (#18)

## [0.3.0] - 2026-11-02

### Changed

- `grammar/gedcom7.abnf` is LR(1) and declares no conflicts, so `tree-sitter generate` rejects any ambiguity. `day` and `year` are aliases of `integer` (the tree has `(day)` where it had `(day (integer))`), the delimiter before `TO`, `AND` and the epoch is part of that token, and a lone extension tag in a date is a calendar. `ParseDate` results are unchanged (#14)

## [0.2.0] - 2026-11-02

### Added

- `gedcom7/registry/`: the specification's registry (`extracted-files/tags`, one YAML per term) vendored at the pinned commit by `task registry:fetch` and embedded; `Standard()` reads it (`Structure`, `Substructure`, `Enumeration`) and `(*Registry).Validate` reports every violation with its line: a tag the registry does not place, a singleton repeated, a required substructure missing, an enumeration payload not in its set. `Structure.Line` is the file line a structure starts on (#10)

## [0.1.0] - 2026-10-29

### Added

- `grammar/gedcom7.abnf`: GEDCOM 7 in YABNF, from the specification's
  `grammar.abnf` vendored at a pinned commit (`task grammar:fetch`);
  `task grammar:generate` runs the tree-sitter chain to a gotreesitter blob
  embedded in `gedcom7`; `task grammar:test` runs the corpus.
- `gedcom7`: `ParseDate` to a typed `DateValue`; `Read` of a file as a tree
  of structures, the first line the grammar rejects being the error.
- `gedcom7`: `ParseTime`, `ParseAge`, `ParsePersonalName`, `ParseLatitude`
  and `ParseLongitude` to the typed `Time`, `Age`, `PersonalName`,
  `Latitude` and `Longitude`, each writing itself back with `String()`.
  A character the grammar cannot place (a second space after the level, a
  tab in a name) is an error, as the C parser reports it.

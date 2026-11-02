# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

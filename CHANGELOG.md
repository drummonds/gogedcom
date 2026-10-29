# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

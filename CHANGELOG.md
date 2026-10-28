# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `abnf`: an RFC 5234 + RFC 7405 reader and matcher (`|` accepted for
  alternation); `Match`, `MatchWhere` with a constraint, `Undefined`.
- `gedcom7`: the specification's `grammar.abnf` vendored at a pinned commit
  (`task grammar:fetch`); `ParseDate` to a typed `DateValue`; `Read` of a
  file as a tree of structures, every line checked against the `Line` rule.

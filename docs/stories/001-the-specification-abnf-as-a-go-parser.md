# 001 — The specification ABNF as a Go parser

**Status:** in progress
**Roadmap:** #1 — the specification's ABNF as a Go parser

## Why

GEDCOM 7's specification ships its grammar as a machine-readable ABNF file,
`extracted-files/grammar.abnf`, generated from the spec text. Every GEDCOM
library we looked at transcribes that grammar into code by hand, so a spec
bump is a code change and a transcription slip is a parser bug. The 7.0.17
release showed the other side of the same coin: a typo in the published
ABNF (`Longitude` with `N`/`S` for `E`/`W`) broke every tool that consumed
the file directly, and the fix reached the extracted file only on `main`,
after the 7.0.18 tag.

So the grammar should be consumed verbatim, pinned to a commit, with a
test suite that would have caught the typo. That needs an ABNF engine in Go:
one that reads RFC 5234 (with RFC 7405 `%s` strings and the house-style `|`
alternation) and matches a string against a named rule, returning the parse
as a tree of rule nodes. The same engine then carries the hand-written 5.5.1
grammar (roadmap #5) and whatever extension grammars come later (#4).

## What changes

- `abnf`: `Parse(src)` reads a grammar; `g.Match(rule, input)` returns the
  derivation as a `Node` tree (rule name, span, children), or an error naming
  the furthest position reached; `g.MatchWhere` adds a constraint that
  rejects a derivation. `g.Undefined()` lists rule names used but not
  defined, so a vendored grammar's external references are explicit.
- `gedcom7/grammar.abnf`, `LICENSE.gedcom`, `grammar.commit`: the upstream
  grammar vendored at a pinned commit by `task grammar:fetch` (Apache 2.0,
  FamilySearch).
- `gedcom7.Grammar()`: the vendored grammar, parsed once.
- `gedcom7.ParseDate(s)`: a `DateValue` from the `DateValue` rule, typed as
  the spec names it: calendar, day, month, year, epoch, and the kind of
  value (date, period, range, approximation).
- `gedcom7.Read(r)`: a file as a tree of structures, each line checked
  against the `Line` rule; `CONT` folded into its parent's payload.

## Acceptance

The engine, on RFC 5234's own constructs:

| Grammar | Input | Outcome |
|---|---|---|
| `a = "ab"` | `AB` | match (bare strings are case-insensitive) |
| `a = %s"ab"` | `AB` | no match |
| `a = %x41-43` | `B` | match |
| `a = %x0D.0A` | CR LF | match |
| `a = 2*3DIGIT` | `12`, `123` | match; `1`, `1234` fail |
| `a = *DIGIT "x"` | `12x`, `x` | match |
| `a = b \| c`, `b = "1"`, `c = "2"` | `2` | match, node `c` under `a` |
| `a = "1" / "12"` | `12` | match (full backtracking, not ordered choice alone) |
| `a = b` `b =/ "z"` | `z` | match (incremental alternative) |
| `a = [ "x" ] "y"` | `y`, `xy` | match |
| `a = b` with `b` undefined | — | `Undefined()` lists `b`; `Match` errors when it reaches `b` |
| `a = b / c`, constraint rejects `b` | `42` | match through `c` |
| any | input left over | error at the furthest position reached |

`ParseDate`:

| Input | Kind | Dates |
|---|---|---|
| `` | none | — |
| `12 JAN 1990` | date | day 12, month `JAN`, year 1990 |
| `JAN 1990` | date | month `JAN`, year 1990 |
| `1990` | date | year 1990 |
| `JULIAN 1 JAN 1700` | date | calendar `JULIAN` |
| `JULIAN 1700` | date | calendar `JULIAN`, year 1700 (calendar wins over month when both fit) |
| `100 BCE` | date | year 100, epoch `BCE` |
| `ABT 1850`, `CAL 1850`, `EST 1850` | approximate, qualifier as given | from 1850 |
| `BET 1850 AND 1860` | range `BET` | from 1850 to 1860 |
| `AFT 1850` | range `AFT` | from 1850 |
| `BEF 1850` | range `BEF` | to 1850 |
| `FROM 1850 TO 1860` | period | from 1850 to 1860 |
| `FROM 1850`, `TO 1860` | period | the one bound |
| `12 JAN`, `1990 JAN 12`, `ABT`, `BET 1850`, `12 jan 1990` | error | — |

`Read`:

| File | Outcome |
|---|---|
| header, one `INDI` with a `NAME`, trailer | one record with one substructure |
| a `NOTE` with two `CONT` lines, the second empty (`1 CONT`, no trailing space) | one payload with two newlines, no `CONT` substructure |
| a leading BOM | accepted |
| a line that fails the `Line` rule (`1  NAME` with two spaces) | error naming the line number |
| no `TRLR` | error |

## Out of scope

- Time, Age, PersonalName, Latitude, Longitude as typed values: the next
  story, on the same pattern as `ParseDate`.
- The structure schema (which tags go where) and enumerations: roadmap #2.
  `Read` accepts any well-formed tag.
- Writing: roadmap #3.
- Performance beyond a 100,000-line file reading in seconds.

## Decisions

| Choice | Taken | Alternative | Why |
|---|---|---|---|
| How the ABNF becomes a parser | An ABNF interpreter in Go: the grammar file is read as it is | tree-sitter: YABNF → `tree-sitter2abnf` → `tree-sitter generate` → `ts2go` → gotreesitter | The upstream file is consumed verbatim, so a spec bump is a file swap and nothing is transcribed. Payloads are short strings: no need for incremental parsing or error recovery. No node or Rust toolchain in the build. `tree-sitter2abnf` reads the same file, so highlighting and railroad diagrams stay available later |
| Matching semantics | Full backtracking over alternatives in written order; the first complete derivation wins | PEG ordered choice; a GLR all-derivations parser | PEG would reject `"1" / "12"` on `12`; all derivations are not needed for a data type. Where the spec is ambiguous (`JULIAN 1990`) the written order decides, and the test pins it |
| The file's lines | Matched one at a time against the `Line` rule | A `*Line` rule over the whole file | Recursion depth is per line, not per file; the error names the line |
| Pinning | A commit on `main`, recorded in `grammar.commit` | A release tag | The extracted file lagged the 7.0.18 tag |
| What the grammar leaves to comments | `MatchWhere(rule, input, accept)`: a constraint called as each rule node completes; a false backtracks | Edit the vendored grammar to enumerate months; validate after the match and retry | `month = stdTag ; constrained by calendar` makes `ABT 1850` a month and a year; the constraint keeps the file verbatim and the derivation right. The months table in `date.go` moves to the registry's YAML in roadmap #2 |

## Done when

- [ ] tests express the acceptance table and pass
- [ ] README
- [ ] changelog entry
- [ ] issue closed with a link to the release

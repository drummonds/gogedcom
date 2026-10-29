# 001 — The specification ABNF as a Go parser (#8)

**Status:** done locally
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

- `grammar/gedcom7.abnf`: the GEDCOM 7 grammar in YABNF, written from the
  vendored `gedcom7/grammar.abnf` (the spec's `extracted-files/grammar.abnf`
  at a pinned commit, Apache 2.0, FamilySearch, refetched by
  `task grammar:fetch`). A line is typed by its tag: `DATE`, `TIME`, `AGE`,
  `NAME`, `LATI` and `LONG` carry their data types; every other tag a
  pointer or text. Three GLR conflicts are declared: day/year,
  calendar/month, and the date before a following keyword.
- `task grammar:generate`: the tree-sitter chain, as the other parsers run
  it: `tree-sitter2abnf` → `grammar/src/grammar.json` →
  `tree-sitter generate` → `grammar/src/parser.c` → `ts2go` →
  `gedcom7/grammar_gen.go` and `gedcom7/grammar_blobs/gedcom7.bin`.
  `task grammar:test` runs the corpus under `grammar/test/corpus`.
- `gedcom7`: `Gedcom7Language()` loads the blob (gzip of the gob-encoded
  `gotreesitter.Language`, as gotreesitter's own registry holds them);
  `ParseDate(s)` reads a `DateValue` typed as the spec names it: calendar,
  day, month, year, epoch, and the kind of value (date, period, range,
  approximation); `Read(r)` reads a file as a tree of structures, the
  first line the grammar rejects being the error, `CONT` folded into its
  parent's payload.

## Acceptance

The grammar, in `grammar/test/corpus` (`task grammar:test`): the dates
below as trees; a header, record, name and trailer; a pointer and `@VOID@`;
text with a doubled `@` and an empty `CONT`; time, age, latitude and
longitude; two spaces after the level as an `ERROR`.

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
| How the ABNF becomes a parser | The tree-sitter chain, as the other parsers (goluca, pta): YABNF → `tree-sitter2abnf` → `tree-sitter generate` → `ts2go` → gotreesitter | An ABNF interpreter reading the spec's file verbatim (built first, then replaced) | One toolchain across the parsers; highlighting, queries and railroad diagrams come with it. The cost is that the YABNF is a transcription: `tree-sitter2abnf` reads `%s` strings and `@pattern`, not `%x` ranges, so the spec's file is the reference the grammar is written from and checked against, not the input |
| Where the grammar lives | In this repo under `grammar/`, its blob embedded by `gedcom7` | A `tree-sitter-gedcom7` repo registered in gotreesitter's manifest | The blob format is gotreesitter's public `Language` type, so no registry change is needed; a grammar repo can be split out when editor support is wanted |
| How a payload is typed | By its tag, in the grammar: `DATE` carries `date_value`, `NAME` `personal_name`, and so on | One grammar per data type; a sentinel prefix selecting a start rule | A tree-sitter grammar has one start rule. The tag is how the spec types payloads, and the whole file parses in one pass. `ParseDate` parses the line `0 DATE <s>` |
| Months | Listed in the grammar per calendar | `month = stdTag ; constrained by calendar` as the spec writes it | Without the list `ABT 1850` is the month `ABT` of 1850. The list is what the registry's YAML says; roadmap #2 generates it from there |
| Day range | Checked in Go after the parse | A `[0-9]{1,2}` token beside `integer` | Two tokens matching the same text is a lexer ambiguity; the grammar keeps one `integer` and the reader rejects a day over 31 |
| The file's lines | A flat sequence in the grammar; nesting by level in Go | Level nesting in the grammar, with an external scanner | Context-sensitive nesting needs a scanner; the reader already needs the levels for `CONT` and `TRLR` |
| Pinning the reference | A commit on `main`, recorded in `grammar.commit` | A release tag | The extracted file lagged the 7.0.18 tag |
| EOL at end of input | Appended by `ParseDate` and `Read` when absent | Report upstream and wait | gotreesitter v0.7.0 returns a wrong root when input ends right after a GLR fork; the C parser does not. To report |

## Done when

- [x] tests express the acceptance table and pass
- [x] README
- [x] changelog entry
- [ ] issue closed with a link to the release

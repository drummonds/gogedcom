# 007 — The 5.5.1 grammar from its specification, and the reader on the 7 reader's tree (#24)

**Status:** proposed
**Roadmap:** #5 — a 5.5.1 ABNF in house style and the reader on the same structure as the 7 reader

## Why

The round trip (story 006, roadmap #7) starts by reading the Geni export
as GEDCOM 5.5.1, and nothing reads 5.5.1 yet. The 5.5.1 specification has
no machine-readable grammar: chapter 1 gives the line syntax and chapter 2
the primitive data types as prose BNF in a PDF. This story writes that
grammar in house style, generates a parser from it by the same chain as
the 7 grammar, and puts a reader on it that returns the tree the 7 reader
returns, so the converter (#6) works on one tree type.

The Geni export decides two things the specification does not. First, it
breaks the 5.5.1 data-type grammars in places (`2 DATE 5  1801` with two
spaces; `1 NAME Louise/Louisa` with one slash; a name with three), and the
file is the contract, so a payload that fails its data type cannot be a
read error. The 5.5.1 reader therefore types nothing: every payload is a
pointer or text, and a data type is parsed on demand. Second, it breaks
the line grammar too (140 lines with a tab; `ADDR humphrey.drummond@gmail.com`
with a lone `@`; a `CONT` payload beginning `@R1203558009@  Ancestry`),
so the line grammar is the specification's with named relaxations, each
one a corpus test.

## The 5.5.1 file, as the grammar reads it

| 5.5.1 line component | Specification (chapter 1) | Grammar | Note |
|---|---|---|---|
| `level` | one or two digits, no leading zero | `0\|[1-9][0-9]?` | 7 allows any number of digits |
| `xref_ID` | `@` alphanum pointer_string `@`; pointer_char is any non-`@` | `@[A-Za-z0-9_][^@\r\n]*@` | spaces, `!` and `:` allowed, as the spec says; 7 allows `[A-Z0-9_]` only |
| `tag` | one or more alphanum (letters, digits, `_`) | `[A-Za-z0-9_]+` | lower case allowed by the spec |
| `pointer` | `@` … `@` alone as the value | as `xref_ID` | no `@VOID@` in 5.5.1 |
| `line_item` (text) | any_char: printable, space, `#`, `@@`; escapes `@#…@` | any character but CR and LF | **relaxed**: tab, control characters and a lone `@` accepted; `@X@` followed by more text is text |
| `terminator` | CR, LF, CR LF, LF CR | the four | 7 has three |
| leading white space, extra terminators | "should be ignored by the reading system" | the reader strips leading spaces and tabs and skips empty lines | 7 rejects both |
| line length ≤ 255, tag ≤ 31, xref ≤ 22 | stated as rules | not enforced | the Geni file has a 279-character line |

The data types, from chapter 2, as typed values with the tag that carries
them: `DATE date_value`, `TIME time_value`, `AGE age_at_event`,
`NAME name_personal`, `LATI place_latitude`, `LONG place_longitude`.
A `DATE_VALUE` differs from 7's `DateValue` in five places: a calendar
escape `@#DJULIAN@` (and `HEBREW`, `FRENCH R`, `GREGORIAN`, `ROMAN`,
`UNKNOWN`) where 7 has a bare word; `B.C.` where 7 has `BCE`; a dual year
`1699/00`; `INT <date> (<phrase>)`; and a bare `(<phrase>)`.

## What changes

- `grammar/` becomes a two-grammar tree-sitter package:
  `grammar/gedcom7/` (the existing grammar, `src/` and corpus moved) and
  `grammar/gedcom551/` (`gedcom551.abnf`, `src/`, `test/corpus`).
  `task grammar:generate GRAMMAR=gedcom551` and
  `task grammar:test GRAMMAR=gedcom551` run the chain and the corpus for
  one grammar; both default to `gedcom7`.
- `grammar/gedcom551/gedcom551.abnf`: the grammar above. Its start rule is
  `document = line / value`: a file line, or a typed value `DATE …`,
  `TIME …` and so on, which is how a data type is parsed on its own. The
  line rule types nothing.
- `gedcom`: a new package holding `Structure`, `Document`, `Find` and
  `All`, the tree both readers return. `gedcom7` keeps the names as type
  aliases, so nothing that imports it changes.
- `gedcom551`: `Gedcom551Language()`, the embedded blob; `Read(r)` on the
  7 reader's contract (`*gedcom.Document`, the first line the grammar
  rejects being the error), with `CONC` folded into its parent's payload
  as written and `CONT` folded with a line break, payloads kept verbatim;
  `ParseDate(s)` reading a `DATE_VALUE` into a `DateValue` with the 5.5.1
  fields (`Calendar` from the escape, `Epoch` `B.C.`, `Dual` year
  modifier, `Phrase`, the kinds `Interpreted` and `Phrase` beside 7's).
- A test that reads the Geni export when `GOGEDCOM_GENI` names it, and
  skips otherwise.
- README: status, vocabulary (5.5.1 terms), packages, toolchain.

## Acceptance

The grammar, in `grammar/gedcom551/test/corpus` (`task grammar:test
GRAMMAR=gedcom551`): each row of the component table above as a line that
parses or fails; each `DATE_VALUE` form below; a time, an age in each
form including `CHILD`, a name in each of the five forms, a latitude and a
longitude.

`Read`:

| File | Outcome |
|---|---|
| header, one `INDI` with a `NAME`, trailer, CR LF line ends, a BOM | one record with one substructure |
| a `NOTE` with a `CONC` whose payload begins with a space, then a `CONT` | one payload: the two joined without a separator, then a line break, then the third |
| a payload ending in a space; a payload with a tab; `ADDR a@b.c`; `CONT @R1@  text` | kept verbatim as text |
| `@@` in a payload | kept as `@@` |
| `1 FAMC @F1@` | a pointer |
| a level-10 line under a level-9 line | accepted |
| a line indented with spaces; an empty line between two lines | the indentation and the empty line ignored |
| LF CR line ends | accepted |
| `1  NAME` (two spaces), `01 NAME` (leading zero), `1 NAME ` (empty value after the delimiter) | error naming the line |
| level 2 directly under level 0 | error naming the line |
| no `TRLR`; a record before `HEAD` | error |
| the Geni export (`GOGEDCOM_GENI`) | reads; 1,976 `INDI`, 722 `FAM`, 164 `NOTE`, 187 `SUBM`; no `CONC` or `CONT` structure remains; every file line is accounted for by one structure or one fold |

`ParseDate`:

| Input | Kind | Fields |
|---|---|---|
| `` | none | — |
| `11 MAR 1961` | date | day 11, month `MAR`, year 1961 |
| `MAR 1961`, `1961` | date | the parts given |
| `@#DJULIAN@ 1 JAN 1700` | date | calendar `JULIAN` |
| `@#DFRENCH R@ 1 VEND 8` | date | calendar `FRENCH R`, month `VEND` |
| `@#DHEBREW@ 1 TSH 5700` | date | calendar `HEBREW`, month `TSH` |
| `100 B.C.` | date | year 100, epoch `B.C.` |
| `15 APR 1699/00` | date | year 1699, dual `00` |
| `ABT 1850`, `CAL 1850`, `EST 1850` | approximate | qualifier as given |
| `BET 1850 AND 1860`, `AFT 1850`, `BEF 1850` | range | as in 7 |
| `FROM 1850 TO 1860`, `FROM 1850`, `TO 1860` | period | as in 7 |
| `INT 1850 (about the war)` | interpreted | from 1850, phrase `about the war` |
| `(spring 1850)` | phrase | phrase `spring 1850` |
| `5  1801`, `12 JAN`, `Louise`, `BCE 100`, `1850 B.C.E.` | error | — |

## Out of scope

- Typed Go values for time, age, name, latitude and longitude: the grammar
  has them and the corpus tests them; the accessors follow as story 002
  followed 001.
- The 5.5.1 structure schema (which tags go where, `LINEAGE-LINKED`),
  enumerations, and a validator. The reader accepts any well-formed tag.
- Decoding `@@` to `@`, or any other interpretation of a payload: the
  payload is the file's text. The converter (#6) interprets.
- Character sets other than UTF-8 (`CHAR ANSEL`, `ANSI`): a transcoding
  step before the reader, if ever needed.
- Writing 5.5.1: roadmap #22.
- The 7 grammar's own relaxations: it stays as it is.

## Decisions

| Choice | Taken | Alternative | Why |
|---|---|---|---|
| Typing payloads | The line grammar types nothing; data types are parsed on demand from a typed `value` entry (`DATE 11 MAR 1961`) in the same grammar | Type lines by tag as the 7 grammar does | The Geni file has a `DATE` and two `NAME` payloads their data types reject. 7 is strict because this library writes it; 5.5.1 is read from elsewhere and the file is the contract. One grammar keeps one chain; a second alternative of the start rule is cheaper than a second grammar |
| Where the typed entry lives | A second alternative of `document`: `line / value`, the value introduced by its tag | A sentinel character; a separate grammar per data type | The tag is how the spec types a payload, so `DATE …` reads as GEDCOM; keyword promotion is context-aware in tree-sitter and gotreesitter, so `DATE` in a line stays a tag |
| Text payloads | Any character but CR and LF; `@X@` followed by text is text | The spec's `any_char` with `@@` and escapes | The Geni file has tabs, lone `@` and pointer-shaped prefixes in text; rejecting them fails the round trip at line one. The relaxations are named in the grammar and tested |
| `@@` | Kept verbatim | Decoded to `@` on read, re-encoded on write | Geni writes a lone `@` for `@`, so decoding would make the writer (#22) turn `a@b` into `a@@b`: a round-trip diff. Interpretation is the converter's |
| Leading white space, empty lines | Ignored, as the spec instructs readers | Reject, as the 7 reader does | The spec says to ignore them |
| `1 NAME ` with nothing after the delimiter | Rejected | Accepted as an empty payload | The spec's `line_item` is non-empty; the Geni file has no such line; accepting would make `1 NAME ` and `1 NAME` the same tree and so a round-trip loss. Revisit when a file needs it |
| Length limits | Not enforced | Errors or warnings | The Geni file exceeds 255; limits are a validator's job (out of scope) |
| The tree's home | New package `gedcom`; `gedcom7` aliases the names | `gedcom551.Read` returning `*gedcom7.Document` | The tree is version-neutral (lines nested by level); a 5.5.1 reader returning a 7 type misnames the converter's input. Aliases keep the API |
| Grammar layout | `grammar/gedcom7/`, `grammar/gedcom551/` under one `tree-sitter.json`; parameterised tasks | A second top-level directory | tree-sitter's own multi-grammar layout; one task per operation |
| Day, year and the dual year | One `integer` token `[0-9]+(/[0-9][0-9])?` aliased to `day` and `year`, as in 7; Go rejects a day over 31 or with a slash | A `year` token beside a `day` token | Two tokens matching `12` is a lexer ambiguity; the 7 grammar's LR(1) reasoning (story 004) carries over |
| The date phrase | One token carrying its optional leading space, `( … )` to the last `)` | `"(" text ")"` as three tokens | After a year the next token decides day against year; a phrase token that begins with its delimiter keeps the grammar LR(1), as `" TO"` and `" AND"` do in 7 |
| The Geni file in tests | Read from `GOGEDCOM_GENI`, skipped when unset | A redacted fixture | Story 006: the file names living people and stays out of the repo |

## Done when

- [ ] tests express the acceptance tables and pass; `task grammar:test GRAMMAR=gedcom551` passes
- [ ] README status, vocabulary, packages, toolchain
- [ ] changelog fragment
- [ ] issue closed with a link to the release

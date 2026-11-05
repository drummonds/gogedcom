# 009 — The 7 to 5.5.1 converter and writer: the inverse of 008, the wording back from PHRASE, 7-native content by one rule (#29)

**Status:** done
**Roadmap:** #22 — the 7 to 5.5.1 converter and the 5.5.1 writer

## Why

The round trip (story 006, roadmap #7) is 5.5.1 → 7 → 5.5.1 with the
trees equal at the end. Story 008 built the first leg and kept, in
`PHRASE`, `_PHRASE`, `EXID.TYPE`, `SOUR @VOID@` and `SCHMA`, everything
the return leg needs. This story builds the return leg: `convert.To551`
turns a 7 tree into a 5.5.1 tree, and `gedcom551.Write` writes it on the
same contract as `gedcom7.Write`, so the trip can be run end to end.

The inverse has two kinds of input. A tree 008 made, where every row of
008's table has an exact inverse and the Geni export must come back
equal. And a 7-native tree, where 7 says things 5.5.1 cannot (`CREA`,
`UID`, `TRAN`, a `PHRASE` under `PEDI`) and `docs/round-trip.md` left one
decision open for the whole list: an extension tag, a `NOTE`, or
dropping. The mapping is the mirror of 008's one line: **5.5.1's line
syntax decides what must change (a tag 5.5.1 does not define must begin
with `_`); the 5.5.1 tag list decides what comes back, by 008's rules
read backwards, the original wording in `PHRASE` first; everything else
passes through as the tree has it.** A 5.5.1 schema violation (`SUBM`
under `INDI`, five `NOTE`s under an inline `OBJE`) is not the converter's
business, as a 7 registry violation was not 008's; 5.5.1 has no registry
here, and the writer writes the tree as it is.

The mapping is written as a decision table before the code, every 7
structure accounted for, a test keeping the code and the table in
agreement.

## The rules

One row per kind of change; a 7 structure matches the first row whose
condition holds. `date7` and `age7` are 008's re-spellings; "the guide"
is [gedcom.io/migrate](https://gedcom.io/migrate).

| Condition on the 7 structure | Outcome in 5.5.1 | Rule |
|---|---|---|
| A `CONT` line | Already folded by `gedcom7.Read`; `gedcom551.Write` splits on line breaks | reader |
| An extension tag `SCHMA` declares with a `https://gedcom.io/terms/v7/` URI the registry has a term for (`_ADDR`, `_EMAIL`, `_PHON`, `_PHRASE`) | The term's standard tag; substructures by their rules, so the rule recurses | relocated |
| Any other extension tag `_X` | Unchanged, with everything under it | — |
| `HEAD` | Its substructures by their rules; `SCHMA` consumed; `GEDC.VERS` `5.5.1`, `GEDC.FORM LINEAGE-LINKED` after it and `CHAR UTF-8` after `GEDC`, each added when absent | header |
| `HEAD.GEDC.VERS` | `5.5.1` | version |
| `HEAD.SCHMA` | Consumed: its `TAG`s drive the relocated row; nothing written | schema |
| An `SNOTE` record; an `SNOTE` pointer | `NOTE` record; `NOTE` pointer | note |
| `EXID` with `TYPE` `…/v7/RFN#prefix`, `…/v7/RFN`, `…/v7/AFN`, `…/v7/RIN`, `…/v7/RIN#system` | `RFN prefix:value`, `RFN value`, `AFN value`, `RIN value`; `TYPE` consumed | exid |
| `TRAN` under `NAME` or `PLAC` whose `LANG` is in the guide's table | `ROMN` or `FONE` with `TYPE` the table's word: `ja-Latn` → `ROMN` + `TYPE romaji`, `und-Latn-pinyin` → `pinyin`, `zh-Latn-wadegile` → `wadegiles`, `ko-hang` → `FONE` + `TYPE hangul`, `ja-hrkt` → `kana`; `TYPE` where `LANG` was | tran |
| `ROLE OTHER` with `PHRASE` under `ASSO`; the same under a citation `EVEN` | `RELA` the phrase; `ROLE` the phrase. Another `ROLE` value: `RELA` (or `ROLE`) the word as 7 spells it | role |
| `INIL` | `WAC` | wac |
| An enumeration payload (`PEDI`, `RESN`, `FAMC.STAT`, `MEDI`, `NAME.TYPE`) | Lower-cased: 5.5.1 spells these sets in lower case. `DNS_CAN` → `DNS/CAN`, `PRE_1970` → `PRE-1970`. `SEX`, `ADOP`, `QUAY` and ordinance `STAT` are spelt alike; a value outside 7's set (`PEDI step`) unchanged | enumeration |
| A `FILE`'s `FORM` in the media-type table | The 5.5.1 word: `image/jpeg` → `jpg`, `image/tiff` → `tif`, `application/pdf` → `pdf`, `text/plain` → `txt`, the rest of 008's table backwards; a type outside the table unchanged | media type |
| A `FORM`'s `MEDI` | `TYPE` under an `OBJE` record's `FORM`, `MEDI` under an inline `OBJE`'s `FORM`: 5.5.1's two placements | medi |
| An `OBJE` pointer whose record exactly one payload in the document names | Inlined: `OBJE` with no payload; its substructures the record's `FILE`s, each `FILE`'s `TITL` hoisted after it, then the pointer's own substructures (`_PRIM`), then the record's remaining substructures (`NOTE`); the record removed | inline obje |
| `SOUR @VOID@` | `SOUR` with no payload; `DATA`, `OBJE`, `NOTE` and the rest stay | citation |
| A `DATE` with a `PHRASE` 008 would have written (`date7(PHRASE)` is the payload) | The `PHRASE` as the payload: `JULIAN 1 JAN 1700` + `@#DJULIAN@ 1 JAN 1700` → the latter; `5  1801`, `(spring)`, `INT 1850 (war)` likewise | date |
| A `DATE` with another `PHRASE` (7-native, under a payload) | 5.5.1's own phrase form: `INT <date> (<phrase>)` | date |
| A `DATE` 5.5.1's data type rejects, no `PHRASE` (`JULIAN …`, `FRENCH_R …`, `100 BCE`, `_ROMAN 71`) | 5.5.1's spelling: `@#DJULIAN@ …`, `@#DFRENCH R@ …`, `100 B.C.`, `@#DROMAN@ 71`; else unchanged | date |
| An `AGE` with a `PHRASE` 008 would have written (`age7(PHRASE)` is the payload) | The `PHRASE` as the payload: `< 8y` + `CHILD` → `CHILD` | age |
| An `AGE` with another `PHRASE` | `AGE` unchanged (7's forms are 5.5.1's), `_PHRASE` the phrase | age |
| A `NAME` with a `PHRASE` (relocated back from `_PHRASE`) | The `PHRASE` as the payload: `Louise Louisa` + `Louise/Louisa` → `NAME Louise/Louisa` | name |
| An `ADDR` whose payload equals the composition of its pieces (`ADR1`…`CTRY`, one per line) | No payload | address |
| `PHRASE` under anything else (`PEDI`, `SEX`, `TYPE`, `RESN`, a `ROLE` from the set) | `_PHRASE`: 5.5.1 has no home for it | extension |
| A standard 7 tag 5.5.1 does not have (`CREA`, `CROP`, `HEIGHT`, `LEFT`, `MIME`, `NO`, `SDATE`, `TOP`, `UID`, `WIDTH`; `EXID` and `TRAN` outside their rows) | `_CREA`, `_UID`, … with everything under it unchanged: 5.5.1 requires a tag it does not define to begin with `_`, and `_UID` is what 5.5.1 files already write | extension |
| A pointer `@VOID@` under `FAMC`, `FAMS`, `HUSB`, `WIFE`, `CHIL`, `ALIA` | Unchanged; 5.5.1 has no void pointer and dropping would lose the substructures | unchanged |
| Otherwise: tag, place and payload 5.5.1 has | Unchanged, payload text verbatim | unchanged |

Rows compose. `_PHRASE` under `NAME` is first relocated back to `PHRASE`
and then consumed by the name row; `_TRAN` from the extension row keeps
its `LANG` and `_TYPE` as they are. A 7-native `DATE 1850` with
`PHRASE 1850` comes back `DATE 1850`, the redundant phrase consumed, and
a 7-native `DATE` with no payload and `PHRASE sometime` comes back
`DATE sometime`: both are what 008 would have made of the result.

## The Geni file under the rules

The 7 tree is 008's; the 5.5.1 tree must equal the original read.

| 7 structure (count) | Row | Comes back as |
|---|---|---|
| `HEAD.GEDC.VERS 7.0`; no `CHAR`, no `FORM` | header | `VERS 5.5.1`, `FORM LINEAGE-LINKED`, `CHAR UTF-8` in the original's places |
| `HEAD.SCHMA` (`_ADDR`, `_EMAIL`, `_PHON`, `_PHRASE`) | schema | gone |
| `_ADDR` (67), `_EMAIL` (59), `_PHON` (8) under `INDI` | relocated | `ADDR`, `EMAIL`, `PHON` |
| `_MARNM` (1,848), `_PRIM` (295) | extension unchanged | as they are |
| `EXID` + `TYPE …/RFN#geni` (1,976) | exid | `RFN geni:…` |
| `SNOTE` records (164), `SNOTE` pointers (164) | note | `NOTE` |
| `OBJE` records (1,448), each pointed to once | inline obje | inline, `FILE TITL _PRIM NOTE` in the original's order |
| `FORM image/jpeg`, `image/png`, `application/pdf`, `text/plain` | media type | `jpg`, `png`, `pdf`, `txt` |
| `SOUR @VOID@` (1,317) | citation | `SOUR` |
| `DATE` + `PHRASE 5  1801` (1) | date | `DATE 5  1801` |
| `NAME` + `_PHRASE` (2) | relocated, name | `Louise/Louisa`, the three-slash name |
| Event `ADDR` with the composed payload (1,326) | address | no payload |
| `SEX M`, `SEX F` | enumeration | unchanged |
| Everything else | unchanged | |

## What changes

- `convert`: `To551(doc *gedcom.Document) *gedcom.Document` returns a new
  tree by the rules; the input is not changed. No error: a 7 tree has
  nothing a 5.5.1 line cannot carry. The rules are a table in
  `convert/rules551.go` keyed as `rules.go` is, by superstructure tag and
  tag.
- `gedcom551`: `Write(w io.Writer, doc *gedcom.Document) error` on
  `gedcom7.Write`'s contract: the header, the records in order, `TRLR`;
  LF line ends, no byte order mark; a line break in a payload becomes a
  `CONT` line; every payload verbatim, no at-sign escape, as the reader
  decodes none; refuses, naming the record, a carriage return in a
  payload, a tag or identifier the 5.5.1 grammar rejects, a document
  with no header. Lines are not split with `CONC`.
- `docs/convert-7-to-551.md`: the decision table above, then one row per
  tag the 7 registry has a structure term for, in its places, naming its
  rule; `TestRules551MatchTheDocument` holds it to `rules551.go` and fails
  on a registry tag with no row.
- `docs/round-trip.md`: the open `OBJE` row decided; the 7-to-5.5.1
  section rewritten as what the code does; stage status updated.
- `README.md`: status (converts both ways, writes 5.5.1), packages,
  vocabulary (inlined record, 7-native extension); `index.md` lists the
  story and the mapping document.

## Acceptance

Each rule, as 7 lines in, 5.5.1 lines out (`gedcom7.Read`, `To551`,
`gedcom551.Write`); the `SCHMA` rows read from a header declaring
`_ADDR`, `_EMAIL`, `_PHON`, `_PHRASE`:

| 7 | 5.5.1 |
|---|---|
| `1 GEDC` / `2 VERS 7.0` under `HEAD`; no `CHAR` | `2 VERS 5.5.1` / `2 FORM LINEAGE-LINKED` / `1 CHAR UTF-8`; `SCHMA` gone; `SOUR`, `DATE`, `SUBM` kept in place |
| `0 @N1@ SNOTE a` / `1 SNOTE @N1@` | `0 @N1@ NOTE a` / `1 NOTE @N1@` |
| `1 EXID 42` / `2 TYPE …/RFN#geni`; `1 EXID 1` / `2 TYPE …/AFN`; `1 EXID 9` / `2 TYPE …/RIN#Geni.com`; `1 EXID x` / `2 TYPE https://example.com/id` | `1 RFN geni:42`; `1 AFN 1`; `1 RIN 9`; `1 _EXID x` / `2 TYPE https://example.com/id` |
| `2 TRAN Tachibana /Hayanari/` / `3 LANG ja-Latn` / `3 SURN Tachibana`; `2 TRAN x` / `3 LANG fr` | `2 ROMN Tachibana /Hayanari/` / `3 TYPE romaji` / `3 SURN Tachibana`; `2 _TRAN x` / `3 LANG fr` |
| `2 ROLE OTHER` / `3 PHRASE Honorary uncle` under `ASSO`; `2 ROLE GODP`; `3 ROLE CHIL` / `4 PHRASE eldest` under a citation `EVEN` | `2 RELA Honorary uncle`; `2 RELA GODP`; `3 ROLE CHIL` / `4 _PHRASE eldest` |
| `1 INIL` | `1 WAC` |
| `2 PEDI BIRTH`; `1 RESN PRIVACY`; `2 STAT CHALLENGED`; `2 PEDI step`; `2 STAT DNS_CAN`; `3 ADOP BOTH`; `1 SEX X` | `2 PEDI birth`; `1 RESN privacy`; `2 STAT challenged`; `2 PEDI step`; `2 STAT DNS/CAN`; `3 ADOP BOTH`; `1 SEX X` |
| `0 @O1@ OBJE` / `1 FILE a.jpg` / `2 FORM image/jpeg` / `3 MEDI PHOTO` / `2 TITL t` / `1 NOTE n` with `1 OBJE @O1@` / `2 _PRIM Y` under `INDI` | `1 OBJE` / `2 FILE a.jpg` / `3 FORM jpg` / `4 MEDI photo` / `2 TITL t` / `2 _PRIM Y` / `2 NOTE n`; no `@O1@` record |
| `0 @O1@ OBJE` / `1 FILE f` / `2 FORM image/png` / `3 MEDI PHOTO` pointed to twice, or never | the record stays, `2 FORM png` / `3 TYPE photo`; the pointers stay |
| `1 SOUR @VOID@` / `2 DATA` / `3 TEXT t`; `1 SOUR @S1@` / `2 PAGE 3` | `1 SOUR` / `2 DATA` / `3 TEXT t`; `1 SOUR @S1@` / `2 PAGE 3` |
| `2 DATE JULIAN 1 JAN 1700` / `3 PHRASE @@#DJULIAN@ 1 JAN 1700`; `2 DATE` / `3 PHRASE 5  1801`; `2 DATE 1850` / `3 PHRASE INT 1850 (war)`; `2 DATE` / `3 PHRASE (spring)` | `2 DATE @#DJULIAN@ 1 JAN 1700`; `2 DATE 5  1801`; `2 DATE INT 1850 (war)`; `2 DATE (spring)` |
| `2 DATE 1850` / `3 PHRASE about the war`; `2 DATE` / `3 PHRASE sometime` | `2 DATE INT 1850 (about the war)`; `2 DATE sometime` |
| `2 DATE JULIAN 1 JAN 1700`; `2 DATE 100 BCE`; `2 DATE FRENCH_R 1 VEND 8`; `2 DATE _ROMAN 71`; `2 DATE BET 1850 AND 1860`; `2 DATE 02 APR 2023` | `2 DATE @#DJULIAN@ 1 JAN 1700`; `2 DATE 100 B.C.`; `2 DATE @#DFRENCH R@ 1 VEND 8`; `2 DATE @#DROMAN@ 71`; unchanged; unchanged |
| `2 AGE < 8y` / `3 PHRASE CHILD`; `2 AGE` / `3 PHRASE twenty`; `2 AGE 25y` / `3 PHRASE about` | `2 AGE CHILD`; `2 AGE twenty`; `2 AGE 25y` / `3 _PHRASE about` |
| `1 NAME Louise Louisa` / `2 GIVN Louise` / `2 _PHRASE Louise/Louisa` | `1 NAME Louise/Louisa` / `2 GIVN Louise` |
| `2 ADDR Errol` / `3 CONT Scotland` / `3 CITY Errol` / `3 CTRY Scotland`; `2 ADDR Here` / `3 CITY There` | `2 ADDR` / `3 CITY Errol` / `3 CTRY Scotland`; `2 ADDR Here` / `3 CITY There` |
| `1 _ADDR x@y` / `2 _EMAIL x@y` / `2 _PHON 1` under `INDI`; `1 _MARNM Smith`; `1 _UID x` / `2 DATE 5  1801` | `1 ADDR x@y` / `2 EMAIL x@y` / `2 PHON 1`; unchanged; unchanged |
| `1 CREA` / `2 DATE 1 JAN 2020`; `1 UID 1234`; `1 FAMC @VOID@` / `2 PEDI ADOPTED` / `3 PHRASE by an aunt` | `1 _CREA` / `2 DATE 1 JAN 2020`; `1 _UID 1234`; `1 FAMC @VOID@` / `2 PEDI adopted` / `3 _PHRASE by an aunt` |
| `1 OCCU`; `1 SEX M` / `1 SEX F`; `1 SUBM @S1@` | unchanged |

`gedcom551.Write`:

| Input | Outcome |
|---|---|
| A header and records with line breaks, trailing spaces, a payload beginning with `@` or `@@`, a `CONT` payload that is empty | `gedcom551.Read` of the output equals the input in every field but `Line`; the `@` payloads are written verbatim |
| A 300-character payload | One line, not split |
| No header; a tag `A-B`; an identifier `@ 1@`; a carriage return in a payload | An error naming the record. A tag `Note` and an identifier `@I 1@` are 5.5.1's and are written |

The Geni export (`GOGEDCOM_GENI`, skipped when unset), `in` the 5.5.1
tree, `back` is `To551(To7(in))`:

| Check | Outcome |
|---|---|
| `back` against `in` | equal in every field but `Line` |
| Records | 1,976 `INDI`, 722 `FAM`, 187 `SUBM`, 164 `NOTE`, 0 `OBJE`, 0 `SNOTE` |
| `gedcom551.Read(gedcom551.Write(back))` | equal to `back` in every field but `Line` |
| The written file | 58,627 lines: the original's 60,126 less its 1,499 `CONC` lines |

The mapping document: every tag the 7 registry has a structure term for
has a row; the test agrees with `convert/rules551.go`.

## What building it taught

One row of the proposal was wrong by its own rule. It said a 7-native
`DATE` with no payload and `PHRASE sometime` would come back
`DATE (sometime)`, but 008 writes exactly `DATE` + `PHRASE sometime` for
a 5.5.1 `DATE sometime` (a payload the data type rejects, like
`5  1801`), so `date7("sometime")` is the empty payload and the
condition for "the phrase is the original" holds. Nothing in the 7 tree
tells the two apart, and the Geni `5  1801` has to come back as written,
so `DATE sometime` it is: a payload 5.5.1's line accepts and its data
type rejects, which 007 made lawful. The `INT … (…)` form is therefore
only for a 7-native phrase under a date with a payload.

The rest built as proposed. The in-memory trip of the Geni export is
equal in every field but `Line` at the first run, which says 008's
conventions carried everything; the written file has 58,627 lines, the
original's less its `CONC` lines. Checking every registry structure tag
against the mapping table is cheap once the tag list is machine-readable
(141 tags) and caught nothing, which is the point of it.

## Out of scope

- The file-level trip and its comparison with `CONC` folding and the
  provenance fields excluded, and the README's list of any remaining
  difference: #7. This story's Geni check is the in-memory trip.
- Splitting lines at 255 characters with `CONC`. The reader does not
  enforce the limit, Geni itself writes a 279-character line, and a
  consumer that needs the limit is a story that asks for it.
- A 5.5.1 schema: placing `NOTE` under `HEAD` as a pointer, `SUBM` under
  `INDI`, five `NOTE`s under an inline `OBJE`. 5.5.1 has no registry in
  this library; the tree is written as it is.
- Semantic downgrades: `SEX X` to `U`, repeated `GIVN` joined with a
  comma, a 7 `TRAN` with no 5.5.1 word made a `NOTE`, BCP 47 `LANG` to
  5.5.1's language names. Each changes meaning or guesses.
- A `_TRAN` whose `LANG und` and `_TYPE` came from a `FONE`/`ROMN` with a
  `TYPE` outside the guide's table: it stays `_TRAN`; which of the two it
  was is not in the 7 tree.
- Documenting `_UID`, `_CREA` and the other 7-native extensions in the
  7 direction: #4.

## Decisions

| Choice | Taken | Alternative | Why |
|---|---|---|---|
| 7-native content with no 5.5.1 home | An extension tag, `_CREA`, `_UID`, `_PHRASE`, with everything under it unchanged | A `NOTE` holding the lines; dropping | 5.5.1's own rule for a tag it does not define; nothing is lost, a 5.5.1 consumer skips what it does not know, and `_UID` is already the convention. A `NOTE` changes meaning; dropping loses |
| Which `OBJE` records come back inline | Any record exactly one payload in the document names, from an `OBJE` pointer | A marker on the records 008 made; a hint from the caller | 5.5.1's inline `OBJE` is a record without an identifier; the inverse takes the identifier away when nothing else needs it. One rule, no invented tag in the 7 file, no option to set. A 5.5.1 file born with single-use `OBJE` records comes back with them inlined: accepted, as the composed `ADDR` edge was in 008; a marker is the story if a file needs its records kept |
| The inlined `OBJE`'s order | `FILE` (with `TITL` hoisted after it), the pointer's substructures, then the record's rest | The record's substructures then the pointer's | Geni writes `FILE TITL _PRIM NOTE`; 5.5.1's link puts `FILE` and `TITL` first, and what sat on the pointer belongs to the link |
| `DATE` + `PHRASE`: when is the phrase the original | When `date7(PHRASE)` reproduces the 7 payload | Always the phrase; only when the 7 payload is empty | Exact inverse of 008's rule, so a 7-native phrase is told apart without a marker; "always" would turn `DATE 1850` + `PHRASE about the war` into `DATE about the war` and lose the year |
| A 7-native `DATE` phrase | 5.5.1's `INT <date> (<phrase>)` and `(<phrase>)` | `_PHRASE` beside the date | 5.5.1 has the form; it is what 008 maps *to* `DATE` + `PHRASE`, so the two directions agree |
| Enumerations | 5.5.1's case per set | 7's upper case everywhere | 5.5.1 spells `PEDI`, `RESN`, `MEDI`, `NAME.TYPE` and `FAMC.STAT` in lower case; 008 lower-to-upper and this story upper-to-lower is the only pair that round-trips the Geni file and the specification's own examples |
| `FORM.MEDI` back | `TYPE` under a record's `FORM`, `MEDI` under an inline `OBJE`'s | `TYPE` everywhere, as 008's `FORM.TYPE` row suggests | 5.5.1 puts `TYPE` in the record and `MEDI` in the link; one spelling everywhere would be off-schema in one of the two places. Geni has neither |
| `To551` and errors | None; returns the tree | `(*Document, error)` like `To7` | 7's tags and identifiers are a subset of 5.5.1's and every payload is text; an error that cannot happen is an API that lies |
| The writer's file form | LF, no byte order mark, as `gedcom7.Write` | CR LF with a BOM, as Geni writes | Form, not content; one convention across the writers; the reader accepts all four terminators and the BOM |
| At-sign in the writer | Verbatim, no `@@` escape | Escape a leading `@` as 5.5.1 says | The reader keeps `@@` and a lone `@` as the file has them (story 007: the file is the contract), so escaping on write would double what it read |
| `HEAD.CHAR`, `GEDC.FORM` | Added by `To551`, not the writer | The writer stamps them | The writer writes the tree as it is, as 7's does; the converter is where 5.5.1's constants belong, and a tree built by hand with its own `CHAR` keeps it |
| The mapping's home | A Go table in `convert/rules551.go` and `docs/convert-7-to-551.md`, a test holding them together and checking the registry's tag list | One table for both directions | The directions have different conditions; a shared table would carry two rules per row. The registry check is new: 7's tag list is machine-readable, 5.5.1's was not |

## Done when

- [x] tests express the acceptance tables and pass, the Geni checks included
- [x] `docs/convert-7-to-551.md` with every 7 registry tag; `docs/round-trip.md` updated; README status, packages, vocabulary; `index.md`
- [x] changelog fragment
- [ ] issue closed with a link to the release

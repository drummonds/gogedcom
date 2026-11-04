# 008 — The 5.5.1 to 7 converter: the mapping as a decision table, every structure accounted for, the wording kept in PHRASE (#26)

**Status:** proposed
**Roadmap:** #6 — the 5.5.1 to 7 converter

## Why

The round trip (story 006, roadmap #7) is 5.5.1 → 7 → 5.5.1 with the
trees equal at the end. The reader (story 007) delivers the first tree;
this story turns it into a GEDCOM 7 tree that `gedcom7.Write` writes,
`gedcom7.Read` reads back, and `Validate` passes, without losing anything
the inverse (#22) will need to put back.

Two specifications pull in different directions. The 7 grammar is strict
and this library writes it, so a payload or placement 7 rejects *must*
change. Story 006 says the file is the contract: an extension tag, an
`ADDR` under `INDI`, a `DATE` spelt `5  1801` must all come back, so
nothing may be dropped or re-spelt without its original wording kept.
The mapping therefore has one line: **7's grammar decides what must
change; the registry decides what may change where the change is lossless
or the specification's migration guide prescribes it; everything else
passes through as the file has it.** What the Geni file got wrong by the
registry's lights (an `OCCU` with no payload) stays wrong, and
`Validate` is how it is seen.

The mapping is written as a decision table before the code, so every
5.5.1 structure is accounted for in a document a reader can check against
the 5.5.1 specification's tag list, and a test keeps the code and the
table in agreement.

## The rules

One row per kind of change; a 5.5.1 structure matches the first row whose
condition holds. "Guide" is [gedcom.io/migrate](https://gedcom.io/migrate).

| Condition on the 5.5.1 structure | Outcome in 7 | Wording kept |
|---|---|---|
| A `CONC` or `CONT` line | Already folded by `gedcom551.Read`; `gedcom7.Write` splits on line breaks | — |
| An xref or tag the 7 grammar rejects (`@I 1@`, lower case) | `To7` returns an error naming the line | — |
| An extension tag `_X`, anywhere | Unchanged, as an undocumented extension | as is |
| A `NOTE` record, or a pointer to one | `SNOTE` record; `SNOTE` pointer | as is |
| `RFN`, `AFN`, `RIN` | `EXID` with `TYPE` per the guide: `RFN geni:123` → `EXID 123`, `TYPE https://gedcom.io/terms/v7/RFN#geni` | in `TYPE` |
| `FONE`, `ROMN` | `TRAN` with `LANG` per the guide's table | — |
| `RELA` | `ROLE` from 7's set, else `ROLE OTHER` | `PHRASE` |
| `WAC` | `INIL` | — |
| `SUBN` record and `HEAD.SUBN`; `HEAD.CHAR`; `HEAD.GEDC.FORM`; `HEAD.FILE` | Dropped: removed from 7, or a constant of the format the inverse writes again | — |
| `HEAD.GEDC.VERS` | `7.0`; `SCHMA` added after `GEDC` with one `TAG` per relocated tag used | — |
| An enumeration payload (`SEX`, `PEDI`, `RESN`, `STAT`, `QUAY`, `MEDI`, `ADOP`, `NAME.TYPE`, ordinance `STAT`) | Upper-cased; `DNS/CAN` → `DNS_CAN`, `PRE-1970` → `PRE_1970`; a value outside 7's set unchanged (so `Validate` reports it) | — |
| A media `FORM` | Media type by table: `jpg`, `jpeg` → `image/jpeg`; `png`; `gif`; `bmp`; `tif`, `tiff`; `pcx`; `wav`; `ole`; `pdf` → `application/pdf`; `txt` → `text/plain`; else unchanged | — |
| An `OBJE` with no pointer (inline) | A new `OBJE` record `@O<n>@`, one per inline `OBJE`, in document order; the pointer in its place; `TITL` moves under `FILE`; `FORM` as above; `NOTE` to the record; extension substructures (`_PRIM`) stay on the pointer | as is |
| A `SOUR` with no pointer and no payload (Geni's citations) | `SOUR @VOID@`; `DATA` (`DATE`, `TEXT`), `OBJE`, `NOTE` stay on the citation | as is |
| A `SOUR` with no pointer and a description payload | A new `SOUR` record `@S<n>@` with `TITL` the description; the rest as the row above | lossy: comes back as a pointer citation |
| A `DATE` whose 7 spelling differs from the text (`@#DJULIAN@ …`, `B.C.`, `1699/00`, `INT …`, `(…)`) | 7's spelling: `JULIAN …`, `BCE`, the first year, the date of `INT`; `(…)` and `@#DROMAN@`/`@#DUNKNOWN@` as the guide says (`_ROMAN`) | `PHRASE` the original |
| A `DATE` the 7 grammar rejects (`5  1801`) | `DATE` with no payload | `PHRASE` the original |
| An `AGE` word (`CHILD`, `INFANT`, `STILLBORN`) or form 7 rejects | `< 8y`, `< 1y`, `0y`; else `AGE` with no payload | `PHRASE` the original |
| A `NAME` the 7 grammar rejects (one slash, three, a tab) | Each `/` and tab a space, runs of spaces one | `_PHRASE` the original |
| An `ADDR` with no payload | Payload composed from the pieces present, `ADR1`…`CTRY`, one per line | — |
| A standard tag in a place the 7 registry has no entry for (`INDI.ADDR`, `ADDR.EMAIL`, `ADDR.PHON`) | Relocated: tag `_ADDR`, `_EMAIL`, `_PHON`, declared in `SCHMA` as `TAG _ADDR https://gedcom.io/terms/v7/ADDR`; substructures judged against the standard type, so the rule recurses | as is |
| Otherwise: tag and place in the 7 registry, payload the 7 grammar accepts with the same meaning | Unchanged, payload text verbatim | as is |

`_PHRASE` is `PHRASE` relocated under a structure 7 gives no `PHRASE` to
(`NAME`); the same mechanism as the row above it. New cross-reference
identifiers skip any the document already has.

## The Geni file under the rules

| Geni structure (count) | Row | Note |
|---|---|---|
| `HEAD.CHAR`, `HEAD.GEDC.FORM` | dropped | |
| `HEAD.SOUR Geni.com`, `.VERS`, `HEAD.DATE.TIME`, `HEAD.SUBM` | unchanged | |
| `INDI.RFN geni:…` (1,976) | `EXID` | |
| `INDI.SUBM` (1,637), `INDI.TITL`, `INDI.RESI`, `INDI.OCCU` (601 with no payload) | unchanged | 7 allows `SUBM` under `INDI`; an empty `OCCU` is not the grammar's concern |
| `INDI.ADDR` (67) with `EMAIL` (59), `PHON` (8) | relocated | `SCHMA` gets `_ADDR`, `_EMAIL`, `_PHON` |
| Event `ADDR` with no payload (1,326) | composed | |
| `INDI.NAME._MARNM` (1,848), `OBJE._PRIM` (295) | extension | undocumented |
| `INDI.NAME` `Louise/Louisa`, the three-slash name | `NAME` repaired | `_PHRASE`; `SCHMA` gets `_PHRASE` |
| `INDI.NOTE @N…@` (164), `NOTE` records (164) | `SNOTE` | |
| `INDI.OBJE` (918), `INDI.SOUR.OBJE` (530) | `OBJE` records | 1,448 records; `jpg`, `png`, `pdf`, `txt` by table |
| `INDI.SOUR` with `DATA.DATE`, `DATA.TEXT` ×2, `OBJE` (1,317) | `SOUR @VOID@` | |
| `DATE 5  1801` (1) | `DATE` + `PHRASE` | every other date is 7 verbatim |
| `CHAN.DATE.TIME`, `FAM.*`, `SUBM.NAME`, `SUBM.ADDR.*`, events, `PLAC`, name pieces | unchanged | |

## What changes

- `convert`: a new package. `To7(doc *gedcom.Document) (*gedcom.Document,
  error)` returns a new tree by the rules; the input is not changed. The
  rules are a table in `convert/rules.go` keyed by superstructure tag and
  tag, consulted in the order above.
- `docs/convert-551-to-7.md`: the decision table above followed by one
  row per 5.5.1 tag (the 5.5.1 specification's Appendix A), each naming
  its rule. A test reads this table and the code's and fails on a tag in
  one but not the other.
- `README.md`: status (converts 5.5.1 to 7), packages (`convert`),
  vocabulary (relocated standard structure, `_PHRASE`); `index.md` lists
  the story and the mapping document.

## Acceptance

Each rule, as 5.5.1 lines in, 7 lines out (`To7` then `gedcom7.Write`):

| 5.5.1 | 7 |
|---|---|
| `0 @N1@ NOTE a` / `1 NOTE @N1@` | `0 @N1@ SNOTE a` / `1 SNOTE @N1@` |
| `1 RFN geni:42` | `1 EXID 42` / `2 TYPE https://gedcom.io/terms/v7/RFN#geni` |
| `1 RELA Honorary uncle` | `1 ROLE OTHER` / `2 PHRASE Honorary uncle` |
| `1 SEX M`; `2 PEDI birth`; `1 RESN privacy` | `1 SEX M`; `2 PEDI BIRTH`; `1 RESN privacy` (reported by `Validate`) |
| `1 OBJE` / `2 FILE a.jpg` / `3 FORM jpg` / `2 TITL t` / `2 _PRIM Y` | `1 OBJE @O1@` / `2 _PRIM Y` … `0 @O1@ OBJE` / `1 FILE a.jpg` / `2 FORM image/jpeg` / `2 TITL t` |
| `1 SOUR` / `2 DATA` / `3 TEXT t` / `2 OBJE` / `3 FILE f` / `4 FORM pdf` | `1 SOUR @VOID@` / `2 DATA` / `3 TEXT t` / `2 OBJE @O1@` … |
| `2 DATE @#DJULIAN@ 1 JAN 1700` | `2 DATE JULIAN 1 JAN 1700` / `3 PHRASE @#DJULIAN@ 1 JAN 1700` |
| `2 DATE 100 B.C.`; `2 DATE 15 APR 1699/00`; `2 DATE INT 1850 (war)`; `2 DATE (spring)` | `100 BCE`; `15 APR 1699`; `1850`; no payload — each with `PHRASE` the original |
| `2 DATE 5  1801` | `2 DATE` / `3 PHRASE 5  1801` |
| `2 DATE 11 MAR 1961`; `ABT 1850`; `BET 1850 AND 1860` | unchanged, no `PHRASE` |
| `2 AGE CHILD` | `2 AGE < 8y` / `3 PHRASE CHILD` |
| `1 NAME Louise/Louisa` | `1 NAME Louise Louisa` / `2 _PHRASE Louise/Louisa` |
| `2 ADDR` / `3 CITY Errol` / `3 CTRY Scotland` | `2 ADDR Errol` / `3 CONT Scotland` / `3 CITY Errol` / `3 CTRY Scotland` |
| `1 ADDR x@y` / `2 EMAIL x@y` under `INDI` | `1 _ADDR x@y` / `2 _EMAIL x@y`; `HEAD` has `1 SCHMA` / `2 TAG _ADDR https://gedcom.io/terms/v7/ADDR` / `2 TAG _EMAIL …/EMAIL` |
| `1 _MARNM Smith` under `NAME` | unchanged; no `SCHMA` entry |
| `1 CHAR UTF-8`, `2 FORM LINEAGE-LINKED` under `HEAD`; `2 VERS 5.5.1` | absent; `2 VERS 7.0` |
| `0 @I 1@ INDI` | error naming line 2 |
| A document with `@O1@` already | the new record is `@O2@` |

The Geni export (`GOGEDCOM_GENI`, skipped when unset):

| Check | Outcome |
|---|---|
| `To7` | no error |
| `Validate` on the result | no violation |
| `gedcom7.Read(gedcom7.Write(result))` | equal to `result` in every field but `Line` |
| Records | 1,976 `INDI`, 722 `FAM`, 187 `SUBM`, 164 `SNOTE`, 1,448 `OBJE`, 0 `NOTE`, 0 `SOUR` |
| `SCHMA` | `_ADDR`, `_EMAIL`, `_PHON`, `_PHRASE`; nothing for `_MARNM`, `_PRIM` |
| `PHRASE` under `DATE` | 1 |
| Every payload of the input not covered by a re-spelling row | found verbatim in the output |

The mapping document: every tag in the 5.5.1 specification's Appendix A
has a row; the test agrees with `convert/rules.go`.

## Out of scope

- The inverse, 7 to 5.5.1 (#22), and the comparison (#7). The `PHRASE`,
  `_PHRASE`, `EXID TYPE` fragment, `SOUR @VOID@`, composed `ADDR` and
  relocation conventions above are what #22 reads back.
- Documenting `_MARNM` and `_PRIM` with URIs in `SCHMA`: roadmap #4.
- Semantic upgrades the guide suggests: `_MARNM` as a second `NAME` with
  `TYPE MARRIED`; splitting comma-separated name pieces; `STILLBORN` as a
  `BIRT TYPE Stillborn` plus `DEAT AGE 0y`; `SEX Male` → `M`. Each changes
  the tree in a way the inverse cannot undo.
- Repairing what 7's registry rejects but its grammar does not: an
  `OCCU` with no payload, an enumeration value outside the set, a
  `{0:1}` substructure twice. `Validate` lists them.
- Sharing one `OBJE` or `SOUR` record between citations of the same file.
- Renaming an xref 7 rejects, instead of the error.
- Character sets other than UTF-8.

## Decisions

| Choice | Taken | Alternative | Why |
|---|---|---|---|
| What must change | Only what the 7 grammar rejects, plus the lossless or guide-prescribed registry changes | Everything `Validate` would report | The file is the contract (006). A grammar rejection means the file cannot be written as 7; a registry violation is information about the source, and repairing it is a story a user asks for |
| Off-schema placements | Relocated standard structures, `_ADDR` with `SCHMA` mapping to the standard URI | Move `INDI.ADDR` into an `INDI.RESI`; or drop | 7's own mechanism for a standard structure in a non-standard place; the inverse reads the URI back to the tag; a `RESI` changes the meaning |
| Original wording of a re-spelt date | `PHRASE` on every re-spelt date, not only the rejected ones | `PHRASE` only when 7 cannot parse the date | 006: a re-spelt date is a difference. `JULIAN 1 JAN 1700` back to `@#DJULIAN@ 1 JAN 1700` could be derived, but one rule is cheaper than two and the dual year cannot be derived |
| Wording under `NAME` | `_PHRASE`, `PHRASE` relocated | `NOTE`; an ad-hoc `_NAME` | Same mechanism as the other relocations; a `NOTE` changes meaning |
| `NAME` repair | Slashes and tabs to spaces | Keep the first `/…/` pair as the surname | A deterministic rule with no guess; the surname is in `SURN` when the writer gave one; the original is in `_PHRASE` either way |
| Empty inline `SOUR` | `SOUR @VOID@` | An empty `SOUR` record per citation | 7 defines `@VOID@` as a citation with no record, which is what Geni wrote; 1,317 empty records would be noise, and 7 forbids a structure with neither payload nor substructure |
| Inline `OBJE` | One record each, no sharing | Share a record per distinct `FILE` | Sharing would make the inverse emit one inline `OBJE` where there were two with different `TITL` or `_PRIM`; order and count are content |
| `_PRIM` | On the pointer | On the record | It is a property of this `INDI`'s link to the file (primary photo), and it comes back inline with the pointer |
| Empty `ADDR` payload | Composed from the pieces | Relocate; leave empty | 7 says the payload is required; a 7 consumer then sees an address; the inverse drops a payload equal to the composition. A 5.5.1 file whose `ADDR` payload equalled the composition would come back without it: accepted, not the Geni file |
| `HEAD.CHAR`, `GEDC.FORM` | Dropped | Relocated | They are constants of the 5.5.1 format, which the inverse writes; #7's comparison excludes them with the provenance fields |
| Extension tags | Pass through undocumented | A URI table parameter | 7 permits undocumented extension tags; a URI table is #4's; nothing in this story needs it |
| The mapping's home | A Go table in `convert/rules.go` and the markdown table in `docs/`, a test holding them together | Generate one from the other | The markdown carries the why and the note per tag; generation would be a tool for one file |
| Package | `convert`, `To7` and later `To551` | `gedcom7.From551`; `gedcom551.To7` | The converter is neither version's; `convert.To7(d)` reads as the roadmap says it |
| `To7` and the input | A new tree, input untouched | In place | The caller may still want the 5.5.1 tree (the round trip compares against it) |

## Done when

- [ ] tests express the acceptance tables and pass, the Geni checks included
- [ ] `docs/convert-551-to-7.md` with every 5.5.1 tag; README status, packages, vocabulary; `index.md`
- [ ] changelog fragment
- [ ] issue closed with a link to the release

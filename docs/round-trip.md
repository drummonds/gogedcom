# GEDCOM round tripping: where it loses

Research for roadmap #6, #22 and #7. The round trip is the Geni export
read as 5.5.1, converted to 7, written, read, converted back and written
as 5.5.1, with the two trees equal by the rules of
[story 006](stories/006-readme-status-and-the-roadmap-the-geni-round-trip.html):
compared after `CONC` and `CONT` folding, header provenance aside, every
tag, payload, placement and order counting. This page lists every point
on the trip where information is lost, comes back changed, or stops the
trip, and marks what the code does today. Each row is a fact to test
against, so a change that moves a row should move its test.

![The round trip](round-trip.svg)

| Verdict | Meaning |
|---|---|
| lossless | the original comes back, by a rule a story names |
| one way | a lawful input comes back different; the Geni file has none of it unless the row says so |
| fails | the trip stops with an error |
| open | the rule is not decided yet |

Stage status: 5.5.1 read is built (story 007); 7 write and read are
built (story 005); 5.5.1 to 7 is built (story 008); 7 to 5.5.1 and
the 5.5.1 writer are not started (#22); the comparison is not started (#7).

## Reading 5.5.1

| Input | Verdict | What happens | In the Geni file |
|---|---|---|---|
| Byte order mark, CR LF, LF CR line ends | lossless | File form, not content; the writer chooses its own | BOM, CR LF |
| Leading white space on a line, an empty line | one way | Ignored as the specification instructs; nothing records them | none |
| `CONC` and `CONT` lines, wherever the split falls | lossless | Folded into the payload; the comparison folds too | 1,499 and 5,700 |
| A `CONC` payload beginning with a space | lossless | Joined without a separator, so the space is content | 279 |
| A payload ending with a space; a tab or control character in a payload | lossless | Kept verbatim | 634 trailing spaces, 140 tabs |
| `@@` in a payload | lossless, one way in meaning | Kept as `@@`; 5.5.1 means `@`, a 7 consumer sees `@@`. Geni writes a lone `@` for `@`, so decoding would make `a@b` come back `a@@b` | none; 138 lone `@` |
| `1 NAME ` with nothing after the delimiter | fails | Rejected: accepting it would make `1 NAME ` and `1 NAME` the same tree | none |
| Lines over 255 characters, tags over 31, identifiers over 22 | lossless | Limits not enforced | a 279-character line |
| `CHAR ANSEL` or `ANSI` | fails | No transcoding; the reader takes UTF-8 | `UTF-8` |
| A `DATE` or `NAME` payload its data type rejects | lossless | The line types nothing; the payload is text | `5  1801`, `Louise/Louisa`, a three-slash name |

## Converting 5.5.1 to 7 (story 008, built)

| Input | Verdict | What happens | In the Geni file |
|---|---|---|---|
| A standard structure in a place the 7 registry allows, payload 7 accepts | lossless | Unchanged | most of the file |
| A standard structure where 7 has no place for it (`INDI.ADDR`, `ADDR.EMAIL`) | lossless | Relocated: `_ADDR` declared in `SCHMA` with the standard URI; the inverse maps the URI back | 67 `ADDR`, 59 `EMAIL`, 8 `PHON` |
| An extension tag | lossless | Passes through undocumented | `_MARNM` 1,848, `_PRIM` 295 |
| A `DATE` 7 spells differently (calendar escape, `B.C.`, dual year, `INT`, a phrase) | lossless | 7 spelling with `PHRASE` the original; the inverse writes the `PHRASE`. A date 7 accepts as written (`02 APR 2023`) passes verbatim | none; 269 zero-padded days pass |
| A `DATE` the 7 grammar rejects | lossless | `DATE` with no payload and `PHRASE` the original | `5  1801` |
| An `AGE` word or a form 7 rejects | lossless | `< 8y` and the like with `PHRASE` the original | none |
| A `NAME` the 7 grammar rejects (one slash, three, a tab) | lossless | Slashes and tabs to spaces; `_PHRASE` (relocated `PHRASE`) the original | 2 |
| A `NOTE` record and pointers to it | lossless | `SNOTE`; 5.5.1 has `NOTE` records, so the inverse is one to one | 164 |
| `RFN prefix:value`, `AFN`, `RIN` | lossless | `EXID value` with `TYPE …/RFN#prefix`; the inverse reads the fragment | 1,976 |
| An inline `OBJE` | lossless for Geni, **open** in general | A new `OBJE` record per inline `OBJE`, `TITL` under `FILE`, `_PRIM` on the pointer. The inverse must fold the record back inline, and nothing marks a record as born inline: a 5.5.1 file with its own `OBJE` records, each pointed to once, would come back with them inlined. Options for #22: inline any record pointed to once; a marker on records the converter made; a hint from the caller | 1,448, no native `OBJE` records |
| An inline `SOUR` with no payload | lossless | `SOUR @VOID@`, `DATA` and `OBJE` on the citation; the inverse writes `SOUR` with no payload | 1,317 |
| An inline `SOUR` with a description payload | one way | A `SOUR` record with `TITL`; comes back as a pointer citation | none |
| A media `FORM` | one way for `jpeg`, `tiff` | `jpg` and `jpeg` both become `image/jpeg`; the inverse writes `jpg`. The same for `tif` and `tiff` | `jpg`, `png`, `pdf`, `txt` |
| An enumeration value (`PEDI birth`) | one way in case | Upper-cased for 7; the inverse writes 5.5.1's lower-case spelling, so `Birth` comes back `birth` | `SEX M`, `SEX F` only |
| An enumeration value outside 7's set | lossless | Unchanged; `Validate` reports it | none |
| An `ADDR` with no payload | one way at an edge | Payload composed from the pieces, one per line, because 7 requires it; the inverse drops a payload equal to the composition, so a 5.5.1 `ADDR` whose payload already equalled it would come back without one | 1,326 |
| `HEAD.CHAR`, `HEAD.GEDC.FORM` | lossless | Dropped; constants of the format the 5.5.1 writer writes again; excluded from the comparison with the provenance fields | both |
| `SUBN` record, `HEAD.SUBN`, `HEAD.FILE` | one way | Dropped: removed from 7 | none |
| `FONE`, `ROMN` | one way | `TRAN` with `LANG` per the migration guide; the inverse would need the reverse table (not planned) | none |
| `RELA` | lossless | `ROLE` from the set, else `ROLE OTHER` with `PHRASE`; the inverse writes the `PHRASE` or the set's word | none |
| A structure 7's registry rejects but its grammar accepts (`OCCU` with no payload, a singleton twice) | lossless | Unchanged; `Validate` reports it | 601 empty `OCCU`, 23 empty `RESI` |
| An xref 7's grammar rejects (a space, lower case, `!`, `:`) or a tag it rejects (lower case) | fails | `To7` returns an error naming the line; 7 cannot carry the identifier and nothing records a renaming | none |

## Writing and reading 7 (story 005, built)

| Input | Verdict | What happens | Probed |
|---|---|---|---|
| A line break in a payload | lossless | `CONT` on write, folded on read | yes |
| A payload or `CONT` line beginning with `@` | lossless | Written `@@`, stripped on read; `Payload` holds the value | yes: `PHRASE @#DJULIAN@ …`, `CONT @R1@  Ancestry` |
| `@@` inside a payload | lossless | Not an escape position; verbatim both ways | yes |
| A trailing space | lossless | Verbatim | yes |
| `DATE` with no payload, with or without `PHRASE` | lossless | Accepted by the grammar and the registry | yes |
| `SOUR @VOID@` with `DATA`; `_ADDR` with `_EMAIL`; `_PHRASE` under `NAME`; `OCCU` with no payload | lossless | Accepted, no violation | yes |
| `AGE CHILD`; `NAME` with one slash or a tab; `RFN` | fails on read or in `Validate` | As 7 says; this is why the converter re-spells them | yes |
| `12 _MONTH 1990`, an extension month after a day | fails | #19, since v0.3.0: the epoch token swallows the extension tag. 7-native only; 5.5.1 has no extension months | the spec's `date.ged`, `extensions.ged` |
| `BET 1850 AND 1860` through `ParseDate` | lossless | Parses; the skip in `gedcom7_test.go` naming gotreesitter's GLR fork is stale (confirmed 2026-11-03) | yes |

## Converting 7 to 5.5.1 and writing it (#22, not started)

The Geni direction is covered above by each rule's inverse. 7-native
content with no 5.5.1 home is the one-way direction of its own; the
inverse has to choose between an extension tag, a `NOTE` and dropping,
and #22 should decide once for the whole list.

| 7 structure | 5.5.1 has | Verdict | Note |
|---|---|---|---|
| `TRAN` under `NAME`, `NOTE`, `PLAC`, `FILE`; `LANG`; `MIME` | `FONE`, `ROMN` for two cases; nothing else | one way | |
| `CREA`, `UID`, `EXID` not from `RFN`/`AFN`/`RIN`, `SDATE`, `NO`, `INIL`, `CROP`, `FACT` | nothing (`WAC` for `INIL`) | one way | |
| `PHRASE` under `DATE`, `AGE` | the payload itself | lossless | the `PHRASE` is written as the payload; a 7-native phrase that is not a 5.5.1 date value comes back as a date phrase `(…)` or stops |
| `PHRASE` under `ROLE`, `PEDI`, `TYPE`, `SEX`, `RESN` | nothing | one way | |
| `@VOID@` under `FAMC`, `FAMS`, `HUSB`, `WIFE`, `CHIL`, `ALIA` | nothing | one way | `SOUR @VOID@` alone has a 5.5.1 form (the inline citation) |
| `BCE`; an extension calendar, month or epoch | `B.C.`; nothing | lossless; one way | |
| `SEX X` | `M`, `F`, `U` | one way | |
| Repeated `GIVN`, `SURN`, `NPFX`, `NSFX`, `SPFX`, `NICK` | one each | one way | join with a comma, as the guide splits |
| `HEAD.LANG`, `SUBM.LANG`, `PLAC.LANG` as BCP 47 | 5.5.1 language names | one way | |
| `SNOTE` under `HEAD` | nothing | one way | |
| `FILE.FORM` as a media type | the 5.5.1 words | one way outside the table | `image/jpeg` to `jpg`; an unknown type has no word |
| `SCHMA` | nothing | lossless | consumed: relocated tags become their standard tags, the rest dropped |
| `OBJE` records with more than one pointer | `OBJE` records | lossless | stays a record; see the inline `OBJE` row above for the one-pointer case |
| A line over 255 characters after `CONC` splitting | the limit | lossless | the writer splits at 255 or not at all; the comparison folds either way |
| `CHAR`, `GEDC.FORM`, BOM | required | lossless | written as constants |

## What the Geni file is expected to show

By the rows above, the Geni export round-trips with no difference: every
one of its structures is on a lossless row. The residual risk is in the
`open` row (the inverse must inline all 1,448 `OBJE` records) and in the
rows marked "none" for the file, which no Geni test exercises and which
need their own small fixtures. The counts in each table are from the
export of 24 October 2026.

# Converting GEDCOM 5.5.1 to 7: the mapping

What `convert.To7` does with every 5.5.1 structure (story 008, roadmap
#6). The rule has one line: **7's grammar decides what must change; the
registry decides what may change where the change is lossless or the
specification's [migration guide](https://gedcom.io/migrate/) prescribes
it; everything else passes through as the file has it.** What the file
got wrong by the registry's lights (an `OCCU` with no payload, a
singleton twice) stays wrong, and `gedcom7.Validate` is how it is seen.
Whatever is re-spelt keeps its original wording in `PHRASE`, so the
inverse (#22) can write it back.

The first table is the decision table: one row per kind of change, a
5.5.1 structure matching the first row whose condition holds. The second
is every tag of the 5.5.1 specification's Appendix A in the places 5.5.1
gives it, naming its rule; it mirrors `tagRules` in `convert/rules.go` row
for row, and `TestRulesMatchTheDocument` fails on a row in one and not the
other.

## The decision table

| Condition on the 5.5.1 structure | Outcome in 7 | Wording kept | Rule |
|---|---|---|---|
| A `CONC` or `CONT` line; `TRLR` | Folded by `gedcom551.Read` (a line break for `CONT`); `gedcom7.Write` splits on line breaks | — | reader |
| A tag the 7 grammar rejects (lower case), anywhere; a record identifier or a pointer it rejects (`@I 1@`) | `To7` returns an error naming the line | — | — |
| An extension tag `_X`, anywhere | Unchanged, with everything under it, as an undocumented extension | as is | — |
| `HEAD` | Its substructures by their rules; `GEDC.VERS` `7.0`, added when absent; `SCHMA` after `GEDC` with one `TAG` per relocated tag used, in tag order | — | header |
| `HEAD.CHAR`, `HEAD.GEDC.FORM`, `HEAD.FILE`, `HEAD.SUBN`, a `SUBN` record | Dropped: removed from 7, or a constant of the format the inverse writes again | — | dropped |
| `HEAD.GEDC.VERS` | `7.0` | — | version |
| A `NOTE` record; a `NOTE` whose payload is a pointer | `SNOTE` record; `SNOTE` pointer. Inline note text stays `NOTE` | as is | snote |
| `RFN`, `AFN`, `RIN` | `EXID` with `TYPE` per the guide: `RFN geni:42` → `EXID 42`, `TYPE https://gedcom.io/terms/v7/RFN#geni`; `AFN` → `…/AFN`; `RIN` → `…/RIN#<HEAD.SOUR>` | the prefix in `TYPE` | exid |
| `FONE`, `ROMN` | `TRAN`; `TYPE` becomes `LANG` by the guide's table, in its place; a `TYPE` outside the table stays, after `LANG und`, and is relocated | as is | tran |
| `RELA`; a citation `EVEN.ROLE` | `ROLE` from 7's set when the value upper-cased is in it, else `ROLE OTHER` | `PHRASE` the original under `OTHER` | role |
| `WAC` | `INIL` | — | inil |
| An enumeration payload (`SEX`, `PEDI`, `RESN`, `STAT`, `QUAY`, `MEDI`, `FAMC.ADOP`, `NAME.TYPE`) | Upper-cased, `DNS/CAN` → `DNS_CAN`, `PRE-1970` → `PRE_1970`, when the result is in 7's set; else unchanged, for `Validate` to report | — | enumeration |
| A `FILE`'s `FORM` | The media type by table: `jpg`, `jpeg` → `image/jpeg`; `png`; `gif`; `bmp`; `tif`, `tiff`; `pcx`; `wav`; `ole`; `pdf` → `application/pdf`; `txt` → `text/plain`; else unchanged | — | media type |
| A `FORM`'s `TYPE` | `MEDI`, an enumeration | — | medi |
| An `OBJE` with no pointer (inline) | A new `OBJE` record `@O<n>@`, one per inline `OBJE`, in document order, appended after the records; the pointer in its place; `TITL` under the record's first `FILE`; `FORM` as above; `NOTE` and every other standard substructure on the record; extension substructures (`_PRIM`) on the pointer | as is | obje record |
| A `SOUR` with no pointer and no payload (Geni's citations) | `SOUR @VOID@`; `DATA` (`DATE`, `TEXT`), `OBJE`, `NOTE`, the rest stay on the citation | as is | citation |
| A `SOUR` with no pointer and a description payload | A new `SOUR` record `@S<n>@` with `TITL` the description; the rest as the row above | lossy: comes back as a pointer citation | citation |
| A `DATE` the 7 grammar accepts as written (`11 MAR 1961`, `02 APR 2023`, `BET 1850 AND 1860`) | Unchanged | as is | date |
| A `DATE` 7 spells differently (`@#DJULIAN@ …`, `@#DFRENCH R@`, `B.C.`, `1699/00`, `INT …`) | 7's spelling: `JULIAN …`, `FRENCH_R …`, `BCE`, the first year, the date of `INT`; `@#DROMAN@` and `@#DUNKNOWN@` as the guide says, `_ROMAN`, `_UNKNOWN` | `PHRASE` the original | date |
| A `DATE` 7 cannot spell (`(spring)`, `5  1801`) | `DATE` with no payload | `PHRASE` the original | date |
| An `AGE` word (`CHILD`, `INFANT`, `STILLBORN`) or a form 7 rejects | `< 8y`, `< 1y`, `0y`; else `AGE` with no payload | `PHRASE` the original | age |
| A `NAME` the 7 grammar rejects (one slash, three, a tab) | Each `/` and tab a space, runs of spaces one; no payload when 7 still rejects it | `_PHRASE` the original | name |
| An `ADDR` with no payload | Payload composed from the pieces present, `ADR1`…`CTRY`, one per line, in their order; no pieces, no payload | — | address |
| A standard tag in a place the 7 registry has no entry for (`INDI.ADDR`, `ADDR.EMAIL`, `ADDR.PHON`; a `PHRASE` under `NAME`; a `TEXT` under a citation) | Relocated: tag `_ADDR`, `_EMAIL`, `_PHON`, `_PHRASE`, declared in `SCHMA` as `TAG _ADDR https://gedcom.io/terms/v7/ADDR` (the plain term for the tag when there is one, else the first by URI); substructures judged against the standard type, so the rule recurses. A tag 7 has no term for at all is relocated undeclared | as is | — |
| Otherwise: tag and place in the 7 registry, payload the 7 grammar accepts with the same meaning | Unchanged, payload text verbatim; a pointer payload checked against 7's identifier rule | as is | unchanged |

The last two rows apply after the others: a rule names the 7 tag and
payload, then the placement rule decides whether it may stand where it
is. `_PHRASE` is `PHRASE` relocated under a structure 7 gives no `PHRASE`
to; new cross-reference identifiers skip any the document already has.

## The tags

Place is where 5.5.1 puts the tag: a record, any superstructure, or the
one superstructure the row is about. A more specific row wins over `any`.

| Tag | Place | Rule | Note |
|---|---|---|---|
| `ABBR` | any | unchanged |  |
| `ADDR` | any | address | no payload: the payload is the `ADR1`, `ADR2`, `ADR3`, `CITY`, `STAE`, `POST`, `CTRY` present, one per line, in their order; no pieces, no payload |
| `ADR1` | any | unchanged | an address piece; composes an empty `ADDR` payload |
| `ADR2` | any | unchanged | as `ADR1` |
| `ADR3` | any | unchanged | as `ADR1` |
| `ADOP` | any | unchanged | the event |
| `ADOP` | `FAMC` | enumeration | `HUSB`, `WIFE`, `BOTH` |
| `AFN` | any | exid | `EXID` with `TYPE …/AFN` |
| `AGE` | any | age | `CHILD` → `< 8y`, `INFANT` → `< 1y`, `STILLBORN` → `0y`, another form 7 rejects → no payload, each with `PHRASE` the original |
| `AGNC` | any | unchanged |  |
| `ALIA` | any | unchanged |  |
| `ANCE` | `SUBN` | dropped | with `SUBN` |
| `ANCI` | any | unchanged |  |
| `ANUL` | any | unchanged |  |
| `ASSO` | any | unchanged |  |
| `AUTH` | any | unchanged |  |
| `BAPL` | any | unchanged |  |
| `BAPM` | any | unchanged |  |
| `BARM` | any | unchanged |  |
| `BASM` | any | unchanged |  |
| `BIRT` | any | unchanged |  |
| `BLES` | any | unchanged |  |
| `BURI` | any | unchanged |  |
| `CALN` | any | unchanged |  |
| `CAST` | any | unchanged |  |
| `CAUS` | any | unchanged |  |
| `CENS` | any | unchanged |  |
| `CHAN` | any | unchanged |  |
| `CHAR` | `HEAD` | dropped | a constant of the 5.5.1 format; the inverse writes `UTF-8` |
| `CHIL` | any | unchanged |  |
| `CHR` | any | unchanged |  |
| `CHRA` | any | unchanged |  |
| `CITY` | any | unchanged |  |
| `CONC` | any | reader | folded by `gedcom551.Read` |
| `CONF` | any | unchanged |  |
| `CONL` | any | unchanged |  |
| `CONT` | any | reader | folded by `gedcom551.Read`; `gedcom7.Write` splits a line break again |
| `COPR` | any | unchanged |  |
| `CORP` | any | unchanged |  |
| `CREM` | any | unchanged |  |
| `CTRY` | any | unchanged |  |
| `DATA` | any | unchanged |  |
| `DATE` | any | date | 7's spelling when the grammar rejects the payload, with `PHRASE` the original: `@#DJULIAN@` → `JULIAN`, `@#DFRENCH R@` → `FRENCH_R`, `@#DROMAN@` → `_ROMAN`, `B.C.` → `BCE`, `1699/00` → `1699`, `INT 1850 (war)` → `1850`, `(spring)` → no payload; a payload 7 accepts (`02 APR 2023`, `BET 1850 AND 1860`) passes verbatim |
| `DEAT` | any | unchanged |  |
| `DESC` | `SUBN` | dropped | with `SUBN` |
| `DESI` | any | unchanged |  |
| `DEST` | any | unchanged |  |
| `DIV` | any | unchanged |  |
| `DIVF` | any | unchanged |  |
| `DSCR` | any | unchanged |  |
| `EDUC` | any | unchanged |  |
| `EMAIL` | any | unchanged | under `ADDR`, where Geni puts it, 7 has no place: relocated `_EMAIL` |
| `EMIG` | any | unchanged |  |
| `ENDL` | any | unchanged |  |
| `ENGA` | any | unchanged |  |
| `EVEN` | any | unchanged |  |
| `FACT` | any | unchanged |  |
| `FAM` | record | unchanged |  |
| `FAMC` | any | unchanged |  |
| `FAMF` | `SUBN` | dropped | with `SUBN` |
| `FAMS` | any | unchanged |  |
| `FAX` | any | unchanged |  |
| `FCOM` | any | unchanged |  |
| `FILE` | any | unchanged | under an `OBJE` record |
| `FILE` | `HEAD` | dropped | removed from 7 |
| `FONE` | any | tran | `TRAN`; `TYPE hangul` → `LANG ko-hang`, `kana` → `ja-hrkt`; another `TYPE` stays after `LANG und` and is relocated |
| `FORM` | any | unchanged | `PLAC.FORM` and `HEAD.PLAC.FORM` |
| `FORM` | `FILE` | media type | `jpg`, `jpeg` → `image/jpeg`; `png`, `gif`, `bmp`, `tif`, `tiff`, `pcx` → `image/…`; `wav` → `audio/wav`; `ole` → `application/ole`; `pdf` → `application/pdf`; `txt` → `text/plain`; else unchanged |
| `FORM` | `GEDC` | dropped | a constant of the 5.5.1 format |
| `GEDC` | any | unchanged |  |
| `GIVN` | any | unchanged |  |
| `GRAD` | any | unchanged |  |
| `HEAD` | record | header | `GEDC.VERS` 7.0, added when absent; `SCHMA` after `GEDC` when a tag was relocated |
| `HUSB` | any | unchanged |  |
| `IDNO` | any | unchanged |  |
| `IMMI` | any | unchanged |  |
| `INDI` | record | unchanged |  |
| `LANG` | any | unchanged | a 5.5.1 language name where 7 wants a BCP 47 tag; the grammar does not check it |
| `LATI` | any | unchanged |  |
| `LONG` | any | unchanged |  |
| `MAP` | any | unchanged |  |
| `MARB` | any | unchanged |  |
| `MARC` | any | unchanged |  |
| `MARL` | any | unchanged |  |
| `MARR` | any | unchanged |  |
| `MARS` | any | unchanged |  |
| `MEDI` | any | enumeration | under `CALN` |
| `NAME` | any | name | when 7 rejects the name (one slash, three, a tab): each `/` and tab a space, runs of spaces one, `_PHRASE` the original |
| `NATI` | any | unchanged |  |
| `NATU` | any | unchanged |  |
| `NCHI` | any | unchanged |  |
| `NICK` | any | unchanged |  |
| `NMR` | any | unchanged |  |
| `NOTE` | record | snote | `0 @N1@ SNOTE` |
| `NOTE` | any | snote | a pointer becomes `SNOTE @N1@`; inline text stays `NOTE` |
| `NPFX` | any | unchanged |  |
| `NSFX` | any | unchanged |  |
| `OBJE` | record | unchanged |  |
| `OBJE` | any | obje record | a pointer passes; an inline `OBJE` becomes `OBJE @O<n>@` and a record: `FILE` with `FORM`, `TITL` under the first `FILE`, `NOTE` and the rest on the record, extension tags (`_PRIM`) on the pointer |
| `OCCU` | any | unchanged |  |
| `ORDI` | `SUBN` | dropped | with `SUBN` |
| `ORDN` | any | unchanged |  |
| `PAGE` | any | unchanged |  |
| `PEDI` | any | enumeration | `birth` → `BIRTH`; `step` stays |
| `PHON` | any | unchanged | as `EMAIL` |
| `PLAC` | any | unchanged |  |
| `POST` | any | unchanged |  |
| `PROB` | any | unchanged |  |
| `PROP` | any | unchanged |  |
| `PUBL` | any | unchanged |  |
| `QUAY` | any | enumeration | `0`–`3`, unchanged in effect |
| `REFN` | any | unchanged |  |
| `RELA` | any | role | `ROLE`, the value upper-cased when in 7's set, else `ROLE OTHER` with `PHRASE` the original |
| `RELI` | any | unchanged |  |
| `REPO` | record | unchanged |  |
| `REPO` | any | unchanged |  |
| `RESI` | any | unchanged |  |
| `RESN` | any | enumeration | a list: each value upper-cased when all are in the set |
| `RETI` | any | unchanged |  |
| `RFN` | any | exid | `RFN geni:42` becomes `EXID 42`, `TYPE …/RFN#geni`; no colon, `TYPE …/RFN` |
| `RIN` | any | exid | `EXID` with `TYPE …/RIN#<HEAD.SOUR>`, or `…/RIN` with no `HEAD.SOUR` |
| `ROLE` | any | role | the citation `EVEN.ROLE`, 5.5.1 free text: as `RELA` |
| `ROMN` | any | tran | `TRAN`; `TYPE pinyin` → `LANG und-Latn-pinyin`, `romaji` → `ja-Latn`, `wadegiles` → `zh-Latn-wadegile`; as `FONE` otherwise |
| `SEX` | any | enumeration | `M`, `F`, `U` pass; another value stays for `Validate` |
| `SLGC` | any | unchanged |  |
| `SLGS` | any | unchanged |  |
| `SOUR` | record | unchanged |  |
| `SOUR` | any | citation | a pointer passes; no pointer and no payload becomes `SOUR @VOID@`; a description becomes `SOUR @S<n>@` and a record with `TITL` the description; the substructures stay on the citation |
| `SOUR` | `HEAD` | unchanged | the system id, not a citation; its `VERS`, `NAME`, `CORP`, `DATA` pass |
| `SPFX` | any | unchanged |  |
| `SSN` | any | unchanged |  |
| `STAE` | any | unchanged |  |
| `STAT` | any | enumeration | `FAMC.STAT` and the ordinance `STAT`; `DNS/CAN` → `DNS_CAN`, `PRE-1970` → `PRE_1970` |
| `SUBM` | record | unchanged |  |
| `SUBM` | any | unchanged | the pointer, under `HEAD`, `INDI` and `FAM` |
| `SUBN` | record | dropped | removed from 7, with everything under it |
| `SUBN` | `HEAD` | dropped | removed from 7 |
| `SURN` | any | unchanged |  |
| `TEMP` | any | unchanged | under an ordinance |
| `TEMP` | `SUBN` | dropped | with `SUBN` |
| `TEXT` | any | unchanged | under `DATA`; directly under a citation 7 has no place: relocated |
| `TIME` | any | unchanged |  |
| `TITL` | any | unchanged |  |
| `TRLR` | record | reader | the reader ends the file on it |
| `TYPE` | any | unchanged | an event, `EVEN`, `FACT`, `IDNO`, `REFN` or `EXID` type: text |
| `TYPE` | `FORM` | medi | 5.5.1's source media type becomes 7's `MEDI`, upper-cased when in the set |
| `TYPE` | `NAME` | enumeration | `birth` → `BIRTH` |
| `VERS` | any | unchanged |  |
| `VERS` | `GEDC` | version | `5.5.1` becomes `7.0` |
| `WAC` | any | inil | not in Appendix A; the guide names it. `INIL` |
| `WIFE` | any | unchanged |  |
| `WILL` | any | unchanged |  |
| `WWW` | any | unchanged |  |

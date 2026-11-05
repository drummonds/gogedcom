# Converting GEDCOM 7 to 5.5.1: the mapping

What `convert.To551` does with every 7 structure (story 009, roadmap
#22): the inverse of [converting 5.5.1 to 7](convert-551-to-7.html). The
rule mirrors 008's: **5.5.1's line syntax decides what must change (a tag
5.5.1 does not define must begin with `_`); the 5.5.1 tag list decides
what comes back, by 008's rules read backwards, the original wording in
`PHRASE` first; everything else passes through as the tree has it.** A
5.5.1 schema violation (`SUBM` under `INDI`, five `NOTE`s under an inline
`OBJE`) is not the converter's business, as a 7 registry violation was
not 008's: 5.5.1 has no registry here, and `gedcom551.Write` writes the
tree as it is.

The first table is the decision table: one row per kind of change, a 7
structure matching the first row whose condition holds. The second is
every tag the 7 registry has a structure term for, in the places 7 gives
it, naming its rule; it mirrors `tagRules551` in `convert/rules551.go`
row for row, and `TestRules551MatchTheDocument` fails on a row in one and
not the other, or on a registry tag with no row.

## The decision table

`date7` and `age7` are 008's re-spellings; "the guide" is
[gedcom.io/migrate](https://gedcom.io/migrate).

| Condition on the 7 structure | Outcome in 5.5.1 | Rule |
|---|---|---|
| A `CONT` line; `TRLR` | Folded by `gedcom7.Read`; `gedcom551.Write` splits on line breaks | reader |
| An extension tag `SCHMA` declares with a `https://gedcom.io/terms/v7/` URI the registry has a term for (`_ADDR`, `_EMAIL`, `_PHON`, `_PHRASE`) | The term's standard tag; substructures by their rules, so the rule recurses | — |
| Any other extension tag `_X` | Unchanged, with everything under it | — |
| `HEAD` | Its substructures by their rules; `SCHMA` consumed; `GEDC.VERS` `5.5.1`, `GEDC.FORM LINEAGE-LINKED` after it and `CHAR UTF-8` after `GEDC`, each added when absent | header |
| `HEAD.GEDC.VERS` | `5.5.1` | version |
| `HEAD.SCHMA` | Consumed: its `TAG`s drive the relocation back; nothing written | schema |
| An `SNOTE` record; an `SNOTE` pointer | `NOTE` record; `NOTE` pointer | note |
| `EXID` with `TYPE` `…/v7/RFN#prefix`, `…/v7/RFN`, `…/v7/AFN`, `…/v7/RIN`, `…/v7/RIN#system` | `RFN prefix:value`, `RFN value`, `AFN value`, `RIN value`; `TYPE` consumed. Another `TYPE`: `_EXID` | exid |
| `TRAN` under `NAME` or `PLAC` whose `LANG` is in the guide's table | `ROMN` or `FONE` with `TYPE` the table's word, in `LANG`'s place: `ja-Latn` → `ROMN` + `TYPE romaji`, `und-Latn-pinyin` → `pinyin`, `zh-Latn-wadegile` → `wadegiles`, `ko-hang` → `FONE` + `TYPE hangul`, `ja-hrkt` → `kana`. Another `TRAN`: `_TRAN` | tran |
| `ROLE OTHER` with `PHRASE` under `ASSO`; the same under a citation `EVEN` | `RELA` the phrase; `ROLE` the phrase. Another `ROLE` value: `RELA` (or `ROLE`) the word as 7 spells it, a `PHRASE` under it `_PHRASE` | role |
| An `INIL` structure | `WAC` | wac |
| An enumeration payload (`PEDI`, `RESN`, `FAMC.STAT`, `MEDI`, `NAME.TYPE`) in 7's set | Lower-cased: 5.5.1 spells these sets in lower case. `DNS_CAN` → `DNS/CAN`, `PRE_1970` → `PRE-1970`. `SEX`, `ADOP`, `QUAY` and ordinance `STAT` are spelt alike; a value outside 7's set (`PEDI step`) unchanged | enumeration |
| A `FILE`'s `FORM` in the media-type table | The 5.5.1 word: `image/jpeg` → `jpg`, `image/tiff` → `tif`, `application/pdf` → `pdf`, `text/plain` → `txt`, the rest of 008's table backwards; a type outside the table unchanged | media type |
| A `FORM`'s `MEDI` | `TYPE` under an `OBJE` record's `FORM`, `MEDI` under an inlined `OBJE`'s `FORM`: 5.5.1's two placements; lower-cased | medi |
| An `OBJE` pointer whose record exactly one payload in the document names | Inlined: `OBJE` with no payload; its substructures the record's `FILE`s, each `FILE`'s `TITL` hoisted after it, then the pointer's own substructures (`_PRIM`), then the record's remaining substructures (`NOTE`); the record leaves the records | inline obje |
| `SOUR @VOID@` | `SOUR` with no payload; `DATA`, `OBJE`, `NOTE` and the rest stay | citation |
| A `DATE` with a `PHRASE` 008 would have written (`date7(PHRASE)` is the payload, "" included) | The `PHRASE` as the payload: `JULIAN 1 JAN 1700` + `@#DJULIAN@ 1 JAN 1700` → the latter; `DATE` + `5  1801`, `(spring)`, `sometime` → the phrase; `1850` + `INT 1850 (war)` → the phrase | date |
| A `DATE` with another `PHRASE` (7-native, under a payload) | 5.5.1's own phrase form: `INT <date> (<phrase>)`, the date spelt as the row below | date |
| A `DATE` 5.5.1's data type rejects, no `PHRASE` (`JULIAN …`, `FRENCH_R …`, `100 BCE`, `_ROMAN 71`) | 5.5.1's spelling: `@#DJULIAN@ …`, `@#DFRENCH R@ …`, `100 B.C.`, `@#DROMAN@ 71`; a payload neither data type accepts unchanged | date |
| An `AGE` with a `PHRASE` 008 would have written (`age7(PHRASE)` is the payload) | The `PHRASE` as the payload: `< 8y` + `CHILD` → `CHILD`; no payload + `twenty` → `twenty` | age |
| An `AGE` with another `PHRASE` | `AGE` unchanged (7's forms are 5.5.1's), `_PHRASE` the phrase | age |
| A `NAME` with a `PHRASE` (relocated back from `_PHRASE`) | The `PHRASE` as the payload: `Louise Louisa` + `Louise/Louisa` → `NAME Louise/Louisa` | name |
| An `ADDR` whose payload equals the composition of its pieces (`ADR1`…`CTRY`, one per line) | No payload | address |
| `PHRASE` under anything else (`PEDI`, `SEX`, `TYPE`, `RESN`, `ALIA`) | `_PHRASE`: 5.5.1 has no home for it | extension |
| A standard 7 tag 5.5.1 does not have (`CREA`, `CROP`, `HEIGHT`, `LEFT`, `MIME`, `NO`, `SDATE`, `TOP`, `UID`, `WIDTH`; `EXID` and `TRAN` outside their rows) | `_CREA`, `_UID`, … with everything under it unchanged: 5.5.1 requires a tag it does not define to begin with `_`, and `_UID` is what 5.5.1 files already write | extension |
| A pointer `@VOID@` under `FAMC`, `FAMS`, `HUSB`, `WIFE`, `CHIL`, `ALIA` | Unchanged; 5.5.1 has no void pointer and dropping would lose the substructures | unchanged |
| Otherwise: tag, place and payload 5.5.1 has | Unchanged, payload text verbatim | unchanged |

Rows compose. `_PHRASE` under `NAME` is first relocated back to `PHRASE`
and then consumed by the name row; `_TRAN` from the extension row keeps
its `LANG` and `_TYPE` as they are. A 7-native `DATE 1850` with
`PHRASE 1850` comes back `DATE 1850`, the redundant phrase consumed, and
a 7-native `DATE` with no payload and `PHRASE sometime` comes back
`DATE sometime`, a payload 5.5.1's data type rejects but its line
accepts: both are what 008 would have made of the result.

## The tags

Place is where 7 puts the tag: a record, any superstructure, or the one
superstructure the row is about. A more specific row wins over `any`. A
row named after its parent's rule (`TYPE` under `EXID`, `LANG` under
`TRAN`, `PHRASE` under `DATE` or `NAME`) is consumed by the parent; left
unconsumed, it is unchanged.

| Tag | Place | Rule | Note |
|---|---|---|---|
| `ABBR` | any | unchanged |  |
| `ADDR` | any | address | a payload equal to its pieces (`ADR1`…`CTRY`, one per line) dropped |
| `ADOP` | any | unchanged | the event |
| `ADOP` | `FAMC` | enumeration | spelt alike |
| `ADR1` | any | unchanged | an address piece |
| `ADR2` | any | unchanged |  |
| `ADR3` | any | unchanged |  |
| `AGE` | any | age | with a `PHRASE` 008 would have written (`CHILD` under `< 8y`, a word under no payload): the `PHRASE` is the payload; another `PHRASE` → `_PHRASE` |
| `AGNC` | any | unchanged |  |
| `ALIA` | any | unchanged |  |
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
| `CHIL` | any | unchanged |  |
| `CHR` | any | unchanged |  |
| `CHRA` | any | unchanged |  |
| `CITY` | any | unchanged |  |
| `CONF` | any | unchanged |  |
| `CONL` | any | unchanged |  |
| `CONT` | any | reader | folded by `gedcom7.Read`; `gedcom551.Write` splits a payload's line breaks into `CONT` lines |
| `COPR` | any | unchanged |  |
| `CORP` | any | unchanged |  |
| `CREA` | any | extension | `_CREA` |
| `CREM` | any | unchanged |  |
| `CROP` | any | extension | `_CROP`, with `TOP`, `LEFT`, `HEIGHT`, `WIDTH` under it |
| `CTRY` | any | unchanged |  |
| `DATA` | any | unchanged |  |
| `DATE` | any | date | with a `PHRASE` 008 would have written (`date7(PHRASE)` is the payload, "" included): the `PHRASE` is the payload; with another `PHRASE`: `INT <date> (<phrase>)`; without: `JULIAN …` → `@#DJULIAN@ …`, `FRENCH_R` → `@#DFRENCH R@`, `BCE` → `B.C.`, `_ROMAN` → `@#DROMAN@` when 5.5.1 rejects the payload as written |
| `DEAT` | any | unchanged |  |
| `DESI` | any | unchanged |  |
| `DEST` | any | unchanged |  |
| `DIV` | any | unchanged |  |
| `DIVF` | any | unchanged |  |
| `DSCR` | any | unchanged |  |
| `EDUC` | any | unchanged |  |
| `EMAIL` | any | unchanged | relocated back from `_EMAIL` under `ADDR` |
| `EMIG` | any | unchanged |  |
| `ENDL` | any | unchanged |  |
| `ENGA` | any | unchanged |  |
| `EVEN` | any | unchanged |  |
| `EXID` | any | exid | `RFN`, `AFN` or `RIN` by the `TYPE`'s term, `RFN` with the fragment as its prefix: `EXID 42` + `TYPE …/RFN#geni` → `RFN geni:42`; another `TYPE`, or none, → `_EXID` |
| `FACT` | any | unchanged |  |
| `FAM` | record | unchanged |  |
| `FAMC` | any | unchanged | `@VOID@` stays |
| `FAMS` | any | unchanged |  |
| `FAX` | any | unchanged |  |
| `FCOM` | any | unchanged |  |
| `FILE` | any | unchanged |  |
| `FORM` | any | unchanged | `PLAC.FORM`, `HEAD.PLAC.FORM` |
| `FORM` | `FILE` | media type | `image/jpeg` → `jpg`, `image/tiff` → `tif`, `application/pdf` → `pdf`, `text/plain` → `txt`, `image/png`, `image/gif`, `image/bmp`, `image/vnd.zbrush.pcx` → `pcx`, `audio/wav`, `application/ole`; another type unchanged |
| `GEDC` | any | unchanged |  |
| `GIVN` | any | unchanged |  |
| `GRAD` | any | unchanged |  |
| `HEAD` | record | header | its substructures by their rules; `VERS` `5.5.1`, `FORM LINEAGE-LINKED` after it, `CHAR UTF-8` after `GEDC`, each added when absent; `SCHMA` consumed |
| `HEIGHT` | any | extension | under `_CROP` |
| `HUSB` | any | unchanged |  |
| `IDNO` | any | unchanged |  |
| `IMMI` | any | unchanged |  |
| `INDI` | record | unchanged |  |
| `INIL` | any | wac | `WAC` |
| `LANG` | any | unchanged | `HEAD`, `SUBM`; one where 5.5.1 places none (`PLAC`, `NOTE`) stays as it is |
| `LANG` | `TRAN` | tran | becomes `TYPE` the table's word, in its place; stays under `_TRAN` |
| `LATI` | any | unchanged |  |
| `LEFT` | any | extension | under `_CROP` |
| `LONG` | any | unchanged |  |
| `MAP` | any | unchanged |  |
| `MARB` | any | unchanged |  |
| `MARC` | any | unchanged |  |
| `MARL` | any | unchanged |  |
| `MARR` | any | unchanged |  |
| `MARS` | any | unchanged |  |
| `MEDI` | any | enumeration | under `CALN`: lower-cased |
| `MEDI` | `FORM` | medi | `TYPE` under an `OBJE` record's `FORM`, `MEDI` under an inlined `OBJE`'s; lower-cased |
| `MIME` | any | extension | `_MIME` |
| `NAME` | any | name | a `PHRASE` (relocated back from `_PHRASE`) is the payload |
| `NATI` | any | unchanged |  |
| `NATU` | any | unchanged |  |
| `NCHI` | any | unchanged |  |
| `NICK` | any | unchanged |  |
| `NMR` | any | unchanged |  |
| `NO` | any | extension | `_NO`, its `DATE` under it |
| `NOTE` | any | unchanged | inline note text |
| `NPFX` | any | unchanged |  |
| `NSFX` | any | unchanged |  |
| `OBJE` | record | unchanged |  |
| `OBJE` | any | inline obje | a pointer to a record that exactly one payload names, this one → inline `OBJE`: the record's `FILE`s each with its `TITL` hoisted after it, the pointer's substructures, the record's rest; the record leaves the records. Another pointer unchanged |
| `OCCU` | any | unchanged |  |
| `ORDN` | any | unchanged |  |
| `PAGE` | any | unchanged |  |
| `PEDI` | any | enumeration | lower-cased when in 7's set: `BIRTH` → `birth`; `step` stays |
| `PHON` | any | unchanged | relocated back from `_PHON` under `ADDR` |
| `PHRASE` | any | extension | `_PHRASE`: under `PEDI`, `SEX`, `TYPE`, `RESN`, `MEDI`, a `ROLE` not consumed, an `AGE` not consumed, `ALIA`, `CHIL`, `HUSB`, `WIFE` |
| `PHRASE` | `DATE` | date | consumed by `DATE` |
| `PHRASE` | `NAME` | name | consumed by `NAME` |
| `PLAC` | any | unchanged |  |
| `POST` | any | unchanged |  |
| `PROB` | any | unchanged |  |
| `PROP` | any | unchanged |  |
| `PUBL` | any | unchanged |  |
| `QUAY` | any | enumeration | spelt alike |
| `REFN` | any | unchanged |  |
| `RELI` | any | unchanged |  |
| `REPO` | record | unchanged |  |
| `REPO` | any | unchanged |  |
| `RESI` | any | unchanged |  |
| `RESN` | any | enumeration | lower-cased, each value of the list |
| `RETI` | any | unchanged |  |
| `ROLE` | any | role | `RELA` under `ASSO`, `ROLE` under a citation `EVEN`; `OTHER` with `PHRASE` → the phrase; another value as 7 spells it |
| `SCHMA` | `HEAD` | schema | consumed; nothing written |
| `SDATE` | any | extension | `_SDATE` |
| `SEX` | any | enumeration | spelt alike; `X` stays |
| `SLGC` | any | unchanged |  |
| `SLGS` | any | unchanged |  |
| `SNOTE` | record | note | `NOTE` record |
| `SNOTE` | any | note | `NOTE` pointer, under `HEAD` too |
| `SOUR` | record | unchanged |  |
| `SOUR` | any | citation | `@VOID@` → no payload; a pointer stays |
| `SPFX` | any | unchanged |  |
| `SSN` | any | unchanged |  |
| `STAE` | any | unchanged |  |
| `STAT` | any | enumeration | ordinance status: `DNS_CAN` → `DNS/CAN`, `PRE_1970` → `PRE-1970`, the rest spelt alike |
| `STAT` | `FAMC` | enumeration | lower-cased |
| `SUBM` | record | unchanged |  |
| `SUBM` | any | unchanged | under `INDI` too |
| `SURN` | any | unchanged |  |
| `TAG` | `SCHMA` | schema | `TAG _ADDR https://gedcom.io/terms/v7/ADDR` maps `_ADDR` back to `ADDR` when the registry has the URI's term; another URI leaves the tag an extension |
| `TEMP` | any | unchanged |  |
| `TEXT` | any | unchanged | under a citation's `DATA` too |
| `TIME` | any | unchanged |  |
| `TITL` | any | unchanged |  |
| `TOP` | any | extension | under `_CROP` |
| `TRAN` | any | tran | under `NAME` or `PLAC` with a `LANG` in the guide's table → `ROMN` (`ja-Latn` romaji, `und-Latn-pinyin` pinyin, `zh-Latn-wadegile` wadegiles) or `FONE` (`ko-hang` hangul, `ja-hrkt` kana); else `_TRAN` |
| `TRLR` | record | reader | the reader ends on it; the writer writes it |
| `TYPE` | any | unchanged | an event or `REFN` type |
| `TYPE` | `EXID` | exid | consumed by `RFN`, `AFN`, `RIN`; stays under `_EXID` |
| `TYPE` | `NAME` | enumeration | lower-cased |
| `UID` | any | extension | `_UID`, the 5.5.1 convention |
| `VERS` | any | unchanged |  |
| `VERS` | `GEDC` | version | `5.5.1` |
| `WIDTH` | any | extension | under `_CROP` |
| `WIFE` | any | unchanged |  |
| `WILL` | any | unchanged |  |
| `WWW` | any | unchanged |  |

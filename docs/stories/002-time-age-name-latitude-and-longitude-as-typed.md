# 002 — Time, age, name, latitude and longitude as typed values (#9)

**Status:** done locally
**Roadmap:** #1 — the specification's ABNF as a Go parser

## Why

Story 001 typed one of the six data types the grammar carries: `DATE` reads
as a `DateValue`. `TIME`, `AGE`, `NAME`, `LATI` and `LONG` parse under the
same grammar, so a bad payload is already rejected at `Read`, but a caller
who wants the hour of a time or the surname of a name reads the payload
string and splits it by hand, which is the transcription the project exists
to avoid. Roadmap #1 names all six; this story finishes it.

## What changes

- `gedcom7`: `ParseTime`, `ParseAge`, `ParsePersonalName`, `ParseLatitude`
  and `ParseLongitude`, each on the `ParseDate` pattern: the payload is
  parsed as the line `0 TAG <s>` and the typed line's value node read into a
  Go value named as the spec names it. `Time` (hour, minute, second,
  fraction, UTC), `Age` (bound, years, months, weeks, days), `PersonalName`
  (given, surname, suffix), `Latitude` and `Longitude` (signed degrees,
  north and east positive). Each has a `String()` that writes GEDCOM.
- The line scaffolding `ParseDate` carried (build the line, parse, find the
  first error, find the typed line) becomes one unexported function the six
  parsers share.
- The reader checks the tree covers every byte of the line. gotreesitter
  v0.7.0 skips a character its lexer cannot match where the C parser
  reports an `ERROR` (gotreesitter issue #9), so `1  NAME x`, a tab in a
  name and `12:30z` were accepted; now each is an error at its column, and
  the "two spaces" case of `TestReadErrors` is no longer skipped.

## Acceptance

`ParseTime`:

| Input | Outcome |
|---|---|
| `12:30` | hour 12, minute 30 |
| `9:05` | hour 9, minute 5 |
| `12:30:05` | second 5 |
| `12:30:05.5Z` | fraction `5`, UTC |
| `23:59:59.999` | fraction `999` |
| `` , `24:00`, `12:60`, `1230`, `12:30z`, `12:30:05.` | error |

`ParseAge`:

| Input | Outcome |
|---|---|
| `` | the zero age |
| `35y` | years 35 |
| `> 35y 2m` | bound `>`, years 35, months 2 |
| `< 8w 3d` | bound `<`, weeks 8, days 3 |
| `1y 2m 3w 4d` | all four |
| `3d` | days 3 |
| `35`, `<35y`, `2m 35y`, `35Y`, `35y2m` | error |

`ParsePersonalName`:

| Input | Outcome |
|---|---|
| `John /Smith/` | given `John`, surname `Smith` |
| `John /Smith/ Jr.` | suffix `Jr.` |
| `/Smith/` | surname only |
| `John` | given only |
| `Lt. Cmndr. Joseph /Allen/ jr.` | given `Lt. Cmndr. Joseph`, surname `Allen`, suffix `jr.` |
| `José /García Márquez/` | the surname with its space |
| `John //` | given `John`, the surname empty |
| `` , `John /Smith`, `John /Smith/ /Jones/`, a name with a tab | error |

`ParseLatitude` and `ParseLongitude`:

| Input | Outcome |
|---|---|
| `N51.5` | 51.5 |
| `S33.86` | -33.86 |
| `N90`, `S0` | 90, 0 |
| `N90.1`, `N91`, `51.5`, `E51.5`, `n51.5`, `N51.` | error |
| `W1.5` | -1.5 |
| `E180`, `E179.99`, `W0.000001` | 180, 179.99, -0.000001 |
| `E180.1`, `E181`, `N1.5`, `1.5` | error |

Every value's `String()` is accepted by its parser and parses to an equal
value.

## Out of scope

- Typed accessors on `Structure` (`s.Date()`): callers parse the payload
  with the function for the tag. Which tag carries which type is the
  schema's knowledge, roadmap #2.
- Calendar arithmetic, time zones beyond the `Z` flag, name parts beyond
  the slashes (`NPFX`, `GIVN`, `NICK`, `SPFX`, `SURN`, `NSFX` are
  substructures, read by `Find`).
- Keeping the text as written: `0d` and `John //` read to values that write
  back as `0y` and `John`. The payload string on the `Structure` is the
  text; the typed value is its meaning.

## Decisions

| Choice | Taken | Alternative | Why |
|---|---|---|---|
| Absent components | Zero values: an age without months has `Months` 0, a name without a suffix `Suffix` "" | A presence flag or pointer per component | Matches `Date` (`Day` 0 is absent). The one meaning lost is an explicit zero, `0d` or `0y`, which writes back as `0y`; the payload keeps the text |
| Name parts | `Given`, `Surname`, `Suffix`: the text before, between and after the slashes, trimmed of the spaces that separate them | `Before`, `After`; the untrimmed text | The spec gives the outer parts no names; `Given` and `Suffix` are what the `GIVN` and `NSFX` substructures hold and how a reader says `John /Smith/ Jr.`. The text before the slashes may also hold a prefix (`Lt. Cmndr. Joseph`) |
| Coordinates | `float64` degrees, north and east positive | A hemisphere letter and unsigned degrees | The convention everything downstream uses; `String()` writes the letter back. `-0` and `0` both write `N0` |
| Fraction of a second | A string of the digits as written | A float or nanoseconds | `.5` and `.50` are different precisions; the digits are what the spec stores |
| `N90.1` | An error, checked in Go after the parse | Accept it, as the spec's ABNF does (`upto90` bounds the whole degrees, not the fraction) | The spec's prose bounds the value; the check sits where `ParseDate`'s day check sits |
| A skipped character | A gap check over the tree's spans, in the reader | Wait for the gotreesitter fix | Silently dropping a byte of a payload is a correctness fault in `Read`, not a cosmetic one; the check is cheap and stays right after the fix |

## Done when

- [x] tests express the acceptance table and pass
- [x] README status, changelog entry
- [ ] issue closed with a link to the release

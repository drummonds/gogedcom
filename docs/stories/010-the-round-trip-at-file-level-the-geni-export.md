# 010 — The round trip at file level: the Geni export through both writers, compared by story 006's rules (#32)

**Status:** done
**Roadmap:** #7 — the round trip of the Geni export

## Why

Story 009 showed the Geni export going 5.5.1 → 7 → 5.5.1 in memory and
coming back equal in every field but `Line`, and that the 5.5.1 writer's
file reads back to the same tree. Roadmap #7 asks for the trip the README
promises: the file read, converted, **written as 7, read again**,
converted back, written as 5.5.1 and read again, the first tree and the
last compared by the rules of
[story 006](006-readme-status-and-the-roadmap-the-geni-round-trip.html),
and any difference that remains listed in the README. Two things are
missing: the chain through the 7 file (the `@` escape, `CONT` splitting
and the 5.5.1 writer's file are each tested alone, not end to end), and
the comparison itself, which so far is `reflect.DeepEqual` with the line
numbers cleared, a yes or no with no list behind it. This story adds the
chain as one test and makes the 006 table executable: a comparison that
reports where two trees differ, so the README's list is the test's
output.

## The trip

![The round trip](../round-trip.svg)

| Stage | Does |
|---|---|
| `gedcom551.Read` | the file to the tree, `CONC` and `CONT` folded, BOM and CR LF gone |
| `convert.To7` | the 5.5.1 tree as a 7 tree (story 008) |
| `gedcom7.Write`, `gedcom7.Read` | through a 7 file: line breaks to `CONT` and back, a leading `@` escaped and stripped |
| `convert.To551` | the 7 tree as a 5.5.1 tree (story 009) |
| `gedcom551.Write`, `gedcom551.Read` | through a 5.5.1 file: `CONT` and back, no `CONC` splitting |
| the comparison | the first tree against the last, by the table below |

## The comparison

Story 006's acceptance table, as the test reads it. Two trees are
compared record by record in order, header first; within a record the
walk stops at the first difference and reports it with its path, so the
list has at most one line per record and a reader can see every record
that differs without one difference hiding the next.

| Difference between the first tree and the last | Reported |
|---|---|
| `HEAD.SOUR` (with its substructures), `HEAD.DATE` (with `TIME`), `HEAD.GEDC.VERS` | no: provenance, the writer's to set |
| `Line` | no: the reader's bookkeeping, not content |
| A payload | yes: `@I1@ INDI/BIRT/DATE: "1 JAN 1900" vs "1 JAN 1901"` |
| A tag or a cross-reference identifier | yes, the same way |
| A substructure more or fewer | yes: `@I1@ INDI/BIRT: 2 vs 3 substructures` |
| Two records swapped, a record more or fewer | yes: order and count are content |
| `HEAD.CHAR`, `HEAD.GEDC.FORM` | yes: `To551` writes them where 5.5.1 puts them, and the Geni file has them there |
| `CONC` split points, the BOM, CR LF | never seen: the readers fold and drop them before the trees exist |

## What changes

- `convert/roundtrip_test.go`: `differences(a, b)` returns the list
  above; `TestDifferences` pins each row with a small pair of trees;
  `TestRoundTripGeni` runs the trip over the file `GOGEDCOM_GENI` names
  and fails with the list if it is not empty.
- `convert/geni_test.go`: the in-memory test reports through
  `differences` too; its own first-difference walker goes.
- `README.md` Status: the trip holds at file level, no difference remains.
- `docs/round-trip.md`: the comparison stage built; the closing section
  says what the file shows; the `CHAR`/`GEDC.FORM` row says they are
  compared, not excluded.
- `index.md`: the story listed.

## Acceptance

The comparison, with `a` the tree of the 5.5.1 lines on the left and `b`
on the right (`0 HEAD`, `1 GEDC`, `2 VERS 5.5.1` and `0 TRLR` in both):

| `a` | `b` | `differences(a, b)` |
|---|---|---|
| `1 SOUR Geni.com` / `2 VERS 1.0` / `1 DATE 24 OCT 2026` / `2 TIME 10:00:00` | `1 SOUR gogedcom` / `1 DATE 5 NOV 2026` / `2 TIME 11:11:11` | none |
| `2 VERS 5.5.1` | `2 VERS 5.5.5` | none |
| `1 CHAR UTF-8` | `1 CHAR ANSEL` | `HEAD/CHAR: "UTF-8" vs "ANSEL"` |
| `0 @I1@ INDI` / `1 BIRT` / `2 DATE 1 JAN 1900` | the same with `2 DATE 1 JAN 1901` | `@I1@ INDI/BIRT/DATE: "1 JAN 1900" vs "1 JAN 1901"` |
| `0 @I1@ INDI` / `1 NAME A` | `0 @I2@ INDI` / `1 NAME A` | `@I1@ INDI: @I1@ vs @I2@` |
| `0 @I1@ INDI` / `1 BIRT` / `2 DATE 1900` / `2 PLAC X` | the same without `PLAC` | `@I1@ INDI/BIRT: 2 vs 1 substructures` |
| `0 @I1@ INDI`, `0 @I2@ INDI` | `0 @I2@ INDI`, `0 @I1@ INDI` | two lines, one per record |
| `0 @I1@ INDI`, `0 @I2@ INDI` | `0 @I1@ INDI` | `2 vs 1 records` |
| the same lines with different `Line` numbers | | none |

The Geni file, with `GOGEDCOM_GENI` set:

| Check | Expected |
|---|---|
| the trip runs with no error at any stage | yes |
| `differences(first, last)` | empty |
| README Status | says the trip holds and no difference remains |

## What building it taught

The trip passed first time, as the in-memory trip of story 009 said it
would: the 7 file's `@` escape and `CONT` lines and the 5.5.1 writer's
file add nothing the stage tests had not already covered. The comparison
earned its place anyway: the in-memory test's first-difference walker was
a yes or no that stopped at one line, and the list is what the README
would carry if a row ever moved.

## Out of scope

- A command to run the trip on another file: the project is a library;
  a reader runs the test with `GOGEDCOM_GENI` set.
- A public comparison API: the only consumer is the test, so the helper
  stays in the test (story 006 said the comparison is the measure, not a
  surface).
- A public fixture for CI (`royal92.ged`): a separate decision, as story
  006 said.
- Byte equality of the files: story 006 decided against it.
- Closing the roadmap issues #6, #22 and #7: `tp release` is meant to,
  and has not; the author's call.

## Decisions

| Choice | Taken | Alternative | Why |
|---|---|---|---|
| Where the comparison lives | a test helper in package `convert` | `gedcom.Diff` as API | one consumer; an API with no second caller is speculative |
| How much to report | one line per record, the first difference in it | a full edit script | bounded, readable in a README, and the next line is a different record so nothing is hidden for long; a second difference in one record shows once the first is fixed |
| What provenance means | story 006's four: `HEAD.SOUR`, `HEAD.DATE`, `HEAD.DATE.TIME`, `HEAD.GEDC.VERS` | `CHAR` and `GEDC.FORM` too, as `docs/round-trip.md` had said | they are content 5.5.1 requires, `To551` puts them where Geni has them, and comparing them costs nothing |
| One release | story and build together | a proposal release first, as 008 and 009 | the rules were fixed by 006; nothing here is a design choice a review could move |

## Done when

- [x] `differences` pins the table; the Geni trip passes at file level
- [x] README Status, `docs/round-trip.md`, `index.md`; changelog fragment
- [ ] issue closed with a link to the release

# 006 — README status and the roadmap: the Geni round trip as the goal (#21)

**Status:** in progress
**Roadmap:** #7 — the round trip of the Geni export

## Why

The README's status line is a list of story numbers, and the roadmap's
5.5.1 epic ends at "completeness measured on real exports" with no
statement of what complete means. Both leave a reader (and the author, at
the start of each 5.5.1 story) to infer the target. The target is one
file: the author's own tree exported from Geni.com as GEDCOM 5.5.1,
`H3export-BloodTree.ged`. The project is done with 5.5.1 when that file
goes 5.5.1 → 7 → 5.5.1 and comes back equal to the original. That goal
decides the open questions in stories #5, #6 and the writer back to 5.5.1:
every Geni structure must survive, extension tags and non-conforming
placements included, so the converter cannot drop what it does not
understand.

This story states the goal in the README and the roadmap and defines the
comparison, so the later stories inherit a fixed finish line.

## The file

| Fact | Value |
|---|---|
| Source | Geni.com export, `SOUR Geni.com`, `VERS 1.0`, 24 October 2026 |
| Form | GEDCOM 5.5.1 `LINEAGE-LINKED`, UTF-8 with a byte order mark, CRLF line ends |
| Size | 60,126 lines; 1,976 `INDI`, 722 `FAM`, 164 `NOTE`, 187 `SUBM` records |
| Tags | 48 standard tags; extensions `_MARNM` (1,848) and `_PRIM` (295) |
| Long text | `CONC` (1,499) and `CONT` (5,700) lines; a `CONC` payload may begin with a space |
| Off the schema | `ADDR` directly under `INDI`, `RFN geni:…`, `SUBM` under `INDI` |
| Where | `~/data/bytestone-corpus/raw/h3geni/`, outside the repo: it names living people |

## What changes

- `README.md` Status: what is built (GEDCOM 7 read, validate, write) and
  what the finish line for 5.5.1 is, in a sentence each; the Geni file
  described by the facts above, not vendored.
- `README.md` Vocabulary: "Round trip" gains the 5.5.1 sense, defined
  against the tree (below), beside the existing `Read(Write(doc))` sense.
- `ROADMAP.md`: the second epic's heading says the goal; #7's line is
  sharpened from "completeness measured" to the round trip and its
  comparison; a new item for the writer back to 5.5.1, the inverse of #6.
- `index.md`: the story listed.

## Acceptance

The round trip, as the roadmap and README state it, holds when the
original and the returned file compare equal by these rules:

| Difference between original and returned file | Counts as |
|---|---|
| Byte order mark, CRLF against LF | Equal: file form, not content |
| A long payload split across `CONC` lines at different points | Equal: compared after `CONC` and `CONT` are folded into the payload |
| Record order, substructure order | Different: order is content in GEDCOM |
| A tag, cross-reference identifier or payload changed | Different |
| An extension tag (`_MARNM`, `_PRIM`) dropped | Different |
| A structure placed where 5.5.1 does not allow it (`ADDR` under `INDI`) dropped or moved | Different: the file is the contract, not the 5.5.1 schema |
| A `DATE` or `NAME` payload re-spelt (`11 MAR 1961` → `11 Mar 1961`) | Different: the original wording comes back from `PHRASE` |
| `HEAD` fields that name the writer (`SOUR`, `DATE`, `TIME`, `GEDC VERS`) | Excluded from the comparison |

The README and ROADMAP after this story:

| Reader asks | Answer in |
|---|---|
| What works today | README Status, one sentence |
| When is 5.5.1 done | README Status and the roadmap epic heading: the Geni round trip, equal under the table above |
| Which roadmap item is the round trip | #7, its line saying so |
| How does 7 get back to 5.5.1 | A roadmap item for the inverse converter |

## Out of scope

- Doing any of the round trip: the 5.5.1 grammar (#5), the converter (#6),
  the writer back and the comparison are their own stories.
- Vendoring the Geni file or a redacted copy. The test for #7 will take
  the file's path from the environment and skip when it is absent; a
  public fixture (`royal92.ged`) may join later for CI, as a separate
  decision.
- A changelog entry beyond the docs fragment: no code changes.

## Decisions

| Choice | Taken | Alternative | Why |
|---|---|---|---|
| Where the goal lives | Sharpen #7 and the epic heading | A new epic "The round trip" | The round trip is what "completeness measured on real exports" meant; one item, not two for the same outcome |
| The writer back to 5.5.1 | Its own roadmap item | Folded into #6 or #7 | It is a surface (a 5.5.1 writer) with its own contract; #6 is the forward mapping, #7 the measure |
| What equal means | Trees compared after `CONC`/`CONT` folding, header provenance excluded, everything else byte for byte | Byte equality of the files | `CONC` split points and the BOM are the writer's choice, not content; nothing else is |
| Non-conforming structures | Must survive | Repair to the 5.5.1 schema on the way | The file is the thing to round-trip; a repaired file is a different file, and repair is a story a user asks for |

## Done when

- [x] README status, vocabulary; ROADMAP epic, #7 sharpened, inverse-converter item filed
- [x] index.md lists the story; changelog fragment
- [ ] issue closed with a link to the release

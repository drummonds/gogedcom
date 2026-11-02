# 005 — A writer: GEDCOM 7 out of the tree, read back to the same tree (#18)

**Status:** done locally
**Roadmap:** #3 — a writer: GEDCOM 7 out of the tree, and GEDZIP with the media

## Why

`Read` gives a tree of structures and `Validate` judges it, but nothing
puts a tree back on disk. The converter (roadmap #6) produces a tree and
has to write it; the completeness measure (#7) is convert, write, read
back, diff, so the writer is the half of that loop we do not have. The
contract is the loop itself: reading what `Write` wrote gives the tree
`Write` was given.

That contract exposes one thing the reader does today that a file-form
detail should not do: a text payload that begins with `@` is written `@@`
in the file (the spec's `lineStr`), and `Read` keeps the doubled form in
`Payload`. The tree should hold the value, so `Read` strips the escape and
`Write` adds it.

## Vocabulary

| Term | Meaning |
|---|---|
| Line | `level [xref] tag [payload]`, as the grammar's `Line`: single spaces, no trailing space when the payload is empty |
| Line break in a payload | A `\n` in `Payload`; in the file it is a `CONT` line one level deeper, its payload the text after the break (`1 CONT` with no payload for an empty remainder) |
| At-sign escape | A text payload whose value begins with `@` is written `@@…`; a pointer (`@I1@`, `@VOID@`) is written as it is |
| Round trip | `Read(Write(doc))` equal to `doc` in every field but `Line`, which is where each structure landed in the written file |

## What changes

- `gedcom7.Write(w io.Writer, doc *Document) error`: the header, the
  records in order, `0 TRLR`; LF line ends, no byte order mark; `CONT`
  lines for line breaks; the at-sign escape on text payloads.
- `Read` strips the at-sign escape: a file's `1 NOTE @@x` is `Payload`
  `@x`. A `CONT` line's payload is escaped the same way, both directions.
- `Write` refuses what no line can carry, naming the structure by its
  record and tag: a `\r` in a payload, a tag or cross-reference identifier
  the grammar's `Tag` and `Xref` rules reject, a document without a header.
  It does not validate against the registry; that is `Validate`.

## Acceptance

`Write`, one row per rule:

| Structure | Written as |
|---|---|
| Header `HEAD` with `GEDC` / `VERS 7.0` | `0 HEAD`, `1 GEDC`, `2 VERS 7.0` |
| Record `@I1@ INDI` with `NAME John /Smith/` | `0 @I1@ INDI`, `1 NAME John /Smith/` |
| Empty payload | `1 GEDC`, no trailing space |
| Payload `first\nsecond\n` on `@N1@ SNOTE` | `0 @N1@ SNOTE first`, `1 CONT second`, `1 CONT` |
| Payload `\nx` | `1 NOTE`, `2 CONT x` |
| Text payload `@x` | `1 NOTE @@x`; `1 CONT @@y` for a line break followed by `@y` |
| Pointer payload `@I1@`, `@VOID@` | written as it is |
| A document | ends `0 TRLR` and a line feed; no byte order mark |
| Every file under `testdata/gedcom70/` (the specification's, from GEDCOM.io) that `Read` accepts | `Read(Write(Read(file)))` equals `Read(file)` but for `Line`; `date.ged` and `extensions.ged` are known rejected (#19) until the reader is fixed |

`Write` errors:

| Condition | Error |
|---|---|
| `Payload` contains `\r` | `gedcom7: @N1@ SNOTE: carriage return in payload` |
| Tag `bad tag` or `""` | `gedcom7: @I1@ INDI: "bad tag" is not a tag` |
| XRef `@a@` (lowercase), `@VOID@`, or missing its at-signs | `gedcom7: @a@ INDI: "@a@" is not a cross-reference identifier` |
| `Header` nil | `gedcom7: no HEAD` |

`Read`:

| Line | `Payload` |
|---|---|
| `1 NOTE @@x` | `@x` |
| `1 NOTE x`, `2 CONT @@y` | `x\n@y` |
| `1 NOTE @I1@` | `@I1@` (a pointer is not escaped) |

## Out of scope

- GEDZIP: a zip of `gedcom.ged` and the media files `FILE` payloads name.
  The next story: it needs a source for the media (an `fs.FS`) and is a
  separate surface.
- Choosing line ends or a byte order mark: LF and none. The spec accepts
  both; a reader that needs CRLF can wrap the writer.
- Ordering, normalising or validating the tree: `Write` writes the tree it
  is given. Substructure order is the caller's; the registry's rules are
  `Validate`'s.
- Writing typed values: `DateValue` and the others already have `String()`;
  the tree holds payload text, so `Write` is strings in, lines out.

## Decisions

| Choice | Taken | Alternative | Why |
|---|---|---|---|
| What `Payload` holds | The value: `@x`, with `Read` stripping `@@` and `Write` adding it | The file form, `@@x`, both ways | A tree built in code (the converter) should not know the file's escape; the spec's grammar puts the doubling in `lineStr`, a file-level rule |
| Line ends and BOM | LF, none | Options on `Write` | No caller asks; the spec accepts both; speculative generality |
| What `Write` checks | What the grammar's `Line` cannot carry: CR, tag, xref, no header | Parse each written line with the grammar | Three regular checks cost nothing; a parse per line doubles the cost of the loop the writer exists for |
| Round-trip equality | Every field but `Line` | `Line` too, by `Write` recording where each structure landed | `Line` is where a structure was read from; writing moves it, and nothing downstream needs the written position |
| GEDZIP | Next story | In this one | One thing at a time; the zip's contract (media source, file name) is its own |

## Done when

- [x] tests express the acceptance tables and pass
- [x] README vocabulary and status; changelog fragment
- [ ] issue closed with a link to the release

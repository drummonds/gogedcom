# 003 — The schema from the registry YAML: structure types, cardinality and enumerations validated with the line (#10)

**Status:** done locally
**Roadmap:** #2 — the structure schema from the specification's YAML registry

## Why

`Read` accepts any well-formed tag anywhere: `1 BIRT` under a `SOUR` record,
three `SEX` lines, a `SEX` of `Q`, an `INDI` without the `GEDC` the header
must have. The grammar cannot say which structures go where; the
specification says it in its registry, `extracted-files/tags`, one YAML
file per URI-identified term: each structure type's standard tag, payload
type, substructures with their cardinality, and enumeration set. Every
library we looked at transcribes that too. The registry should be consumed
verbatim at the pinned commit, as the grammar is, and a file checked
against it with every violation reported at its line, because that is
what a converter (roadmap #6) and a loss list (#7) will be measured with.

## Vocabulary

| Term | Meaning |
|---|---|
| Registry | The specification's `extracted-files/tags`, vendored at the pinned commit as `gedcom7/registry/`: one YAML file per term |
| Term | Anything the registry gives a URI: a structure type, enumeration value, enumeration set, calendar, month or data type |
| Structure type | A term of type `structure` (`https://gedcom.io/terms/v7/record-INDI`): its standard tag, payload type, substructures and superstructures, each with a cardinality. One tag is several types: `DATE` is `DATE`, `DATE-exact` or `HEAD-DATE` by its superstructure |
| Cardinality | `{0:1}`, `{0:M}`, `{1:1}`, `{1:M}`: whether a substructure is required, and whether it may repeat |
| Document-level type | A structure type with no superstructures: `HEAD`, `TRLR` and the seven record types |
| Enumeration set | A term listing the terms a `type-Enum` payload may name; the payload text is the value's standard tag (`enum-M` is `M`; `enumset-EVEN` lists structure types, so `BIRT` is a value) |
| Violation | A structure that breaks the registry: its line and what it breaks |

## What changes

- `gedcom7/registry/`: the 321 term files, vendored verbatim by
  `task registry:fetch` at `GEDCOM7_COMMIT`, embedded in the package.
- `gedcom7`: `Standard()` is the registry read from those files:
  `Structure(uri)`, `Substructure(super, tag)` (the type a tag has under a
  superstructure, or at document level), `Enumeration(set)`. The types
  `StructureType`, `Cardinality`, `Violation`.
- `Structure.Line`: the file line a structure starts on, set by `Read`.
- `(*Registry).Validate(doc)`: every violation in the document, in line
  order, by the decision table below.

## Acceptance

The registry:

| Query | Answer |
|---|---|
| `Structure("…/record-INDI")` | tag `INDI`, no superstructures, `SEX` substructure `{0:1}`, `FAMC` `{0:M}` |
| `Substructure(nil, "INDI")` | `record-INDI` |
| `Substructure(nil, "HEAD")` | `HEAD` |
| `Substructure(nil, "BIRT")` | nil |
| `Substructure(INDI, "FAMC")` | `INDI-FAMC`; under `ADOP`, `ADOP-FAMC` |
| `Substructure(CHAN, "DATE")` | `DATE-exact`, payload `type-Date#exact` |
| `Structure("…/SEX").EnumerationSet` | `enumset-SEX` |
| `Enumeration("…/enumset-SEX")` | `M F X U` |
| `Enumeration("…/enumset-EVEN")` | includes `CENS` and `BIRT` |

`Validate`, one row per rule; the line reported is the structure's unless
said otherwise:

| Structure | Condition | Violation |
|---|---|---|
| Level 0, standard tag | not a document-level type (`0 BIRT`) | `line N: BIRT is not a record type` |
| Standard tag under type X | not a substructure of X (`1 BIRT` under `SOUR`) | `line N: BIRT is not a substructure of record-SOUR` |
| Extension tag (`_X`) anywhere | — | none, and its substructures are not checked (roadmap #4) |
| Substructures of an unknown structure | — | not checked: one violation, not a cascade |
| Type with cardinality `{0:1}` or `{1:1}` under X | a second occurrence | `line N: record-INDI has more than one SEX`, at the second |
| Type with cardinality `{1:1}` or `{1:M}` under X | absent | `line N: HEAD lacks GEDC`, at the superstructure's line |
| Payload of a `type-Enum` structure | not a value of its set (`SEX Q`) | `line N: Q is not in enumset-SEX` |
| Payload of a `type-Enum` structure | empty (`1 SEX`) | `line N: SEX has no payload` |
| Payload of a `type-List#Enum` structure | any comma-separated item not in the set | the same, per item |
| A file with none of these | — | no violations |

## Out of scope

- Payload types other than enumerations: `DATE-exact` holding a period, a
  pointer to a record of the wrong type, `nonNegativeInteger`, `Y|<NULL>`.
  The next story, with the typed values of 001 and 002.
- Extension tags and `SCHMA`: roadmap #4; they pass unchecked here.
- The grammar's month list regenerated from the calendar terms: a grammar
  story once the registry is in.
- Enforcing the violations: `Validate` reports, `Read` still reads.

## Decisions

| Choice | Taken | Alternative | Why |
|---|---|---|---|
| How the registry becomes Go | The YAML files embedded and read at first use | Go tables generated at build time from the TSVs, as the grammar is | Roadmap #4 defines extensions "in a YAML of the registry's shape", so a reader of that shape is needed anyway; one reader serves both. 321 small files read in milliseconds; generation stays open if that changes |
| The YAML parser | goccy/go-yaml, a YAML 1.2 parser with no dependencies of its own | yaml.v3 or its go.yaml.in successors, which the other projects use | The files declare `%YAML 1.2` and the yaml.v3 family refuses that directive outright; reading them verbatim beats stripping a line first, and the registry's values hold nothing 1.1 and 1.2 read differently either way |
| Which files | `tags/` only, verbatim, names as the spec gives them | Also the five TSVs | The README says the TSVs are look-ups derived from the YAML |
| Typing a structure | By superstructure type and tag, walking from the document level | By tag alone | `DATE` is three types and `FAMC` two; the registry keys substructures by URI |
| What an unknown structure costs | One violation; its subtree is skipped | Check the subtree against nothing | A cascade hides the cause |
| Where a missing required substructure is reported | At the superstructure's line | At the end of the superstructure | The line the fix goes under |

## Done when

- [x] tests express the acceptance tables and pass
- [x] README, changelog entry
- [ ] issue closed with a link to the release

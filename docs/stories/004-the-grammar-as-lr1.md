# 004 — The grammar as LR(1) (#14)

**Status:** done locally
**Roadmap:** —

## Why

`grammar/gedcom7.abnf` declared three GLR conflicts: `(day year)`,
`(calendar month)` and `(date)`. Each one is the single-space delimiter
sitting between a decision and the token that decides it, so the parser
needed two tokens of lookahead and tree-sitter forked the parse instead.
The fork is where gotreesitter has been wrong (story 001's EOL
workaround), and the pure-Go parser planned to replace it is LR(1) only.
A conflict-free grammar also means `tree-sitter generate` rejects any
ambiguity a later edit introduces.

## What changes

- `calendar` gets `@prec(1)`: a lone extension tag is a calendar, the
  reading the spec implies since extension months only occur under an
  extension calendar. Before, GLR's tie-break chose.
- `day` and `year` are `@alias(~day) integer` and `@alias(~year) integer`
  in `date`, so the parser shifts the integer and decides from the token
  after the space. The tree has `(day)` where it had `(day (integer))`;
  node text is unchanged.
- The delimiter is folded into `" TO"`, `" AND"` and `epoch`, so the lexer's
  longest match decides. The `epoch` node's text carries its leading space;
  `toDate` trims it.
- `; @conflicts` goes.

## Acceptance

`task grammar:test`: the corpus of story 001 with `(day)` and `(year)` as
leaves, plus `0 DATE 12 JAN` as `(date (year))` followed by `ERROR (month)`,
still an error. `task test`: the `ParseDate` table of story 001 unchanged,
`100 BCE` giving epoch `BCE`.

## Decisions

| Choice | Taken | Alternative | Why |
|---|---|---|---|
| Extension tag as calendar or month | `@prec(1)` on `calendar` | Split `date` into standard and extension-calendar forms | One annotation states the spec's rule; the split duplicates the date shape |
| Day and year | Aliases of `integer` | Keep `day = integer` and `year = integer` | Two nonterminals over one token force a reduce on the space; an alias names the position without a reduce |
| Two-token lookahead through a space | Fold the space into the following token | Make the space an extra | Whitespace is significant everywhere else in a GEDCOM line; a local fold keeps that |
| Error tree for `12 JAN` | Accept the new shape | Keep the old `ERROR (day month)` | Recovery shape is the parser's, not the grammar's; both report an error |

## Done when

- [x] corpus and Go tests express the acceptance and pass
- [x] changelog entry
- [ ] issue closed with a link to the release

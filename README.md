# ctworldgen

A Go CLI that generates rules-accurate Classic Traveller subsectors per
Book 3's Worlds chapter (© 1977 text): the eighty-hex world occurrence
scan, starport types and their bases, commercial space lanes, and the eight
characteristics of every world found. Sibling to
[ctchargen](https://github.com/philoserf/ctchargen).

[`docs/PRD.md`](docs/PRD.md) is the v1 contract; all four milestones are
implemented. [`docs/COVERAGE.md`](docs/COVERAGE.md) maps every rule of
Book 3 pp. 1–12 to its page cite, implementation, and test;
[`docs/ERRATA.md`](docs/ERRATA.md) records every interpretation.

## Usage

```sh
ctworldgen new --seed 42 --name Vega        # one subsector, JSON to stdout
ctworldgen new --occurrence-dm -1 -o rim.json   # a sparse region of the galaxy (p. 1)
ctworldgen batch --count 16 -o sector/      # a sector's worth of subsectors
ctworldgen render subsector.json            # Markdown subsector listing
ctworldgen render --history subsector.json  # the full generation transcript
ctworldgen replay subsector.json            # verify a record reproduces exactly
ctworldgen version
```

Every record carries its seed, versions, inputs, and the complete event
log — every throw of every hex; `replay` re-runs the engine from the seed
and exits non-zero at the first mismatch. The one thing it does not verify
is the name you give a world: the book asks you to name each one (p. 12)
and prints no table, so names are yours to write in afterwards and replay
leaves them alone.

## What a world looks like

Characteristics are written as the string of digits of p. 4 — starport,
size, atmosphere, hydrographics, population, government, law level,
technological index, in the order the book's own Planetary Characteristics
box lists them:

```
| Hex  | Name | Profile  | Bases        |
| 0306 |      | B8579987 | naval, scout |
```

Eight characters, no separator. The hyphen most people remember before the
technological index is not in the held 1977 text, so it is not here either;
[`docs/ERRATA.md`](docs/ERRATA.md) E005 sets out the reading in full, and
`render` prints every characteristic in labelled form beside it.

## Differences from ctchargen

Character generation is a procedure of decisions, so ctchargen has a
`Decider` interface, an auto policy, an interactive mode, and a
`POLICY.md`. World generation has no choice points at all — walk pp. 1–12
and every step is a throw and a table lookup. The referee's latitude is
exercised before the run, not during it, so it lives in the record's
`inputs`. There is therefore no `--auto` flag, no interactive mode, and no
`policy_version`. See docs/PRD.md, "The architectural delta from
ctchargen".

## Development

```sh
task deps   # install the toolchain (brew bundle + NilAway at tip)
task hooks  # install the pre-push gate
task        # check (modernize + fmt + vet + lint + nilaway) + test
```

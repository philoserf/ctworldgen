# ctworldgen

A Go CLI that generates rules-accurate Classic Traveller subsectors and
sectors from Book 3 _Worlds and Adventures_, the Worlds chapter (pp. 1-12,
© 1977 text). It writes a JSON record from a recorded seed and renders it
as the Markdown listing a referee runs from, or as a printable PDF booklet
with the hex map drawn and the routes joined.

## Install

```sh
go install github.com/philoserf/ctworldgen/cmd/ctworldgen@latest
```

**Until `v1.0.0-alpha.2` is tagged, use `@main` instead.** The newest tag is
still `v1.0.0-alpha.1`, which predates `sector`, the drawn map and
`--lanes`, so `@latest` installs a tool this README does not describe. This
paragraph goes away with the tag; the line above it does not change.

## What is built, and what is not

The engine walks the whole of Book 3 pp. 1-12 -- the eighty-hex occurrence
scan, starport types, naval and scout bases, commercial routes, and the
eight characteristics of every world -- all from one seed. `render` turns a
record into the listing, opening with a text map of the p. 3 hex grid, or
into the booklet. `sector` lays sixteen subsectors on one 32x40 grid and
throws for the routes at their seams.

Two things are not built, and both are open:

- **The technological levels tables of pp. 10-11.** Every world carries a
  technological index and no gloss of it. The listing says why the line is
  bare rather than leaving it silently so
  ([#4](https://github.com/philoserf/ctworldgen/issues/4)).
- **Per-area occurrence DMs** -- a rift in one corner, a cluster in the
  other ([#5](https://github.com/philoserf/ctworldgen/issues/5)).

`docs/COVERAGE.md` is the live map of rule to code to test, and it carries a
row for what is not built as well as for what is. Ask it whether a rule is
in here; this section only says what the shape of the thing is.

What the tool should do comes from
[issue 1](https://github.com/philoserf/ctworldgen/issues/1), the first
alpha's report: it was played as a referee rather than as a developer, and
what that found is the backlog. A second alpha is close, and it will be
read the same way. `docs/PRD.md` was the contract through
`v1.0.0-alpha.1` and governs nothing now -- read it for why a thing is the
way it is, never for whether a thing may be built.

## Use

```sh
ctworldgen new    [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
ctworldgen sector [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
ctworldgen render [--format markdown|pdf] [--lanes legible|all] [-o file] [--force] record.json
ctworldgen version
```

```sh
$ ctworldgen new --seed 1977 --name Aramis --occurrence-dm -1
```

`sector` writes one record covering sixteen subsectors on a single 32x40
grid, 0101 through 3240, and throws for the commercial routes that cross
between them -- the routes sixteen independent subsectors can never have.
Every member is unchanged: member _i_ of `sector --seed N` is exactly the
subsector `new --seed N+i` writes, so the subsector you read is the
subsector you get. The seams are one further reading (`ERRATA.md` E006),
stamped on every sector record.

`render` reads either record, a subsector or a sector, and writes the
Markdown listing by default. `--format pdf` writes the booklet instead: the
map beside its roster on the first page, then the routes and a page of
detail per world. It is the one output that draws p. 2's "line connecting
the two worlds on the map", which a monospace grid has nowhere to put. A
booklet is a binary, so `--format pdf` needs `-o`, and it reproduces byte
for byte from the same record.

`--occurrence-dm` takes -1, 0 or +1 and nothing else, and defaults to 0.
Without `--seed`, a seed is drawn from OS entropy and written into the
record, so a run is reproducible after the fact; `--seed 0` is therefore an
explicit and distinct choice rather than a request for a random one.
Existing files are never overwritten without `--force`, and flags precede
any filename.

`docs/examples/complete.json` is a full record with every field populated;
`task regenerate` rewrites it, so it never drifts from what the tool
writes. `docs/examples/minimal.json` is the smallest record the schema
admits, and is held by hand.

### What a seed fixes, and what it does not

A seed and the inputs reproduce a subsector exactly, and only exactly. All
the dice come from one stream in procedure order, so anything that changes
how many throws are made before a given world changes that world.

**Changing `--occurrence-dm` regenerates the subsector; it does not thin
the one you have.** The occurrence scan throws one die per hex over all
eighty before anything else, so the same seed gives the same eighty faces
and the star fields nest: every hex placed at -1 is placed at 0, and every
hex placed at 0 is placed at +1. The map looks like a dial. But each extra
world consumes dice for its starport, its bases and its characteristics,
so the hexes two runs share keep almost nothing else — different starports,
different populations, a different string of digits. "The same subsector, a
touch sparser" is not available: change the seed for another subsector, and
hold the DM to keep this one.

**Separate subsector files are a loop over `new`.** `sector` is the answer
when the sixteen belong on one grid, because it throws for the routes at
their seams. When you want them as independent records instead, their
seeds are simply consecutive:

```sh
for i in $(seq 0 15); do
  ctworldgen new --seed $((1977 + i)) --name Aramis -o "aramis-$i.json"
done
```

**Sector seeds closer together than sixteen share subsectors.** A sector's
members run on `seed` through `seed + 15`, so `sector --seed 100` and
`sector --seed 110` have six subsectors in common -- the same worlds with
the same digits, sitting in different corners of the two maps. Nothing
warns about it, because each sector is individually correct. Leave a gap
of at least sixteen when generating a second sector to set beside the
first.

### Which lanes are drawn

A dense subsector throws a hundred and sixty commercial routes over
forty-six worlds, and a map with all of them on it cannot be read. P. 2
offers the map-drawer a way out in the book's own voice -- a connection
already present "may be ignored" -- and `render` takes it by default: a
lane whose two worlds are already joined by shorter lanes is not drawn and
not listed (`ERRATA.md` E007).

It removes about 46% of the lanes and changes the reachability of nothing,
because a lane is only dropped when its ends are already joined. The
summary line says how many were drawn, the route section says how many were
not, and `--lanes all` draws every one.

**The record is unchanged and carries every lane.** This is a decision
about ink, not about dice: the engine still examines every pair and
consumes every die (E003), so no seed's meaning moves.

### Writing in the record

The record is the referee's notebook page, so there is a place to write in
it. `notes` on a world, and `notes` on the record as a whole, belong to the
referee: the tool never generates either and never reads them back. Both
survive re-rendering -- a world's note becomes a line in its detail block
and the record's becomes a paragraph under the heading, in the listing and
in the booklet alike.

```json
{
  "name": "Tessarane",
  "notes": "The rift campaign. Players start at 0602.",
  "worlds": [{ "hex": "0602", "name": "Reagan", "notes": "dust storms" }]
}
```

Nothing else is admitted. Every other key the record does not define is
still refused, which is what makes the generated fields worth trusting;
`notes` is a field the record names rather than a hole in that rule.

## Where the rules come from

Rules come only from the held PDFs of the FFE reprints, never from memory.
Training-data Traveller is mostly the 1981 revision and later editions, and
the held © 1977 page governs even where it differs -- most visibly in the
string of digits, which carries no hyphen (`A867A698`, not `A867A69-8`).

Book 3 pp. 1-12 is the ruleset. Book 1 pp. 2-3 and p. 8 supply the die roll
conventions and the hexadecimal notation that Book 3 uses without
restating. Nothing else is in authority: not Books 2 and 4+, not the
supplements, the Starter Edition, The Traveller Book, JTAS, the
_Consolidated Errata_, or the rest of Book 3.

Two habits follow, and they are the ones worth knowing from outside:

- **Every table is transcribed from a visual read of the page**, then
  transcribed a second time inside the table package's tests, so the two
  must agree. The held PDFs' embedded font maps the em-dash to the glyph
  `4` and the minus sign to `3`, so a text extraction renders the jump
  routes table's empty cells as the digit 4 and the size formula `2D − 2`
  as `2D32`. Both readings are wrong and both look like data.
- **No ambiguity is resolved in silence.** Where the page does not settle
  something, the reading goes in `docs/ERRATA.md` with its page cite and
  the condition under which a record stamps it, and every record carries
  the stamps that applied to it.

## The documents

`docs/ERRATA.md` records every reading of an ambiguous or silent page.
`docs/COVERAGE.md` maps the rules to the implementation and the tests.
`docs/record.schema.json` is the record's schema, with a minimal and a
complete example beside it. `CLAUDE.md` carries the authority model in
full, and the traps -- the font, the hex grid parity, the dice-stream
consumption order -- that a change to this code has to respect.
`docs/PRD.md` is historical; see above.

## Development

```sh
task              # the whole gate: tidy, vet, lint (formatting included), nilaway, test -race, coverage ratchet
task regenerate   # rewrite the golden fixtures and the complete example, then read the diff
```

CI runs exactly `task`. The toolchain is deliberately unpinned: the gate is
meant to fail when a tool moves rather than drift behind it, so a red gate
on code you did not touch is the signal working. Answer the finding; do not
pin a tool or add a linter disable to silence it.

`main` is protected -- pull request, a green gate, linear history, no
bypass -- so work lands on a branch and merges through CI.

## Licence

MIT. See `LICENSE`. Traveller is © Far Future Enterprises; this tool
implements the rules and reproduces none of the text beyond the table
labels a listing needs.

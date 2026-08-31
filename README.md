# ctworldgen

A Go CLI that generates rules-accurate Classic Traveller subsectors from
Book 3 _Worlds and Adventures_, the Worlds chapter (pp. 1-12, © 1977
text). It writes a JSON record and renders it as a Markdown listing a
referee can run from.

## Status

All four milestones are complete. The engine walks the whole of Book 3
pp. 1-12 -- the eighty-hex occurrence scan, starport types, naval and
scout bases, commercial space lanes, and the eight characteristics of
every world -- all from a recorded seed. `render` turns a record into the
Markdown listing a referee can run from, and `batch` produces a sector's
worth of subsectors at a time.

`docs/COVERAGE.md` is the current map of rule to code to test.

## Use

```sh
ctworldgen new   [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
ctworldgen batch --count N [--seed N] [--name X] [--occurrence-dm N] [-o dir|file.jsonl] [--force]
ctworldgen render [-o file] [--force] subsector.json
ctworldgen version
```

`batch` emits JSONL to stdout; given a directory it writes one file per
subsector, named for the subsector and its index -- `aramis-00.json`.
Members are numbered from zero, so `batch --count 1 --seed N` produces
exactly the subsector `new --seed N` produces.

`--occurrence-dm` takes -1, 0 or +1 and nothing else, and defaults to 0.
Without `--seed`, a seed is drawn from OS entropy and written into the
record, so a run is reproducible after the fact; `--seed 0` is therefore an
explicit and distinct choice rather than a request for a random one.
Existing files are never overwritten without `--force`, and flags precede
any filename.

```sh
$ go run ./cmd/ctworldgen new --seed 1977 --name Aramis --occurrence-dm -1
```

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

**`batch` member seeds are consecutive.** Member _i_ runs on `seed + i`,
which is what makes `batch --count 1 --seed N` produce exactly `new --seed
N`. It also means two batches whose base seeds sit closer together than
their counts share members: `--seed 100 --count 16` and `--seed 110 --count
16` have six subsectors in common.

## The documents

`docs/PRD.md` is the contract: the authority model, the requirements, the
determinism rules, the record shape and the milestones. `docs/ERRATA.md`
records every reading of an ambiguous or silent page, with its page cite
and the condition under which a record stamps it. `docs/COVERAGE.md` maps
the rules to the implementation. `docs/subsector.schema.json` is the
record's schema, with a minimal and a complete example beside it.

Rules come only from the held PDFs of the FFE reprints. Training-data
Traveller is mostly the 1981 revision and later editions, and the held page
governs even where it differs — most visibly in the string of digits, which
carries no hyphen (`A867A698`, not `A867A69-8`).

## Development

```sh
task              # the whole gate: tidy, vet, lint (formatting included), nilaway, test -race, coverage ratchet
task regenerate   # rewrite the golden fixtures and the complete example, then read the diff
```

CI runs exactly `task`. The toolchain is deliberately unpinned: the gate is
meant to fail when a tool moves rather than drift behind it, so a red gate
on code you did not touch is the signal working. Answer the finding; do not
pin a tool or add a linter disable to silence it.

## Licence

MIT. See `LICENSE`. Traveller is © Far Future Enterprises; this tool
implements the rules and reproduces none of the text beyond the table
labels a listing needs.

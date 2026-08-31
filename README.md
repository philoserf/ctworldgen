# ctworldgen

A Go CLI that generates rules-accurate Classic Traveller subsectors from
Book 3 _Worlds and Adventures_, the Worlds chapter (pp. 1-12, © 1977
text). It writes a JSON record and renders it as a Markdown listing a
referee can run from.

## Status

Milestone 1: the domain and its edges. The eighty-hex world occurrence
scan and starport types generate, deterministically, from a recorded seed.
Bases and space lanes are milestone 2; the eight characteristics of each
world are milestone 3; the listing and `batch` are milestone 4.

`docs/COVERAGE.md` is the current map of rule to code to test.

## Use

```sh
ctworldgen new [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
ctworldgen version
```

`--occurrence-dm` takes -1, 0 or +1 and nothing else, and defaults to 0.
Without `--seed`, a seed is drawn from OS entropy and written into the
record, so a run is reproducible after the fact; `--seed 0` is therefore an
explicit and distinct choice rather than a request for a random one.
Existing files are never overwritten without `--force`, and flags precede
any filename.

```sh
$ go run ./cmd/ctworldgen new --seed 1977 --name Aramis --occurrence-dm -1
```

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
task              # the whole gate: tidy, fmt, vet, lint, nilaway, test -race, coverage ratchet
task regenerate   # rewrite the golden fixtures, then read the diff
```

CI runs exactly `task`. The toolchain is deliberately unpinned: the gate is
meant to fail when a tool moves rather than drift behind it, so a red gate
on code you did not touch is the signal working. Answer the finding; do not
pin a tool or add a linter disable to silence it.

## Licence

MIT. See `LICENSE`. Traveller is © Far Future Enterprises; this tool
implements the rules and reproduces none of the text beyond the table
labels a listing needs.

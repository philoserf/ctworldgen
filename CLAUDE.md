# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`ctworldgen`: a Go CLI that generates rules-accurate Classic Traveller
subsectors from Book 3's Worlds chapter (pp. 1–12, © 1977 text), sibling to
`philoserf/ctchargen`.

**Status: all four PRD milestones complete** — the eighty-hex occurrence
scan, starport types and bases, space lanes, the seven basic
characteristics and the technological index, `batch`, `render`, and
`replay` verification.
`docs/PRD.md` is the v1 contract — read it before doing any work here.
`docs/COVERAGE.md` maps rules to implementation; `docs/ERRATA.md` holds the
recorded readings. Golden fixtures move only via `task goldens`, never by
hand.

## Commands

```sh
task          # the full gate: check (modernize + gofumpt + prettier + vet + golangci-lint + nilaway) + test
task fmt      # format Go (gofumpt -extra) and JSON/Markdown (prettier)
task test     # go test -race ./... with the per-package uncovered-statement ceilings
task goldens  # regenerate the golden fixtures, then run the gate
task deps     # install the toolchain (brew bundle + NilAway at tip)
task hooks    # install the tracked pre-push hook (runs `task`)
go test -race ./worldgen -run TestGolden   # a single test
```

CI (`.github/workflows/ci.yml`) runs exactly `task` — never add checks to CI
that the local `task` gate doesn't run, and never add a tool to the gate
without also installing it there. golangci-lint runs with `default: all`
and a curated disable list (`.golangci.yml`); fix findings rather than
adding disables.

**The toolchain is deliberately unpinned**, in the Brewfile and in CI
alike: the gate is meant to fail when a tool moves rather than drift behind
it. So a red gate on code you did not touch is expected occasionally and is
the signal working. Answer the finding, or adjust `.golangci.yml`; do not
pin a tool to make it go away.

The `test` task ratchets on **uncovered statements per package**, not on a
coverage percentage — a percentage can stay put while a guarded branch adds
one covered statement and one uncovered. It fails in both directions: raise
a ceiling only with a reason worth writing in the commit, and lower it
whenever coverage improves.

## Authority model — the most important rule

Rules come **only** from the held PDFs in `~/Documents/Traveller/Classic/`
(FFE reprints of the © 1977 text): Book 3 _Worlds and Adventures_ pp. 1–12,
plus Book 1 _Characters and Combat_ pp. 2–3 (die roll conventions) and p. 8
(hexadecimal digit notation), which Book 3 uses without restating.

- **Never implement a Traveller rule from memory.** Training-data Traveller
  is mostly the 1981 revision and later editions. The held page governs
  even where it differs — the string of digits has no hyphen before the
  technological index (E005), and a population-0 world still gets one.
- Every implemented rule carries a printed-page cite. Where the text is
  ambiguous or silent, the chosen reading goes in `docs/ERRATA.md` with the
  page cite and is stamped in every record it governed — never applied
  silently.
- Page N as printed is PDF page N+5 in Book 3, N+6 in Book 1.
- Everything else in that collection (Consolidated Errata, Starter Edition,
  Books 2 and 4+, the rest of Book 3, …) is out of authority.

### Read the tables off the page, never out of `pdftotext`

The held PDFs' embedded font maps the em-dash to the glyph `4` and the
minus sign to `3`. A text extraction therefore renders the jump routes
table's empty cells as the digit 4 — `B-E 4 4 4 4` for `4 — — —` — and the
size formula `2D − 2` as `2D32`. Both readings are wrong and both look like
data. Every table in `tables/data/` was transcribed from a **visual** read
of the page (`Read` the PDF with a `pages` range), and each is transcribed
a second time inside `tables/tables_test.go` so the two must agree. Any
future table work does the same.

## Clean room

Do **not** read, import from, or copy the sibling repos
`philoserf/ctchargen`, `philoserf/t5chargen`, `philoserf/t5`, or
`philoserf/traveller` unless the user explicitly asks. The _contracts_
ctchargen proved (replay/provenance, the event log, the COVERAGE and ERRATA
documents, the unpinned toolchain gate) are adopted; its code is not.

## Architecture

- Packages: `dice`, `worldgen` (the record, the event log, the engine, the
  hex grid, replay), `tables` (data-driven Book 3 charts), `render`,
  `cmd/ctworldgen`. Plus `internal/fixture`, test support rather than
  architecture: the one definition of the golden-fixture roster, so the
  `worldgen` and `render` golden trees cannot come to describe different
  subsectors under the same name.
- **There is no `Decider`, no auto policy, no interactive mode, and no
  `policy_version`.** That is a deliberate delta from the sibling, not a
  gap: pp. 1–12 have no choice points, and the referee's latitude is an
  input rather than a decision. Adding a choice point later means adding
  all of that back. See docs/PRD.md, "The architectural delta from
  ctchargen".
- Tables/thresholds/labels are embedded data files (`go:embed`) with
  load-time validation; procedural mechanics are typed Go. No rules
  language.
- The JSON subsector record is the source of truth; Markdown listings are
  renders of it. Every record carries full provenance (seed, versions,
  inputs, event log) and must replay byte-identically — dice-stream
  consumption order is load-bearing, so procedure-order changes are
  replay-breaking and bump `EngineVersion`. So does editing the text of any
  event, because `Replay` compares whole `Event` values.
- Two throws are deliberately **not** made: a size-0 world's atmosphere and
  a size-0-or-1 world's hydrographics are automatic (p. 4). Rolling and
  discarding a die there would shift every later world.
- Testing: golden fixtures at each occurrence DM, hex-distance cases
  measured by hand off the printed p. 3 grid, replay round-trips, schema
  validation, and property sweeps over many seeds asserting the book's
  invariants.

### The hex grid is the quiet trap

`worldgen/hex.go`'s offset-to-cube conversion has to match the parity the
p. 3 grid prints (even-numbered columns sit half a hex lower). Getting it
backwards leaves every distance internally consistent and wrong by one for
half the map — which **replay cannot catch**, because replay compares a
record against a record. `TestDistanceMatchesThePrintedGrid` measures
against the page instead. Never change that conversion without re-measuring
it there.

## Project documents

All project documents live in `docs/`:

- `docs/PRD.md` — the contract; milestones live there.
- `docs/ERRATA.md`, `docs/COVERAGE.md`, `docs/subsector.schema.json` with a
  minimal and a complete example beside it.

Unlike the sibling's, those two examples are hand-written rather than
engine output and so are not replay-verified: a real record runs past a
third of a megabyte. The golden fixtures under `worldgen/testdata` are the
engine output that carries that guarantee.

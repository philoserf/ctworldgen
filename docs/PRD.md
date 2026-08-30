# PRD: Classic Traveller World Generator (Go CLI)

2026-08-30. Status: draft — the v1 contract, before any code.

## Problem

Classic Traveller world generation is the referee's mapping chore: eighty
hexes checked one at a time for a star system, a starport graded for each
one found, commercial space lanes drawn between them, and then eight
characteristics rolled for every world — size, atmosphere, hydrographics,
population, government, law level, and a technological index derived from
all of the above. A subsector runs to forty worlds. Doing it by hand is
slow and error-prone. Build a Go CLI that generates rules-accurate Classic
Traveller subsectors, as a sibling to `philoserf/ctchargen`.

Ruleset baseline: **Book 3 pp. 1–12 only** — Book 3 _Worlds and
Adventures_, as held in `~/Documents/Traveller/Classic/` (the FFE reprints
of the © 1977 text). Book 1 _Characters and Combat_ is consulted only for
the die roll conventions (pp. 2–3) and the hexadecimal digit notation
(p. 8), which Book 3 uses without restating. All page cites are to those
artifacts and nothing else.

## User

Mark: solo referee and developer. Secondary: any Classic Traveller referee
who needs a subsector to run in, or a sector's worth of them.

## Goals

1. Generate a complete subsector per Book 3's Worlds chapter (pp. 1–12):
   the eighty-hex world occurrence scan, starport types and their bases,
   space lanes, and the eight characteristics of every world found.
2. Deterministic replay: re-running the engine from the recorded seed and
   inputs reproduces the identical subsector (see Replay and provenance
   contract).
3. Output a subsector record as JSON (canonical) and a Markdown subsector
   listing in the style the book's own records call for — the world's
   name, its hex, and its characteristics as a string of digits (p. 4).
4. Emit a generation record: the full chronological history — every throw
   and outcome — embedded in the JSON and renderable as a Markdown
   transcript.

## Non-goals (v1)

- Anything outside Book 3 pp. 1–12. Equipment (pp. 13–18), encounters
  (pp. 19–23), animal encounters (pp. 24–32), and psionics (pp. 33–42) are
  in-play material, not world generation.
- The technological level tables (pp. 10–11). They say what a technological
  index _means_ — what weapons, computers, and drives a world can build —
  which is referee reference during play, not a step of generation. Only
  the index's range is read from them. This is the same line ctchargen
  draws at Book 1's skill descriptions.
- Books 1, 2, and everything else in the collection: Books 4+, the
  supplements, the Starter Edition, The Traveller Book, JTAS, and the
  _Consolidated Errata_ PDF — all out of authority. Book 1 pp. 2–3 and
  p. 8 are in, and only because Book 3's procedure is written on top of
  them.
- Later printings' rules. The held text is the © 1977 text; where later
  printings famously differ, the held page governs. The two that bite
  here: the held technological index matrix gives a population-0 world no
  DM and therefore a technological index like any other world's, and the
  held text has no hyphen in the string of digits (see ERRATA E005).
- **Stellar detail.** The book generates "the single inhabited world in a
  star system" (p. 2) and leaves the star, the other planets, satellites,
  and gas giants to the referee. So does this tool.
- **Single-world generation.** The unit of generation is the subsector,
  because two of the three star-mapping steps are subsector-scoped: a
  starport distribution is drawn for a whole subsector, and a space lane
  needs a neighbour. A lone world has no starport throw of its own and no
  routes to determine.
- **World names.** P. 12 step 3 says to name each world and prints no
  table; the book's naming material is advice, not a generator. Records
  carry an empty `name` for every world and the hex identifies it.
  `--name` names the subsector.
- Graphical hex maps. The record carries hexes and routes; drawing them is
  a downstream concern.

## Functional requirements

Rule citations are printed page numbers of Book 3 unless marked B1.

**FR1 — The subsector grid.** Eighty hexes, eight columns of ten rows,
identified by the four-digit column-and-row number the p. 3 grid prints
(0101 through 0810). One hex is one parsec (p. 1). Even-numbered columns
sit half a hex lower than odd-numbered ones, as printed; distance between
two hexes is the hex-grid distance that layout gives.

**FR2 — World occurrence.** Check each hex in turn, throwing one die: 4, 5,
or 6 places a world and its attendant star system (p. 1). The referee may
impose a DM of +1 or −1 on the whole subsector to make worlds more or less
frequent; that DM is an input, not a decision the tool makes. The scan
order is unstated and stream-load-bearing (ERRATA E002).

**FR3 — Starport type.** Two dice for each world, read against the
starports table: 2–4 A, 5–6 B, 7–8 C, 9 D, 10–11 E, 12 X (p. 1).

**FR4 — Bases.** The starport chart prints the throws the p. 12 checklist
omits (p. 5): naval base 8+ at starport A and B; scout base 10+ at A, 9+ at
B, 8+ at C, 7+ at D. E and X have neither. Two dice each, per B1 pp. 2–3.
Their position in the dice stream is a reading (ERRATA E001).

**FR5 — Space lanes.** For every pair of worlds four or fewer hexes apart,
throw one die against the jump routes table cell for the pair's starport
types and their distance; equal or greater means a lane exists (p. 2). The
table has rows for A through E only, and cells marked with a dash at which
no lane is possible. Each pair is examined once. Enumeration order and the
treatment of X starports are readings (ERRATA E003).

**FR6 — Planetary size.** 2D − 2, range 0 to 10; zero is an
asteroid/planetoid complex (pp. 4, 12).

**FR7 — Planetary atmosphere.** 2D − 7 + planetary size. A planet of size
zero automatically has an atmosphere of zero (pp. 4, 12).

**FR8 — Hydrographic percentage.** 2D − 7 + planetary size, with a further
DM of −4 if the atmosphere is 0, 1, or greater than 9. A planetary size of
0 or 1 gives an automatic 0 (pp. 4, 12).

**FR9 — Population.** 2D − 2, range 0 to 10, an exponent of 10 (pp. 8, 12).

**FR10 — Planetary government.** 2D − 7 + the population digit (pp. 8, 12).

**FR11 — Law level.** 2D − 7 + the government type (pp. 8, 12).

**FR12 — Technological index.** One die, modified by the sum of the DMs the
technological index matrix gives for the world's starport, size,
atmosphere, hydrographics, population, and government values (p. 9). The
index the book describes runs zero to 18.

**FR13 — Value ranges.** Every generated value is clamped to the printed
range of its own table; the arithmetic of FR6–FR12 can leave it. Law level
is the exception and is not capped above. The reading, and the warrant
p. 8 gives for it, are in ERRATA E004. The technological index matrix
depends on the clamp: it has a row for every value in a table's printed
range and none for a value outside it.

**FR14 — The string of digits.** Each world's characteristics are recorded
as a string, "in much the same manner as the Universal Personality Profile
is used for the easy identification of persons" (p. 4), in the order the
p. 4 Planetary Characteristics box lists them. The digit convention and the
absence of a separator are a reading (ERRATA E005).

**FR15 — Subsector record.** Track the subsector's name, the occurrence DM
it was generated under, every world (hex, name, starport, bases, the seven
basic characteristics, the technological index, and the string of digits),
and every space lane (the two hexes and the jump distance). The JSON record
is the source of truth; the Markdown listing is a render of it.

**FR16 — Dice engine.** One- and two-die throws with DMs against `N+`/`N−`/
exact targets per the die roll conventions (B1 pp. 2–3). Book 3 uses
one-die targets the character procedure never does — world occurrence is
4+ on one die, and the jump routes table's cells are one-die targets from 1
to 6 — so the target notation admits values below 2. All rolls are consumed
from a seeded stream and logged for replay.

**FR17 — Generation record.** An ordered event log of the entire
generation: each procedure step entered, each throw (dice, target, DMs,
result), and each consequence (world placed, starport graded, lane drawn,
characteristic set). Stored as an `events` array in the subsector JSON with
monotonic sequence numbers; consequence events reference the causing
throw. Serves audit (verify any subsector against Book 3 by walking the
log), replay (verification data for goal 2), and narrative.

## The architectural delta from ctchargen: no choice points

Character generation is a procedure of decisions — which service, submit to
the draft, which skills table, take the title. Its engine is built around a
`Decider` interface so that interactive play, the auto policy, and replay
share one procedure, and a `POLICY.md` documents what the policy decides.

**World generation has none of that.** Walk pp. 1–12 and every step is a
throw and a table lookup; the referee's latitude is exercised before the
run, not during it — the occurrence DM (p. 1), a house starports table
(p. 1), a deliberately imposed world (p. 8), a name (p. 12). Those are
_inputs_, and inputs belong in the `inputs` block where replay already
carries them.

So, deliberately and by contrast with the sibling:

- No `Decider` interface, no auto policy, no `POLICY.md`, no
  `policy_version` in the record.
- No interactive mode and no `--auto` flag. `new` is seeded generation and
  that is the whole of it.
- Replay is the simpler thing it becomes when nothing but the seed and the
  inputs feeds the engine.

Adding a choice point later means adding all of that back; it is not a
missing feature but an absent one.

## Replay and provenance contract

Adopted from ctchargen, which adopted it from t5chargen:

- Every subsector JSON carries `schema_version`, `ruleset` (pinned: Book 3
  pp. 1–12, © 1977 text, FFE reprints as held), `engine_version`, `rng`
  (algorithm + seed), an `inputs` block, and the ERRATA.md readings that
  governed it.
- RNG: Go `math/rand/v2` PCG, named in the record. Changing algorithm or
  dice-stream consumption order is an engine version bump.
- Replay re-runs the engine from the recorded seed and inputs, recomputing
  every throw; the stored event log is verification data, not input.
  `ctworldgen replay subsector.json` exits non-zero at the first mismatch,
  reporting the diverging event's sequence number.
- **World names are outside the contract.** Naming each world is step 3 of
  the p. 12 checklist and the book prints no table for it, so a name is
  written into the record after generation and appears in neither the
  inputs nor the event log. The engine can only ever regenerate it as
  empty. Replay therefore ignores the `name` of every world — otherwise a
  referee who did what the book asks would find his own record reported as
  diverged. Everything else about a world is verified. The subsector's own
  name is an input and is verified like one.
- `replay --ignore-provenance` waives the version match and nothing else.

## JSON conventions

Characteristics stored numeric with the string of digits derived and stored
alongside. Hexes are the four-character strings the p. 3 grid prints, not
column/row pairs: that is the identifier a referee writes in a notebook
(p. 4). Routes store the two hexes and the jump distance. Derived values
are stored and recomputed on replay. The schema is
`docs/subsector.schema.json` (draft 2020-12) with a minimal and a complete
example beside it, versioned by `schema_version`, which tracks the shape of
the records the engine writes: a constraint that only narrows the schema to
what the engine already produced is a clarification; one that would
invalidate a record the current engine writes is a bump.

## CLI sketch

```
ctworldgen new [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
ctworldgen batch --count 16 [--seed N] [-o dir|file.jsonl] [--force]
ctworldgen render [--history] subsector.json
ctworldgen replay [--ignore-provenance] subsector.json
ctworldgen version
```

`--occurrence-dm` takes −1, 0, or +1 and nothing else: p. 1 offers those
two modifiers and no others. `new` writes JSON to stdout unless `-o`;
`batch` emits JSONL (or one file per subsector with `-o dir`) and derives
each member's seed from the base seed + index, recorded in each record.
Sixteen is the default `--count` suggestion for a reason — a sector is
sixteen subsectors — but no sector-level structure is implied or recorded;
`batch` produces independent subsectors. Existing files are never
overwritten without `--force`. Flags precede the filename (Go `flag` stops
at the first non-flag argument). `version` reports the build and the
versions a record stamps, read from the toolchain's embedded build info.

## Decisions (2026-08-30)

- **The subsector is the record.** Not the world. Star mapping is
  subsector-scoped and lanes need neighbours.
- **The held printing governs.** Where the © 1977 text differs from the
  1981 revision most people remember, implement the page as held; do not
  adjudicate editions from memory. Readings go in ERRATA.md with the page
  cite, never applied silently. The string of digits (E005) is the one a
  reader will notice first.
- **An empty subsector is a result.** A run whose eighty throws place no
  world produces a valid record with no worlds and no lanes, the way a
  character killed by a survival throw produces a valid record in the
  sibling. Nothing rerolls.
- **Clean room.** Sibling repos `philoserf/ctchargen`, `philoserf/t5chargen`,
  `philoserf/t5`, and `philoserf/traveller` are not imported from. The
  _contracts_ ctchargen proved — replay/provenance, the event log, the
  COVERAGE and ERRATA documents, the unpinned toolchain gate — are
  adopted; its code is not.

## Architecture notes

- Repo: new `philoserf/ctworldgen`; code does not live in the document
  collection.
- Packages: `dice`, `worldgen` (the record, the event log, the engine that
  walks pp. 1–12, replay), `tables` (data-driven Book 3 charts: the
  starports table and chart, the jump routes table, the technological index
  matrix, the descriptive tables), `render`, `cmd/ctworldgen`. Plus
  `internal/fixture`, test support rather than architecture: the one
  definition of the golden-fixture roster, so the `worldgen` and `render`
  golden trees cannot come to describe different subsectors under the same
  name.
- Data/logic boundary: tables, thresholds, and labels are embedded data
  files loaded with `go:embed` plus load-time validation; the procedure's
  orchestration and its exceptional mechanics (size 0 forces atmosphere 0;
  size 0 or 1 forces hydrographics 0) are typed Go. No rules language.
- Testing: golden subsector fixtures across the occurrence DMs, hex-distance
  tests measured by hand off the printed p. 3 grid, replay round-trips,
  schema validation, property tests on the dice engine and on the generated
  ranges. Fixtures move only via regeneration, never by hand.

## Milestones

1. Dice engine + the hex grid and its distance metric + the subsector
   record and render, with the generation event log wired in from the start
   (end-to-end walking skeleton: occurrence scan and starport types only —
   a map with no worlds detailed on it).
2. The rest of star mapping: bases and space lanes. Exit criterion: a
   living `COVERAGE.md` mapping every step of pp. 1–12 to its page cite,
   implementation, and golden test.
3. World creation: the seven basic characteristics and the technological
   index, with the clamps and the string of digits.
4. Batch mode; replay verification; the render's full subsector listing and
   history transcript.

## Sources

- `~/Documents/Traveller/Classic/Book 3 Worlds and Adventures.pdf`
  (authoritative for the whole procedure; Worlds pp. 1–12).
- `~/Documents/Traveller/Classic/Book 1 Characters and Combat.pdf` (die
  roll conventions pp. 2–3, hexadecimal digit notation p. 8 — consulted
  only because Book 3's procedure is written on top of them).
- Page cites are printed page numbers. In the held PDFs, printed page N is
  PDF page N+5 in Book 3 and N+6 in Book 1.
- **The printed tables are transcribed from a visual read of the page, not
  from a text extraction.** The held PDFs' embedded font maps the em-dash
  to the glyph `4` and the minus sign to `3`, so `pdftotext` renders the
  jump routes table's empty cells as the digit 4 and `2D−2` as `2D32`. Both
  produce data that is wrong and plausible. Any future table work reads the
  page.
- Everything else in the collection — the _Consolidated Errata_, the
  Starter Edition, the facsimile, the _Rules Companion_, the _Guide to
  Classic Traveller_ — is explicitly **out of authority** for this tool.

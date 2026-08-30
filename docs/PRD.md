# PRD: Classic Traveller Subsector Generator (Go CLI), second take

2026-08-30. Status: draft — the contract for a fresh, empty repository,
before any code.

This is not a migration plan. It is what the project would be if it were
started again today, keeping what the first take proved and reversing what
it got expensive. Two changes carry the weight: the domain is typed rather
than stringly-typed, and the generator generates — the chronological
event log stops being cargo in the record and becomes a view that can be
re-derived from any record on demand.

## Problem

Classic Traveller world generation is the referee's mapping chore: eighty
hexes checked one at a time for a star system, a starport graded for each
one found, commercial space lanes drawn between them, and then eight
characteristics rolled for every world — size, atmosphere, hydrographics,
population, government, law level, and a technological index derived from
all of the above. A subsector runs to forty worlds. Doing it by hand is
slow and error-prone.

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
2. Model the book's own vocabulary as types. A hex, a starport type, and a
   characteristic digit are things with an identity, not strings that
   happen to look right.
3. Determinism: a seed and the inputs reproduce a subsector exactly, and a
   record says which engine made it.
4. Output a subsector record as JSON (canonical) and a Markdown subsector
   listing in the style the book's own records call for — the world's
   name, its hex, and its characteristics as a string of digits (p. 4).
5. Show the work. The full transcript of throws, DMs, and outcomes, with
   page cites, is available for any record — derived, not stored.

## Non-goals

- Anything outside Book 3 pp. 1–12. Equipment (pp. 13–18), encounters
  (pp. 19–23), animal encounters (pp. 24–32), and psionics (pp. 33–42) are
  in-play material, not world generation.
- The technological level tables (pp. 10–11). They say what a technological
  index _means_ — what weapons, computers, and drives a world can build —
  which is referee reference during play, not a step of generation. Only
  the index's range is read from them.
- Books 1, 2, and everything else in the collection: Books 4+, the
  supplements, the Starter Edition, The Traveller Book, JTAS, and the
  _Consolidated Errata_ PDF — all out of authority. Book 1 pp. 2–3 and
  p. 8 are in, and only because Book 3's procedure is written on top of
  them.
- Later printings' rules. The held text is the © 1977 text; where later
  printings famously differ, the held page governs. The two that bite
  here: the held technological index matrix gives a population-0 world no
  DM and therefore a technological index like any other world's, and the
  held text has no hyphen in the string of digits (E005).
- **Stellar detail.** The book generates "the single inhabited world in a
  star system" (p. 2) and leaves the star, the other planets, satellites,
  and gas giants to the referee. So does this tool.
- **Single-world generation.** The unit of generation is the subsector:
  a starport distribution is drawn for a whole subsector, and a space lane
  needs a neighbour. A lone world has no starport throw of its own and no
  routes to determine.
- **World names.** P. 12 step 3 says to name each world and prints no
  table; the book's naming material is advice, not a generator. Records
  carry an empty `name` for every world and the hex identifies it.
  `--name` names the subsector.
- **A stored event log.** See "The transcript is a view".
- **A record verifier.** No `verify` and no `replay` subcommand. See
  "Verification lives in the suite".
- **Types that carry rule invariants.** See "The domain".
- **Choice points.** No `Decider`, no auto policy, no interactive mode, no
  `policy_version`. Walk pp. 1–12 and every step is a throw and a table
  lookup; the referee's latitude — the occurrence DM (p. 1), a house
  starports table (p. 1), a deliberately imposed world (p. 8), a name
  (p. 12) — is exercised before the run, not during it. Those are
  _inputs_. Adding a choice point later means adding all of that back; it
  is not a missing feature but an absent one.
- Graphical hex maps. The record carries hexes and routes; drawing them is
  a downstream concern.

## Authority model — the most important rule

Rules come **only** from the held PDFs named above.

- **Never implement a Traveller rule from memory.** Training-data
  Traveller is mostly the 1981 revision and later editions. The held page
  governs even where it differs.
- Every implemented rule carries a printed-page cite, in the code and in
  `COVERAGE.md`.
- Where the text is ambiguous or silent, the chosen reading goes in
  `ERRATA.md` with the page cite, and is recorded in every record it
  governed — never applied silently.
- Page N as printed is PDF page N+5 in Book 3, N+6 in Book 1.

### Read the tables off the page, never out of `pdftotext`

The held PDFs' embedded font maps the em-dash to the glyph `4` and the
minus sign to `3`. A text extraction therefore renders the jump routes
table's empty cells as the digit 4 — `B-E 4 4 4 4` for `4 — — —` — and
the size formula `2D − 2` as `2D32`. Both readings are wrong and both look
like data. Every table is transcribed from a **visual** read of the page
(`Read` the PDF with a `pages` range), and then transcribed a second time
inside the table package's tests, so the two must agree. That second
transcription is the check the font trap needs, and it is where a new
numeric table belongs.

The one exception is the descriptive **labels** of pp. 5–7 — "Feudal
Technocracy", "Dense, tainted". Those are editorial abbreviations of the
book's printed prose rather than transcriptions of it, so retyping them in
a test compares an abbreviation against itself. They are checked by
re-reading the page; the suite only asserts that every value in range has
a non-empty label.

## The domain

The types are the deliverable of this take. The rule dividing them is:
**types carry identity, never rule invariants.**

```go
type Hex struct{ Col, Row int }   // 0101–0810, marshals as "0101"
type Starport byte                // A B C D E X, marshals as "A"
type Digit byte                   // the p. 2 alphabet, marshals as "A"
type Characteristic int           // Size … TechIndex; a closed enum
type Parsecs int                  // a jump distance, 1–4 for a lane
```

Each parses at the program's edges, marshals to the string the record
prints, and makes a whole class of defect uncompilable rather than
runtime-checked: a hex the engine wrote cannot need re-parsing before it
can be measured, a starport cannot be `"a"` or `""`, and the technological
index matrix cannot be asked for a column that does not exist.

**Characteristic values stay `int`.** A `type Atmosphere int` with a
compile-time range of 0–12 would put the table's ceiling in Go source,
next to the data file and the test that transcribes it — a third copy of
a number the book prints once, recreating exactly the drift this design
exists to prevent. The ceiling is read off the loaded table
(`len(rows)-1`) so the clamp and the chart cannot come apart. A type's
invariant is also a rules claim, and a rules claim belongs on a page with
a cite, not in a struct definition where no reader will look for it.

So: `Hex`, `Starport`, `Digit`, `Characteristic`, `Parsecs`, and the
dice package's `Target` are types. Size, atmosphere, hydrographics,
population, government, law level, and the technological index are `int`,
clamped against their own printed tables.

## Functional requirements

Rule citations are printed page numbers of Book 3 unless marked B1.

**FR1 — The subsector grid.** Eighty hexes, eight columns of ten rows,
identified by the four-digit column-and-row number the p. 3 grid prints
(0101 through 0810). One hex is one parsec (p. 1). Even-numbered columns
sit half a hex lower than odd-numbered ones, as printed; distance between
two hexes is the hex-grid distance that layout gives.

**FR2 — World occurrence.** Check each hex in turn, throwing one die: 4, 5,
or 6 places a world and its attendant star system (p. 1). The referee may
impose a DM of +1 or −1 on the whole subsector; that DM is an input, and
+1, 0, and −1 are the whole range p. 1 offers. The scan order is unstated
and stream-load-bearing (E002).

**FR3 — Starport type.** Two dice for each world, read against the
starports table: 2–4 A, 5–6 B, 7–8 C, 9 D, 10–11 E, 12 X (p. 1).

**FR4 — Bases.** The starport chart prints the throws the p. 12 checklist
omits (p. 5): naval base 8+ at starport A and B; scout base 10+ at A, 9+ at
B, 8+ at C, 7+ at D. E and X have neither. Two dice each, per B1 pp. 2–3.
Their position in the dice stream is a reading (E001).

**FR5 — Space lanes.** For every pair of worlds four or fewer hexes apart,
throw one die against the jump routes table cell for the pair's starport
types and their distance; equal or greater means a lane exists (p. 2). The
table has rows for A through E only, and cells marked with a dash at which
no lane is possible; a pair with no cell consumes no die. Each pair is
examined once. Enumeration order and the treatment of X starports are
readings (E003).

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

**FR13 — The two throws not made.** A size-0 world's atmosphere and a
size-0-or-1 world's hydrographics are automatic (p. 4). No die is thrown
and none is discarded: rolling and dropping one there would shift every
later world in the stream.

**FR14 — Value ranges.** Every generated value is clamped to the printed
range of its own table — floor at 0, ceiling at the last row the book
prints — because the arithmetic of FR6–FR12 can leave it, and because the
technological index matrix has a row for every value in a table's printed
range and none outside it. Law level is the exception and is floored only:
its table ends at 9, but the note beneath it is written for higher levels
and law level feeds no matrix (E004). A clamp that actually binds is
recorded on the world it bound (see JSON conventions), so the reading is
never applied silently.

**FR15 — The string of digits.** Each world's characteristics are recorded
as a string, "in much the same manner as the Universal Personality Profile
is used for the easy identification of persons" (p. 4), in the order the
p. 4 Planetary Characteristics box lists them: starport, size, atmosphere,
hydrographics, population, government, law level, technological index.
Eight characters, one per characteristic, in the p. 2 notation, with
nothing between them (E005).

**FR16 — Dice engine.** One- and two-die throws with cumulative DMs against
`N+`/`N−`/exact targets per the die roll conventions (B1 pp. 2–3). Book 3
uses one-die targets the character procedure never does — world occurrence
is 4+ on one die, and the jump routes table's cells are one-die targets
from 1 to 6 — so the target notation admits values below 2. All dice are
consumed from one seeded stream, in procedure order.

**FR17 — Subsector record.** The subsector's name, the occurrence DM it was
generated under, every world (hex, name, starport, bases, the seven basic
characteristics, the technological index, the string of digits, and any
clamp that bound), and every space lane (the two hexes and the jump
distance). Plus provenance: schema version, ruleset, engine version, RNG
algorithm and seed, inputs, and the readings that governed. The JSON
record is the source of truth; every Markdown output is a render of it.

**FR18 — Trace.** The engine accepts an optional trace sink and reports
every step entered, every throw (dice, DMs, target, total, result), and
every outcome, each with its page cite. **Tracing is observation, never
input**: a run with a sink attached and the same run without produce
identical records, and a test asserts it.

**FR19 — The transcript.** `render --history record.json` re-runs the
engine from the record's own seed and inputs with a trace sink attached,
and renders the transcript. Nothing is read back from the record but seed,
inputs, and version stamps. Available for any record whose engine version
matches the build; refused, by name, for one that does not. There is no
flag to waive that: a transcript that does not describe the record it was
asked about is worse than no transcript.

**FR20 — Batch.** `batch --count N` produces N independent subsectors,
each member's seed derived from the base seed plus its index and recorded
in its own record. Sixteen is the suggested count because a sector is
sixteen subsectors, but no sector-level structure is implied or recorded.

**FR21 — Listing.** The Markdown subsector listing: the world roster with
hexes, names, and strings of digits; the space lanes; and a page of detail
per world with the labels its Book 3 tables give — "at least one (and
preferably several) pages in a central notebook maintained by the referee"
(p. 4), which is what this stands in for.

## The transcript is a view

The first take stored the event log in the record. It cost 92% of every
record's bytes — 258KB of a 280KB subsector, 1,164 events for 41 worlds —
and it version-pinned prose: because verification compared whole events,
editing the wording of an outcome invalidated every record written before
it, and the engine version had to be bumped for a typo fix.

None of that is necessary here, because the procedure has no choice
points. Seed and inputs determine everything, so the transcript is
**recomputable**: attach a sink, re-run, render (FR19). The record carries
outcomes; the transcript carries the reasoning; both come from one engine.

What the record still carries, because it cannot be recomputed from a
value, is a clamp that bound (FR14) — small structured data on the world
it governed rather than a prose event, which keeps E004 non-silent at a
cost of a few dozen bytes.

## Verification lives in the suite

There is no `verify` and no `replay` subcommand. With no choice points, a
verifier is a re-run and a diff — which is exactly what the golden
fixtures already do, on every `task` run, against records this repository
controls. A shipped command would add only the ability to run that diff
against a _user's_ file, and that is the case where it is least welcome: a
record is the referee's notebook, and the book tells him to write in it
(p. 12 step 3). The first take had already carved world names out of the
comparison for that reason, and every further annotation a referee might
want would need another carve-out.

So the guarantee is kept where it is cheap and dropped where it fights the
book. A test re-runs the engine from each golden record's own seed and
inputs and asserts that it reproduces that record's worlds and routes.
Nothing ships that audits a record it did not write.

What this gives up, plainly: nothing detects a hand-edited record, or a
record made by an older engine, at the moment it is read. The version
stamps still say which engine wrote it, and `render --history` refuses a
record it cannot re-derive (FR19) — that refusal is the one place a stale
record is caught, and it is caught by the feature that needs the answer
rather than by a verifier standing beside it.

## Determinism and provenance

- Determinism is unchanged from the first take and is not weakened by
  dropping the stored log. **Dice-stream consumption order stays
  load-bearing**: it fixes what a seed means, and the golden fixtures pin
  it. Procedure-order changes are still engine-version changes.
- RNG: Go `math/rand/v2` PCG, named in the record.
- **The engine version bumps for three things and no others**: a rule
  change, a change to dice-stream consumption order, or a change to the RNG
  construction. Prose — outcome wording, transcript layout, a page cite's
  phrasing — is no longer part of the contract and never bumps it. This is
  the headline improvement over the first take.
- The schema version tracks the shape of the records the engine writes. A
  constraint that only narrows the schema to what the engine already
  produced is a clarification; one that would invalidate a record the
  current engine writes is a bump.
- Records carry the `ERRATA.md` readings that governed them, in document
  order.
- The stamps are not vestigial with the verifier gone. `render --history`
  needs them to know whether it may re-derive a record's transcript
  (FR19), and a record that cannot say which engine wrote it cannot be
  reasoned about at all.

## JSON conventions

Characteristics stored numeric, with the string of digits derived and
stored alongside. Hexes are the four-character strings the p. 3 grid
prints, not column/row pairs: that is the identifier a referee writes in a
notebook (p. 4), and the domain type marshals to it. Starports and digits
are the single characters the book prints. Routes store the two hexes and
the jump distance, the lower hex identifier first. Derived values — the string of
digits above all — are stored, and recomputed whenever the engine is
re-run from a record's seed.

A clamp that bound is recorded on its world:

```json
"clamps": [{ "characteristic": "atmosphere", "raw": 15, "value": 12 }]
```

Absent where no clamp bound, which is most worlds. The record-level
`errata` list still names E004 when any clamp bound anywhere.

The schema is `subsector.schema.json` (draft 2020-12) with a minimal and a
complete example beside it. Unknown fields are rejected when reading a
record, so a record from a newer schema fails loudly rather than silently
dropping data.

## CLI sketch

```
ctworldgen new [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
ctworldgen batch --count 16 [--seed N] [-o dir|file.jsonl] [--force]
ctworldgen render [--history] subsector.json
ctworldgen version
```

`--occurrence-dm` takes −1, 0, or +1 and nothing else. `new` writes JSON to
stdout unless `-o`; `batch` emits JSONL, or one file per subsector with
`-o dir`. Existing files are never overwritten without `--force`. Flags
precede the filename (Go `flag` stops at the first non-flag argument).
`version` reports the build and the versions a record stamps, read from the
toolchain's embedded build info.

## Architecture notes

- Packages: `dice` (the seeded stream, targets, throws); `subsector` (the
  domain types and the record, with their marshalers); `tables` (the
  embedded Book 3 charts and their typed lookups); `gen` (the engine that
  walks pp. 1–12 and the trace sink); `render` (listing and transcript);
  `cmd/ctworldgen`. Plus `internal/fixture`, test support rather than
  architecture: the one definition of the golden-fixture roster, so two
  golden trees cannot come to describe different subsectors under the same
  name.
- Data/logic boundary: tables, thresholds, and labels are embedded JSON
  loaded with `go:embed` plus load-time validation; the procedure's
  orchestration and its exceptional mechanics (FR13) are typed Go. No rules
  language.
- **The hex grid is the quiet trap.** The offset-to-cube conversion must
  match the parity the p. 3 grid prints — 0101 at the top-left, 0201 half a
  hex below it, so even-numbered columns are the ones pushed down (standard
  odd-q in zero-based indices). Getting it backwards leaves every distance
  internally consistent and wrong by one for half the map, which no
  record-against-record check can catch. A test measures against the page
  instead, on distances taken by hand off the printed grid. Never change
  the conversion without re-measuring there.
- Testing: golden subsector fixtures across the occurrence DMs; hex
  distances measured by hand off p. 3; the second transcription of every
  numeric table; the trace-does-not-affect-generation identity (FR18);
  the regeneration round-trip, each golden reproduced from its own
  recorded seed and inputs; schema validation; and property sweeps over many
  seeds asserting the book's invariants — every characteristic inside its
  table's printed range, every lane four parsecs or fewer between worlds
  with cells, no lane touching an X starport. Fixtures move only by
  regeneration, never by hand.
- The gate: `task` runs formatting, `go vet`, golangci-lint with
  `default: all` and a curated disable list, NilAway, and `go test -race`.
  CI runs exactly `task`. **The toolchain is deliberately unpinned** — the
  gate is meant to fail when a tool moves rather than drift behind it, so a
  red gate on untouched code is the signal working. Answer the finding; do
  not pin a tool to silence it.
- Coverage ratchets on **uncovered statements per package**, not a
  percentage: a percentage can hold still while a guarded branch adds one
  covered statement and one uncovered. It fails in both directions.

## Delta from the first take

Carried unchanged: the authority model and the visual-read rule; ERRATA
and COVERAGE as living documents; the readings E001–E005 with their
existing numbers, so records and citations remain comparable across the
rebuild; provenance stamps; the absence of choice points; goldens,
property sweeps, and the printed-grid distance test; the unpinned gate and
the uncovered-statement ratchet; an empty subsector as a valid result, with
nothing rerolled.

Reversed:

1. **The domain is typed.** Hexes, starports, digits, and characteristic
   identities are types with marshalers, not strings. Value ranges stay in
   table data.
2. **The event log is not stored.** It is a trace sink at generation time
   and a re-derived render afterwards (FR18–FR19).
3. **Replay is retired.** No verifier ships. The re-run-and-diff guarantee
   moves into the test suite, where the golden fixtures already provide
   it; see "Verification lives in the suite".
4. **Engine-version bumps shrink** to rules, stream order, and RNG.
5. **Records shrink** by roughly an order of magnitude, which makes a
   golden diff reviewable by eye.

## Decisions

- **The subsector is the record.** Not the world. Star mapping is
  subsector-scoped and lanes need neighbours.
- **The held printing governs.** Where the © 1977 text differs from the
  1981 revision most people remember, implement the page as held; do not
  adjudicate editions from memory. Readings go in ERRATA.md with the page
  cite, never applied silently. The string of digits (E005) is the one a
  reader will notice first.
- **An empty subsector is a result.** A run whose eighty throws place no
  world produces a valid record with no worlds and no lanes. Nothing
  rerolls.
- **Types for identity, data for ranges.** Stated in full under "The
  domain"; it is the decision this take exists to make.
- **The log is a view.** Stated in full under "The transcript is a view".

## Milestones

1. The domain and its edges: `dice`, the typed grid with its printed-grid
   distance test, `tables` loaded and validated with the second
   transcription in place, and the record with its marshalers. Walking
   skeleton: the occurrence scan and starport types, written out as JSON —
   a map with no worlds detailed on it.
2. The rest of star mapping: bases and space lanes, with the trace sink
   wired through from the start. Exit criterion: a living `COVERAGE.md`
   mapping every step of pp. 1–12 to its page cite, implementation, and
   test.
3. World creation: the seven basic characteristics and the technological
   index, with the clamps, the recorded clamp data, and the string of
   digits.
4. The renders (listing and re-derived transcript) and `batch`.

## Deliverable documents

`PRD.md` (this document), `ERRATA.md`, `COVERAGE.md`,
`subsector.schema.json` with a minimal and a complete example, and
`CLAUDE.md` carrying the authority model, the visual-read rule, the hex
parity trap, and the clean-room boundary.

## Sources

Book 3 _Worlds and Adventures_ pp. 1–12; Book 1 _Characters and Combat_
pp. 2–3 and p. 8. FFE reprints of the © 1977 text, as held in
`~/Documents/Traveller/Classic/`. Nothing else.

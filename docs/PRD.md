# PRD: Classic Traveller Subsector Generator (Go CLI)

2026-08-30. Status: the contract for this repository, before any code.

## Problem

Classic Traveller world generation is the referee's mapping chore: eighty
hexes checked one at a time for a star system, a starport graded for each
one found, commercial space lanes drawn between them, and then eight
characteristics rolled for every world — size, atmosphere, hydrographics,
population, government, law level, and a technological index derived from
all of the above. A subsector runs to forty worlds. Doing it by hand is
slow and error-prone.

`ctworldgen` generates a subsector from a seed, writes it as JSON, and
renders it as the Markdown listing a referee can run from.

## User

Mark: solo referee and developer. Secondary: any Classic Traveller referee
who needs a subsector to run in, or a sector's worth of them.

## Authority

Rules come **only** from the held PDFs in `~/Documents/Traveller/Classic/`
— the FFE reprints of the © 1977 text:

- **Book 3 _Worlds and Adventures_ pp. 1–12**, the Worlds chapter. This is
  the ruleset.
- **Book 1 _Characters and Combat_ pp. 2–3** (the die roll conventions)
  and **p. 8** (the hexadecimal digit notation), which Book 3 uses without
  restating.

Nothing else in that collection is in authority — not Books 2 and 4+, not
the supplements, the Starter Edition, The Traveller Book, JTAS, or the
_Consolidated Errata_ PDF, and not the rest of Book 3.

- **Never implement a Traveller rule from memory.** Training-data
  Traveller is mostly the 1981 revision and later editions. The held page
  governs even where it differs, and where it famously differs it is
  recorded: the string of digits carries no hyphen (E005), and a
  population-0 world receives a technological index like any other
  (`ERRATA.md`, Noted discrepancies).
- Every implemented rule carries a printed-page cite, in the code and in
  `COVERAGE.md`.
- Where the text is ambiguous or silent, the chosen reading goes in
  `ERRATA.md` with its page cite and its stamping condition, and is named
  in every record it governed — never applied silently.
- Printed page N is PDF page N+5 in Book 3, N+6 in Book 1.

This section is the single statement of the authority model. `CLAUDE.md`
points here rather than restating it, so the two cannot drift.

### Read the tables off the page, never out of `pdftotext`

The held PDFs' embedded font maps the em-dash to the glyph `4` and the
minus sign to `3`. A text extraction therefore renders the jump routes
table's empty cells as the digit 4 — `B-E 4 4 4 4` for `4 — — —` — and
the size formula `2D − 2` as `2D32`. Both readings are wrong and both look
like data.

Every table is transcribed from a **visual** read of the page (`Read` the
PDF with a `pages` range), and then transcribed a second time inside the
table package's tests, so the two must agree. That second transcription is
the check the font trap needs, and it is where a new numeric table
belongs.

The one exception is the descriptive **labels** of pp. 5–7 — "Feudal
Technocracy", "Dense, tainted". Those are editorial abbreviations of the
book's printed prose rather than transcriptions of it, so retyping them in
a test compares an abbreviation against itself. They are checked by
re-reading the page; the suite only asserts that every value in range has
a non-empty label.

## Scope

The tool generates a subsector: the eighty-hex world occurrence scan,
starport types and their bases, space lanes, and the eight characteristics
of every world found. It writes a JSON record and renders a Markdown
listing from it.

### Not in scope

- **Anything outside Book 3 pp. 1–12.** Equipment (pp. 13–18), encounters
  (pp. 19–23), animal encounters (pp. 24–32), and psionics (pp. 33–42) are
  in-play material, not world generation.
- **The technological level tables (pp. 10–11).** They say what a
  technological index _means_ — what weapons, computers, and drives a
  world can build — which is referee reference during play, not a step of
  generation.
- **Later printings' rules.** The held text is the © 1977 text; where a
  later printing differs, the held page governs.
- **Stellar detail.** The book generates "the single inhabited world in a
  star system" (p. 2) and leaves the star, the other planets, satellites,
  and gas giants to the referee. So does this tool.
- **Single-world generation.** The unit of generation is the subsector: a
  starport distribution is drawn for a whole subsector, and a space lane
  needs a neighbour. A lone world has no starport throw of its own and no
  routes to determine.
- **World names.** P. 12 step 3 says to name each world and prints no
  table; the book's naming material is advice, not a generator. Records
  carry an empty `name` for every world and the hex identifies it.
  `--name` names the subsector.
- **The referee's latitude that the book leaves as prose.** His own
  starports table, "perhaps as many as one for each subsector" (p. 1); an
  occurrence DM applied "on broad areas within a subsector" rather than to
  the whole of it (p. 1) — the subsector-wide DM is supported and the
  per-area one is not; worlds "imposed … deliberately (rather than
  randomly) generated" (p. 8); the separate government of each territory
  on a balkanized world (p. 8); and the alternate world forms — rosettes,
  ringworlds, sphereworlds — which p. 8 states are "not included in the
  world creation sequence." None of these is a throw against a table.
- **In-play reference the generation pages happen to carry.** The law
  enforcement throw and the starport's extraterritoriality (pp. 1, 7), the
  protective equipment each atmosphere requires (p. 5), and the population
  density comparisons (p. 9). The tables' labels are carried; the prose
  beneath them is not.
- **Showing the work.** The tool records what the dice produced, not the
  reasoning that produced it. There is no event log, stored or derived,
  and no transcript render. A record is a notebook page (p. 4), which is
  what the book asks a referee to keep.
- **Choice points.** No `Decider`, no auto policy, no interactive mode, no
  `policy_version`. Walk pp. 1–12 and every step is a throw and a table
  lookup; the referee's latitude is exercised before the run, not during
  it, which makes it an _input_. Adding a choice point later means adding
  all of that back; it is not a missing feature but an absent one.
- **A record verifier.** No `verify` and no `replay` subcommand. See
  "Verification lives in the suite".
- **Graphical hex maps.** The record carries hexes and routes; drawing
  them is a downstream concern.

## The domain

The rule dividing the types from the data is: **types carry identity,
never rule invariants.**

```go
type Hex struct{ Col, Row int }   // 0101–0810, marshals as "0101"
type Starport byte                // A B C D E X, marshals as "A"
type Digit byte                   // the p. 2 alphabet, marshals as "A"
type Characteristic int           // Size … TechIndex; a closed enum
type Parsecs int                  // a jump distance
```

Each parses at the program's edges, marshals to the string the record
prints, and makes a class of defect uncompilable rather than
runtime-checked: a hex the engine wrote cannot need re-parsing before it
can be measured, a starport cannot be `"a"` or `""`, and the technological
index matrix cannot be asked for a column that does not exist.

**Characteristic values stay `int`.** A `type Atmosphere int` with a
compile-time range of 0–12 would put the p. 5 table's last row into Go
source, next to the data file and the test that transcribes it — a third
copy of a number the book prints once, recreating exactly the drift this
design exists to prevent. It would also be **wrong**: R14 caps nothing at
a descriptive table's last row, so a generated atmosphere reaches 15 and
the type would reject a legal value. A range in a type is a rules claim,
and a rules claim belongs on a page with a cite, not in a struct
definition where no reader will look for it — least of all one the rules
turn out not to make.

**`Hex` and `Digit` are the deliberate boundary cases**, and they are on
the type side. `Hex` enforces the eight columns and ten rows of the p. 3
grid, and `Digit` enforces the p. 2 alphabet — both printed rules. They
are types because an identifier and a notation are identity: a hex outside
0101–0810 is not a hex, and a character outside the alphabet is not a
digit. Neither is a _value range that a table also prints_, which is the
thing the paragraph above refuses. If the grid or the notation ever
becomes data, that is a different tool.

So: `Hex`, `Starport`, `Digit`, `Characteristic`, `Parsecs`, and the dice
package's `Target` are types. Size, atmosphere, hydrographics, population,
government, law level, and the technological index are `int`, bounded only
by R14 — a floor of 0, and the technological index's printed cap.

## Requirements

In the order the book walks them. Rule citations are printed page numbers
of Book 3 unless marked B1. A requirement that rests on a reading names
its `ERRATA.md` entry.

### Star mapping (pp. 1–3)

**R1 — The subsector grid.** Eighty hexes, eight columns of ten rows,
identified by the four-digit column-and-row number the p. 3 grid prints
(0101 through 0810). One hex is one parsec (p. 1). Even-numbered columns
sit half a hex lower than odd-numbered ones, as printed; distance between
two hexes is the hex-grid distance that layout gives.

**R2 — World occurrence.** Check each hex in turn, throwing one die
against a target of 4+: the page marks the hex "if the result is a 4, 5,
or 6" (p. 1), which is that target and not a set of three faces — the same
paragraph lets the referee make worlds "more frequent or less frequent" by
a DM, and only a target reading makes the DM do that (`ERRATA.md`, Noted
discrepancies). A success places a world and its attendant star system.
The referee's DM applies to the whole subsector; it is an input, and +1,
0, and −1 are the only values p. 1 offers. The scan order is unstated and
fixes the dice stream (E002).

**R3 — Starport type.** Two dice for each world, read against the
starports table: 2–4 A, 5–6 B, 7–8 C, 9 D, 10–11 E, 12 X (p. 1).

**R4 — Naval and scout bases.** The starport chart prints the throws the
p. 12 checklist omits (p. 5): naval base on 8+ at starport A and B; scout
base on 10+ at A, 9+ at B, 8+ at C, 7+ at D. Starports E and X have
neither. Two dice each, per B1 pp. 2–3. Their position in the dice stream
is a reading (E001).

**R5 — Space lanes.** For every pair of worlds four or fewer hexes apart,
throw one die against the jump routes table cell for the pair's starport
types and their distance; "if the one die throw is equal to, or greater
than the number, a space lane exists" (p. 2). The table prints rows for
A–A through E–E and none for X, and prints an em-dash in the cells at
which no lane is possible. Neither an absent row nor a dash cell consumes
a die: the throw is made against a stated number, and neither states one.
Each pair is examined once. Which pairs are examined, the treatment of X
starports, the dash cells, and the enumeration order are readings (E003).

### World creation (pp. 4–9)

**R6 — Planetary size.** 2D − 2, ranging 0 to 10; zero is an
asteroid/planetoid complex (pp. 4, 12).

**R7 — Planetary atmosphere.** 2D − 7 + planetary size. A planet of size
zero automatically has an atmosphere of zero (pp. 4, 12).

**R8 — Hydrographic percentage.** 2D − 7 + planetary size, with a further
DM of −4 if the atmosphere is 0, 1, or greater than 9. A planetary size of
0 or 1 gives an automatic 0 (pp. 4, 12).

**R9 — Population.** 2D − 2, ranging 0 to 10, an exponent of 10
(pp. 8, 12).

**R10 — Planetary government.** 2D − 7 + the population digit (pp. 8, 12).

**R11 — Law level.** 2D − 7 + the government type (pp. 8, 12).

**R12 — Technological index.** One die, modified by the sum of the DMs the
technological index matrix gives for the world's starport, size,
atmosphere, hydrographics, population, and government values (p. 9). The
index "may vary from zero to 18" (p. 9).

**R13 — The two throws not made.** A size-0 world's atmosphere and a
size-0-or-1 world's hydrographics are automatic (p. 4). No die is thrown
and none is discarded: rolling and dropping one there would shift every
later world in the stream.

**R14 — Value ranges.** Every generated value is **floored at 0**, which
the notation forces rather than a rule choosing it: R15 requires one
character per characteristic and neither B1 p. 8's hexadecimal nor p. 2's
letters has a character for a negative number. R7, R8, R10, R11 and R12
can all fall below zero; R6 and R9 cannot. A floored value is the value —
it is what the record carries and what every later step consumes, so
government feeds law level already floored, and the raw negative is
recorded beside it and used for nothing.

**Nothing is capped because a table stops describing it.** A value no
table names is a gap in the descriptive table, not a bound on the throw,
and p. 8's remedy for one is the referee's own description rather than a
clamp (E004). So atmosphere, hydrographics and government run 0 to 15, and
law level 0 to 20.

**One cap: the technological index at 18**, because p. 9 prints a range
for the value itself — "may vary from zero to 18" — where the other
characteristics get only a table that happens to end. It binds: the
matrix's DMs reach +14, which with a die of 6 gives 20.

That maximum is starport A, size 0, population A, government 5 — the
atmosphere and hydrographics of a size-0 world being automatic (R13).

A floor or the cap that actually binds is recorded on the world it bound,
so the reading is never applied silently.

**R15 — The string of digits.** Each world's characteristics are recorded
as a string, "in much the same manner as the Universal Personality Profile
is used for the easy identification of persons" (p. 4), in the order the
p. 4 Planetary Characteristics box lists them: starport, size, atmosphere,
hydrographics, population, government, law level, technological index.
Eight characters, one per characteristic, with nothing between them
(E005). The notation is B1 p. 8's hexadecimal extended by p. 2's letter
set: 10 is A, 15 is F, 16 is G, 17 is H, 18 is J — I being omitted — 19 is
K and 20 is L. It must reach 20, which is law level's maximum under R14.

**R16 — The descriptive tables.** Every value the engine generates has a
label the book prints, and the listing carries it: the planetary size
(0–C), atmosphere (0–C), hydrographic percentage (0–A) and population
(0–A) tables of pp. 5–6, the governmental type table (0–D) of p. 6, and
the law levels table (0–9) of p. 7. The p. 5 starport chart supplies the
same for a starport type — quality, fuel, annual maintenance overhaul, and
shipyard — alongside the base throws of R4. These are data, embedded and
validated at load like every other table, and every value in a table's
printed range must have a label.

R14 lets a generated value exceed that range — an atmosphere of 13, a
hydrographic percentage of 12, a government of 14 — and the book prints no
label for one. The listing prints the digit and no description. That is a
gap in the page, not an error to correct, and the load-time check is
unaffected: it covers the range the table prints.

### The tool

**R17 — Dice engine.** One- and two-die throws with cumulative DMs against
`N+`/`N−`/exact targets, per the die roll conventions (B1 pp. 2–3). Book 3
uses one-die targets the character procedure never does — world occurrence
is 4+ on one die, and the jump routes table's cells are one-die targets
from 1 to 6 — so the target notation admits values below 2. All dice are
consumed from one seeded stream, in procedure order.

**R18 — The subsector record.** The subsector's name, the occurrence DM it
was generated under, every world (hex, name, the seven basic
characteristics — starport, size, atmosphere, hydrographics, population,
government, law level — the technological index, the naval and scout
bases, the string of digits, and any clamp that bound), and every space
lane (the two hexes and the jump distance). Plus provenance: schema version, ruleset, engine version, RNG
algorithm and seed, inputs, and the readings that governed. The JSON
record is the source of truth; the Markdown listing is a render of it.

**R19 — Batch.** `batch --count N` produces N independent subsectors, each
member's seed derived from the base seed and its index and recorded in its
own record. Sixteen is the suggested count because a sector is sixteen
subsectors, but no sector-level structure is implied or recorded.

**R20 — The subsector listing.** Markdown: the world roster with hexes,
names, and strings of digits; the space lanes; and a page of detail per
world with the labels its Book 3 tables give (R16) — "at least one (and
preferably several) pages in a central notebook maintained by the referee"
(p. 4), which is what this stands in for.

## Determinism and provenance

A seed and the inputs reproduce a subsector exactly. **Dice-stream
consumption order is load-bearing**: it fixes what a seed means, and the
golden fixtures pin it.

- **RNG.** Go `math/rand/v2` PCG. The single recorded seed fills both
  words of the PCG state — `rand.New(rand.NewPCG(seed, seed))` — so one
  seed field reproduces the stream. The record stamps the algorithm as the
  exact string `go-math-rand-v2-pcg`.
- **A seed is always recorded.** Without `--seed`, one is drawn from the
  OS entropy source and written into the record, so a run is reproducible
  after the fact. That draw is the one exception to the seeded-stream
  rule, and it happens before the engine starts. `--seed 0` is therefore
  an explicit and distinct choice, not a request for a random one.
- **Batch derivation.** Member _i_ of a batch is seeded with `base + i` as
  an unsigned 64-bit addition, wrapping if it overflows, and that derived
  seed — not the base — is what its record carries. Each member is a
  complete, independently reproducible record.
- **The engine version bumps for three things and no others**: a rule
  change, a change to dice-stream consumption order, or a change to the
  RNG construction. Nothing else can alter what a seed produces. The
  engine emits no prose at all, so there is no wording anywhere in the
  contract for a version to track.
- **The schema version tracks the shape of the records the engine
  writes.** A constraint that only narrows the schema to what the engine
  already produced is a clarification; one that would invalidate a record
  the current engine writes is a bump.
- **The provenance stamps are for a reader and for the suite.** No shipped
  command reads them back. They are recorded because a record that cannot
  say which engine, ruleset and seed made it cannot be reasoned about by
  the person holding it, and because the regeneration test reads each
  golden's own seed and inputs to reproduce it (see below). That is their
  whole job, and it is enough.

## Verification lives in the suite

There is no `verify` and no `replay` subcommand. With no choice points, a
verifier is a re-run and a diff — which is what the golden fixtures
already do, on every `task` run, against records this repository controls.
A shipped command would add only the ability to run that diff against a
_user's_ file, and that is the case where it is least welcome: a record is
the referee's notebook, and the book tells him to write in it (p. 12
step 3). Every annotation a referee might want would need a carve-out from
the comparison.

So the guarantee is kept where it is cheap and dropped where it fights the
book. A test re-runs the engine from each golden record's own seed and
inputs and asserts that it reproduces that record's worlds and routes.
Nothing ships that audits a record it did not write.

What this gives up, plainly: nothing detects a hand-edited record, or a
record made by an older engine, at the moment it is read. The version
stamps say which engine wrote it, and a reader who cares can re-run the
seed and compare. There is no automated moment at which that happens.

## The record

Characteristics are stored numerically, with the string of digits derived
and stored alongside. Hexes are the four-character strings the p. 3 grid
prints, not column/row pairs: that is the identifier a referee writes in a
notebook (p. 4), and the domain type marshals to it. Starports and digits
are the single characters the book prints. Routes store the two hexes and
the jump distance, the lower hex identifier first.

A floor or the cap that bound is recorded on its world. Both are called
clamps in the record, which is the one word for the one mechanism:

```json
"clamps": [{ "characteristic": "government", "raw": -3, "value": 0 }]
```

Absent where nothing bound, which is most worlds. `raw` admits negative
numbers — under R14 almost every clamp that binds is a floor. It is the
one thing the record carries that cannot be recomputed from a value, which
is why it is there.

**Stamping is conditional, and the condition is part of the contract.** A
record's `errata` array carries, in document order, the identifiers of the
readings that actually governed it: E002 on every record, E001 where any
base throw was made, E003 where there were two or more worlds, E004 where
a floor or the cap actually bound, E005 where there was at least one
world. Each
entry in `ERRATA.md` states its own condition, and those statements are
the specification.

The schema is `subsector.schema.json` (draft 2020-12) with a minimal and a
complete example beside it. **Rejecting unknown fields is two
obligations**: `"additionalProperties": false` at every level of the
schema, and `Decoder.DisallowUnknownFields` on the Go side. A schema
alone rejects nothing at read time. Both are required, so a record from a
newer schema fails loudly rather than silently dropping data.

## CLI

```
ctworldgen new   [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
ctworldgen batch --count N [--seed N] [--name X] [--occurrence-dm N] [-o dir|file.jsonl] [--force]
ctworldgen render [-o file] [--force] subsector.json
ctworldgen version
```

- `--occurrence-dm` takes −1, 0, or +1 and nothing else; it defaults to 0.
- `--seed` defaults to a value drawn from OS entropy and recorded (see
  "Determinism and provenance").
- `--name` names the subsector and defaults to empty. In `batch` it names
  every member.
- `new` and `render` write to stdout unless `-o` is given. `batch` emits
  JSONL to stdout; with `-o`, a path that names an existing directory or
  ends in a separator gets one file per subsector, and any other path is a
  single JSONL file.
- Existing files are never overwritten without `--force`.
- Flags precede the filename (Go `flag` stops at the first non-flag
  argument).
- `version` reports the build and the versions a record stamps, read from
  the toolchain's embedded build info.

## Architecture

- **Packages:** `dice` (the seeded stream, targets, throws); `subsector`
  (the domain types and the record, with their marshalers); `tables` (the
  embedded Book 3 charts and their typed lookups); `gen` (the engine that
  walks pp. 1–12); `render` (the subsector listing); `cmd/ctworldgen`.
  Plus `internal/fixture`, test support rather than architecture: the one
  definition of the golden-fixture roster, so two golden trees cannot come
  to describe different subsectors under the same name.
- **Data/logic boundary:** tables, thresholds, and labels are embedded
  JSON loaded with `go:embed` plus load-time validation; the procedure's
  orchestration and its exceptional mechanics (R13) are typed Go. No rules
  language.
- **The hex grid is the quiet trap.** The offset-to-cube conversion must
  match the parity the p. 3 grid prints — 0101 at the top-left, 0201 half
  a hex below it, so even-numbered columns are the ones pushed down
  (standard odd-q in zero-based indices). Getting it backwards leaves
  every distance internally consistent and wrong by one for half the map,
  which no record-against-record check can catch. A test measures against
  the page instead, on distances taken by hand off the printed grid. Never
  change the conversion without re-measuring there.

## Testing and the gate

- Golden subsector fixtures across the occurrence DMs. Fixtures move only
  by regeneration, never by hand.
- Hex distances measured by hand off the p. 3 grid.
- The second transcription of every numeric table, inside the `tables`
  tests: the starports table, the jump routes table, the technological
  index matrix, and the p. 5 chart's base throws.
- Every value in every table's printed range has a non-empty label (R16).
- The regeneration round-trip: each golden reproduced from its own
  recorded seed and inputs.
- Schema validation of the goldens and of the two examples.
- Property sweeps over many seeds asserting the book's invariants — every
  characteristic at or above 0 and no higher than its own formula allows
  (size 10, population 10, atmosphere 15, hydrographics 15, government 15,
  law level 20, technological index 18 after its cap), every lane four parsecs
  or fewer between worlds whose starport pair has a cell, no lane touching
  an X starport, no lane at a dash cell.
- Both directions on the readings: every `E00N` cited in the code and the
  documents resolves to a heading in `ERRATA.md`, and every heading is
  cited at least once.
- **The gate:** `task` runs formatting, `go vet`, golangci-lint with
  `default: all` and a curated disable list, NilAway, and `go test -race`.
  CI runs exactly `task`. **The toolchain is deliberately unpinned** — the
  gate is meant to fail when a tool moves rather than drift behind it, so
  a red gate on untouched code is the signal working. Answer the finding;
  do not pin a tool to silence it.
- **Coverage ratchets on uncovered statements per package**, not a
  percentage: a percentage can hold still while a guarded branch adds one
  covered statement and one uncovered. It fails in both directions.

## Decisions

- **The subsector is the record.** Not the world. Star mapping is
  subsector-scoped and lanes need neighbours.
- **The held printing governs.** Where the © 1977 text differs from the
  edition most people remember, implement the page as held; do not
  adjudicate editions from memory. Readings go in `ERRATA.md` with the
  page cite, never applied silently. The string of digits (E005) is the
  one a reader will notice first.
- **An empty subsector is a result.** A run whose eighty throws place no
  world produces a valid record with no worlds and no lanes. Nothing
  rerolls.
- **Types for identity, data for ranges.** Stated in full under "The
  domain"; it is the decision this design turns on.
- **There is no log.** The record carries outcomes. A referee who wants
  the throws shown makes them himself.

## Milestones

0. **The documents that govern the code**: `ERRATA.md`, carrying E001–E005
   with their page cites and stamping conditions, and `CLAUDE.md`. Both
   precede the first line of Go: a guard written after the code it guards
   is not a guard.
1. **The domain and its edges**: `dice`, the typed grid with its
   printed-grid distance test, `tables` loaded and validated with the
   second transcription in place, and the record with its marshalers and
   `subsector.schema.json`. Walking skeleton: the occurrence scan and
   starport types, written out as JSON — a map with no worlds detailed on
   it.
2. **The rest of star mapping**: bases and space lanes. Exit criterion: a
   living `COVERAGE.md` mapping every step of pp. 1–12 to its page cite,
   implementation, and test.
3. **World creation**: size, atmosphere, hydrographics, population,
   government and law level, then the technological index, with the floors
   and the one cap, the recorded clamp data, and the string of digits.
4. **The listing render and `batch`.**

## Deliverable documents

`PRD.md` (this document), `ERRATA.md`, `COVERAGE.md`,
`subsector.schema.json` with a minimal and a complete example, and
`CLAUDE.md`, which points here for the authority model and adds only what
an agent needs and a human reader does not.

## Sources

Book 3 _Worlds and Adventures_ pp. 1–12; Book 1 _Characters and Combat_
pp. 2–3 and p. 8. FFE reprints of the © 1977 text, as held in
`~/Documents/Traveller/Classic/`. Nothing else.

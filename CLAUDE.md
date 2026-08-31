# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working
with code in this repository.

## What this is

`ctworldgen`: a Go CLI that generates rules-accurate Classic Traveller
subsectors from Book 3's Worlds chapter (pp. 1–12, © 1977 text).

**What this tool is for is what a referee needs at the table.** The
alpha was played ([issue 1]) and its report is the working specification:
what held up is what must not break, what it did not want is not to be
built, and what it lacked is the backlog. `docs/ERRATA.md` holds the
recorded readings; `docs/COVERAGE.md` maps rules to implementation and
tests. Both are live.

`docs/PRD.md` is **historical**. It was the contract through
`v1.0.0-alpha.1` and governs nothing now; read it for why a thing is the
way it is, never for whether a thing may be built. Its scope fence in
particular is retired — several of its "Not in scope" bullets are open
backlog items.

[issue 1]: https://github.com/philoserf/ctworldgen/issues/1

**Status.** The engine walks the whole of pp. 1-12; `render` writes the
listing, which opens with a text map of the p. 3 grid, or a printable PDF
booklet (`--format pdf`) with the grid drawn and p. 2's lane lines joined;
`batch` produces N independent subsectors; and `sector` lays sixteen of
them on one 32x40 grid and throws for the lanes at their seams (E006). A sector's members
are unchanged -- member *i* of `sector --seed N` is the subsector `new
--seed N+i` writes -- which is the property that keeps a sector
trustworthy and is tested directly.

`docs/COVERAGE.md` is the live map of rule to code to test. The
technological levels tables of pp. 10-11 are the one thing inside pp. 1-12
that is not built (issue 1 #4).

There are two golden trees now -- the JSON records in `gen/testdata` and
the Markdown listings in `render/testdata` -- and both are driven from the
one roster in `internal/fixture`, so they cannot come to describe
different subsectors under the same name. `task regenerate` rewrites
both.

## The rule that matters most

Rules come only from the held PDFs in `~/Documents/Traveller/Classic/`.
**Never implement a Traveller rule from memory** — training-data
Traveller is mostly the 1981 revision and later editions, and the held
© 1977 page governs even where it differs.

This outlived the contract that first stated it, because it is not a
contract term. It is what the alpha report singles out as the reason the
output could be trusted at all: "The numbers are the page's numbers",
"Error messages cite the page — this bought trust in the rest of the
output before I had checked any of it", and "The errata loop works …
this is the best thing in the alpha." Page accuracy is a user need. What
retired with the PRD was its scope fence, not this.

**What is in authority**, and nothing else is:

- **Book 3 _Worlds and Adventures_ pp. 1-12**, the Worlds chapter. This
  is the ruleset. (pp. 10-11 are reference for play rather than a step of
  generation, and are not transcribed; that is a backlog item, not a
  boundary.)
- **Book 1 _Characters and Combat_ pp. 2-3** (the die roll conventions)
  and **p. 8** (the hexadecimal digit notation), which Book 3 uses
  without restating.

Not Books 2 and 4+, not the supplements, the Starter Edition, The
Traveller Book, JTAS, or the _Consolidated Errata_ PDF, and not the rest
of Book 3. Where a later printing differs from the held © 1977 text, the
held page governs — most visibly in the string of digits, which carries
no hyphen (E005), and in a population-0 world receiving a technological
index like any other (`ERRATA.md`, Noted discrepancies).

**Printed page N is PDF page N+5 in Book 3, and N+6 in Book 1.**

### The font trap

The held PDFs' embedded font maps the em-dash to the glyph `4` and the
minus sign to `3`. A text extraction therefore renders the jump routes
table's empty cells as the digit 4 — `B-E 4 4 4 4` for `4 — — —` — and
the size formula `2D − 2` as `2D32`. Both readings are wrong and both
look like data.

So every table is transcribed from a **visual** read of the page, and
then transcribed a second time inside the table package's tests, so the
two must agree. That second transcription is the check the font trap
needs, and it is where a new numeric table belongs.

The one exception is the descriptive **labels** of pp. 5-7 — "Feudal
Technocracy", "Dense, tainted". Those are editorial abbreviations of the
book's prose rather than transcriptions of it, so retyping them in a test
compares an abbreviation against itself. They are checked by re-reading
the page; the suite only asserts that every value in a printed range has
a non-empty label.

Three habits follow, and they are what an agent gets wrong:

- **Read a table off the page, visually** — `Read` the PDF with a `pages`
  range and look at it. Never `pdftotext`, never `pdfplumber`, never a
  grep of extracted text. The embedded font maps the em-dash to `4` and
  the minus sign to `3`, so an extraction produces wrong numbers that
  look like right ones.
- **Cite the printed page, in the code**, for every rule implemented.
- **Never resolve an ambiguity in silence.** If the page does not settle
  something, the reading goes in `docs/ERRATA.md` with its page cite and
  its stamping condition, and gets an E-number. If you find yourself
  choosing, you are writing errata.

## Clean room

Do not read, import from, or copy the sibling repos
`philoserf/ctchargen`, `philoserf/t5chargen`, `philoserf/t5`, or
`philoserf/traveller` unless the user explicitly asks. The _contracts_
those repos proved — provenance stamps, the COVERAGE and ERRATA
documents, the unpinned toolchain gate — are adopted; their code is not.

## When the gate goes red on code you did not touch

The toolchain is deliberately unpinned, so the gate is meant to fail when
a tool moves rather than drift behind it. That is the signal working.
Answer the finding; do not pin a tool or add a linter disable to silence
it.

## Two traps the PRD names and an agent will still hit

- **The hex grid parity.** Getting the offset-to-cube conversion
  backwards leaves every distance internally consistent and wrong by one
  for half the map, which no record-against-record test can catch. The
  test that catches it measures against the printed p. 3 grid. Never
  change the conversion without re-measuring there.

  The parity now lives in **three** places, and they must agree:
  `subsector.Hex.cube`, the text map's `render.gridLine`, and the drawn
  map's `render.mapFit.hexCenter`. Each has its own measurement against
  the page, because each can be flipped without the other two noticing.
  The drawn one is measured off the PDF it produced rather than off its
  own constants -- a check fed those agrees with a map drawn upside down
  -- and adjacency alone is symmetric under a flip, so the direction is
  anchored by two further assertions.
- **Dice-stream consumption order is load-bearing.** It fixes what a seed
  means. Two throws are deliberately _not_ made — a size-0 world's
  atmosphere and a size-0-or-1 world's hydrographics (p. 4) — and rolling
  and discarding a die there would shift every later world. So would
  reordering any step.

## A test that cannot fail looks exactly like one that passes

This has bitten five times now, and never once showed up as a failing
suite:

- The base and lane throws could have their sense inverted and every
  invariant still passed, because the checks only asked whether the lanes
  and bases that exist are legal.
- Five world-creation assertions were written, and the edit meant to call
  them from the sweep silently matched nothing. They sat in the file,
  defined and dead, and the suite went green.
- A roster check searched the whole listing for each world's hex, which
  the world's own detail page satisfies, so dropping half the roster
  passed.
- A label check had the same shape: another world carrying the same value
  answered for the one under test.
- The map's parity check ran on a record with **no worlds**, where every
  cell is the same width, so a bug that shifts a row only where a starport
  letter is drawn could not be expressed by the fixture at all. Its
  companion check read each cell at the position it found the label at, so
  it held content and never placement. Between them the two looked
  complete.

So the habit: **a new invariant is not done until a deliberate mutation
has been shown to kill it.** Invert a target, drop a column from a sum,
halve a loop -- then run the suite and read the failure. If it does not
name the thing you broke, the check is not holding what you think.

Three things make that harder here than elsewhere, and each has already
disguised a dead check:

- **Regenerate the goldens under the mutation.** `TestGoldens` compares
  against fixtures the code under test wrote, so a mutation moves them and
  the suite fails for the wrong reason. Run `task regenerate` first; if
  only the goldens complain, the invariant proved nothing.
- **Check that the mutation applies at all.** A mutation aimed at a hex
  that is not a world in any fixture is a no-op, and reads as a surviving
  mutant. Assert the edit changed the file, and prefer breaking something
  every fixture exercises.
- **The fixture has to be able to express the bug.** The mutation can
  apply and the assertion can be live, and the check is still blind if the
  record it was handed has no instance of the thing that breaks: a
  world-less map cannot show a mis-drawn world. Run a new invariant over
  the populated goldens, not only over the hand-built minimum. And when a
  mutation dies on some fixtures and survives on others, find out why
  before trusting the ones that died -- the map's first version was caught
  on three fixtures by luck, because their probe hex happened to carry a
  letter, and survived outright on the fourth.

## Commands

`task` is the whole gate — tidy, vet, golangci-lint, NilAway, `go test
-race`, and the coverage ratchet — and CI runs exactly `task`. Never add
a check to CI that the local gate does not run, and never add a tool to
the gate without also installing it there.

`task regenerate` rewrites both golden trees, the sector's seam golden,
and the shipped example, then you read the diff. `task ratchet:update`
records a new uncovered-statement baseline, which is for deliberately
unreachable code and test-support packages, not for skipping a test.

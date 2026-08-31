# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working
with code in this repository.

## What this is

`ctworldgen`: a Go CLI that generates rules-accurate Classic Traveller
subsectors from Book 3's Worlds chapter (pp. 1–12, © 1977 text).

**Read `docs/PRD.md` before doing any work here.** It is the contract:
the authority model, the requirements, the determinism rules, the record
shape, and the milestones all live there, and this file does not restate
them. `docs/ERRATA.md` holds the recorded readings; `docs/COVERAGE.md`
maps rules to implementation and tests.

**Status: all four milestones are complete.** The engine walks the whole
of pp. 1-12, `render` writes the subsector listing from a record, and
`batch` produces N independent subsectors. Every rule of pp. 1-12 has a
row marked done in `docs/COVERAGE.md`, and all five readings stamp the
records they govern.

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

`docs/PRD.md`, "Authority", is the full statement: which books and pages
are in authority, the printed-page-to-PDF offsets, and the `pdftotext`
font trap that makes a text extraction of any table unsafe. Read it
rather than working from this summary.

Three habits follow from it, and they are what an agent gets wrong:

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
- **Dice-stream consumption order is load-bearing.** It fixes what a seed
  means. Two throws are deliberately _not_ made — a size-0 world's
  atmosphere and a size-0-or-1 world's hydrographics (p. 4) — and rolling
  and discarding a die there would shift every later world. So would
  reordering any step.

## A test that cannot fail looks exactly like one that passes

This bit four times across milestones 2 to 4, and never once showed up as
a failing suite:

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

So the habit: **a new invariant is not done until a deliberate mutation
has been shown to kill it.** Invert a target, drop a column from a sum,
halve a loop -- then run the suite and read the failure. If it does not
name the thing you broke, the check is not holding what you think.

Two things make that harder here than elsewhere, and both have already
disguised a dead check:

- **Regenerate the goldens under the mutation.** `TestGoldens` compares
  against fixtures the code under test wrote, so a mutation moves them and
  the suite fails for the wrong reason. Run `task regenerate` first; if
  only the goldens complain, the invariant proved nothing.
- **Check that the mutation applies at all.** A mutation aimed at a hex
  that is not a world in any fixture is a no-op, and reads as a surviving
  mutant. Assert the edit changed the file, and prefer breaking something
  every fixture exercises.

## Commands

The Taskfile does not exist yet. When it does, `task` is the whole gate
and CI runs exactly `task` — never add a check to CI that the local gate
does not run, and never add a tool to the gate without also installing it
there.

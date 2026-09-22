# Changelog

## Unreleased

Bullets as the work lands. The release pass rewrites this into whatever the
release turns out to be.

## v1.0.0-beta.3 — 2026-09-08

The third beta. One commit since beta.2, and it is the last unbuilt clause
of Book 3 pp. 1-12.

What a referee sees: he can lay out his own geography. P. 1 offers the
world occurrence DM "on the whole subsector, or on broad areas within a
subsector", and only the first half was built; `new --occurrence-area
-1@0101-0805` is the second. An area is a rectangle of the grid's own
numbering with its own DM, the flag repeats, and two areas that share a hex
are refused before any die is thrown -- p. 1 offers +1 and -1 and gives no
basis for combining two of them, so the tool never invents one. Both
documents name the areas a record was generated under and cite the reading.
That reading is ERRATA E012, and with it every page of pp. 1-12 is built.

What a referee's existing records do: still reproduce. EngineVersion is
still 1, the schema is still 1, and no throw, no consumption order and no
RNG construction moved. A broad area varies the number each throw is read
against and never the throw itself -- the occurrence scan draws the same
one die per hex over all eighty in the same grid order it always did -- so
the four goldens that predate areas regenerate byte-identical, and so does
the shipped example. A record written by beta.1 still decodes and renders
here.

What changed underneath: `occurrence_areas` is a new record field, additive
and absent from a record with none, which is what keeps that byte-identity.
`starmap` owns the p. 1 DM rule in one place now, for the subsector DM and
an area's alike, and `Grid.HoldAreas` is the one definition of a legal set
of areas -- called by the engine's inputs and by the read path, because a
schema alone rejects nothing at read time, and because the corners and the
overlap are things JSON Schema cannot state at all.

What is not built: broad areas on a sector. An area is read against one
grid and a sector's sixteen members are each generated on their own p. 3
grid, so `sector` refuses them by name rather than dropping them quietly.
Issue 59 carries the two decisions that work turns on.

## v1.0.0-beta.2 — 2026-09-08

The second beta. Sixteen commits since `v1.0.0-beta.1`, and one of them is the change a referee will notice.

### What a referee sees

**Every world's technological index now carries a gloss.** Book 3 pp. 10-11 are read downward with their holes left as holes (ERRATA E009, E010), and T5 Core Book 2 pp. 230-232 supplies the band and the era **for description alone** (E011) — named in the document, changing no throw, no record field and no stamp. That was the last thing [issue 1](https://github.com/philoserf/ctworldgen/issues/1) asked for, and the one page of pp. 1-12 the engine had not read.

A malformed chart also now reports the same complaint on every run, rather than whichever of two the map iteration reached first.

### What your existing records do

**They still reproduce.** `EngineVersion` is still `1`, `schema_version` is still `1`, no throw and no consumption order moved, and the shipped example is **byte-identical** to the one beta.1 carried. A record written by beta.1 decodes and renders here.

### Underneath

- The read path holds a record to the whole of the schema it claims, and the engine's own output is held to the same contract across **six hundred seeds** rather than four goldens.
- The retired contract document is deleted, and the four decisions it alone recorded are kept in `CLAUDE.md`.
- About 200 lines of unreachable code are gone — along with **two checks that could not fail**, both found by mutating for something else: the sector index map had no golden at all, and `fitMap`'s height term was live but unasserted.
- `internal/audit` holds no non-test file, so production code cannot reach it even by mistake — the import fails to compile.
- A pass over every comment in the tree: comments now say what the code **is** rather than what it was, and the doc comments read as Go doc comments — short synopsis lines, and doc links that pkg.go.dev will render.

### New documents

- **`THEORY.md`** — what a maintainer has to hold in mind to change this without damaging it, and where that account is uncertain.
- **`walkthrough.md`** — a tour from entry point to output. Every fenced block is executable; `uvx showboat verify walkthrough.md` re-runs them all.

### Still open

Per-area occurrence DMs — a rift in one corner, a cluster in the other ([#5](https://github.com/philoserf/ctworldgen/issues/5)).

## v1.0.0-beta.1 — 2026-09-08

The first beta. Three fixes, and the first of them is a crash on ordinary input: a note carrying an apostrophe typed on a Mac panicked the PDF renderer.

```sh
go install github.com/philoserf/ctworldgen/cmd/ctworldgen@latest
```

### Fixed

- **`render --format pdf` no longer panics on the referee's own note** ([#17](https://github.com/philoserf/ctworldgen/issues/17)). macOS and iOS turn every typed apostrophe into U+2019, so `notes: "the world's capital"` written in any Apple text field wrote a record that crashed the renderer outright.

  `render` crossed into fpdf by two doors that had to agree and did not. Drawing goes through Windows-1252 **bytes**, which the core fonts are indexed by; measuring hands `SplitText` **runes**, which it indexes into the same 256-entry table. The guard meant to keep measuring safe asked whether Windows-1252 could _carry_ a rune rather than whether the rune was below 256, so the whole 0x80-0x9F block -- curly quotes, the em-dash, the ellipsis -- passed through untouched and indexed `cw[8217]` into a slice of 256. A rune the encoding _cannot_ carry took the question-mark branch and was fine: the crashing set was exactly the characters it handles best.

  One function now owns the alphabet the page can draw, and both doors use it. Wrapping measures a proxy whose runes are the ordinals of the bytes that will be inked, so the apostrophe stays an apostrophe and is measured at its own width, rather than becoming the question mark a guard that merely stopped the panic would have left in every note a Mac wrote.

- **`--force` no longer truncates the previous record before the new one is known to write** ([#19](https://github.com/philoserf/ctworldgen/issues/19)). The target was opened `O_TRUNC` and only then written, so a write that failed partway -- a full disk, a signal, an I/O error -- left a truncated or empty file where the record had been. That is the one file the tool asks a referee to keep: it is what `render` reads, and it may carry names and notes written into it over several sessions. The new record is now written beside the old one and renamed over it, so a write that fails leaves the record he already had.

- **A malformed tail is no longer reported as a second document** ([#20](https://github.com/philoserf/ctworldgen/issues/20)). Anything `Decode` found past the record that was not `io.EOF` became "more than one document in the record read; a record is one JSON document" -- a specific and confident claim about a file that may hold no second document at all, with the decoder's own error and its byte offset discarded. `{...}]` said the record was duplicated. The three outcomes are now separated: a syntax error comes back as itself with the offset that says where to look, a reader failure comes back as itself, and a genuine second document keeps the sentinel and names the token it found.

### Changed

Nothing a referee sees.

- The coverage ratchet is eleven lines of `awk` and a `diff` rather than 391 lines of Go ([#16](https://github.com/philoserf/ctworldgen/issues/16)).
- `CLAUDE.md` writes down the ten things that look removable and are not ([#34](https://github.com/philoserf/ctworldgen/pull/34)) -- each has a visible cost and an invisible benefit, which is how they come to be deleted.
- `golang.org/x/text` 0.14.0 to 0.41.0 ([#14](https://github.com/philoserf/ctworldgen/pull/14)).

### Compatibility

**No seed's meaning moves.** `engine 1`, `schema 1`, `ruleset ct-1977-book3-pp1-12` are unchanged, no golden record or listing moved, and every record written by an earlier alpha reproduces and renders. The em-dashes the documents set themselves are drawn exactly as before; the encoding fix reaches only text a referee wrote.

Two differences in `--force`, both the cost of writing beside the target and renaming:

- a rename needs write permission on the **directory** rather than on the file, so `--force` into a directory that cannot be written now fails where a truncating open succeeded -- and that case could not have kept the old file either;
- where the target is a **symbolic or hard link**, a truncating open wrote _through_ it into the file it named; a rename replaces the link with the new record and leaves what it pointed at alone. This one the old code handled and this one does not.

### Still not built

- **The technological levels tables of pp. 10-11** ([#4](https://github.com/philoserf/ctworldgen/issues/4)). Every world carries a technological index and no gloss of it. This is the last untranscribed thing inside the declared pp. 1-12, and the only outstanding finding from the alpha report.
- **Per-area occurrence DMs** ([#5](https://github.com/philoserf/ctworldgen/issues/5)).

`docs/COVERAGE.md` is the live map of rule to code to test and carries a row for what is not built as well as for what is.

**Full changelog**: https://github.com/philoserf/ctworldgen/compare/v1.0.0-alpha.3...v1.0.0-beta.1

## v1.0.0-alpha.3 — 2026-09-01

The third alpha. One change, and it is the follow-on the second alpha's own thread ranked: a sector's documents are no longer one large subsector.

```sh
go install github.com/philoserf/ctworldgen/cmd/ctworldgen@latest
```

### Changed

- **A sector renders as its sixteen sub-sectors.** `render` now opens a sector with an index of the whole grid and a table of its sixteen, then carries the sixteen sub-sector listings themselves -- each on its own p. 3 grid, ringed by one hex of its neighbours so a lane crossing a seam has a far end that can be looked up, and each headed with the seed that writes it standalone. The alpha referee read sixteen listings and ran one of them; `sector` had taken that away.

  |                         | Before                                                    | After                                                  |
  | ----------------------- | --------------------------------------------------------- | ------------------------------------------------------ |
  | Text map                | 161 columns, 82 lines                                     | index 39 columns; member maps 51                       |
  | Drawn sector map        | 1,280 hexes, four-digit numbers overrunning their own hex | index: no numbers, six heavy seams, each band numbered |
  | Roster / lanes / detail | one flat 662 / 1,879 / 662                                | partitioned across the sixteen                         |
  | Booklet                 | 143 pages                                                 | 168                                                    |

  Member _i_ of `sector --seed 42` carries exactly the worlds `new --seed 47` writes -- the identity E006 has always promised, now visible in the document rather than only in a test.

- **Type inside a hex is a fraction of the hex** rather than a fixed size, calibrated so a sub-sector's map is drawn exactly as it was. A sector booklet used to draw 4.5pt numbers in 17pt hexes.

### Fixed

- **The booklet names ERRATA E007 where it suppresses.** E007's own reading says both documents say what they suppressed and how many, and _both name this entry_. The listing did; the booklet never had. This is the only change to a sub-sector's output.

### The reading

`docs/ERRATA.md` **E008**, in four parts: how a member is identified, the one-hex ring and why one and not four, a crossing lane appearing under both its sub-sectors, and the sector map being an index. Like E007 it governs the documents rather than the generation, so it is never stamped -- the record is unchanged and carries no member field.

The familiar A-through-P lettering of subsectors is deliberately **not** used: it comes from the 1981 revision and later, not from the held © 1977 pages. A member is named by its index, its hex range and its seed.

### Two holes the suite had been passing

- **Nothing pinned the member numbering.** `Place` and `MemberOf` are exact inverses, so every test that put a hex through one and read it back through the other passed under a consistent transpose -- and both band widths are even, so the p. 3 parity survived it too. Only a golden moved, which is not a check. E006 part 1 is now transcribed by hand.
- **A sector could drop every interior route and the engine's own package stayed green.** The member identity compared worlds; the seams golden pins only the lanes that cross a border; the interior lanes fell between them. They are compared now, as are the two sort orders.

### Compatibility

**No seed's meaning moves.** `engine 1`, `schema 1`, `ruleset ct-1977-book3-pp1-12` are unchanged, no golden record moved, and every record written by an earlier alpha reproduces and renders. A sub-sector's Markdown listing is byte-identical.

### Still not built

- **The technological levels tables of pp. 10-11** ([#4](https://github.com/philoserf/ctworldgen/issues/4)). Every world carries a technological index and no gloss of it.
- **Per-area occurrence DMs** ([#5](https://github.com/philoserf/ctworldgen/issues/5)).

`docs/COVERAGE.md` is the live map of rule to code to test and carries a row for what is not built as well as for what is.

**Full changelog**: https://github.com/philoserf/ctworldgen/compare/v1.0.0-alpha.2...v1.0.0-alpha.3

## v1.0.0-alpha.2 — 2026-09-01

The second alpha. Everything the first alpha's report ([#1](https://github.com/philoserf/ctworldgen/issues/1)) asked for, except the technological levels tables.

```sh
go install github.com/philoserf/ctworldgen/cmd/ctworldgen@latest
```

### Breaking

- **`batch` is gone.** `sector` replaces it when the sixteen belong on one grid, because it throws for the routes at their seams -- the routes sixteen independent subsectors can never have. When you want independent records instead, the seeds are simply consecutive; the README shows the loop. (#3)

### New

- **`sector`** lays sixteen subsectors on one 32x40 grid, 0101 through 3240, and throws for the commercial routes crossing between them (ERRATA E006). Members are unchanged: member _i_ of `sector --seed N` is exactly what `new --seed N+i` writes.
- **The map is drawn.** The listing opens with a text map of the p. 3 hex grid, and `render --format pdf` writes a printable booklet -- map beside its roster, then the routes and a page of detail per world. The booklet is the one output that draws p. 2's "line connecting the two worlds on the map". (#2)
- **Legible lanes.** A dense subsector throws 160 routes over 46 worlds. By default a lane whose worlds are already joined by shorter lanes is not drawn -- p. 2's own "may be ignored" (ERRATA E007). About 46% fewer lanes, reachability unchanged. `--lanes all` draws every one. The record always carries them all: this is a decision about ink, not dice, so no seed's meaning moves. (#8)
- **`notes`** on a world and on the record are the referee's. Never generated, never read back, and they survive re-rendering into both the listing and the booklet. (#7)

### Answering the first alpha

- **Names reach everywhere** -- detail headings and both lane columns, not just the roster table, so a listing can be read aloud at the table.
- **A bare characteristic line says why it is bare.** 802 of them in the alpha, for two different reasons; the scrupulous ones no longer read as bugs.

### Still not built

- **The technological levels tables of pp. 10-11** ([#4](https://github.com/philoserf/ctworldgen/issues/4)). Every world carries a technological index and no gloss of it.
- **Per-area occurrence DMs** ([#5](https://github.com/philoserf/ctworldgen/issues/5)).

`docs/COVERAGE.md` is the live map of rule to code to test and carries a row for what is not built as well as for what is. `docs/PRD.md` is retired -- read it for why a thing is the way it is, never for whether a thing may be built.

**Full changelog**: https://github.com/philoserf/ctworldgen/compare/v1.0.0-alpha.1...v1.0.0-alpha.2

## v1.0.0-alpha.1 — 2026-08-31

Tagged without published release notes.

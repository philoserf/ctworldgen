# A Theory of ctworldgen

What you need to hold in mind to change this program without damaging it.
It is not a tour of the packages; `CLAUDE.md` argues the decisions and
`WALKTHROUGH.md` walks the call chain. This says what the thing _is_.

## What is being modelled

Not "a random world generator". The subject is a **procedure printed on
twelve pages in 1977**, and the artifact is **the referee's notebook page**
that procedure tells him to keep. Book 3 p. 4 asks him to maintain "at
least one (and preferably several) pages in a central notebook"; this tool
writes that page for him and hands it back as something he can still write
in.

That framing decides almost everything downstream, and two consequences are
worth stating before anything else.

**The unit of record is the subsector, not the world.** A world's
characteristics are thrown in isolation, but the two facts that make a
subsector a subsector are not: whether a hex holds a world at all is thrown
against the eighty-hex grid p. 3 prints, and a commercial route is a
statement about a _pair_ of worlds and cannot be evaluated until both
exist. A world-shaped record could carry neither. So `starmap.Record` is
one map, and a world is a row inside it. Everything that looks like an
architectural choice here — `gen` filling a whole record before returning
it, `render` taking a record rather than a world, `sector` layering sixteen
records rather than 1,280 worlds — is that one decision propagating.

**The referee outranks the tool.** `Record.Notes`, `World.Notes` and
`World.Name` are fields the engine writes into and never reads back. The
record refuses every JSON key it does not define, which is what makes its
provenance stamps worth anything, so the places a referee may write had to
be _named_ rather than left as an escape hatch. `Marshal` deliberately does
not call `Validate`, and `carriesThisToolsProvenance` deliberately does not
recompute `digits` from the values beside it: he is allowed to adjust a
world to suit his campaign, and the digits are what he reads. The tool is
scrupulous about what it generates and incurious about what he does to it
afterwards.

## The load-bearing ideas

### The seed is the record's entire causal history, and consumption order is that history's schema

Every throw in a run comes from one `dice.Stream` seeded from one number,
and that number is written into the record. There is no log — deliberately.
A log would be a second, driftable copy of something the seed already
determines exactly.

The price of having no log is that **the order in which dice are consumed
is part of the public contract**, as much as the record's field names are.
Add a throw, remove one, or reorder two, and every record anyone holds
still parses, still renders, and silently no longer reproduces. That is the
one corruption undetectable from the file itself, which is why it is the
one thing `EngineVersion` exists to announce.

This is why the code refuses dice it does not need rather than rolling and
discarding them. A size-0 world's atmosphere and a size-0-or-1 world's
hydrographics are not thrown (`gen.detail`). A jump-routes cell printing a
dash states no target, so no die is thrown at it (`gen.routes`). A broad
area varies the _number a throw is read against_ and never the throw, its
order, or its eighty dice (`gen.scan`, `gen.dmAt`) — which is why a record
generated with no areas is byte-for-byte what the tool wrote before areas
existed. `dice.Stream.Die` is documented as _exactly_ one `IntN(6)` draw
plus one, because `IntN(36)` under the same seed would be the same
generator producing an entirely different subsector.

`gen.seams` being O(n²) is the same idea wearing a different costume. The
loop's visit order _is_ the order the seam stream is drawn in, so a spatial
index reaching the same pairs faster in a different order writes a
different sector from the same seed. The safe optimisation is an early-out
inside the existing loop; anything that reorders is a rule change.

If you learn one thing from this document: **a performance change in `gen`
is a rules change until proven otherwise.**

### Where the page is silent, the silence gets a name

The held 1977 text is ambiguous or silent in about a dozen places that
matter — where the base throws sit in the procedure, what order hexes are
scanned in, what "already present" means for a route, what a hole in the
technological levels table signifies. The rule of this codebase is that
**you may not resolve such a silence quietly**. A reading goes in
`docs/ERRATA.md` with its page cite, gets an E-number, and states the
condition under which a record stamps it.

The `errata` array on a record is therefore not metadata. It is the record
saying which readings actually governed _it_ — E004 only where a clamp
bound, E003 only where there were two worlds to pair, E012 only where the
referee drew an area. `internal/audit/errata_test.go` enforces the loop in
both directions: every E-number cited anywhere in the code or the prose
resolves to a reading, and every reading is cited somewhere.

**The stamping condition carries a distinction that is easy to lose.** A
reading that governs _generation_ is a property of the record and is
stamped. A reading that governs the _documents_ — E007's legible lanes,
E008's member sections, E009 and E010's downward reads, E011's borrowed
gloss — is never stamped, because the same record renders two ways
depending on what the referee asked `render` for. Stamping one would make
the record a claim about ink. If you add a reading, deciding which kind it
is _is_ the design work.

### Two authorities, and only one of them is open

For **generation** the source list is closed: Book 3 pp. 1–12, plus Book 1
pp. 2–3 and p. 8 for conventions Book 3 uses without restating. Nothing
else. Not later printings, not the supplements, not the consolidated
errata. This is not pedantry about editions — a seed's meaning is defined
by that list, so widening it silently invalidates every record in
existence.

For **description** there is exactly one opening, taken once and on the
record (E011): T5 Core Book 2 pp. 230–232 may gloss a technological index,
because pp. 10–11 read faithfully still give a referee ten proper nouns of
1977 shorthand and no sense of the world in front of him. The permission
comes with three conditions — named in the document where used, written
down in the errata with pages, and changing no throw, no record field and
no stamp. `tables.Borrowed` and `tables.Held` are separate Go types so that
this boundary is visible in the code and not only in the prose.

The pressure this design is under is constant and one-directional: a held
page is terse, a later edition is clear, and reaching for the later edition
is always locally reasonable. Terseness is not the test. The test is
whether the held page can be used at the table at all.

### The page is read with eyes, not with a parser

The held PDFs' embedded font maps the em-dash to the glyph `4` and the
minus sign to `3`. A text extraction of the jump routes table renders its
empty cells as the digit 4 — `B-E 4 4 4 4` for `4 — — —` — and reads the
size formula `2D − 2` as `2D32`. Both readings are wrong and, crucially,
**both look like data**.

There is no way to detect this from the extracted text. So every table is
transcribed from a visual read of the page and then transcribed a _second_
time inside `tables`' own tests, and the two must agree. That duplication
is not redundancy for its own sake; it is the only check the font trap
admits. The single exception is the descriptive _labels_ of pp. 5–7, which
are editorial abbreviations of the book's prose rather than transcriptions
of it — retyping an abbreviation compares it against itself. Proper nouns
are never that exception.

### Types carry identity; data carries ranges

The dividing rule is sharp and worth memorising, because it looks like an
oversight from either side.

A **type** exists where something either is or is not the thing:
`starmap.Hex` (an identifier outside 0101–3240 is not a hex),
`starmap.Digit` (a character outside the p. 2 alphabet is not a digit),
`Starport`, `Characteristic`, `Parsecs`, `dice.Target`. Each parses at the
program's edge, marshals to the string the record prints, and makes a class
of defect uncompilable.

A **value range** stays `int`. Size, atmosphere, hydrographics, population,
government, law level and the technological index are plain integers, and
`type Atmosphere int` bounded 0–12 would be wrong twice over: it would put
the p. 5 table's last row into Go source beside the data file and the test
that transcribes it — a third copy of a number the book prints once — and
R14 caps nothing at a descriptive table's last row, so a generated
atmosphere legitimately reaches 15 and the type would reject a legal value.

**A range in a type is a rules claim, and a rules claim belongs on a page
with a cite.** That sentence resolves every borderline case here. `Hex` and
`Digit` sit on the type side despite encoding printed rules because a grid
position and a notation are identity, not value.

### The record is the entire interface between throwing and drawing

`render` does not import `gen`, and that is the architecture rather than an
accident of layering. Everything that throws a die is upstream of the JSON
record; everything that reads a descriptive table is downstream. A record
written before a feature existed still renders — `Decode` fills in the p. 3
grid for records written before grids were recorded, because the alpha's
referee has sixteen such files.

Every other package edge is one the compiler already refuses as an import
cycle, which is why `.golangci.yml` writes depguard rules for none of them
and only for the two constraints that would otherwise compile.

## The seams

**The JSON record** is the one that matters, and it is defended twice over.
`docs/record.schema.json` states the shape and `starmap.Record.Validate`
re-states it in Go, because _a schema alone rejects nothing at read time_.
That doubling is deliberate and load-bearing: the schema ships for other
tools, the Go checks bind this one, and `internal/audit/schema_test.go`
holds the two together by validating everything the engine writes against
the published schema.

Unknown-field rejection looks like the whole of the read-path defence and
is not. It catches a record from a _newer_ schema that added a field. A
record claiming a different schema version, ruleset, or generator parses
perfectly cleanly — every field reads, the listing renders — and would
report a subsector under provenance stamps that are not true of it. That is
what `carriesThisToolsProvenance` exists for, and it is the check most
likely to look redundant to someone tidying up.

**The two grids.** There are exactly two shapes a record takes: the p. 3
sub-sector grid, and the sector of sixteen (E006). `Grid.IsSector()` owns
that question so that `render` does not compare structs to `SectorGrid()`
at four sites, where a dimension check would read as arithmetic and mean a
kind.

**The sector layer** is the seam under the most tension, and its integrity
rests on one arithmetic accident: a sub-sector is **eight** columns wide,
and eight is even, so the p. 3 column parity survives translation onto the
sector grid. An odd band width would flip the parity of every second band
and silently change interior distances. That is why `starmap.Place` is
exported — so the property can be asserted against the translation the
engine actually uses rather than against a second copy of the arithmetic in
a test — and why no member field is stored anywhere: `MemberOf` reads the
band straight back off the hex, so the decomposition cannot disagree with
the grid.

**The hex parity itself** lives in three _measured_ places that must agree
and that each can be flipped without the other two noticing:
`starmap.Hex.cube`, `render.gridLine`, and `render.mapFit.hexCenter`. Each
has its own measurement against the printed page, and the three test
harnesses stay separate on purpose — a merged harness compares the three
against each other and passes when all three are flipped together. Getting
this wrong leaves every distance internally consistent and wrong by one for
half the map.

Correction, from later in the same review that produced this document:
there are **four** encodings, not three. `render.indexLine` — the sector
index map — is a fourth, and nothing measures it. `code-reduction`
mutation-tested it: flipping its indent to the wrong parity leaves
`go test ./render` green, which is precisely the failure this section says
cannot be caught by comparing copies against each other. `CLAUDE.md` says
three and is wrong on the count for the same reason this paragraph
originally was. See `.issues/indexline-is-a-fourth-copy-of-the-page-three-parity-that-nothing-measures.md`,
which argues the right answer is to delete the copy rather than measure it.

**The PDF font**, covered above, is a seam with the outside world that no
amount of tooling makes safe.

**The fences that would otherwise compile** are the two depguard rules: the
command may not reach `tables` or `dice`, and production code may not reach
`internal/fixture`. `internal/audit` holds no non-test file at all, which
makes it unimportable by construction — a firmer fence than a lint rule.

## What the shape accommodates, and what it resists

**Accommodates easily.** A new descriptive table (add JSON under
`tables/data/`, a loader, and the second transcription in the tests). A new
output format (`render` gains a typesetter; the record is untouched). New
referee-authored fields (name them in the record, mark them `omitempty`, and
old records stay byte-identical). A new reading of a silent page (errata
entry, E-number, stamping condition, cite at the point of use). Reading
older records — that is what `Decode`'s grid backfill is for.

**Resists, and should.** Anything that moves the dice stream. Concurrency —
the sixteen members are generated in sequence _because_ their independence
is a correctness property, and the program has no concurrency model to
extend. Bounding a characteristic in a type. Widening the description
authority past T5. Splitting `render` into two packages: the two
typesetters share a middle (`bullets`, `member`, `legible`, `summary`,
`named`, `bases`) and that middle is what stops the documents diverging —
they did diverge once, when the bullet list was written out twice and a
change to one was invisible to the other's tests.

**Where a maintainer who understood this would look first**, for the two
changes most likely to be asked for next:

- _A new rule from a page not yet implemented_ — `docs/COVERAGE.md` maps
  rule to code to test, and the p. 12 checklist order in `gen.Generate` is
  where a new step has to be placed. Placing it anywhere other than where
  the checklist puts it is a stream change.
- _Broad areas on a sector_ (issue #59) — the clipping is mechanical, but
  it lands on E006 part 1's identity ("member _i_ is `new --seed N+i`") and
  on E008's member headings, which currently print a bare seed. Both are
  errata text, not code, and both are read by a referee as promises. That
  is the actual work; the rectangle arithmetic is not.

**Where someone who did _not_ understand it would cause damage.** Deleting
`render.memberSeed`'s unreachable bounds guard (it is the proof gosec's
G115 accepts, and removing it costs a lint disable the project forbids).
Inlining the unmatched error sentinels (what they buy is one definition per
rule of a message carrying a page cite, and the alpha report names
page-citing errors as what bought trust in the output before anything else
had been checked). Merging the three parity harnesses. Adding `-race`'s
absence back. Optimising `gen.seams`. Every one of these is locally
defensible and globally wrong, which is why `CLAUDE.md` keeps a standing
list of them.

## A structural hazard peculiar to this codebase

Most of what this program asserts is an invariant over dice, and **an
invariant test that cannot fail looks exactly like one that passes.** This
has happened repeatedly here and has never once surfaced as a red suite:
checks that only asked whether the routes that exist are legal (so
inverting the throw's sense passed), assertions defined and never called, a
roster check satisfied by the world's own detail page, a map parity check
run on a record with no worlds where every cell is the same width.

The habit that follows is not optional: **a new invariant is not done until
a deliberate mutation has been shown to kill it** — and three things make
that harder here than elsewhere. `TestGoldens` compares against fixtures
the code under test wrote, so a mutation moves them and the suite fails for
the wrong reason (regenerate first). A mutation aimed at something no
fixture exercises is a no-op that reads as a surviving mutant. And the
fixture has to be able to _express_ the bug — a world-less map cannot show
a mis-drawn world.

## Uncertainties and tensions

Marked clearly, because these are where I am inferring from code and could
be wrong.

**Five of Book 3 pp. 10–11's ten columns never reach a document.**
`render.technological` prints Personal, Armor, Computers, Air and Space;
Special, Communication, Water, Land and Fuels are transcribed, validated at
load, transcribed a second time in the tests, and dropped. I cannot tell
from the code whether this is an editorial choice about line length or an
unfinished pass, and nothing in the errata or `CLAUDE.md` claims it. Filed
as `pp-ten-eleven-gloss-drops-five-held-columns`.

**The `internal/fixture` fence does not cover `render`.** The depguard rule
lists `dice`, `starmap`, `tables`, `gen` and `cmd/ctworldgen`, not
`render` — the largest production package. No production file currently
imports the fixtures, so this is latent. Filed as
`depguard-fixture-fence-omits-render`.

**`gen.CrossingRoutes` and `render.ends` are two implementations of the same
predicate**, on opposite sides of the record boundary, and they must agree.
This is a _consequence_ of the boundary rather than a defect of it —
`render` may not import `gen`, so the predicate has to exist twice — but
`CrossingRoutes`'s doc comment justifies its export partly as "what the
listing would highlight", and the listing does not use it. Not filed: the
duplication is forced by an architecture I think is right, and the comment
is written in the subjunctive.

**I could not verify the transcriptions against the source PDFs.** Every
claim here about what a page prints is taken from the code, the errata and
the tests, which agree with each other. If all three were transcribed from
the same bad read, nothing in this repository would notice. The visual-read
discipline is the only defence and it lives outside the tooling.

**The `--force` / `replaceFile` trade-off is stated but I did not test it.**
The rename-over-temp pattern means `--force` into a directory the referee
cannot write fails where a truncating open would succeed, and replaces a
symlink rather than writing through it. Both are argued in the source
comment as acceptable; I have no evidence either way about whether a
referee has hit them.

**The theory of `Notes` may be thinner than it looks.** Both `Notes` fields
are write-only from the tool's perspective and both are flattened to one
line at render time. That is coherent, but nothing enforces that a
hand-edited record's notes survive a regenerate cycle, and I could not find
a test that round-trips a referee-annotated record through
`render --format pdf`. It may exist and I may have missed it.

## Index

| #   | Severity | Issue                                         | Primary location                              |
| --- | -------- | --------------------------------------------- | --------------------------------------------- |
| 1   | medium   | `pp-ten-eleven-gloss-drops-five-held-columns` | `render/render.go:545`, `docs/COVERAGE.md:58` |
| 2   | medium   | `depguard-fixture-fence-omits-render`         | `.golangci.yml:46`                            |

**Total: 2 issues (0 critical, 0 high, 2 medium, 0 low)**

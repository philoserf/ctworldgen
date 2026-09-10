# ctworldgen Walkthrough

*2026-09-10T12:16:45Z by Showboat 0.6.1*
<!-- showboat-id: a6f5904d-7504-4533-98f7-20648e5723dc -->

## Overview

`ctworldgen` generates Classic Traveller subsectors from the Worlds chapter of
Book 3, pages 1-12 of the 1977 text. It is a Go CLI with no network, no
database, no concurrency and no configuration file. You give it a seed; it
throws dice against the book's tables and writes a JSON record. A second
subcommand turns that record into a Markdown listing or a printable PDF
booklet.

Two facts shape everything below, and it is worth holding them from the start.

**The record is the subsector, not the world.** Whether a hex holds a world at
all is thrown against the eighty-hex grid page 3 prints, and a commercial
route is a statement about a *pair* of worlds. Neither fits inside a
world-shaped record, so one record is one map and a world is a row inside it.

**The order in which dice are drawn is part of the contract.** There is no log
of throws — the seed reproduces them exactly — so adding, removing or
reordering a single throw changes what every existing seed means. You will see
the code go out of its way to *not* draw a die several times.

The tour follows one seed from the command line to a rendered page, then
doubles back for the sector layer.

## Architecture

Six packages. The dependency graph is the design statement, so read it before
anything else — it is derived from the source rather than drawn by hand:

```bash
go list -f '{{.Name}} {{join .Imports ","}}' ./dice ./starmap ./tables ./gen ./render ./cmd/ctworldgen | sed 's|github.com/philoserf/ctworldgen/|@|g' | awk '{n=$1;d="";c=split($2,a,",");for(i=1;i<=c;i++)if(a[i]~/^@/){sub(/^@/,"",a[i]);d=d" "a[i]}printf "%-6s ->%s\n",n,(d==""?" (nothing of ours)":d)}'
```

```output
dice   -> (nothing of ours)
starmap -> dice
tables -> dice starmap
gen    -> dice starmap tables
render -> starmap tables
main   -> gen render starmap
```

Two things in that graph are load-bearing.

`render` does not import `gen`. The JSON record is the *entire* interface
between throwing dice and drawing pages: everything upstream of the record
generates, everything downstream describes. A record written before a feature
existed still renders, and no renderer can accidentally re-throw a die.

`dice` imports nothing of ours. It is Book 1's die-roll conventions and one
seeded stream, and it knows nothing about worlds.

The three `internal/` packages are omitted above because they are not part of
the product: `internal/fixture` is the one roster both golden trees are
generated from, `internal/cmd/regenerate` rewrites those goldens, and
`internal/audit` holds *no non-test file at all* — which makes it unimportable
by construction, a firmer fence than a lint rule.

## The entry point

`cmd/ctworldgen` is flags and file I/O and nothing else. `run` is the whole
dispatch:

```bash
sed -n '81,102p' cmd/ctworldgen/main.go
```

```output
func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, usage)

		return errNoSubcommand
	}

	switch args[0] {
	case "new":
		return newCmd(args[1:], stdout, stderr)
	case "sector":
		return sectorCmd(args[1:], stdout, stderr)
	case "render":
		return renderCmd(args[1:], stdout, stderr)
	case "version":
		return versionCmd(stdout)
	default:
		_, _ = io.WriteString(stderr, usage)

		return fmt.Errorf("%w %q", errUnknownSubcommand, args[0])
	}
}
```

`new` and `sector` share a body, `singleRecordCmd`, because they differ in only
two ways: which pass of the engine fills the record, and whether broad areas
may be given at all. Both must record a seed, and this is the single place in
the program where a number comes from anywhere but the stream:

```bash
sed -n '189,199p' cmd/ctworldgen/main.go
```

```output
	// A seed is always recorded, so a run is reproducible after the fact.
	// Drawing one from OS entropy is the single exception to the
	// seeded-stream rule, and it happens before the engine starts.
	if !isSet(flags, "seed") {
		drawn, drawErr := entropySeed()
		if drawErr != nil {
			return drawErr
		}

		*seed = drawn
	}
```

So `--seed 0` is an explicit choice and not a request for a random one — the
flag's *presence* is what is tested, via `flags.Visit`, not its value.

The drawn seed is masked into a smaller range than `uint64` allows, and the
reason is worth reading in full because it is the kind of hazard that never
shows up in a test:

```bash
sed -n '27,33p' cmd/ctworldgen/main.go
```

```output
// maxSafeSeed is 2^53 - 1, the largest integer an IEEE-754 double holds
// exactly. A drawn seed is kept inside it because a record whose seed has
// been rounded by a reader that parses JSON numbers as doubles reproduces
// a different subsector -- the one corruption this record cannot afford,
// and a silent one. An explicit --seed is deliberately not bounded: it is
// the operator's own number, and Go reads it back exactly.
const maxSafeSeed = 1<<53 - 1
```

## The dice

`dice` is 58 lines and every one of them is a contract. The whole package:

```bash
sed -n '29,58p' dice/dice.go
```

```output
// NewStream seeds a PCG generator from the seed a record carries. Both
// words take that seed, so the one number reproduces the whole stream.
func NewStream(seed uint64) *Stream {
	// A deterministic generator is the point: a record must reproduce from
	// its recorded seed, which a cryptographic source cannot do.
	return &Stream{r: rand.New(rand.NewPCG(seed, seed))} //nolint:gosec // reproducibility, not secrecy
}

// Die throws one die.
//
// One die is exactly one IntN(6) draw plus one. This is as load-bearing
// as the consumption order it feeds: IntN(36), or a masked Uint64, would
// be the same generator under the same seed and would produce an entirely
// different subsector.
func (s *Stream) Die() int { return s.r.IntN(faces) + 1 }

// D2 throws two dice, the first die and then the second. B1 p. 2 makes
// two dice the unqualified throw.
func (s *Stream) D2() int { return s.Die() + s.Die() }

// Target is a throw target of the form N+: the throw succeeds when it is
// equal to or greater than N.
//
// N+ is the only target kind Book 3 pp. 1-12 uses. World occurrence is
// 4+, the base throws are 7+ through 10+ (p. 5), and a jump routes cell
// states a number the one die throw must equal or exceed (p. 2).
type Target int

// Met reports whether a throw meets the target.
func (t Target) Met(throw int) bool { return throw >= int(t) }
```

Note what `Die` is *not* allowed to become. `IntN(36)` would be the same
generator, under the same seed, drawing the same underlying bits — and would
produce an entirely different subsector, because the mapping from bits to
faces changed. `D2` is deliberately two `Die` calls rather than one draw, for
the same reason.

`Target` is the only kind of throw the Worlds chapter uses: N-or-better. World
occurrence is 4+, the base throws run 7+ through 10+, and every stated cell of
the jump routes table is a one-die target. `Met` is the entire rule.

## Generation, pass by pass

`gen.Generate` walks the page 12 checklist. Each pass is a complete sweep over
the subsector before the next begins, and within a pass the hexes run in
ascending grid number. That ordering is a *reading* of a page that does not
say — which is why the first thing the function does after building the record
is stamp it:

```bash
sed -n '100,155p' gen/gen.go
```

```output
func (e *Engine) Generate(inputs Inputs) (*starmap.Record, error) {
	err := inputs.Validate()
	if err != nil {
		return nil, err
	}

	record := starmap.New(inputs.Seed, inputs.Name, inputs.OccurrenceDM, inputs.OccurrenceAreas)
	stream := dice.NewStream(inputs.Seed)

	// The order of the passes, and of the hexes within them, is a reading.
	record.Stamp("E002")

	// 1.A. Throw for each hex; 4, 5, or 6 indicates a world is present.
	// The referee's DM applies to the whole subsector, or to broad areas
	// within one (p. 1, ERRATA E012).
	hexes, err := scan(stream, inputs.OccurrenceDM, inputs.OccurrenceAreas)
	if err != nil {
		return nil, err
	}

	// A broad area is a reading of a silence, and there was one to read
	// only where the referee drew an area.
	if len(inputs.OccurrenceAreas) > 0 {
		record.Stamp("E012")
	}

	// 1.B. Determine starport type; two dice throw and consult the
	// starports table (pp. 1, 12). The base throws follow immediately,
	// naval then scout, because a base is a property of the starport and
	// the starport chart is where the throw is printed (ERRATA E001).
	err = e.starports(stream, record, hexes)
	if err != nil {
		return nil, err
	}

	// 1.C. "Determine space lanes; check all possible jump routes"
	// (pp. 2, 12). Quoted from the p. 12 checklist, which is why this one
	// says "space lanes" where the rest of the code says routes: the page
	// prints both words and the code picked one, but a quotation keeps the
	// page's.
	record.Routes = e.routes(stream, record.Worlds)

	// A pair is the thing the reading governs, so it takes two worlds to
	// have governed anything.
	if len(record.Worlds) >= worldsInAPair {
		record.Stamp("E003")
	}

	// 2. Generate specific worlds.
	err = e.createWorlds(stream, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}
```

The `Stamp` calls are the mechanism that makes this tool trustworthy, and they
are conditional on purpose. A record's `errata` array is not metadata about the
program; it is the record saying which readings of a silent page actually
governed *it*. E012 is stamped only where the referee drew a broad area, E003
only where there were two worlds to pair. When you add a step, deciding its
stamping condition is part of the work.

### Pass 1.A — where the worlds are

Eighty hexes, one die each, in ascending grid number:

```bash
sed -n '349,366p' gen/gen.go
```

```output
func scan(stream *dice.Stream, occurrenceDM int, areas []starmap.Area) ([]starmap.Hex, error) {
	var found []starmap.Hex

	for col := 1; col <= starmap.Columns; col++ {
		for row := 1; row <= starmap.Rows; row++ {
			hex, err := starmap.NewHex(col, row)
			if err != nil {
				return nil, fmt.Errorf("hex %d,%d: %w", col, row, err)
			}

			if occurrenceTarget.Met(stream.Die() + dmAt(occurrenceDM, areas, hex)) {
				found = append(found, hex)
			}
		}
	}

	return found, nil
}
```

Every hex draws a die, including the ones that fail. A hex that fails is left
blank and consumes its die like any other — skipping the draw would shift
everything after it.

The broad areas of page 1 hang off this loop by exactly one addend. They vary
the *number the throw is read against* and never the throw, its order, or its
eighty dice:

```bash
sed -n '376,384p' gen/gen.go
```

```output
func dmAt(occurrenceDM int, areas []starmap.Area, hex starmap.Hex) int {
	for _, area := range areas {
		if area.Contains(hex) {
			return area.DM
		}
	}

	return occurrenceDM
}
```

That is why a record generated with no areas is byte-for-byte what the tool
wrote before the feature existed: with an empty list `dmAt` walks nothing and
returns the DM it was handed. Scanning area by area instead — the obvious
implementation — would visit the hexes out of grid order and silently move
every seed's meaning.

At most one area can answer, because overlapping areas are refused before any
die is thrown. Page 1 gives no basis for combining two DMs, so a set that would
need one is rejected rather than resolved.

### Pass 1.B — starports, and the throws the checklist forgot

Page 12's checklist has three steps and no base throw. But page 1 says
starports "will be accompanied by naval or scout bases," and the starport chart
on page 5 prints the throws as rules. Bases are generated; the checklist is a
summary that omits them. What the pages do not fix is *where* the throws sit,
and the reading (E001) puts them immediately after each starport:

```bash
sed -n '290,305p' gen/gen.go
```

```output
		if target, printed := e.charts.StarportChart.NavalBase(port); printed {
			world.NavalBase = target.Met(stream.D2())
			basesThrown = true
		}

		if target, printed := e.charts.StarportChart.ScoutBase(port); printed {
			world.ScoutBase = target.Met(stream.D2())
			basesThrown = true
		}

		record.Worlds = append(record.Worlds, world)
	}

	if basesThrown {
		record.Stamp("E001")
	}
```

The `printed` return is the important half. The chart prints a naval throw only
at starports A and B and a scout throw at A through D; where the chart prints
nothing, no die is drawn. `basesThrown` tracks whether any throw was actually
made, so a subsector of nothing but E and X starports does not claim a reading
that governed none of it.

### Pass 1.C — the commercial routes

Every pair of worlds, examined once:

```bash
sed -n '318,337p' gen/gen.go
```

```output
func (e *Engine) routes(stream *dice.Stream, worlds []starmap.World) []starmap.Route {
	routes := []starmap.Route{}

	for i, first := range worlds {
		for _, second := range worlds[i+1:] {
			distance := first.Hex.Distance(second.Hex)

			target, stated := e.charts.JumpRoutes.Target(first.Starport, second.Starport, distance)
			if !stated {
				continue
			}

			if target.Met(stream.Die()) {
				routes = append(routes, starmap.Route{From: first.Hex, To: second.Hex, Distance: distance})
			}
		}
	}

	return routes
}
```

The `stated` guard is the same discipline as the base throws. A pair with an X
starport has no row in the jump routes table, and a dash cell states no number.
Page 2 describes the throw as being made *against a stated number*, so where
none is stated the pair is skipped without drawing. This is also where the font
trap bites hardest — a text extraction of that table renders its dashes as the
digit 4, which would turn every empty cell into a target of 4+.

### Pass 2 — the worlds themselves

Each world is finished before the next is begun. The six characteristics of
pages 4 through 8, then the technological index of page 9:

```bash
sed -n '197,224p' gen/gen.go
```

```output
	// 2.C. Planetary atmosphere. 2D-7 + planetary size. A planet of size
	// zero automatically has an atmosphere of zero, and no die is thrown:
	// rolling and dropping one would shift every later world (R13).
	if world.Size > 0 {
		world.Atmosphere = clamp(world, starmap.Atmosphere, stream.D2()-automaticMinusSeven+world.Size, uncapped)
	}

	// 2.D. Hydrographic percentage. 2D-7 + planetary size, with a further
	// DM of -4 if the atmosphere is 0, 1, or greater than 9. A size of 0 or
	// 1 gives an automatic 0, and again no die is thrown (R13).
	if world.Size > 1 {
		raw := stream.D2() - automaticMinusSeven + world.Size
		if world.Atmosphere <= 1 || world.Atmosphere > 9 {
			raw -= dryAtmosphereDM
		}

		world.Hydrographics = clamp(world, starmap.Hydrographics, raw, uncapped)
	}

	// 2.E. Population. 2D-2, an exponent of 10 (pp. 8, 12).
	world.Population = clamp(world, starmap.Population, stream.D2()-automaticMinusTwo, uncapped)

	// 2.F. Planetary government. 2D-7 + the population digit (pp. 8, 12).
	world.Government = clamp(world, starmap.Government, stream.D2()-automaticMinusSeven+world.Population, uncapped)

	// 2.G. Law level. 2D-7 + the government type (pp. 8, 12). Government
	// feeds this already floored: a clamped value is the value.
	world.LawLevel = clamp(world, starmap.LawLevel, stream.D2()-automaticMinusSeven+world.Government, uncapped)
```

The two `if` guards are the most important lines in the package. A size-0 world
gets an atmosphere of 0 automatically, and a size-0-or-1 world a hydrographics
of 0, and in neither case is a die thrown. Rolling one and discarding it would
be simpler and would shift every later world in the subsector. The same
temptation recurs everywhere in this codebase and the answer is always the
same.

Note also that `Government` feeds `LawLevel` *already floored*. A clamped value
is the value; the raw throw is recorded but never propagated.

`clamp` is that flooring, and it records itself:

```bash
sed -n '259,267p' gen/gen.go
```

```output
func clamp(world *starmap.World, which starmap.Characteristic, raw, high int) int {
	value := min(max(raw, 0), high)

	if value != raw {
		world.Clamps = append(world.Clamps, starmap.Clamp{Characteristic: which, Raw: raw, Value: value})
	}

	return value
}
```

## The record

That is the whole of generation. What comes out is a `starmap.Record`, and it
is worth looking at a real one before reading the type. The command is
reproducible — the seed is given explicitly, so this is the same subsector
every time:

```bash
go run ./cmd/ctworldgen new --seed 1977 --name Aramis --occurrence-dm -1 | head -34
```

```output
{
  "schema_version": 1,
  "ruleset": "ct-1977-book3-pp1-12",
  "engine_version": "1",
  "rng_algorithm": "go-math-rand-v2-pcg",
  "seed": 1977,
  "errata": [
    "E001",
    "E002",
    "E003",
    "E004",
    "E005"
  ],
  "name": "Aramis",
  "occurrence_dm": -1,
  "grid": {
    "columns": 8,
    "rows": 10
  },
  "worlds": [
    {
      "hex": "0105",
      "name": "",
      "starport": "X",
      "naval_base": false,
      "scout_base": false,
      "size": 9,
      "atmosphere": 10,
      "hydrographics": 2,
      "population": 5,
      "government": 0,
      "law_level": 0,
      "tech_index": 1,
      "digits": "X9A25001"
```

Four things to notice.

The first four fields are provenance: schema version, ruleset, engine version,
and the exact RNG construction. These are what a referee trusts without
checking, and the read path verifies every one of them — a record claiming a
different generator parses perfectly cleanly and would report a subsector this
tool cannot vouch for.

`errata` lists the readings that governed *this* record. E012 is absent because
no broad area was given; E006 is absent because this is not a sector.

`digits` — `X9A25001` — is eight characters, the starport followed by the seven
characteristics with nothing between them. The familiar hyphen before the
technological index is a later printing's and is not added here. Every value is
stored numerically beside it, so the format loses nothing.

`grid` says which of two shapes this record is: the page 3 subsector grid, or
the sector of sixteen. There is no third.

### Types carry identity; data carries ranges

`starmap` holds a small set of types — `Hex`, `Starport`, `Digit`,
`Characteristic`, `Parsecs` — and a deliberate absence. The seven
characteristics are plain `int`.

The dividing rule is worth memorising because it looks like an oversight from
either side: **a type exists where something either is or is not the thing; a
value range stays `int`.** A hex outside the grid is not a hex and a character
outside the alphabet is not a digit, so both are types. An atmosphere of 13 is
a perfectly legal atmosphere that no table describes, so `type Atmosphere int`
bounded 0-12 would reject a value the engine legitimately writes — and would
put the page 5 table's last row into Go source, a third copy of a number the
book prints once.

The single most dangerous function in the package is the offset-to-cube
conversion behind `Hex.Distance`:

```bash
sed -n '240,260p' starmap/hex.go
```

```output
// coordinates.
//
// The grid prints 0101 at the top left with 0201 half a hex below it, so
// the even-numbered printed columns are the ones pushed down. In
// zero-based indices that is the standard odd-q vertical layout. Getting
// this backwards leaves every distance internally consistent and wrong by
// one for half the map, which no record-against-record check can catch --
// hex_test.go measures against the printed page instead. Never change
// this without re-measuring there.
func (h Hex) cube() (int, int, int) {
	// Every second column is pushed down half a hex, so a column
	// contributes half its index to the row offset (p. 3).
	const columnsPerRowStep = 2

	q := h.Col - 1
	r := h.Row - 1
	x := q
	z := r - (q-(q&1))/columnsPerRowStep

	return x, -x - z, z
}
```

Flip that parity and every distance stays internally consistent and is wrong by
one for half the map. No record-against-record test can catch it, because both
sides of the comparison move together. The test that catches it measures
against the grid printed on page 3.

The same parity is encoded in two other places — the text map's line builder
and the drawn map's hex centre — and each has its own independent measurement
against the page, because each can be flipped without the other two noticing.
Their three test harnesses are kept separate on purpose: a merged harness
compares the three against each other and passes when all three are flipped
together.

One subtlety that will confuse you if you meet it cold: `Hex` bounds itself by
the *largest* grid there is, not by the page 3 grid.

```bash
sed -n '262,267p' starmap/hex.go
```

```output
// valid bounds a hex by the largest grid there is, because a hex is an
// identifier and the identifier is four digits either way. Whether a hex
// is on *this* record's grid is [Grid.Contains], and the record checks it.
func (h Hex) valid() bool {
	return h.Col >= 1 && h.Col <= SectorColumns && h.Row >= 1 && h.Row <= SectorRows
}
```

An identifier is four digits whether it names a subsector hex or a sector hex,
so `0910` parses successfully. Only the *record's own grid* refuses it, in
`Grid.hold`. That split is what lets one `Hex` type serve both grids.

### The read path

`render` starts by decoding a record, and the decode is defensive in a way that
repays reading:

```bash
sed -n '222,251p' starmap/record.go
```

```output
func Decode(r io.Reader) (*Record, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var record Record

	err := dec.Decode(&record)
	if err != nil {
		return nil, fmt.Errorf("decoding the record: %w", err)
	}

	// A record written before grids were recorded is a subsector, which is
	// the only shape the tool wrote then. The reporter of issue 1 has
	// sixteen such files; they still read.
	if record.Grid.Zero() {
		record.Grid = PageThreeGrid()
	}

	err = record.Validate()
	if err != nil {
		return nil, err
	}

	err = pastTheRecord(dec)
	if err != nil {
		return nil, err
	}

	return &record, nil
}
```

`DisallowUnknownFields` catches a record from a *newer* schema that added a
field. That looks like the whole of the defence and is not — a record claiming
a different schema version, ruleset, or generator parses perfectly cleanly,
every field reads, and the listing renders. `Validate` is what refuses those,
and it is the check most likely to look redundant to someone tidying up.

The grid backfill is backward compatibility with a real user: records written
before grids were recorded carry none, and the reporter of the alpha has
sixteen such files.

`pastTheRecord` refuses trailing content, so a file holding two concatenated
records fails loudly rather than decoding the first and discarding the rest.

The published schema in `docs/record.schema.json` states the same rules, and
both exist on purpose — *a schema alone rejects nothing at read time*. The
schema ships for other tools; the Go checks bind this one; and
`internal/audit` holds the two together by validating everything the engine
writes against the published schema.

You can watch the provenance check fire. Take the record generated above,
corrupt one provenance field, and try to render it:

```bash
go run ./cmd/ctworldgen new --seed 1977 | sed 's|ct-1977-book3-pp1-12|ct-1981-book3|' | go run ./cmd/ctworldgen render /dev/stdin 2>&1 | head -1
```

```output
ctworldgen: reading /dev/stdin: not a record of the ruleset this tool implements: the record says "ct-1981-book3" and this tool implements "ct-1977-book3-pp1-12"
```

The error names the page-level fact and both sides of the disagreement. That
habit — errors that cite the rule and the value that offended it — runs through
the whole codebase, and the alpha report singled it out as what bought trust in
the output before anything had been checked.

The unknown-field half behaves the same way:

```bash
go run ./cmd/ctworldgen new --seed 1977 | sed 's|"seed": 1977,|"seed": 1977, "house_rule": true,|' | go run ./cmd/ctworldgen render /dev/stdin 2>&1 | head -1
```

```output
ctworldgen: reading /dev/stdin: decoding the record: json: unknown field "house_rule"
```

## The tables, and the font trap

`tables` holds every chart of pages 1-12 as embedded JSON, validated at load.
Behind it sits the most unusual constraint in this repository.

The held PDFs' embedded font maps the em-dash to the glyph `4` and the minus
sign to `3`. A text extraction of the jump routes table therefore renders its
empty cells as the digit 4, and the size formula `2D − 2` as `2D32`. Both
readings are wrong and — this is the part that matters — **both look like
data**. Nothing downstream can detect it.

So every table is transcribed from a *visual* read of the page and then
transcribed a second time inside the package's own tests, and the two must
agree. Look at the row the trap would corrupt:

```bash
sed -n '1,20p' tables/data/jump_routes.json
```

```output
{
  "table": "JUMP ROUTES",
  "page": 2,
  "comment": "targets are indexed by jump distance 1 through 4; null is the em-dash the page prints, at which no route is possible and no die is thrown (ERRATA E003)",
  "rows": [
    { "pair": "A-A", "targets": [1, 2, 4, 5] },
    { "pair": "A-B", "targets": [1, 3, 4, 5] },
    { "pair": "A-C", "targets": [1, 4, 6, null] },
    { "pair": "A-D", "targets": [1, 5, null, null] },
    { "pair": "A-E", "targets": [2, null, null, null] },
    { "pair": "B-B", "targets": [1, 3, 4, 6] },
    { "pair": "B-C", "targets": [2, 4, 6, null] },
    { "pair": "B-D", "targets": [3, 6, null, null] },
    { "pair": "B-E", "targets": [4, null, null, null] },
    { "pair": "C-C", "targets": [3, 6, null, null] },
    { "pair": "C-D", "targets": [4, null, null, null] },
    { "pair": "C-E", "targets": [4, null, null, null] },
    { "pair": "D-D", "targets": [4, null, null, null] },
    { "pair": "D-E", "targets": [5, null, null, null] },
    { "pair": "E-E", "targets": [6, null, null, null] }
```

`B-E` is `[4, null, null, null]` — a target of 4 at one parsec and an em-dash at
two, three and four. A `pdftotext` of that row reads `B-E 4 4 4 4`, which is a
plausible-looking table in which every long jump is possible on a 4+. Nothing
but a second pair of eyes on the printed page catches that.

The `null`s then have to survive into the lookup, which is why `Target` returns
two values rather than a sentinel number:

```bash
sed -n '189,205p' tables/tables.go
```

```output
func (j *JumpRoutes) Target(a, b starmap.Starport, distance starmap.Parsecs) (dice.Target, bool) {
	if distance < 1 || distance > starmap.MaxJump {
		return 0, false
	}

	row, ok := j.targets[pairKey(a, b)]
	if !ok {
		return 0, false
	}

	target := row[distance-1]
	if target == nil {
		return 0, false
	}

	return dice.Target(*target), true
}
```

Three different "no" answers — out of range, no row for X, dash cell — all
collapse into `false`, and `gen.routes` skips without drawing. A sentinel
target of 0 would have been met by every throw.

The one exception to the double transcription is the descriptive *labels* of
pages 5-7 — "Feudal Technocracy", "Dense, tainted". Those are editorial
abbreviations of the book's prose rather than transcriptions of it, so retyping
them in a test compares an abbreviation against itself. Tables of proper nouns
are never that exception.

## Rendering

`render` takes a record and writes one of two documents. Nothing in the package
throws a die or decides a rule. Here is the head of the listing for the record
we generated above:

```bash
go run ./cmd/ctworldgen new --seed 1977 --name Aramis --occurrence-dm -1 | go run ./cmd/ctworldgen render /dev/stdin | sed -n '1,30p'
```

````output
# Aramis

27 worlds, 44 routes, 31 drawn. Generated from seed 1977 at occurrence DM -1.

## The map

The p. 3 sub-sector hex grid. The odd-numbered columns sit high and the even-numbered ones half a hex below them, which is how the page prints it. A world carries the letter of its starport -- p. 1 marks the hex with the letter the starports table gives -- and a hex with no world is left blank, which is what p. 1 says to leave it. P. 2 also asks for a line drawn between the worlds a route joins; a monospace grid has nowhere to put one, so this map draws none and the route table below carries them instead. `render --format pdf` draws them.

```text
0101      0301 C    0501      0701
     0201      0401 A    0601      0801
0102      0302      0502      0702 A
     0202      0402 B    0602 E    0802
0103      0303      0503 E    0703 B
     0203      0403 C    0603      0803
0104      0304      0504      0704 C
     0204      0404      0604      0804
0105 X    0305      0505      0705
     0205      0405 E    0605      0805
0106 E    0306 A    0506 B    0706
     0206 A    0406      0606      0806
0107 B    0307      0507      0707
     0207      0407 A    0607      0807 E
0108      0308      0508      0708
     0208 E    0408      0608 A    0808
0109      0309 B    0509 C    0709
     0209 B    0409 E    0609      0809
0110      0310      0510      0710 C
     0210 B    0410      0610      0810 C
```
````

Read the summary line: **27 worlds, 44 routes, 31 drawn.** The record carries
44 routes; the map draws 31. That gap is a reading, and the document is
required to announce it.

Read the map alongside `Hex.cube` above: the odd-numbered columns sit on their
row's first line and the even-numbered ones are pushed half a slot right and
half a row down, which is how page 3 prints the grid. This is the second of the
three places the parity lives.

### The legible lane rule

Page 2, having observed that a jump-2 can be made over two jump-1 links, tells
the map-drawer that some connections may be ignored because they are "already
present". That is permission addressed to whoever draws the map, not a rule
addressed to whoever throws the dice — so the engine examines every pair and
the record carries every lane, and only the *drawing* is thinned.

The implementation looks like a spanning forest and is deliberately not one:

```bash
sed -n '40,75p' render/lanes.go
```

```output
func legible(routes []starmap.Route) []starmap.Route {
	joined := newGroups()
	keep := make(map[starmap.Route]bool, len(routes))

	for distance := starmap.Parsecs(1); distance <= starmap.MaxJump; distance++ {
		layer := make([]starmap.Route, 0, len(routes))

		for _, route := range routes {
			if route.Distance == distance {
				layer = append(layer, route)
			}
		}

		// Examined against what the shorter lanes joined ...
		for _, route := range layer {
			if !joined.same(route.From, route.To) {
				keep[route] = true
			}
		}

		// ... and only then joining anything itself.
		for _, route := range layer {
			joined.merge(route.From, route.To)
		}
	}

	drawn := make([]starmap.Route, 0, len(keep))

	for _, route := range routes {
		if keep[route] {
			drawn = append(drawn, route)
		}
	}

	return drawn
}
```

The two loops over `layer` are the whole rule, and separating them is the
point. A distance is examined *in full* against what the shorter distances
joined, and only then does it join anything itself.

Union each lane as it is examined instead — the ordinary greedy spanning forest
— and the result stops being a function of the record: which of several
equal-length lanes survives becomes whichever the loop happened to reach first.
It would draw fewer lanes and look better, and the drawn map would quietly
depend on the order the routes sit in. Measured over two hundred shufflings of
one record's routes, the layered reading gives one result and the greedy form
gives two hundred.

A jump-1 is therefore never dropped — nothing is shorter than it — and two
equal-length lanes never suppress each other.

Both readings of the same record, side by side:

```bash
go run ./cmd/ctworldgen new --seed 1977 --name Aramis --occurrence-dm -1 > /tmp/ctwg-walkthrough.json && for mode in legible all; do printf '%-8s ' "$mode"; go run ./cmd/ctworldgen render --lanes $mode /tmp/ctwg-walkthrough.json | sed -n '3p'; done; rm -f /tmp/ctwg-walkthrough.json
```

```output
legible  27 worlds, 44 routes, 31 drawn. Generated from seed 1977 at occurrence DM -1.
all      27 worlds, 44 routes. Generated from seed 1977 at occurrence DM -1.
```

## The sector layer

The book charts a subsector and stops. But its route rule speaks of a world's
"neighbors" and the jump routes table is read on the starport pair and the
distance — neither knows where a subsector ends. The border is an artifact of
generating one subsector at a time, not a term in the rule. So `sector`
assembles sixteen subsectors on one 32x40 grid and throws only for the pairs
that straddle two of them.

The whole of it:

```bash
sed -n '24,71p' gen/sector.go
```

```output
func (e *Engine) Sector(inputs Inputs) (*starmap.Record, error) {
	// Before Validate, which holds an area against the p. 3 grid: a
	// sector-grid rectangle refused there would be reported as off an 8x10
	// grid, which is a true sentence about the wrong thing.
	//
	// A broad area is a rectangle of one grid's numbering, and the sixteen
	// members are each generated on their own p. 3 grid: a sector-grid
	// rectangle would have to be clipped into sixteen local ones, which is
	// not built. Refused rather than dropped -- a sector that quietly
	// ignored them would carry a DM that governed nothing.
	if len(inputs.OccurrenceAreas) > 0 {
		return nil, ErrSectorTakesNoAreas
	}

	err := inputs.Validate()
	if err != nil {
		return nil, err
	}

	record := starmap.New(inputs.Seed, inputs.Name, inputs.OccurrenceDM, nil)

	record.Grid = starmap.SectorGrid()

	// The members consume the streams of seeds base through base+15, in
	// order (ERRATA E006 part 4). Counting the seed alongside the index
	// keeps the derivation in one place and out of a conversion.
	seed := inputs.Seed

	for index := range starmap.Members {
		err := e.member(record, inputs, index, seed)
		if err != nil {
			return nil, err
		}

		seed++
	}

	// E002's order, now read across the whole grid.
	slices.SortFunc(record.Worlds, func(a, b starmap.World) int { return a.Hex.Number() - b.Hex.Number() })

	record.Routes = append(record.Routes,
		e.seams(dice.NewStream(inputs.Seed+seamSeedOffset), record.Worlds)...)
	slices.SortFunc(record.Routes, routeOrder)

	record.Stamp("E006")

	return record, nil
}
```

Nothing is re-thrown at sector level. No world is placed, no starport graded,
no characteristic generated. `e.member` calls `Generate` — the same function
the `new` subcommand calls — sixteen times on consecutive seeds, and then the
seam pass consumes a seventeenth stream at `base + 16`.

The translation onto the sector grid rests on one arithmetic accident, and the
comment on `Place` says which:

```bash
sed -n '41,59p' starmap/hex.go
```

```output
// Place translates a member's local hex onto the sector grid. Member
// index sits at column band index mod 4 and row band index div 4, and a
// local hex moves by whole bands (ERRATA E006 parts 1 and 2).
//
// A sub-sector is eight columns wide and eight is even, so a column's
// odd-or-even parity survives this -- which is what makes an interior
// pair measure the same distance on the sector grid as it did at home.
// It is exported because that property is worth asserting against the
// translation the engine actually uses, rather than against a second copy
// of the arithmetic written in a test.
//
// It lives here rather than in gen because it is grid geometry and not a
// rule: the engine needed it first, and the renderer needs the same one.
func Place(index int, hex Hex) Hex {
	across := index % SectorAcross
	down := index / SectorAcross

	return Hex{Col: across*Columns + hex.Col, Row: down*Rows + hex.Row}
}
```

A subsector is **eight** columns wide, and eight is even, so a column's
odd-or-even parity survives translation. That is what makes an interior pair
measure the same distance on the sector grid as it did at home. An odd band
width would have flipped the parity of every second band and quietly changed
interior distances — the same trap `Hex.cube` carries, arriving by a second
road.

Note also what is *not* stored. The record carries no member field. Which
subsector a world belongs to is read straight back off its hex by `MemberOf`,
so the decomposition cannot come to disagree with the grid.

### The property that makes a sector trustworthy

Member *i* of `sector --seed N` is exactly the subsector `new --seed N+i`
writes. Nothing about being in a sector changes a world. That is the identity a
referee relies on when he picks one subsector out of a sector and runs a
session in it, and it is checked directly in `gen/sector_test.go` — but it is
also checkable from outside the program.

Member 5 sits at column band 1 and row band 1, so its local hex *(c, r)* lands
at *(c+8, r+10)* on the sector grid. Translate the band back and compare:

```bash
go run ./cmd/ctworldgen sector --seed 1000 > /tmp/ctwg-sec.json && go run ./cmd/ctworldgen new --seed 1005 > /tmp/ctwg-m5.json && python3 -c '
import json
sec=json.load(open("/tmp/ctwg-sec.json")); m5=json.load(open("/tmp/ctwg-m5.json"))
def local(h): return "%02d%02d"%(int(h[:2])-8, int(h[2:])-10)
band=[w for w in sec["worlds"] if 9<=int(w["hex"][:2])<=16 and 11<=int(w["hex"][2:])<=20]
a=[(local(w["hex"]), w["digits"]) for w in band]
b=[(w["hex"], w["digits"]) for w in m5["worlds"]]
print("member 5 of sector --seed 1000 :", len(a), "worlds")
print("      new --seed 1005          :", len(b), "worlds")
print("identical hex and digit strings:", a==b)
'; rm -f /tmp/ctwg-sec.json /tmp/ctwg-m5.json
```

```output
member 5 of sector --seed 1000 : 42 worlds
      new --seed 1005          : 42 worlds
identical hex and digit strings: True
```

### The seam pass

Page 2 says each specific pair of worlds should be examined for jump routes
only once. An interior pair was already examined inside its own member, so the
sector pass examines only pairs whose two worlds are in *different* members:

```bash
sed -n '118,146p' gen/sector.go
```

```output
func (e *Engine) seams(stream *dice.Stream, worlds []starmap.World) []starmap.Route {
	routes := []starmap.Route{}

	for index, first := range worlds {
		// Hoisted: the first world's band is fixed for the whole inner
		// loop, and deriving it per pair is a division per pair over the
		// two hundred thousand a sector has.
		firstMember := starmap.MemberOf(first.Hex)

		for _, second := range worlds[index+1:] {
			if firstMember == starmap.MemberOf(second.Hex) {
				continue
			}

			distance := first.Hex.Distance(second.Hex)

			target, stated := e.charts.JumpRoutes.Target(first.Starport, second.Starport, distance)
			if !stated {
				continue
			}

			if target.Met(stream.Die()) {
				routes = append(routes, starmap.Route{From: first.Hex, To: second.Hex, Distance: distance})
			}
		}
	}

	return routes
}
```

Everything else is the subsector route pass unchanged — same table, one die
against a stated number, no row for an X starport and no throw at a dash cell.

This loop is O(n²): a sector examines about two hundred thousand pairs to find
the few within four parsecs, and a spatial index is the obvious fix. It stays
quadratic on purpose. The loop's visit order **is** the order the seam stream
is drawn in, so an index reaching the same pairs in a different order writes a
different sector from the same seed, silently. The safe optimisation is an
early-out on column distance *inside* this loop, which preserves the order
exactly.

This is the single best example of the rule stated at the top of this document:
**a performance change in `gen` is a rules change until proven otherwise.**

### A sector's documents

A sector rendered as one large subsector would be 662 worlds in one roster,
1,879 lanes in one table, and a map of 1,280 hexes whose four-digit numbers do
not fit inside them. So the listing is an index of the whole grid followed by
the sixteen subsector listings its members would have had, each on its own page
3 grid.

The contents table names each member by index, hex range, and the seed that
writes it standalone:

```bash
go run ./cmd/ctworldgen sector --seed 1000 > /tmp/ctwg-sec.json && go run ./cmd/ctworldgen render /tmp/ctwg-sec.json | grep -A 5 '^| Subsector | Hexes'; rm -f /tmp/ctwg-sec.json
```

```output
| Subsector | Hexes | Worlds | Lanes within | Crossing | Seed |
| --- | --- | --- | --- | --- | --- |
| 0 | 0101 to 0810 | 35 | 51 | 28 | 1000 |
| 1 | 0901 to 1610 | 29 | 42 | 40 | 1001 |
| 2 | 1701 to 2410 | 39 | 60 | 68 | 1002 |
| 3 | 2501 to 3210 | 37 | 85 | 26 | 1003 |
```

And each member section opens by making the identity actionable — the referee
who wants this one subsector and nothing else can generate it from the heading:

```bash
go run ./cmd/ctworldgen sector --seed 1000 > /tmp/ctwg-sec.json && go run ./cmd/ctworldgen render /tmp/ctwg-sec.json | grep -A 4 '^## Subsector 5 '; rm -f /tmp/ctwg-sec.json
```

```output
## Subsector 5 &mdash; 0911 to 1620

Member 5 of this sector. `ctworldgen new --seed 1005` writes this sub-sector on its own p. 3 grid, where its first hex is 0101; here it is laid on the sector's, and that hex is 0911 (ERRATA E006).

42 worlds, 108 lanes within it, 54 crossing into its neighbours.
```

The A-through-P lettering familiar from later Traveller editions is
deliberately not used: it comes from the 1981 revision, and only the held 1977
pages govern. An index and a seed are things the tool itself established.

Two counts in that summary are worth distinguishing. "Lanes within it" are the
routes both of whose ends are at home; "crossing" are the ones a referee
generating sixteen subsectors one at a time could never have found. A crossing
lane is listed under *both* the subsectors it joins, so the sixteen tables
together carry more rows than the record has lanes — and the document says so
where the tables begin.

## The gate

There is one gate and CI runs exactly it — tidy, vet, golangci-lint, NilAway,
`go test -race`, and a coverage ratchet. The workflow is a single line, which
is the point:

```bash
sed -n '45,49p' .github/workflows/ci.yml
```

```output

      # CI runs exactly `task`. Never add a check here that the local gate
      # does not run, and never add a tool to the gate without also
      # installing it above.
      - run: task
```

The toolchain is deliberately unpinned. The gate is meant to fail when a tool
*moves* rather than drift behind it, so a red gate on code you did not touch is
the signal working. Answer the finding; do not pin a tool or add a linter
disable to silence it.

`go test -race` guards no product concurrency — there is none — only the
parallel test harness. It stays because it is the check that catches the day
someone adds concurrency, which the sixteen independent members make a live
temptation rather than a theoretical one.

The coverage ratchet counts *uncovered statements per package* rather than a
percentage, and fails in both directions. A percentage holds still while a
guarded branch adds one covered statement and one uncovered, and it grows more
forgiving as the repository grows.

### A test that cannot fail looks exactly like one that passes

This is the hazard peculiar to this codebase and it deserves the last word.
Almost everything the program asserts is an invariant over dice, and a broken
invariant check is indistinguishable from a passing one. It has happened
repeatedly here and has never once surfaced as a red suite:

- The base and route throws could have had their sense *inverted* and every
  invariant still passed, because the checks only asked whether the routes and
  bases that exist are legal.
- Five world-creation assertions were written, and the edit meant to call them
  from the sweep silently matched nothing. They sat in the file, defined and
  dead, and the suite went green.
- A roster check searched the whole listing for each world's hex — which the
  world's own detail page satisfies — so dropping half the roster passed.
- The map's parity check ran on a record with *no worlds*, where every cell is
  the same width, so a bug that shifts a row only where a starport letter is
  drawn could not be expressed by the fixture at all.

The habit that follows is not optional: **a new invariant is not done until a
deliberate mutation has been shown to kill it.** Invert a target, drop a column
from a sum, halve a loop — then run the suite and read the failure. If it does
not name the thing you broke, the check is not holding what you think.

Three things make that harder here than elsewhere. `TestGoldens` compares
against fixtures the code under test wrote, so a mutation moves them and the
suite fails for the wrong reason — run `task regenerate` first. A mutation
aimed at something no fixture exercises is a no-op that reads as a surviving
mutant. And the fixture has to be able to *express* the bug: a world-less map
cannot show a mis-drawn world.

## Where to go next

`THEORY.md` explains why the system is shaped this way and where its theory is
thinnest. `docs/ERRATA.md` holds every reading of a silent page, with its cite
and its stamping condition. `docs/COVERAGE.md` maps rule to code to test. And
`CLAUDE.md` keeps the standing list of things that look like slack and are not
— the changes that are locally defensible and globally wrong.

## Findings from this pass

Tracing the code end to end turned up two things a reader of this document
should not have to rediscover. Both are filed in `.issues/`.

## Index

| #   | Severity | Issue                                                         | Primary location                                   |
| --- | -------- | ------------------------------------------------------------- | -------------------------------------------------- |
| 1   | low      | `hex-less-is-the-named-ordering-rule-and-nothing-production-calls-it` | `starmap/hex.go:186`, `gen/sector.go:62`     |
| 2   | low      | `render-sector-go-holds-both-typesetters-breaking-the-package-file-split` | `render/sector.go:174`, `render/sector.go:276` |

**Total: 2 issues (0 critical, 0 high, 0 medium, 2 low)**

### Corrections to this document, from later passes in the same review

Two claims above were written before the reduction and refactor passes ran over
the same tree, and both should be read with these attached.

**The parity is encoded in four places, not three.** Sections above say the page
3 column parity lives in `starmap.Hex.cube`, `render.gridLine` and
`render.mapFit.hexCenter`, each with its own measurement against the printed
page. `CLAUDE.md` says the same. There is a fourth — `render.indexLine`, which
draws the sector index map — and nothing measures it. `code-reduction`
mutation-tested it: flipping its indent to the wrong parity leaves
`go test ./render` green. That is exactly the failure the "three separate
harnesses" argument exists to prevent, occurring in the one copy the argument
did not count. Its finding argues the right answer is to delete the copy rather
than measure it, by folding the index map into the listing's own grid drawer.

**Finding 1 below is contested and both later passes rejected it.**
`code-reduction` and `code-refactor` each concluded that closed GitHub issue #32
("Exported surface that exists only for tests") settled the exported-for-tests
question deliberately rather than by oversight, and that `Hex.Number` already
carries the ERRATA E002 cite in its own doc comment — which undercuts the
finding's premise that the production sorts carry no cite. Read
`.issues/000-reduction.md` and `.issues/000-refactor.md` before acting on it.
The file is left in place because the issues protocol does not delete findings,
not because the objection is weak.


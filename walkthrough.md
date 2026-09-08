# ctworldgen Walkthrough

*2026-09-08T13:15:44Z by Showboat 0.6.1*
<!-- showboat-id: e04036cd-7e27-4745-a998-04e35f2e8011 -->

## Overview

`ctworldgen` is a Go CLI that generates Classic Traveller subsectors from
the Worlds chapter of Book 3 *Worlds and Adventures* — pages 1 to 12 of the
© 1977 text — and renders them as the documents a referee runs a session
from.

Three facts shape the whole codebase:

- **The specification is a printed page.** The rules come only from held
  PDFs. No test can tell you a table was mis-transcribed, so every table is
  typed in twice — once as data, once inside the test that reads it — and
  the two must agree.
- **The dice stream's consumption order is the meaning of a seed.** Every
  throw happens in the same order on every run. Two throws are deliberately
  *not* made, and adding one would shift every world after it.
- **The subsector is the record.** Not the world. Star mapping is
  subsector-scoped, and a lane cannot be drawn until its neighbours exist.

Four subcommands, two output formats.

```bash
go run ./cmd/ctworldgen 2>&1 | head -8
```

```output
ctworldgen generates Classic Traveller subsectors from Book 3 pp. 1-12.

usage:
  ctworldgen new    [--seed N] [--name X] [--occurrence-dm N] [--occurrence-area DM@FROM-TO]... [-o file] [--force]
  ctworldgen sector [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
  ctworldgen render [--format markdown|pdf] [--lanes legible|all] [-o file] [--force] record.json
  ctworldgen version
ctworldgen: no subcommand
```

## Architecture

Eight packages, depending on one another in one direction only.

| Package            | Holds                                                    | Depends on                       |
| ------------------ | -------------------------------------------------------- | -------------------------------- |
| `dice`             | The PCG stream, one and two dice, N+ targets (B1 pp. 2-3) | stdlib                           |
| `starmap`          | The domain types and the record itself                    | stdlib                           |
| `tables`           | Book 3's charts as embedded JSON, validated at load       | `starmap`, `dice`                |
| `gen`              | The generation procedure of pp. 1-12                      | `starmap`, `tables`, `dice`      |
| `render`           | The Markdown listing and the PDF booklet                  | `starmap`, `tables`, `gen`       |
| `cmd/ctworldgen`   | The four subcommands                                      | `gen`, `render`, `starmap`       |
| `internal/fixture` | The one roster both golden trees are generated from       | `starmap`                        |
| `internal/audit`   | Repository checks: schema conformance, errata resolution  | test-only                        |

Most of those edges are import cycles, which the compiler already refuses.
`.golangci.yml` writes depguard rules only for the two that would otherwise
compile: the command may not reach `tables` or `dice`, and production code
may not reach `internal/fixture`.

`internal/audit` needs no rule at all, because it holds no non-test file —
so the import fails to compile.

```bash
go list -f '{{.ImportPath}} GoFiles={{.GoFiles}}' ./internal/audit/
```

```output
github.com/philoserf/ctworldgen/internal/audit GoFiles=[]
```

## 1. The grid, and the parity hiding in it

Everything sits on the hex grid printed on Book 3 page 3: eight columns,
ten rows, hexes numbered `0101` through `0810`. A `Hex` is that identifier,
and it is a type rather than a pair of ints because a hex outside the grid
is not a hex.

```bash
sed -n '137,152p' starmap/hex.go
```

```output
// Hex identifies one hex of the subsector grid. It marshals to the
// four-digit column-and-row number the p. 3 grid prints -- "0101" -- which
// is the identifier a referee writes in a notebook (p. 4).
//
// The zero value is not a hex, and marshaling it is an error.
type Hex struct{ Col, Row int }

// NewHex returns the hex at a column and row of the p. 3 grid, both
// one-based.
func NewHex(col, row int) (Hex, error) {
	h := Hex{Col: col, Row: row}
	if !h.valid() {
		return Hex{}, fmt.Errorf("%w: hex %d,%d, and the grid is %dx%d", ErrOffGrid, col, row, SectorColumns, SectorRows)
	}

	return h, nil
```

Distance is where the trap lives. The page prints `0101` at the top left
with `0201` half a hex **below** it, so even-numbered columns are pushed
down. Get that backwards and every distance stays internally consistent and
is wrong by one for half the map — which no record-against-record test can
catch, because both sides of the comparison are wrong together.

```bash
sed -n '239,258p' starmap/hex.go
```

```output
// cube converts the offset coordinates of the p. 3 grid to cube
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

```

The only check that can catch a flip is one measured against the printed
page by hand, which is what this is — hex pairs and the distances a person
read off page 3.

```bash
sed -n '10,32p' starmap/hex_test.go
```

```output
// TestDistanceAgainstPrintedGrid measures against the sub-sector hex grid
// printed on Book 3 p. 3, not against another calculation.
//
// This is the test the offset-to-cube parity needs. The grid prints 0101
// at the top left with 0201 half a hex below it, so the even-numbered
// printed columns are pushed down. Getting that backwards leaves every
// distance internally consistent and wrong by one for half the map, which
// no record-against-record check can catch. Never change the conversion
// without re-measuring here, on the page.
func TestDistanceAgainstPrintedGrid(t *testing.T) {
	t.Parallel()

	// Distances taken by hand off the printed p. 3 grid.
	cases := []struct {
		a, b string
		want starmap.Parsecs
		note string
	}{
		{"0101", "0101", 0, "a hex is no distance from itself"},
		{"0101", "0102", 1, "straight down the same column"},
		{"0101", "0201", 1, "down and to the right: 0201 sits half a hex below 0101"},
		{"0102", "0201", 1, "up and to the right; a flipped parity gives 2"},
		{"0101", "0202", 2, "0202 is below 0201; a flipped parity gives 1"},
```

## 2. The tables, and the font trap

Book 3's charts live as JSON beside the loader, embedded into the binary.

```bash
sed -n '43,52p' tables/tables.go
```

```output
)

//go:embed data/*.json
var files embed.FS

// Tables is every chart of pp. 1-12 that generation consults.
type Tables struct {
	Starports       Starports
	JumpRoutes      JumpRoutes
	StarportChart   StarportChart
```

```bash
cat tables/data/jump_routes.json
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
  ]
}
```

That file says a route is possible at 1 to 4 parsecs and impossible beyond,
with `null` for the em-dashes the page prints. Which brings up the reason
every table in this repository is typed in twice.

**The embedded font in the held PDFs maps the em-dash to the glyph `4` and
the minus sign to `3`.** A text extraction of the jump routes table renders
`4 — — —` as `4 4 4 4`, and the size formula `2D − 2` as `2D32`. Both
readings are wrong, both look like data, and nothing downstream can tell.

So tables are read visually, and then transcribed a second time inside the
test package, so that the data file and the test must agree. This is the
same table, retyped by hand:

```bash
grep -n -A14 'func TestJumpRoutesTable' tables/tables_test.go | head -22
```

```output
109:func TestJumpRoutesTable(t *testing.T) {
110-	t.Parallel()
111-
112-	want := jumpRoutesTranscription()
113-	if len(want) != 15 {
114-		t.Fatalf("the transcription has %d rows, and the page prints 15", len(want))
115-	}
116-
117-	routes := load(t).JumpRoutes
118-	dashes := 0
119-
120-	for pair, cells := range want {
121-		a, b := pairStarports(t, pair)
122-
123-		for i, cell := range cells {
--
156:func TestJumpRoutesTableIsSymmetric(t *testing.T) {
157-	t.Parallel()
158-
159-	routes := load(t).JumpRoutes
160-
161-	for pair := range jumpRoutesTranscription() {
```

## 3. The dice

One PCG stream per record, seeded from the number the record carries. `Die`
is one die, `D2` is two — Book 1 page 2 makes two dice the unqualified
throw — and `Target` is the `N+` form, the only target kind pages 1-12 use.

```bash
sed -n '27,60p' dice/dice.go
```

```output
type Stream struct{ r *rand.Rand }

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

## 4. The engine: the procedure, in the order the book prints it

`Generate` walks the passes of page 12's summary. Pass 1 covers the whole
grid before pass 2 details any world, because that is the order the book
gives and the order the seed's meaning depends on.

```bash
sed -n '112,145p' gen/gen.go
```

```output
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
```

Then each world is detailed, 2.B through 2.H, in the order the page 4
Planetary Characteristics box lists them.

**The two throws that are not made are the most load-bearing lines in the
package.** Page 4 says a size-0 world has no atmosphere and a world of size
0 or 1 has no hydrographics. The engine does not throw and discard — it does
not throw at all. Rolling a die there would shift every subsequent world in
the stream, and every record anyone holds would stop reproducing.

```bash
sed -n '194,224p' gen/gen.go
```

```output
	// 2.B. Planetary size. 2D-2 (pp. 4, 12).
	world.Size = clamp(world, starmap.Size, stream.D2()-automaticMinusTwo, uncapped)

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

## 5. A sector is sixteen subsectors, and one more pass

ERRATA E006 reads the book's own "additional subsectors will have to be
charted" as sixteen subsectors laid on one 32x40 grid. Each member is
generated whole, drawing its own stream, so member *i* of a sector is
exactly the subsector `new --seed N+i` writes. That identity is what makes
a sector trustworthy, and it is tested directly.

What a member cannot do is throw for a route to a world in another member,
because it never saw one. So there is a seam pass, with a stream of its own.

```bash
sed -n '102,123p' gen/sector.go
```

```output
// seams throws for the pairs the members could not examine: those whose
// two worlds sit in different members (ERRATA E006 part 3). An interior
// pair was examined inside its member, and p. 2 examines each pair once.
//
// Everything else is E003 unchanged -- the same table, one die against a
// stated number, no row for an X starport and no throw at a dash cell,
// and the same order, read now in sector coordinates.
//
// Which member a world came from is not remembered: starmap.Place puts
// member i's hexes in band i, so starmap.MemberOf reads the band straight
// back off the hex (ERRATA E006 part 1). That is the same fact the
// translation states, held in one place rather than two that must agree.
func (e *Engine) seams(stream *dice.Stream, worlds []starmap.World) []starmap.Route {
	routes := []starmap.Route{}

	for index, first := range worlds {
		// Hoisted: the first world's band is fixed for the whole inner
		// loop, and deriving it per pair is a division per pair over the
		// two hundred thousand a sector has.
		firstMember := starmap.MemberOf(first.Hex)

		for _, second := range worlds[index+1:] {
```

## 6. The record

The record is the durable artifact: what `new` writes, what `render` reads,
and the one file the tool asks a referee to keep. It is a subsector — or a
sector — and a world is a row inside it.

```bash
sed -n '41,85p' starmap/record.go
```

```output
type Record struct {
	SchemaVersion int      `json:"schema_version"`
	Ruleset       string   `json:"ruleset"`
	EngineVersion string   `json:"engine_version"`
	RNGAlgorithm  string   `json:"rng_algorithm"`
	Seed          uint64   `json:"seed"`
	Errata        []string `json:"errata"`
	Name          string   `json:"name"`

	// Notes is the referee's, like Name, and about the map as a whole. It
	// is never generated and never read back: the engine writes nothing
	// here and no later step consults it.
	//
	// It is a field rather than an escape hatch. The record refuses every
	// key it does not define, which is what makes it trustworthy, so the
	// place to write had to be named rather than carved out (issue 1 #6).
	// omitempty keeps a record without one byte-identical to what the tool
	// wrote before this field existed.
	Notes string `json:"notes,omitempty"`

	OccurrenceDM int `json:"occurrence_dm"`

	// OccurrenceAreas are the broad areas of p. 1 (ERRATA E012): each a
	// rectangle of the grid's numbering with its own DM, overriding
	// OccurrenceDM at every hex it covers. A hex in no area takes
	// OccurrenceDM, which is the whole-subsector form of the same sentence.
	//
	// They are carried in the order the referee gave them and nothing sorts
	// them. Sorting would make the record a function of the geography
	// rather than of the typing, which is the nicer property and one no
	// test could fail; the property that matters is bought by refusing
	// overlaps, because the order of a set of areas that cannot overlap
	// changes no die.
	//
	// omitempty keeps a record without one byte-identical to what the tool
	// wrote before this field existed, as Notes does.
	OccurrenceAreas []Area `json:"occurrence_areas,omitempty"`

	// Grid is what the hexes below are numbered on: the p. 3 sub-sector
	// grid, or the sector grid of sixteen of them (ERRATA E006).
	Grid Grid `json:"grid"`

	Worlds []World `json:"worlds"`
	Routes []Route `json:"routes"`
}
```

`docs/record.schema.json` states the shape, and `Validate` holds a record to
it in code, because a schema alone rejects nothing at read time. `Decode`
calls it; `Marshal` deliberately does not, because the file is the referee's
notebook page and he is allowed to hand-edit it.

```bash
sed -n '339,362p' starmap/record.go
```

```output
func (s *Record) Validate() error {
	// The schema names these two grids and nothing else, and unknown-shape
	// rejection is two obligations: the schema, and this.
	if !s.Grid.IsSector() && s.Grid != PageThreeGrid() {
		return fmt.Errorf("%w: %dx%d", ErrNotAGrid, s.Grid.Columns, s.Grid.Rows)
	}

	err := s.Grid.HoldAreas(s.OccurrenceAreas)
	if err != nil {
		return err
	}

	err = s.carriesThisToolsProvenance()
	if err != nil {
		return err
	}

	err = s.carriesTheFieldsTheSchemaRequires()
	if err != nil {
		return err
	}

	return s.onItsOwnGrid()
}
```

## 7. Rendering: two documents, one middle

`render` writes both the Markdown listing and the PDF booklet, and stays one
package on purpose. The middle they share is what stops them diverging: the
per-world bullet list was once written out twice, agreed by convention, and
a change to one was a change the other's tests could not see.

```bash
sed -n '644,664p' render/render.go
```

```output
//
// One list, two typesetters. The Markdown listing sets these as bold
// labels through markdown() above, and the booklet lays the same slice out
// as a block of wrapped lines. They agree by construction because there is
// one slice. Two copies would have to agree by convention instead -- two
// suites checking two lists -- and a change to one would be a change the
// other's tests could not see.
func bullets(charts *tables.Tables, world starmap.World) []bullet {
	starport := "no chart row"

	row, err := charts.StarportChart.Row(world.Starport)
	if err == nil {
		starport = row.Description
	}

	// One line for the starport, six for the characteristics of pp. 5-8,
	// one for the technological index, one for the bases, and one for
	// every clamp that bound.
	const fixedLines = 9

	lines := make([]bullet, 0, fixedLines+len(world.Clamps))
```

One reading gets its own file. Page 2 says a route between two worlds
already joined by shorter routes "may be ignored in the drawing" — so the
default documents draw legible lanes, `--lanes all` draws every one, and the
record carries them all either way (ERRATA E007).

```bash
sed -n '40,60p' render/lanes.go
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
```

## 8. Seeing it run

A subsector from a fixed seed. The record is JSON; the listing opens with a
text map of the page 3 grid.

```bash
go run ./cmd/ctworldgen new --seed 1977 --name Aramis --occurrence-dm -1 -o /tmp/ctw-demo.json --force && head -32 /tmp/ctw-demo.json
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
```

```bash
go run ./cmd/ctworldgen render -o /tmp/ctw-demo.md --force /tmp/ctw-demo.json && sed -n '1,30p' /tmp/ctw-demo.md
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

And the same seed a second time, to make the point the whole design turns
on:

```bash
go run ./cmd/ctworldgen new --seed 1977 --name Aramis --occurrence-dm -1 -o /tmp/ctw-again.json --force && diff /tmp/ctw-demo.json /tmp/ctw-again.json && echo 'identical'
```

```output
identical
```

## 9. How the project holds itself together

`task` is the whole gate — tidy, vet, golangci-lint, NilAway, `go test
-race`, and a coverage ratchet — and CI runs exactly `task`. The toolchain
is deliberately unpinned, so the gate fails when a tool moves rather than
drifting behind it.

The ratchet is an integer count of uncovered statements per package, and it
fails in **both** directions: a rising number is lost coverage, a falling one
is coverage to record.

```bash
cat coverage.baseline
```

```output
# Uncovered statements per package. Written by `task ratchet:update`.
# The gate fails in both directions: gaining uncovered statements is
# lost coverage, and losing them is coverage the baseline should record.
github.com/philoserf/ctworldgen/cmd/ctworldgen 36
github.com/philoserf/ctworldgen/dice 0
github.com/philoserf/ctworldgen/gen 11
github.com/philoserf/ctworldgen/internal/cmd/regenerate 127
github.com/philoserf/ctworldgen/internal/fixture 26
github.com/philoserf/ctworldgen/render 4
github.com/philoserf/ctworldgen/starmap 19
github.com/philoserf/ctworldgen/tables 27
```

Two documents govern the code. `docs/ERRATA.md` records every place the page
is silent or ambiguous, with its page cite and its stamping condition; each
record carries the readings that governed it. `docs/COVERAGE.md` maps every
rule to its implementation and its test.

`internal/audit` checks the errata loop in both directions — every `E00N`
cited anywhere resolves to a heading, and every heading is cited at least
once — so a reading cannot be quietly orphaned or quietly invented.

```bash
grep -o '^## E0[0-9]*' docs/ERRATA.md | sed 's/## //' | paste -sd' ' -
```

```output
E001 E002 E003 E004 E005 E006 E007 E008 E009 E010 E011 E012
```

## Where to look first

- **A distance wrong on half the map** — the three parity encodings
  (`starmap.Hex.cube`, `render.gridLine`, `render.mapFit.hexCenter`) and
  their separate measurements against page 3. Never the conversion alone.
- **A record that renders but no longer reproduces** — whether a step in
  `gen` was added, removed or reordered, and whether `EngineVersion` moved
  with it.
- **A number that disagrees with the book** — the data file *and* the second
  transcription in the test, then the page, read visually.
- **A check that looks suspiciously green** — mutate what it claims to hold
  and see whether it dies. Regenerate the goldens first, or it fails for the
  wrong reason.
- **Why a thing is shaped as it is** — `THEORY.md` for the design, `CLAUDE.md`
  for the traps, `docs/ERRATA.md` for every reading of a silent page.

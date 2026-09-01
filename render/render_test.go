package render_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/render"
	"github.com/philoserf/ctworldgen/starmap"
	"github.com/philoserf/ctworldgen/tables"
)

// emptyWorldDigits is a world whose every characteristic threw zero,
// which is what a record built by hand for a question about the listing
// rather than about the dice carries.
const emptyWorldDigits = "A0000000"

func listing(t *testing.T, record *starmap.Record) string {
	t.Helper()

	return listingWith(t, record, render.LegibleLanes)
}

func listingWith(t *testing.T, record *starmap.Record, lanes render.Lanes) string {
	t.Helper()

	renderer, err := render.New(lanes)
	if err != nil {
		t.Fatal(err)
	}

	var built strings.Builder

	err = renderer.Listing(&built, record)
	if err != nil {
		t.Fatal(err)
	}

	return built.String()
}

func generated(t *testing.T, golden fixture.Golden) *starmap.Record {
	t.Helper()

	engine, err := gen.New()
	if err != nil {
		t.Fatal(err)
	}

	record, err := engine.Generate(gen.Inputs{
		Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM,
	})
	if err != nil {
		t.Fatal(err)
	}

	return record
}

// TestListings pins the Markdown. Both golden trees are driven from the
// one roster in internal/fixture, so they cannot come to describe
// different subsectors under the same name.
func TestListings(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			want, err := os.ReadFile(filepath.Join("testdata", golden.File+".md"))
			if err != nil {
				t.Fatalf("%v (run `task regenerate` to create it)", err)
			}

			if listing(t, generated(t, golden)) != string(want) {
				t.Errorf("%s does not match the golden.\n"+
					"If this change was intended, run `task regenerate` and read the diff.", golden.File)
			}
		})
	}
}

// section returns one "## " section of the listing, so a check on the
// roster cannot be satisfied by the detail pages further down. Searching
// the whole document is what let a mutation that dropped half the roster
// survive: every hex still appeared, under its own detail heading.
func section(t *testing.T, written, heading string) string {
	t.Helper()

	start := strings.Index(written, "## "+heading+"\n")
	if start < 0 {
		t.Fatalf("the listing has no %q section", heading)
	}

	rest := written[start+len(heading)+4:]

	body, _, _ := strings.Cut(rest, "\n## ")

	return body
}

// TestEveryWorldAndRouteReachesTheListing is the check a golden cannot
// make: a golden is regenerated from the code under test, so dropping
// half the roster would simply produce a shorter golden.
func TestEveryWorldAndRouteReachesTheListing(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			record := generated(t, golden)
			written := listing(t, record)
			roster := section(t, written, "Worlds")
			details := section(t, written, "The worlds in detail")

			for _, world := range record.Worlds {
				row := "| " + world.Hex.String() + " | " + world.Name + " | " + world.Digits + " |"
				if !strings.Contains(roster, row) {
					t.Errorf("the roster has no row for the world at %s", world.Hex)
				}

				if !strings.Contains(details, "### "+world.Hex.String()+" ") {
					t.Errorf("the detail pages have none for the world at %s", world.Hex)
				}
			}

			if rows := tableRows(roster); rows != len(record.Worlds) {
				t.Errorf("the roster has %d rows and the subsector has %d worlds", rows, len(record.Worlds))
			}

			if pages := strings.Count(details, "\n### "); pages != len(record.Worlds) {
				t.Errorf("there are %d detail pages and %d worlds", pages, len(record.Worlds))
			}

			// The route table is checked under --lanes=all, which is the
			// mode that still promises every lane. What the default draws
			// is TestTheDefaultListingDrawsLegibleLanes.
			all := section(t, listingWith(t, record, render.AllLanes), "Routes")

			for _, route := range record.Routes {
				row := "| " + route.From.String() + " | " + route.To.String() + " |"
				if !strings.Contains(all, row) {
					t.Errorf("--lanes=all has no row for %s-%s", route.From, route.To)
				}
			}

			if rows := tableRows(all); rows != len(record.Routes) {
				t.Errorf("--lanes=all has %d rows and the subsector has %d routes", rows, len(record.Routes))
			}
		})
	}
}

// tableRows counts the body rows of a Markdown table: the lines that open
// with a hex. Counting "\n| 0" instead worked only while every hex began
// with a zero, and would have undercounted in silence on a sector grid,
// where they run to 3240.
func tableRows(section string) int {
	rows := 0

	for line := range strings.SplitSeq(section, "\n") {
		if strings.HasPrefix(line, "| ") && len(line) > 2 && line[2] >= '0' && line[2] <= '9' {
			rows++
		}
	}

	return rows
}

// detailPages splits the detail section into one page per world, keyed by
// hex. A check on a world's own bullets must be made against that world's
// page: searching the whole document lets another world carrying the same
// value satisfy it, which is the mistake the roster check was written to
// avoid.
func detailPages(t *testing.T, written string) map[string]string {
	t.Helper()

	pages := make(map[string]string)

	for _, page := range strings.Split(section(t, written, "The worlds in detail"), "\n### ")[1:] {
		hex, _, _ := strings.Cut(page, " ")

		pages[hex] = page
	}

	return pages
}

// TestLabelsComeFromTheTables: every description the listing prints for a
// value in a table's printed range is that table's own label, and a value
// beyond the range gets no description at all (R16).
func TestLabelsComeFromTheTables(t *testing.T) {
	t.Parallel()

	charts, err := tables.Load()
	if err != nil {
		t.Fatal(err)
	}

	beyond := 0

	for _, golden := range fixture.Goldens() {
		record := generated(t, golden)
		pages := detailPages(t, listing(t, record))

		for _, world := range record.Worlds {
			page, ok := pages[world.Hex.String()]
			if !ok {
				t.Fatalf("%s: the listing has no detail page for the world at %s", golden.File, world.Hex)
			}

			for _, line := range []struct {
				name   string
				value  int
				labels tables.Labels
			}{
				{"Government", world.Government, charts.Government},
				{"Law level", world.LawLevel, charts.LawLevels},
				{"Atmosphere", world.Atmosphere, charts.Atmosphere},
			} {
				label, printed := line.labels.Label(line.value)
				bullet := "- **" + line.name + " " + digitOf(t, line.value) + ".**"

				if printed {
					if !strings.Contains(page, bullet+" "+label+"\n") {
						t.Fatalf("%s: %s at %s should carry the label %q for %d",
							golden.File, line.name, world.Hex, label, line.value)
					}

					continue
				}

				beyond++

				// The note is transcribed a second time here rather than
				// shared with the renderer: a check that read the same
				// constant would assert only that a constant is itself.
				if !strings.Contains(page, bullet+" Above the last row its table prints; p. 8 leaves") {
					t.Fatalf("%s: %s %d at %s is beyond the printed table and should say so, not print a bare digit",
						golden.File, line.name, line.value, world.Hex)
				}
			}
		}
	}

	// R16's gap is the interesting half, so say that it was exercised.
	if beyond == 0 {
		t.Error("no value in the fixtures ran past its table's printed range, so the gap proved nothing")
	}
}

func digitOf(t *testing.T, value int) string {
	t.Helper()

	written, err := starmap.NewDigit(value)
	if err != nil {
		t.Fatal(err)
	}

	return written.String()
}

// TestARefereesOwnNameStaysInOneCell: a world's name is the one field the
// referee writes in himself (p. 12 step 3 prints no table for it), so the
// roster escapes what would otherwise break its row into columns the
// table does not have.
// TestTheRefereesNotesReachBothDocuments: the record is his notebook page,
// so what he writes in it has to come back out when he re-renders (issue 1
// #6). Both documents are asserted in one test because they are required
// to carry the same lines, and checking only one is how the characteristic
// labels came to be held on the Markdown side and by nothing on the
// booklet's.
//
// The awkward text is deliberate. A pipe would break a Markdown table row,
// so the roster escapes it -- but a note is not in a table, and the escape
// must not reach a document with no pipes to escape. A line break would end
// the bullet list. And the booklet draws Windows-1252, which has no
// character for the arrow, so encode promises a question mark rather than a
// silent deletion.
func TestTheRefereesNotesReachBothDocuments(t *testing.T) {
	t.Parallel()

	const (
		onTheMap   = "The rift campaign | players start at 0602"
		onTheWorld = "dust storms;\nask about the yard → 0704"
	)

	hex, err := starmap.NewHex(1, 1)
	if err != nil {
		t.Fatal(err)
	}

	record := starmap.New(1, "Aramis", 0)

	record.Notes = onTheMap
	record.Worlds = append(record.Worlds, starmap.World{
		Hex: hex, Name: "Regina", Notes: onTheWorld, Starport: starmap.StarportA,
		NavalBase: false, ScoutBase: false,
		Size: 0, Atmosphere: 0, Hydrographics: 0, Population: 0,
		Government: 0, LawLevel: 0, TechIndex: 0,
		Digits: emptyWorldDigits, Clamps: nil,
	})

	// The Markdown listing.
	written := listing(t, record)

	if !strings.Contains(written, "\n"+onTheMap+"\n") {
		t.Errorf("the listing does not carry the map's note:\n%s", written)
	}

	if strings.Contains(written, `\|`) {
		t.Error("the listing escaped a pipe outside a table; that escape belongs to Markdown's table syntax alone")
	}

	if !strings.Contains(written, "- **Notes.** dust storms; ask about the yard → 0704\n") {
		t.Errorf("the listing does not carry the world's note on one line:\n%s", written)
	}

	// The booklet, from the same record.
	drawnStamps := everyStamp(t, drawn(t, record))

	if !anyStampWith(drawnStamps, "The rift campaign | players start at 0602") {
		t.Error("the booklet does not carry the map's note")
	}

	if !anyStamp(drawnStamps, "Notes.") {
		t.Error("the booklet does not carry the world's note bullet")
	}

	// The arrow has no Windows-1252 character, so it is drawn as a question
	// mark rather than dropped -- the note keeps the shape he gave it.
	if !anyStampWith(drawnStamps, "ask about the yard ? 0704") {
		t.Errorf("the booklet did not draw the world's note as one line with the unmappable character kept:\n%v",
			drawnStamps)
	}
}

func TestARefereesOwnNameStaysInOneCell(t *testing.T) {
	t.Parallel()

	hex, err := starmap.NewHex(1, 1)
	if err != nil {
		t.Fatal(err)
	}

	record := starmap.New(1, "Aramis", 0)

	record.Worlds = append(record.Worlds, starmap.World{
		Hex: hex, Name: "Regina | the\nold capital", Notes: "", Starport: starmap.StarportA,
		NavalBase: false, ScoutBase: false,
		Size: 0, Atmosphere: 0, Hydrographics: 0, Population: 0,
		Government: 0, LawLevel: 0, TechIndex: 0,
		Digits: emptyWorldDigits, Clamps: nil,
	})

	var rows []string

	for line := range strings.SplitSeq(section(t, listing(t, record), "Worlds"), "\n") {
		if strings.HasPrefix(line, "| 0101 ") {
			rows = append(rows, line)
		}
	}

	if len(rows) != 1 {
		t.Fatalf("the world at 0101 has %d roster rows, want 1: %q", len(rows), rows)
	}

	if columns := strings.Count(rows[0], "|") - strings.Count(rows[0], `\|`); columns != 5 {
		t.Errorf("the row divides into %d cells and the table has four columns: %s", columns-1, rows[0])
	}

	assertTheRouteTableStaysInThreeCells(t, record, hex)
}

// assertTheRouteTableStaysInThreeCells: the route table carries the same
// name in both of its ends, so it needs the same escape the roster does.
//
// It is asserted separately because the escape lives in a different place
// from the roster's. named() does not escape -- it feeds the PDF booklet
// too, which has no pipes to escape and drew a backslash the referee
// never typed while named() did -- so the Markdown route emitter applies
// cell() itself, and this is what holds it there.
func assertTheRouteTableStaysInThreeCells(t *testing.T, record *starmap.Record, from starmap.Hex) {
	t.Helper()

	other, err := starmap.NewHex(1, 2)
	if err != nil {
		t.Fatal(err)
	}

	record.Worlds = append(record.Worlds, starmap.World{
		Hex: other, Name: "Efate", Notes: "", Starport: starmap.StarportA,
		NavalBase: false, ScoutBase: false,
		Size: 0, Atmosphere: 0, Hydrographics: 0, Population: 0,
		Government: 0, LawLevel: 0, TechIndex: 0,
		Digits: emptyWorldDigits, Clamps: nil,
	})
	record.Routes = append(record.Routes, starmap.Route{From: from, To: other, Distance: 1})

	var rows []string

	for line := range strings.SplitSeq(section(t, listing(t, record), "Routes"), "\n") {
		if strings.HasPrefix(line, "| 0101 ") {
			rows = append(rows, line)
		}
	}

	if len(rows) != 1 {
		t.Fatalf("the route from 0101 has %d rows, want 1: %q", len(rows), rows)
	}

	if columns := strings.Count(rows[0], "|") - strings.Count(rows[0], `\|`); columns != 4 {
		t.Errorf("the row divides into %d cells and the table has three columns: %s", columns-1, rows[0])
	}
}

// TestAnEmptySubsectorRenders: a run whose eighty throws place no world
// produces a valid record, and the listing says so rather than printing an
// empty table.
func TestAnEmptySubsectorRenders(t *testing.T) {
	t.Parallel()

	written := listing(t, starmap.New(7, "", 0))

	for _, want := range []string{"# Subsector", "No world was placed", "No route was drawn"} {
		if !strings.Contains(written, want) {
			t.Errorf("the listing of an empty subsector has no %q:\n%s", want, written)
		}
	}

	if strings.Contains(written, "## The worlds in detail") {
		t.Error("the listing of an empty subsector has a detail section")
	}
}

// annotated is a record with names written in on some worlds and not
// others, which is the state a referee's file is actually in: he names
// the worlds he has used and leaves the rest as hexes.
func annotated(t *testing.T) *starmap.Record {
	t.Helper()

	record := starmap.New(1, "Aramis", 0)

	for _, world := range []struct {
		col, row int
		name     string
	}{{1, 1, "Regina"}, {1, 2, "Efate"}, {1, 3, ""}} {
		hex, err := starmap.NewHex(world.col, world.row)
		if err != nil {
			t.Fatal(err)
		}

		record.Worlds = append(record.Worlds, starmap.World{
			Hex: hex, Name: world.name, Notes: "", Starport: starmap.StarportA,
			NavalBase: false, ScoutBase: false,
			Size: 0, Atmosphere: 0, Hydrographics: 0, Population: 0,
			Government: 0, LawLevel: 0, TechIndex: 0,
			Digits: emptyWorldDigits, Clamps: nil,
		})
	}

	record.Routes = append(record.Routes,
		starmap.Route{From: record.Worlds[0].Hex, To: record.Worlds[1].Hex, Distance: 1},
		starmap.Route{From: record.Worlds[1].Hex, To: record.Worlds[2].Hex, Distance: 1})

	return record
}

// TestARefereesNameReachesEveryPlaceTheHexAppears: a name written into
// the record reaches the roster, the world's own detail heading, and both
// columns of the route table -- a hex the referee has named should not go
// on reading as a bare number anywhere he meets it.
//
// Every assertion is against one exact line, not against the document.
// Searching the whole listing for a name is satisfied by the roster alone,
// and searching a route table for "Regina" is satisfied by either column;
// both are the shape of check that has already gone dead here twice.
func TestARefereesNameReachesEveryPlaceTheHexAppears(t *testing.T) {
	t.Parallel()

	written := listing(t, annotated(t))
	pages := detailPages(t, written)

	for hex, want := range map[string]string{"0101": "0101 Regina", "0102": "0102 Efate", "0103": "0103"} {
		if !strings.HasPrefix(pages[hex], want+" &mdash; ") {
			t.Errorf("the detail heading at %s should read %q and reads %q",
				hex, want, line(pages[hex]))
		}
	}

	// Both route rows are pinned whole. The first names both endpoints;
	// the second names only its From, so a change that filled the To
	// column from the wrong world would show here.
	routes := section(t, written, "Routes")
	for _, want := range []string{"| 0101 Regina | 0102 Efate | 1 |", "| 0102 Efate | 0103 | 1 |"} {
		if !strings.Contains(routes, "\n"+want+"\n") {
			t.Errorf("the route table has no row %q:\n%s", want, routes)
		}
	}

	roster := section(t, written, "Worlds")
	if !strings.Contains(roster, "\n| 0101 | Regina | ") {
		t.Errorf("the roster row for 0101 does not carry its name:\n%s", roster)
	}
}

// line returns the first line of a string, for an error message that
// quotes a heading rather than the page beneath it.
func line(s string) string {
	first, _, _ := strings.Cut(s, "\n")

	return first
}

// TestTheListingSaysWhyTheTechnologicalIndexIsBare: pp. 10-11 are not
// transcribed, so every world's technological index prints its digit
// alone. That is 670 bare lines in a sector, and the
// listing says once why, rather than in every one of them or nowhere.
func TestTheListingSaysWhyTheTechnologicalIndexIsBare(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			record := generated(t, golden)
			written := listing(t, record)

			if len(record.Worlds) == 0 {
				return
			}

			details := section(t, written, "The worlds in detail")
			if said := strings.Count(details, "technological levels tables of pp. 10-11"); said != 1 {
				t.Errorf("the detail section explains the bare technological index %d times, want once", said)
			}

			for _, page := range detailPages(t, written) {
				start := strings.Index(page, "- **Technological index ")
				if start < 0 {
					t.Fatalf("a detail page has no technological index line:\n%s", page)
				}

				if got := line(page[start:]); !strings.HasSuffix(got, ".**") {
					t.Errorf("the technological index line carries a description: %q", got)
				}
			}
		})
	}
}

// hexPlace is where a hex's label was drawn on the map, in characters.
type hexPlace struct{ line, char int }

// mapBlock returns the fenced drawing alone, so the prose above it cannot
// answer a question asked of the map.
func mapBlock(t *testing.T, body string) string {
	t.Helper()

	_, after, found := strings.Cut(body, "```text\n")
	if !found {
		t.Fatal("the map section has no drawing")
	}

	block, _, found := strings.Cut(after, "```")
	if !found {
		t.Fatal("the map's drawing is not closed")
	}

	return block
}

// mapPlaces reads back where every hex label was drawn.
func mapPlaces(t *testing.T, block string) map[starmap.Hex]hexPlace {
	t.Helper()

	places := make(map[starmap.Hex]hexPlace)

	for number, text := range strings.Split(block, "\n") {
		for char := 0; char+4 <= len(text); char++ {
			hex, err := starmap.ParseHex(text[char : char+4])
			if err != nil {
				continue
			}

			if where, drawn := places[hex]; drawn {
				t.Fatalf("%s is drawn twice, at line %d and line %d", hex, where.line, number)
			}

			places[hex] = hexPlace{line: number, char: char}
		}
	}

	return places
}

func hexOf(t *testing.T, col, row int) starmap.Hex {
	t.Helper()

	hex, err := starmap.NewHex(col, row)
	if err != nil {
		t.Fatal(err)
	}

	return hex
}

// mapGeometry measures the drawing's slot and row off the drawing itself,
// rather than taking the renderer's constants. A check fed the constants
// the renderer drew with would agree with a map drawn upside down.
func mapGeometry(t *testing.T, places map[starmap.Hex]hexPlace) (int, int) {
	t.Helper()

	// One slot across is two printed columns; one row down is two lines.
	// Measured over the drawing rather than read off named hexes: a
	// member's map begins at its own band and its ring a hex outside that,
	// so 0101 is not on most of the maps this holds.
	slot, step := 0, 0

	for _, place := range places {
		for _, against := range places {
			slot, step = closerPlace(place, against, slot, step)
		}
	}

	if slot == 0 || step == 0 {
		t.Fatal("this map has no two hexes to measure a slot and a row step between")
	}

	if slot <= 0 || step <= 0 {
		t.Fatalf("the map's slot is %d characters and its row %d lines; both run forwards", slot, step)
	}

	if slot%2 != 0 || step%2 != 0 {
		t.Fatalf("the map's slot is %d characters and its row %d lines; a hex sits on the half of each", slot, step)
	}

	return slot, step
}

// drawnTouching returns the hexes the drawing puts against this one: half
// a slot across and half a row down, or squarely one row above or below.
func drawnTouching(places map[starmap.Hex]hexPlace, hex starmap.Hex, slot, step int) map[starmap.Hex]bool {
	touching := make(map[starmap.Hex]bool)
	from := places[hex]

	for other, where := range places {
		if other == hex {
			continue
		}

		across, down := abs(where.char-from.char), abs(where.line-from.line)
		if (across == slot/2 && down == step/2) || (across == 0 && down == step) {
			touching[other] = true
		}
	}

	return touching
}

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}

// TestTheMapIsTheGridPrintedOnPageThree is the parity check the map needs,
// and it is the trap Hex.Distance has: draw the even-numbered columns high
// instead of low and every hex still lands in a tidy grid, the drawing
// still looks like p. 3, and it is wrong by one for half the subsector.
//
// So the drawing is measured against Hex.Distance, which
// TestDistanceAgainstPrintedGrid anchors to the printed page by hand. Two
// hexes drawn touching must be one parsec apart, and two hexes one parsec
// apart must be drawn touching. An inverted parity fails both ways: it
// draws 0603 against 0502, which is two parsecs off, and pulls 0603 and
// 0704 apart.
//
// It runs on a populated map as well as an empty one, and the empty one is
// not enough on its own: a cell padded to the width of its bare hex rather
// than to the width it drew shifts every later column of a row carrying a
// starport letter, which an empty map cannot show. That mutation survived
// the marking test outright, because the marking test reads a slot at the
// position the label was found at and so cannot see a label in the wrong
// place. Here it pulls hexes apart and the check fails.
func TestTheMapIsTheGridPrintedOnPageThree(t *testing.T) {
	t.Parallel()

	// The empty record pins that all eighty hexes are drawn with no world
	// on the map at all; the goldens carry the starport letters.
	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		assertTheMapIsThePrintedGrid(t, section(t, listing(t, starmap.New(1, "Aramis", 0)), "The map"),
			everyHexOfGrid(starmap.PageThreeGrid()))
	})

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()
			assertTheMapIsThePrintedGrid(t, section(t, listing(t, generated(t, golden)), "The map"),
				everyHexOfGrid(starmap.PageThreeGrid()))
		})
	}
}

// assertTheMapIsThePrintedGrid holds one drawing against Hex.Distance.
func assertTheMapIsThePrintedGrid(t *testing.T, body string, want map[starmap.Hex]bool) {
	t.Helper()

	if len(want) == 0 {
		t.Fatal("this check was given no hexes to look for, so it could only pass")
	}

	places := mapPlaces(t, mapBlock(t, body))

	assertTheseAreTheHexesDrawn(t, places, want)

	slot, step := mapGeometry(t, places)

	for hex := range places {
		touching := drawnTouching(places, hex, slot, step)

		for other := range touching {
			if apart := hex.Distance(other); apart != 1 {
				t.Errorf("the map draws %s touching %s, and the p. 3 grid puts them %d parsecs apart",
					hex, other, apart)
			}
		}

		for other := range places {
			if hex.Distance(other) == 1 && !touching[other] {
				t.Errorf("%s and %s are one parsec apart on the p. 3 grid and the map draws them apart",
					hex, other)
			}
		}
	}
}

// closerPlace narrows the smallest slot and row step seen so far by one
// pair of drawn labels.
func closerPlace(place, against hexPlace, slot, step int) (int, int) {
	if place.line == against.line && against.char > place.char && (slot == 0 || against.char-place.char < slot) {
		slot = against.char - place.char
	}

	if place.char == against.char && against.line > place.line && (step == 0 || against.line-place.line < step) {
		step = against.line - place.line
	}

	return slot, step
}

// assertTheseAreTheHexesDrawn holds a drawing to an exact set of hexes.
func assertTheseAreTheHexesDrawn(t *testing.T, places map[starmap.Hex]hexPlace, want map[starmap.Hex]bool) {
	t.Helper()

	for hex := range want {
		if _, drew := places[hex]; !drew {
			t.Fatalf("the map does not draw %s, and it should (%d of %d drawn)", hex, len(places), len(want))
		}
	}

	for hex := range places {
		if !want[hex] {
			t.Fatalf("the map draws %s, and it should not (%d drawn, %d wanted)", hex, len(places), len(want))
		}
	}
}

// TestTheMapMarksWhatPageOneSaysToMark: a world's hex carries the letter
// of its starport (p. 1 step 2) and a hex with no world is left blank
// (p. 1 step 1). Each hex's own slot is checked, so another world's letter
// cannot answer for it.
func TestTheMapMarksWhatPageOneSaysToMark(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			record := generated(t, golden)
			block := mapBlock(t, listing(t, record))
			lines := strings.Split(block, "\n")
			places := mapPlaces(t, block)
			slot, _ := mapGeometry(t, places)

			marked := make(map[starmap.Hex]starmap.Starport, len(record.Worlds))
			for _, world := range record.Worlds {
				marked[world.Hex] = world.Starport
			}

			for hex, where := range places {
				want := hex.String()
				if starport, ok := marked[hex]; ok {
					want += " " + starport.String()
				}

				if got := drawnSlot(lines[where.line], where.char, slot); got != want {
					t.Errorf("the map draws %s as %q, want %q", hex, got, want)
				}
			}
		})
	}
}

// drawnSlot returns one hex's slot of a map line, its padding removed. The
// last slot on a line is short, that padding having been trimmed when the
// line was written.
func drawnSlot(line string, char, slot int) string {
	return strings.TrimSpace(line[char:min(char+slot, len(line))])
}

// everyHexOfGrid is the set a map of a whole grid draws.
func everyHexOfGrid(grid starmap.Grid) map[starmap.Hex]bool {
	want := map[starmap.Hex]bool{}

	for col := 1; col <= grid.Columns; col++ {
		for row := 1; row <= grid.Rows; row++ {
			want[starmap.Hex{Col: col, Row: row}] = true
		}
	}

	return want
}

// memberAndItsRingHexes is the set one member's map draws: its own eighty
// hexes, and every hex of the sector one parsec outside them (ERRATA E008
// part 2).
//
// Measured against every hex of the band, which is the slow road; the
// renderer clamps a hex into the band and measures once, which is the
// fast one. Two roads to one set, so a ring drawn a row out of place
// cannot pass.
func memberAndItsRingHexes(index int) map[starmap.Hex]bool {
	grid := starmap.SectorGrid()
	first, last := starmap.MemberBounds(index)

	own := []starmap.Hex{}

	for col := first.Col; col <= last.Col; col++ {
		for row := first.Row; row <= last.Row; row++ {
			own = append(own, starmap.Hex{Col: col, Row: row})
		}
	}

	want := map[starmap.Hex]bool{}
	for _, hex := range own {
		want[hex] = true
	}

	for col := 1; col <= grid.Columns; col++ {
		for row := 1; row <= grid.Rows; row++ {
			hex := starmap.Hex{Col: col, Row: row}
			if want[hex] {
				continue
			}

			for _, at := range own {
				if hex.Distance(at) == 1 {
					want[hex] = true

					break
				}
			}
		}
	}

	return want
}

// sectorListing generates the sector the listing tests read.
func sectorListingRecord(t *testing.T) *starmap.Record {
	t.Helper()

	engine, err := gen.New()
	if err != nil {
		t.Fatal(err)
	}

	record, err := engine.Sector(gen.Inputs{Seed: 1, Name: "Aramis", OccurrenceDM: 0})
	if err != nil {
		t.Fatal(err)
	}

	return record
}

// memberSection cuts one of the sixteen "## Subsector N" sections out of a
// sector's listing, so a check on one member cannot be answered by
// another's.
func memberSection(t *testing.T, written string, index int) string {
	t.Helper()

	heading := fmt.Sprintf("## Subsector %d &mdash; ", index)

	start := strings.Index(written, heading)
	if start < 0 {
		t.Fatalf("the listing has no section for member %d", index)
	}

	body, _, _ := strings.Cut(written[start+len(heading):], "\n## ")

	return body
}

// subsection cuts one "### " part out of a member's section.
func subsection(t *testing.T, body, heading string) string {
	t.Helper()

	start := strings.Index(body, "### "+heading+"\n")
	if start < 0 {
		t.Fatalf("this member's section has no %q part", heading)
	}

	part, _, _ := strings.Cut(body[start+len(heading)+5:], "\n### ")

	return part
}

// TestTheSectorSliceGolden pins one member's section of a sector's
// listing, which is where the decomposition, the ring of neighbours and
// the doubled crossing lanes can be read at a size a human will read.
//
// It catches drift and nothing else. A golden is regenerated from the
// code under test, so what holds the decomposition itself is the live
// assertions above it.
func TestTheSectorSliceGolden(t *testing.T) {
	t.Parallel()

	var whole strings.Builder

	renderer, err := render.New(render.LegibleLanes)
	if err != nil {
		t.Fatal(err)
	}

	golden := fixture.SectorGolden()

	engine, err := gen.New()
	if err != nil {
		t.Fatal(err)
	}

	record, err := engine.Sector(gen.Inputs{
		Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = renderer.Listing(&whole, record)
	if err != nil {
		t.Fatal(err)
	}

	got := fixture.SectorSlice(whole.String())
	if got == "" {
		t.Fatalf("the sector listing has no section for member %d", fixture.SectorSliceMember)
	}

	want, err := os.ReadFile(filepath.Join("..", fixture.SectorSlicePath()))
	if err != nil {
		t.Fatal(err)
	}

	if got != string(want) {
		t.Errorf("subsector %d's section has changed; run `task regenerate` and read the diff",
			fixture.SectorSliceMember)
	}
}

// TestARefereesNoteReachesASectorsIndexPage: the note goes under the
// summary and above everything the tool generated (issue 1 #6), on a
// sector's index page as on a subsector's first page.
func TestARefereesNoteReachesASectorsIndexPage(t *testing.T) {
	t.Parallel()

	record := sectorListingRecord(t)

	record.Notes = "the Spinward Marches, such as they are"

	if written := listing(t, record); !strings.Contains(written, record.Notes) {
		t.Error("the referee's note about the sector is not in its listing")
	}

	renderer, err := render.New(render.LegibleLanes)
	if err != nil {
		t.Fatal(err)
	}

	var booklet bytes.Buffer

	err = renderer.Booklet(&booklet, record)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(booklet.Bytes(), []byte("Spinward Marches")) {
		t.Error("the referee's note about the sector is not in its booklet")
	}
}

// TestASectorsMemberMapsAreThePrintedGrid: a sector's listing draws
// sixteen p. 3 grids rather than one map of 1,280 hexes, and each is held
// to the parity the same way a subsector's is (ERRATA E008).
//
// The ring of neighbours is what keeps the sixteen in one frame. Member
// 5's map draws 0810, which is member 4's, one hex from its own 0911, so
// every seam relation the single grid stated is stated again here.
func TestASectorsMemberMapsAreThePrintedGrid(t *testing.T) {
	t.Parallel()

	written := listing(t, sectorListingRecord(t))

	for index := range starmap.Members {
		assertTheMapIsThePrintedGrid(t,
			subsection(t, memberSection(t, written, index), "The map"),
			memberAndItsRingHexes(index))
	}
}

// TestASectorsMemberSectionsPartitionItsWorlds: each of the sixteen
// rosters carries exactly the worlds of its own band, and between them
// they carry every world once (ERRATA E008 part 1).
//
// A count would not do this. Sixteen rosters summing to 662 rows survives
// member 5's worlds being printed under member 6, which is the mutation
// the band arithmetic can make.
func TestASectorsMemberSectionsPartitionItsWorlds(t *testing.T) {
	t.Parallel()

	record := sectorListingRecord(t)
	written := listing(t, record)

	seen := map[starmap.Hex]int{}

	for index := range starmap.Members {
		roster := subsection(t, memberSection(t, written, index), "Worlds")

		for _, world := range record.Worlds {
			at := starmap.MemberOf(world.Hex)
			carries := strings.Contains(roster, "| "+world.Hex.String()+" |")

			if carries != (at == index) {
				t.Fatalf("subsector %d's roster carries %s = %v, and that world is in subsector %d",
					index, world.Hex, carries, at)
			}

			if carries {
				seen[world.Hex]++
			}
		}
	}

	for _, world := range record.Worlds {
		if seen[world.Hex] != 1 {
			t.Errorf("%s appears in %d of the sixteen rosters", world.Hex, seen[world.Hex])
		}
	}
}

// TestACrossingLaneIsListedUnderBothItsSubsectors: a lane whose two ends
// are in different members is in both their tables, because a referee
// reading one sub-sector needs to see the road out of it (ERRATA E008
// part 3). A lane at home is in one.
func TestACrossingLaneIsListedUnderBothItsSubsectors(t *testing.T) {
	t.Parallel()

	record := sectorListingRecord(t)
	written := listingWith(t, record, render.AllLanes)

	tables := make([]string, starmap.Members)
	for index := range tables {
		tables[index] = subsection(t, memberSection(t, written, index), "Routes")
	}

	rows := 0
	crossing := 0

	for _, route := range record.Routes {
		row := fmt.Sprintf("| %s | %s | %d |", route.From, route.To, route.Distance)
		home := starmap.MemberOf(route.From)
		across := starmap.MemberOf(route.To)

		for index, table := range tables {
			listed := strings.Contains(table, row)
			want := index == home || index == across

			if listed != want {
				t.Fatalf("subsector %d lists %s-%s = %v, and that lane joins subsectors %d and %d",
					index, route.From, route.To, listed, home, across)
			}
		}

		rows++

		if home != across {
			rows++

			crossing++
		}
	}

	// The consequence, worth stating because the document states it: the
	// sixteen tables carry more rows between them than the record has
	// lanes, by exactly the number that cross.
	counted := 0
	for _, table := range tables {
		counted += tableRows(table)
	}

	if counted != rows {
		t.Errorf("the sixteen lane tables carry %d rows; %d lanes and %d crossings make %d",
			counted, len(record.Routes), crossing, rows)
	}
}

// TestTheListingSaysWhichGridItDrew: the listing is a render of the
// record, and a record on the 32x40 grid is not a subsector. An unnamed
// sector headed "# Subsector", or a map note calling sixteen grids "the
// p. 3 sub-sector hex grid", reads perfectly well while misreporting its
// own subject -- which is the only way this can go wrong.
func TestTheListingSaysWhichGridItDrew(t *testing.T) {
	t.Parallel()

	engine, err := gen.New()
	if err != nil {
		t.Fatal(err)
	}

	// Unnamed, because the heading under test is the one a record with no
	// referee's name of its own falls back to.
	asSector, err := engine.Sector(gen.Inputs{Seed: 1, Name: "", OccurrenceDM: 0})
	if err != nil {
		t.Fatal(err)
	}

	sectorListing := listing(t, asSector)

	if !strings.HasPrefix(sectorListing, "# Sector\n") {
		t.Errorf("an unnamed sector is headed %q", line(sectorListing))
	}

	if note := noteLine(t, sectorListing, "The sector"); !strings.Contains(note, "3240") {
		t.Errorf("the map note of a sector does not say which grid was drawn: %q", note)
	}

	subsectorListing := listing(t, starmap.New(1, "", 0))

	if !strings.HasPrefix(subsectorListing, "# Subsector\n") {
		t.Errorf("an unnamed subsector is headed %q", line(subsectorListing))
	}

	if note := noteLine(t, subsectorListing, "The map"); !strings.HasPrefix(note, "The p. 3 sub-sector hex grid.") {
		t.Errorf("the map note of a subsector no longer names the p. 3 grid: %q", note)
	}
}

// noteLine is the map section's prose: its first non-empty line, which is
// the sentence naming the grid.
func noteLine(t *testing.T, written, heading string) string {
	t.Helper()

	return line(strings.TrimLeft(section(t, written, heading), "\n"))
}

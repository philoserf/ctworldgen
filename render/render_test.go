package render_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/render"
	"github.com/philoserf/ctworldgen/subsector"
	"github.com/philoserf/ctworldgen/tables"
)

func listing(t *testing.T, record *subsector.Subsector) string {
	t.Helper()

	renderer, err := render.New()
	if err != nil {
		t.Fatal(err)
	}

	var built strings.Builder

	err = renderer.Subsector(&built, record)
	if err != nil {
		t.Fatal(err)
	}

	return built.String()
}

func generated(t *testing.T, golden fixture.Golden) *subsector.Subsector {
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

// TestEveryWorldAndLaneReachesTheListing is the check a golden cannot
// make: a golden is regenerated from the code under test, so dropping
// half the roster would simply produce a shorter golden.
func TestEveryWorldAndLaneReachesTheListing(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			record := generated(t, golden)
			written := listing(t, record)
			roster := section(t, written, "Worlds")
			lanes := section(t, written, "Space lanes")
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

			if rows := strings.Count(roster, "\n| 0"); rows != len(record.Worlds) {
				t.Errorf("the roster has %d rows and the subsector has %d worlds", rows, len(record.Worlds))
			}

			if pages := strings.Count(details, "\n### "); pages != len(record.Worlds) {
				t.Errorf("there are %d detail pages and %d worlds", pages, len(record.Worlds))
			}

			for _, route := range record.Routes {
				row := "| " + route.From.String() + " | " + route.To.String() + " |"
				if !strings.Contains(lanes, row) {
					t.Errorf("the lanes table has no row for %s-%s", route.From, route.To)
				}
			}

			if rows := strings.Count(lanes, "\n| 0"); rows != len(record.Routes) {
				t.Errorf("the lanes table has %d rows and the subsector has %d lanes", rows, len(record.Routes))
			}
		})
	}
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

				if !strings.Contains(page, bullet+"\n") {
					t.Fatalf("%s: %s %d at %s is beyond the printed table and should carry the digit alone",
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

	written, err := subsector.NewDigit(value)
	if err != nil {
		t.Fatal(err)
	}

	return written.String()
}

// TestARefereesOwnNameStaysInOneCell: a world's name is the one field the
// referee writes in himself (p. 12 step 3 prints no table for it), so the
// roster escapes what would otherwise break its row into columns the
// table does not have.
func TestARefereesOwnNameStaysInOneCell(t *testing.T) {
	t.Parallel()

	hex, err := subsector.NewHex(1, 1)
	if err != nil {
		t.Fatal(err)
	}

	record := subsector.New(1, "Aramis", 0)

	record.Worlds = append(record.Worlds, subsector.World{
		Hex: hex, Name: "Regina | the\nold capital", Starport: subsector.StarportA,
		NavalBase: false, ScoutBase: false,
		Size: 0, Atmosphere: 0, Hydrographics: 0, Population: 0,
		Government: 0, LawLevel: 0, TechIndex: 0,
		Digits: "A0000000", Clamps: nil,
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
}

// TestAnEmptySubsectorRenders: a run whose eighty throws place no world
// produces a valid record, and the listing says so rather than printing an
// empty table.
func TestAnEmptySubsectorRenders(t *testing.T) {
	t.Parallel()

	written := listing(t, subsector.New(7, "", 0))

	for _, want := range []string{"# Subsector", "No world was placed", "No space lane was drawn"} {
		if !strings.Contains(written, want) {
			t.Errorf("the listing of an empty subsector has no %q:\n%s", want, written)
		}
	}

	if strings.Contains(written, "## The worlds in detail") {
		t.Error("the listing of an empty subsector has a detail section")
	}
}

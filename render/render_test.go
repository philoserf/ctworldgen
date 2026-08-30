package render_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/render"
	"github.com/philoserf/ctworldgen/worldgen"
)

var update = flag.Bool("update", false, "rewrite the golden renders")

// TestGoldenRenders holds both renders of every fixture to the file on
// disk. The subsector listing is the readable view of a record whose JSON
// runs to a third of a megabyte, so it is also where a semantic change
// shows up legibly.
func TestGoldenRenders(t *testing.T) {
	for _, f := range fixture.All() {
		sub, err := worldgen.Generate(f.Config)
		if err != nil {
			t.Fatalf("Generate(%s): %v", f.Name, err)
		}

		listing, err := render.Listing(sub)
		if err != nil {
			t.Fatalf("Listing(%s): %v", f.Name, err)
		}

		for _, out := range []struct {
			suffix string
			got    string
		}{
			{"listing", listing},
			{"history", render.History(sub)},
		} {
			t.Run(f.Name+"/"+out.suffix, func(t *testing.T) {
				checkGolden(t, filepath.Join("testdata", f.Name+"."+out.suffix+".md"), out.got)
			})
		}
	}
}

func checkGolden(t *testing.T, path, got string) {
	t.Helper()

	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil { //nolint:gosec // test fixture
			t.Fatalf("writing %s: %v", path, err)
		}

		return
	}

	want, err := os.ReadFile(path) //nolint:gosec // a testdata path this test built
	if err != nil {
		t.Fatalf("reading %s (run `task goldens` to create it): %v", path, err)
	}

	if got != string(want) {
		t.Errorf("%s does not match; run `task goldens` if the change was intended", path)
	}
}

// TestListingOfAnEmptySubsector covers the one outcome no fixture can
// reach. Eighty occurrence throws that all fail is a valid record
// (docs/PRD.md, Decisions), but at a half chance per hex it has a
// probability of 2^-80, so no seed will ever produce one and the record is
// built here directly.
func TestListingOfAnEmptySubsector(t *testing.T) {
	sub := &worldgen.Subsector{Name: "Void", Worlds: []worldgen.World{}, Routes: []worldgen.Route{}}

	out, err := render.Listing(sub)
	if err != nil {
		t.Fatalf("Listing: %v", err)
	}

	for _, want := range []string{
		"# Void Subsector",
		"0 worlds in 80 hexes",
		"No world is present in any hex",
		"No space lane connects any pair of worlds",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("listing of an empty subsector does not mention %q:\n%s", want, out)
		}
	}

	if strings.Contains(out, "## World details") {
		t.Error("listing of an empty subsector has a World details section")
	}
}

// TestListingOfAnUnnamedSubsector: --name is optional, and the heading has
// to read as a heading without it.
func TestListingOfAnUnnamedSubsector(t *testing.T) {
	out, err := render.Listing(&worldgen.Subsector{})
	if err != nil {
		t.Fatalf("Listing: %v", err)
	}

	if !strings.HasPrefix(out, "# Subsector\n") {
		heading, _, _ := strings.Cut(out, "\n")
		t.Errorf("unnamed subsector heading is %q", heading)
	}
}

// TestListingNamesAWorldThatHasOne. The book asks the referee to name each
// world (p. 12 step 3) and prints no table, so the engine leaves the field
// empty — but a record a referee has filled in must render the name.
func TestListingNamesAWorldThatHasOne(t *testing.T) {
	sub := &worldgen.Subsector{
		Worlds: []worldgen.World{{Hex: "0101", Name: "Regina", Profile: "A788899-A", Starport: "A"}},
	}

	out, err := render.Listing(sub)
	if err != nil {
		t.Fatalf("Listing: %v", err)
	}

	if !strings.Contains(out, "### Regina (0101 A788899-A)") {
		t.Errorf("named world has no named detail heading:\n%s", out)
	}
}

// TestListingOfAValueOutsideItsTable: law level is floored but never
// capped (docs/ERRATA.md E004), so a value above the p. 7 table's last row
// is a real state and says so. A value below zero is not — only a
// hand-edited record reaches it — and it must not be reported as being
// above the table.
func TestListingOfAValueOutsideItsTable(t *testing.T) {
	sub := &worldgen.Subsector{
		Name:   "Edge",
		Worlds: []worldgen.World{{Hex: "0101", Profile: "A000000J", Starport: "A", LawLevel: 18, Size: -1}},
		Routes: []worldgen.Route{},
	}

	out, err := render.Listing(sub)
	if err != nil {
		t.Fatalf("Listing: %v", err)
	}

	if !strings.Contains(out, "Law level 18: above the last row the printed table gives") {
		t.Errorf("an uncapped law level is not reported as above the table:\n%s", out)
	}

	if !strings.Contains(out, "Planetary size -1: outside the printed table") {
		t.Errorf("a below-range value is not reported as outside the table:\n%s", out)
	}
}

// TestListingEscapesARefereesName: the name is hand-written into the record
// after generation, so it can hold a pipe, which would otherwise shift
// every column boundary in its row.
func TestListingEscapesARefereesName(t *testing.T) {
	sub := &worldgen.Subsector{
		Name:   "Marches",
		Worlds: []worldgen.World{{Hex: "0101", Name: "Ally | Fortress\nsecond line", Profile: "A0000000", Starport: "A"}},
		Routes: []worldgen.Route{},
	}

	out, err := render.Listing(sub)
	if err != nil {
		t.Fatalf("Listing: %v", err)
	}

	if !strings.Contains(out, `| Ally \| Fortress second line |`) {
		t.Errorf("the name was not escaped into one cell:\n%s", out)
	}

	// The escaped pipe still is a pipe, so count only the unescaped ones:
	// those are the cell boundaries, and the table has four columns.
	for line := range strings.SplitSeq(out, "\n") {
		if !strings.HasPrefix(line, "| 0101 ") {
			continue
		}

		if got := strings.Count(strings.ReplaceAll(line, `\|`, ""), "|"); got != 5 {
			t.Errorf("the world row has %d cell boundaries, want the table's 5: %q", got, line)
		}
	}
}

// TestHistoryReportsTheReadings: the transcript is the audit trail, so it
// has to say which docs/ERRATA.md readings governed the run.
func TestHistoryReportsTheReadings(t *testing.T) {
	sub, err := worldgen.Generate(worldgen.Config{Seed: 7})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	out := render.History(sub)
	for _, id := range sub.Errata {
		if !strings.Contains(out, id) {
			t.Errorf("history does not report the stamped reading %s", id)
		}
	}
}

// TestHistoryOfEventsWithoutAStep: a record whose log begins before any
// step event is malformed, but `render` reads files a user names, so the
// transcript must show those events rather than swallow them.
func TestHistoryOfEventsWithoutAStep(t *testing.T) {
	sub := &worldgen.Subsector{
		Events: []worldgen.Event{
			{Seq: 1, Kind: worldgen.KindOutcome, Step: "orphan", Text: "an outcome with no step"},
			{Seq: 2, Kind: "something else", Step: "orphan", Text: "an event of no known kind"},
		},
	}

	out := render.History(sub)
	for _, want := range []string{"## (no step)", "an outcome with no step", "an event of no known kind"} {
		if !strings.Contains(out, want) {
			t.Errorf("history does not contain %q:\n%s", want, out)
		}
	}
}

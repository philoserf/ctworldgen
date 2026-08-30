package tables_test

import (
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/dice"
	"github.com/philoserf/ctworldgen/tables"
)

func load(t *testing.T) *tables.Charts {
	t.Helper()

	charts, err := tables.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	return charts
}

// TestStarportsTableMatchesThePage transcribes Book 3 p. 1's two-dice
// starports table a second time, here, and checks the data file against
// it. Two independent transcriptions of the same eleven rows is the point:
// the held PDF's font makes a text extraction of these pages unreliable
// (docs/ERRATA.md, Noted discrepancies), so the table is worth reading
// twice.
func TestStarportsTableMatchesThePage(t *testing.T) {
	charts := load(t)

	page := map[int]string{
		2: "A", 3: "A", 4: "A",
		5: "B", 6: "B",
		7: "C", 8: "C",
		9:  "D",
		10: "E", 11: "E",
		12: "X",
	}

	for total, want := range page {
		got, err := charts.StarportForThrow(total)
		if err != nil {
			t.Errorf("StarportForThrow(%d): %v", total, err)

			continue
		}

		if got != want {
			t.Errorf("a throw of %d gives starport %q, p. 1 prints %q", total, got, want)
		}
	}

	// A total two dice cannot show has no row, and says so.
	for _, total := range []int{1, 13, 0, -1} {
		if _, err := charts.StarportForThrow(total); err == nil {
			t.Errorf("StarportForThrow(%d) succeeded", total)
		}
	}
}

// TestBaseThrowsMatchTheStarportChart (p. 5). These are the throws the
// p. 12 checklist omits and docs/ERRATA.md E001 restores.
func TestBaseThrowsMatchTheStarportChart(t *testing.T) {
	charts := load(t)

	for _, c := range []struct {
		starport string
		naval    string
		scout    string
	}{
		{"A", "8+", "10+"},
		{"B", "8+", "9+"},
		{"C", "", "8+"},
		{"D", "", "7+"},
		{"E", "", ""},
		{"X", "", ""},
	} {
		checkBaseThrow(t, "naval", c.starport, c.naval, charts.NavalBaseTarget)
		checkBaseThrow(t, "scout", c.starport, c.scout, charts.ScoutBaseTarget)
	}
}

func checkBaseThrow(t *testing.T, base, starport, want string, get func(string) (dice.Target, bool)) {
	t.Helper()

	target, printed := get(starport)

	if want == "" {
		if printed {
			t.Errorf("starport %s has a %s base throw of %s; p. 5 prints none", starport, base, target)
		}

		return
	}

	if !printed {
		t.Errorf("starport %s has no %s base throw; p. 5 prints %s", starport, base, want)

		return
	}

	if got := target.String(); got != want {
		t.Errorf("starport %s %s base throw is %s, p. 5 prints %s", starport, base, got, want)
	}
}

// TestJumpRoutesTableMatchesThePage is the second transcription of the one
// table the held PDF's font damages worst: pdftotext renders every em-dash
// in it as the digit 4, which is also a real target in seven of its cells.
// Reading it wrong produces a subsector that is plausible and not the
// book's, so it is transcribed again here from the page.
func TestJumpRoutesTableMatchesThePage(t *testing.T) {
	charts := load(t)

	// 0 is the printed em-dash: no lane possible at that distance.
	page := map[string][4]int{
		"A-A": {1, 2, 4, 5},
		"A-B": {1, 3, 4, 5},
		"A-C": {1, 4, 6, 0},
		"A-D": {1, 5, 0, 0},
		"A-E": {2, 0, 0, 0},
		"B-B": {1, 3, 4, 6},
		"B-C": {2, 4, 6, 0},
		"B-D": {3, 6, 0, 0},
		"B-E": {4, 0, 0, 0},
		"C-C": {3, 6, 0, 0},
		"C-D": {4, 0, 0, 0},
		"C-E": {4, 0, 0, 0},
		"D-D": {4, 0, 0, 0},
		"D-E": {5, 0, 0, 0},
		"E-E": {6, 0, 0, 0},
	}

	for pair, row := range page {
		a, b := string(pair[0]), string(pair[2])

		for i, want := range row {
			distance := i + 1

			target, possible, err := charts.RouteTarget(a, b, distance)
			if err != nil {
				t.Errorf("RouteTarget(%s, %d): %v", pair, distance, err)

				continue
			}

			if want == 0 {
				if possible {
					t.Errorf("%s at jump-%d gives %s; p. 2 prints an em-dash", pair, distance, target)
				}

				continue
			}

			if !possible {
				t.Errorf("%s at jump-%d gives no lane; p. 2 prints %d+", pair, distance, want)

				continue
			}

			if target.Value != want || target.Mode != dice.Plus {
				t.Errorf("%s at jump-%d gives %s, p. 2 prints %d+", pair, distance, target, want)
			}
		}
	}
}

// TestRouteTableIsSymmetricAndBounded: p. 2 prints each unordered pair
// once, so a lookup must not care which world is named first, and no lane
// exists past the table's four columns.
func TestRouteTableIsSymmetricAndBounded(t *testing.T) {
	charts := load(t)

	for _, a := range tables.StarportTypes() {
		for _, b := range tables.StarportTypes() {
			checkRouteCell(t, charts, a, b)
		}
	}

	if _, _, err := charts.RouteTarget("A", "Q", 1); err == nil {
		t.Error("RouteTarget accepted a starport type the chart does not know")
	}
}

func checkRouteCell(t *testing.T, charts *tables.Charts, a, b string) {
	t.Helper()

	for distance := range 8 {
		forward, okF, errF := charts.RouteTarget(a, b, distance)
		back, okB, errB := charts.RouteTarget(b, a, distance)

		if errF != nil || errB != nil {
			t.Fatalf("RouteTarget(%s,%s,%d): %v / %v", a, b, distance, errF, errB)
		}

		// P. 2 prints each unordered pair once, so a lookup must not care
		// which world is named first.
		if okF != okB || forward != back {
			t.Errorf("RouteTarget is not symmetric for %s-%s at %d", a, b, distance)
		}

		if okF {
			checkLaneIsPossible(t, a, b, distance)
		}
	}
}

// checkLaneIsPossible asserts the two ways a cell must not yield a lane:
// past the p. 2 table's four distance columns, and for a pair the table
// prints no row for (docs/ERRATA.md E003).
func checkLaneIsPossible(t *testing.T, a, b string, distance int) {
	t.Helper()

	if distance < 1 || distance > tables.MaxJumpDistance {
		t.Errorf("%s-%s has a lane at distance %d, outside the p. 2 table", a, b, distance)
	}

	if a == tables.StarportX || b == tables.StarportX {
		t.Errorf("%s-%s has a lane, but p. 2 prints no X row", a, b)
	}
}

// TestTechnologicalIndexMatrixMatchesThePage: the third table the font
// damages — every em-dash reads as a 4, and +4 is a real cell in it.
func TestTechnologicalIndexMatrixMatchesThePage(t *testing.T) {
	charts := load(t)

	// Columns in the p. 9 order: starport, size, atm, hyd, pop, govt.
	page := map[byte][6]int{
		'0': {0, 2, 1, 0, 0, 1},
		'1': {0, 2, 1, 0, 1, 0},
		'2': {0, 1, 1, 0, 1, 0},
		'3': {0, 1, 1, 0, 1, 0},
		'4': {0, 1, 0, 0, 1, 0},
		'5': {0, 0, 0, 0, 1, 1},
		'6': {0, 0, 0, 0, 0, 0},
		'7': {0, 0, 0, 0, 0, 0},
		'8': {0, 0, 0, 0, 0, 0},
		'9': {0, 0, 0, 1, 2, 0},
		'A': {6, 0, 1, 2, 4, 0},
		'B': {4, 0, 1, 0, 0, 0},
		'C': {2, 0, 1, 0, 0, 0},
		'D': {0, 0, 1, 0, 0, -2},
		'E': {0, 0, 1, 0, 0, 0},
		'X': {-4, 0, 0, 0, 0, 0},
	}

	columns := tables.TechColumns()
	if len(columns) != 6 {
		t.Fatalf("TechColumns has %d entries, p. 9 prints 6", len(columns))
	}

	for value, row := range page {
		for i, column := range columns {
			got, err := charts.TechDM(column, value)
			if err != nil {
				t.Errorf("TechDM(%s, %q): %v", column, string(value), err)

				continue
			}

			if got != row[i] {
				t.Errorf("matrix cell [%q][%s] is %+d, p. 9 prints %+d", string(value), column, got, row[i])
			}
		}
	}

	if _, err := charts.TechDM("law_level", '5'); err == nil {
		t.Error("TechDM accepted law level, which p. 9 prints no column for")
	}

	if _, err := charts.TechDM(tables.Size, 'Z'); err == nil {
		t.Error("TechDM accepted a value the matrix has no row for")
	}
}

// TestEveryClampedValueHasAMatrixRow is the load-bearing link between
// docs/ERRATA.md E004 and the p. 9 matrix: the clamp exists partly so that
// every value the engine can present to the matrix has a row there.
func TestEveryClampedValueHasAMatrixRow(t *testing.T) {
	charts := load(t)

	clamped := []string{
		tables.Size, tables.Atmosphere, tables.Hydrographics,
		tables.Population, tables.Government,
	}

	for _, table := range clamped {
		maximum, ok := charts.MaxValue(table)
		if !ok {
			t.Fatalf("no %s table", table)
		}

		for v := 0; v <= maximum; v++ {
			d, ok := tables.Digit(v)
			if !ok {
				t.Fatalf("%s value %d has no digit", table, v)
			}

			if _, err := charts.TechDM(table, d); err != nil {
				t.Errorf("%s may be %d, but the p. 9 matrix has no row for %q", table, v, string(d))
			}
		}
	}

	for _, starport := range tables.StarportTypes() {
		if _, err := charts.TechDM(tables.StarportColumn, starport[0]); err != nil {
			t.Errorf("starport %s has no row in the p. 9 matrix", starport)
		}
	}
}

// TestMaxValuesAreThePrintedRanges pins the clamp ceilings E004 reads off
// the tables. They are asserted here rather than only derived, so a row
// accidentally added to or dropped from a data file moves a clamp and
// fails loudly.
func TestMaxValuesAreThePrintedRanges(t *testing.T) {
	charts := load(t)

	for table, want := range map[string]int{
		tables.Size:          10, // p. 5, 0 through A
		tables.Atmosphere:    12, // p. 5, 0 through C
		tables.Hydrographics: 10, // p. 6, 0 through A
		tables.Population:    10, // p. 6, 0 through A
		tables.Government:    13, // p. 6, 0 through D
		tables.LawLevel:      9,  // p. 7, 0 through 9 — floored only, never capped (E004)
	} {
		got, ok := charts.MaxValue(table)
		if !ok {
			t.Errorf("no %s table", table)

			continue
		}

		if got != want {
			t.Errorf("%s table ends at %d, the book prints %d", table, got, want)
		}
	}

	if _, ok := charts.MaxValue("gas_giants"); ok {
		t.Error("MaxValue answered for a table the book does not print")
	}
}

// TestLabelsCoverEveryValueAndStopThere.
func TestLabelsCoverEveryValueAndStopThere(t *testing.T) {
	charts := load(t)

	for _, table := range tables.CharacteristicTables() {
		maximum, ok := charts.MaxValue(table)
		if !ok {
			t.Fatalf("no %s table", table)
		}

		for v := 0; v <= maximum; v++ {
			if label, ok := charts.Label(table, v); !ok || label == "" {
				t.Errorf("%s %d has no label", table, v)
			}
		}

		for _, beyond := range []int{-1, maximum + 1} {
			if _, ok := charts.Label(table, beyond); ok {
				t.Errorf("%s answered a label for %d, past its printed rows", table, beyond)
			}
		}
	}

	if _, ok := charts.Label("gas_giants", 0); ok {
		t.Error("Label answered for a table the book does not print")
	}
}

// TestDigitNotation is docs/ERRATA.md E005's alphabet: hexadecimal to F,
// then p. 2's letters with I and O omitted, so 18 is J.
func TestDigitNotation(t *testing.T) {
	for value, want := range map[int]byte{
		0: '0', 9: '9', 10: 'A', 15: 'F', 16: 'G', 17: 'H', 18: 'J',
	} {
		got, ok := tables.Digit(value)
		if !ok || got != want {
			t.Errorf("Digit(%d) = %q, %v; want %q", value, string(got), ok, string(want))
		}

		back, ok := tables.Value(want)
		if !ok || back != value {
			t.Errorf("Value(%q) = %d, %v; want %d", string(want), back, ok, value)
		}
	}

	for _, absent := range []byte{'I', 'O', '-', ' '} {
		if _, ok := tables.Value(absent); ok {
			t.Errorf("Value(%q) succeeded; p. 2 omits it", string(absent))
		}
	}

	if _, ok := tables.Digit(-1); ok {
		t.Error("Digit(-1) succeeded")
	}

	if _, ok := tables.Digit(1000); ok {
		t.Error("Digit(1000) succeeded")
	}
}

// TestStarportFacilitiesMatchThePage is the second transcription of the
// p. 5 chart's facility columns, read off the page the way the base throws
// beside them are (docs/ERRATA.md, Noted discrepancies).
//
// It is the fuel grade render actually branches on, and the chart states
// each type's in a sentence rather than a cell: A and B have "Refined fuel
// available", C and D "Only unrefined fuel available", E is "a bare spot of
// bedrock with no fuel, facilities, or bases", and X makes "no provision
// ... for any starship landings". Overhaul is the "Annual maintenance
// overhaul available" sentence, which only A and B carry, and a shipyard is
// present only where the chart says one is — C's "reasonable repair
// facilities" are not one.
func TestStarportFacilitiesMatchThePage(t *testing.T) {
	charts := load(t)

	for _, page := range []struct {
		starport string
		fuel     string
		overhaul bool
		shipyard bool
	}{
		{tables.StarportA, "refined", true, true},
		{tables.StarportB, "refined", true, true},
		{tables.StarportC, "unrefined", false, false},
		{tables.StarportD, "unrefined", false, false},
		{tables.StarportE, "none", false, false},
		{tables.StarportX, "none", false, false},
	} {
		sp, err := charts.Starport(page.starport)
		if err != nil {
			t.Errorf("Starport(%q): %v", page.starport, err)

			continue
		}

		if sp.Fuel != page.fuel {
			t.Errorf("starport %s fuel is %q, p. 5 says %q", page.starport, sp.Fuel, page.fuel)
		}

		if sp.Overhaul != page.overhaul {
			t.Errorf("starport %s overhaul is %v, p. 5 says %v", page.starport, sp.Overhaul, page.overhaul)
		}

		// "none" is how the chart's no-shipyard rows are encoded; C's row
		// names its repair facilities in the same field and is still not a
		// shipyard, so the test asks whether the value starts with "none".
		if got := !strings.HasPrefix(sp.Shipyard, "none"); got != page.shipyard {
			t.Errorf("starport %s shipyard present = %v (%q), p. 5 says %v",
				page.starport, got, sp.Shipyard, page.shipyard)
		}
	}
}

// TestStarportChartIsComplete: every type has a chart row, and the
// accessors report the ones the book does not print.
func TestStarportChartIsComplete(t *testing.T) {
	charts := load(t)

	for _, tp := range tables.StarportTypes() {
		sp, err := charts.Starport(tp)
		if err != nil {
			t.Errorf("Starport(%q): %v", tp, err)

			continue
		}

		if sp.Type != tp || sp.Quality == "" {
			t.Errorf("starport %q chart row is %+v", tp, sp)
		}
	}

	if _, err := charts.Starport("Q"); err == nil {
		t.Error("Starport answered for a type the p. 5 chart does not print")
	}
}

// TestStarportChartRowsAreCopies is TestExportedSlicesAreCopies' other
// half: Starport hands back a pointer rather than a slice, so it is the one
// accessor a caller could write straight through into the loaded chart.
func TestStarportChartRowsAreCopies(t *testing.T) {
	charts := load(t)

	sp, err := charts.Starport(tables.StarportA)
	if err != nil {
		t.Fatalf("Starport: %v", err)
	}

	want := sp.Quality
	sp.Quality = "clobbered"

	again, err := charts.Starport(tables.StarportA)
	if err != nil {
		t.Fatalf("Starport: %v", err)
	}

	if again.Quality != want {
		t.Errorf("writing to a returned chart row changed the chart: %q", again.Quality)
	}
}

// TestExportedSlicesAreCopies: callers get their own, so one that appends
// or sorts cannot reach the package's tables.
func TestExportedSlicesAreCopies(t *testing.T) {
	for name, get := range map[string]func() []string{
		"StarportTypes":        tables.StarportTypes,
		"CharacteristicTables": tables.CharacteristicTables,
		"TechColumns":          tables.TechColumns,
	} {
		first := get()
		if len(first) == 0 {
			t.Fatalf("%s is empty", name)
		}

		first[0] = "clobbered"

		if get()[0] == "clobbered" {
			t.Errorf("%s hands out the package's own slice", name)
		}
	}
}

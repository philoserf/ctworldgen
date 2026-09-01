package tables_test

import (
	"testing"

	"github.com/philoserf/ctworldgen/starmap"
	"github.com/philoserf/ctworldgen/tables"
)

// The transcriptions below are the second reading of each page. They were
// typed from a visual read of the PDF, not copied from the data files,
// and the point is that the two must agree.
//
// The held PDFs' embedded font maps the em-dash to the glyph 4 and the
// minus sign to 3. A text extraction of the jump routes table therefore
// renders "4 - - -" as "4 4 4 4", and the size formula "2D - 2" as
// "2D32". Both are wrong and both look like data, which is why no table
// in this repository may be produced by extraction.

func load(t *testing.T) *tables.Tables {
	t.Helper()

	loaded, err := tables.Load()
	if err != nil {
		t.Fatalf("loading the embedded tables: %v", err)
	}

	return loaded
}

// TestStarportsTable is the starports table of Book 3 p. 1.
func TestStarportsTable(t *testing.T) {
	t.Parallel()

	want := map[int]string{
		2: "A", 3: "A", 4: "A",
		5: "B", 6: "B",
		7: "C", 8: "C",
		9:  "D",
		10: "E", 11: "E",
		12: "X",
	}

	starports := load(t).Starports
	for throw, wantType := range want {
		got, err := starports.Type(throw)
		if err != nil {
			t.Fatalf("throw %d: %v", throw, err)
		}

		if got.String() != wantType {
			t.Errorf("a throw of %d gives starport %s, want %s", throw, got, wantType)
		}
	}

	for _, outside := range []int{1, 0, 13, -1} {
		_, err := starports.Type(outside)
		if err == nil {
			t.Errorf("a throw of %d gave a starport; the table runs 2 to 12", outside)
		}
	}
}

// jumpRoutesTranscription is the second reading of the jump routes table
// of Book 3 p. 2. A nil cell is the em-dash the page prints: no number is
// stated there, so there is no target to throw against and no die is
// consumed (ERRATA E003).
func jumpRoutesTranscription() map[string][4]*int {
	ptr := func(v int) *int { return &v }

	return map[string][4]*int{
		"A-A": {ptr(1), ptr(2), ptr(4), ptr(5)},
		"A-B": {ptr(1), ptr(3), ptr(4), ptr(5)},
		"A-C": {ptr(1), ptr(4), ptr(6), nil},
		"A-D": {ptr(1), ptr(5), nil, nil},
		"A-E": {ptr(2), nil, nil, nil},
		"B-B": {ptr(1), ptr(3), ptr(4), ptr(6)},
		"B-C": {ptr(2), ptr(4), ptr(6), nil},
		"B-D": {ptr(3), ptr(6), nil, nil},
		"B-E": {ptr(4), nil, nil, nil},
		"C-C": {ptr(3), ptr(6), nil, nil},
		"C-D": {ptr(4), nil, nil, nil},
		"C-E": {ptr(4), nil, nil, nil},
		"D-D": {ptr(4), nil, nil, nil},
		"D-E": {ptr(5), nil, nil, nil},
		"E-E": {ptr(6), nil, nil, nil},
	}
}

// pairStarports splits a transcription key such as "A-C".
func pairStarports(t *testing.T, pair string) (starmap.Starport, starmap.Starport) {
	t.Helper()

	first, err := starmap.ParseStarport(pair[0:1])
	if err != nil {
		t.Fatal(err)
	}

	b, err := starmap.ParseStarport(pair[2:3])
	if err != nil {
		t.Fatal(err)
	}

	return first, b
}

// TestJumpRoutesTable holds the embedded table to the second transcription
// of the page.
func TestJumpRoutesTable(t *testing.T) {
	t.Parallel()

	want := jumpRoutesTranscription()
	if len(want) != 15 {
		t.Fatalf("the transcription has %d rows, and the page prints 15", len(want))
	}

	routes := load(t).JumpRoutes
	dashes := 0

	for pair, cells := range want {
		a, b := pairStarports(t, pair)

		for i, cell := range cells {
			distance := starmap.Parsecs(i + 1)

			got, stated := routes.Target(a, b, distance)

			if cell == nil {
				dashes++

				if stated {
					t.Errorf("%s at jump-%d states a target of %d; the page prints an em-dash", pair, distance, got)
				}

				continue
			}

			if !stated {
				t.Errorf("%s at jump-%d states no target; the page prints %d", pair, distance, *cell)

				continue
			}

			if int(got) != *cell {
				t.Errorf("%s at jump-%d = %d, want %d", pair, distance, got, *cell)
			}
		}
	}

	if dashes != 29 {
		t.Errorf("the transcription has %d em-dash cells; the page prints 29 of 60", dashes)
	}
}

// TestJumpRoutesTableIsSymmetric: a pair is a pair either way round.
func TestJumpRoutesTableIsSymmetric(t *testing.T) {
	t.Parallel()

	routes := load(t).JumpRoutes

	for pair := range jumpRoutesTranscription() {
		a, b := pairStarports(t, pair)

		forward, statedForward := routes.Target(a, b, 1)

		backward, statedBackward := routes.Target(b, a, 1)
		if statedForward != statedBackward || forward != backward {
			t.Errorf("%s reads %d (%v) forwards and %d (%v) backwards",
				pair, forward, statedForward, backward, statedBackward)
		}
	}
}

// TestJumpRoutesHasNoRowForX is the second part of ERRATA E003: the table
// prints rows for A-A through E-E and none for X, and p. 5 gives starport
// X "no provision ... for any starship landings". A pair with an X at
// either end is not examined and consumes no die.
func TestJumpRoutesHasNoRowForX(t *testing.T) {
	t.Parallel()

	routes := load(t).JumpRoutes

	for _, port := range starmap.Starports() {
		for distance := starmap.Parsecs(1); distance <= tables.MaxJump; distance++ {
			if _, stated := routes.Target(starmap.StarportX, port, distance); stated {
				t.Errorf("X-%s at jump-%d states a target; the table prints no row for X", port, distance)
			}

			if _, stated := routes.Target(port, starmap.StarportX, distance); stated {
				t.Errorf("%s-X at jump-%d states a target; the table prints no row for X", port, distance)
			}
		}
	}

	// Nothing is stated beyond the four columns the page prints.
	for _, distance := range []starmap.Parsecs{0, 5, 6, -1} {
		if _, stated := routes.Target(starmap.StarportA, starmap.StarportA, distance); stated {
			t.Errorf("A-A at jump-%d states a target; the page prints four distance columns", distance)
		}
	}
}

// TestStarportChartBaseThrows is the p. 5 starport chart's base throws,
// which the p. 12 checklist omits (ERRATA E001).
func TestStarportChartBaseThrows(t *testing.T) {
	t.Parallel()

	ptr := func(v int) *int { return &v }
	want := []struct {
		typ          string
		naval, scout *int
	}{
		{"A", ptr(8), ptr(10)},
		{"B", ptr(8), ptr(9)},
		{"C", nil, ptr(8)},
		{"D", nil, ptr(7)},
		{"E", nil, nil},
		{"X", nil, nil},
	}

	chart := load(t).StarportChart

	for _, wantRow := range want {
		port, err := starmap.ParseStarport(wantRow.typ)
		if err != nil {
			t.Fatal(err)
		}

		naval, hasNaval := chart.NavalBase(port)
		if (wantRow.naval == nil) != !hasNaval {
			t.Errorf("starport %s: naval base throw printed = %v, want %v", wantRow.typ, hasNaval, wantRow.naval != nil)
		} else if wantRow.naval != nil && int(naval) != *wantRow.naval {
			t.Errorf("starport %s: naval base on %d+, want %d+", wantRow.typ, naval, *wantRow.naval)
		}

		scout, hasScout := chart.ScoutBase(port)
		if (wantRow.scout == nil) != !hasScout {
			t.Errorf("starport %s: scout base throw printed = %v, want %v", wantRow.typ, hasScout, wantRow.scout != nil)
		} else if wantRow.scout != nil && int(scout) != *wantRow.scout {
			t.Errorf("starport %s: scout base on %d+, want %d+", wantRow.typ, scout, *wantRow.scout)
		}
	}
}

// TestTechnologicalIndexMatrix is the matrix of Book 3 p. 9. A nil cell is
// the em-dash the page prints: no modifier.
func TestTechnologicalIndexMatrix(t *testing.T) {
	t.Parallel()

	ptr := func(v int) *int { return &v }
	// Value, then the Size, Atm, Hyd, Pop and Govt columns. The Starport
	// column is checked separately below: it is indexed by starport type,
	// not by a number.
	want := map[string][5]*int{
		"0": {ptr(2), ptr(1), nil, nil, ptr(1)},
		"1": {ptr(2), ptr(1), nil, ptr(1), nil},
		"2": {ptr(1), ptr(1), nil, ptr(1), nil},
		"3": {ptr(1), ptr(1), nil, ptr(1), nil},
		"4": {ptr(1), nil, nil, ptr(1), nil},
		"5": {nil, nil, nil, ptr(1), ptr(1)},
		"6": {nil, nil, nil, nil, nil},
		"7": {nil, nil, nil, nil, nil},
		"8": {nil, nil, nil, nil, nil},
		"9": {nil, nil, ptr(1), ptr(2), nil},
		"A": {nil, ptr(1), ptr(2), ptr(4), nil},
		"B": {nil, ptr(1), nil, nil, nil},
		"C": {nil, ptr(1), nil, nil, nil},
		"D": {nil, ptr(1), nil, nil, ptr(-2)},
		"E": {nil, ptr(1), nil, nil, nil},
	}
	columns := [5]tables.Column{
		tables.ColSize, tables.ColAtmosphere, tables.ColHydrographics,
		tables.ColPopulation, tables.ColGovernment,
	}

	matrix := load(t).TechIndexMatrix

	for value, cells := range want {
		digit, err := starmap.ParseDigit(value)
		if err != nil {
			t.Fatal(err)
		}

		for index, cell := range cells {
			got := matrix.DM(columns[index], digit.Value())

			wantDM := 0
			if cell != nil {
				wantDM = *cell
			}

			if got != wantDM {
				t.Errorf("matrix value %s, column %d: DM %+d, want %+d", value, index, got, wantDM)
			}
		}
	}

	starportDMs := map[string]int{"A": 6, "B": 4, "C": 2, "D": 0, "E": 0, "X": -4}
	for typ, wantDM := range starportDMs {
		p, err := starmap.ParseStarport(typ)
		if err != nil {
			t.Fatal(err)
		}

		if got := matrix.StarportDM(p); got != wantDM {
			t.Errorf("starport %s contributes %+d, want %+d", typ, got, wantDM)
		}
	}
}

// TestMatrixMaximumBinds is the arithmetic of ERRATA E004: the DMs reach
// +14, which with a die of 6 gives 20 against a printed cap of 18. The cap
// is rare but reachable, so it binds.
func TestMatrixMaximumBinds(t *testing.T) {
	t.Parallel()

	matrix := load(t).TechIndexMatrix

	sum := matrix.StarportDM(starmap.StarportA) +
		matrix.DM(tables.ColSize, 0) +
		matrix.DM(tables.ColAtmosphere, 0) +
		matrix.DM(tables.ColHydrographics, 0) +
		matrix.DM(tables.ColPopulation, 10) +
		matrix.DM(tables.ColGovernment, 5)
	if sum != 14 {
		t.Errorf("the greatest DM total is %+d, want +14 (starport A, size 0, population A, government 5)", sum)
	}

	if sum+6 <= 18 {
		t.Error("the technological index cap of 18 does not bind; ERRATA E004 says it does")
	}
}

// TestValueWithNoMatrixRowContributesNothing is the last part of ERRATA
// E004. The Value column runs 0 through 9, A through E and X, so 15 is the
// only generated value with no row at all, and an absent row contributes
// nothing -- which is what the printed dashes already mean.
func TestValueWithNoMatrixRowContributesNothing(t *testing.T) {
	t.Parallel()

	matrix := load(t).TechIndexMatrix

	all := []tables.Column{
		tables.ColSize, tables.ColAtmosphere, tables.ColHydrographics,
		tables.ColPopulation, tables.ColGovernment,
	}
	for _, col := range all {
		if got := matrix.DM(col, 15); got != 0 {
			t.Errorf("a value of 15 contributes %+d in column %d, want 0", got, col)
		}
	}
}

// TestEveryPrintedValueHasALabel is R16. Only the range each table prints
// is covered: a generated value beyond it has no label, which is a gap in
// the page and not an error to correct.
func TestEveryPrintedValueHasALabel(t *testing.T) {
	t.Parallel()

	loaded := load(t)
	for _, table := range []struct {
		name       string
		labels     tables.Labels
		printedMax int
	}{
		{"planetary size (p. 5)", loaded.Size, 12},
		{"planetary atmosphere (p. 5)", loaded.Atmosphere, 12},
		{"hydrographic percentage (p. 6)", loaded.Hydrographics, 10},
		{"population (p. 6)", loaded.Population, 10},
		{"governmental type (p. 6)", loaded.Government, 13},
		{"law levels (p. 7)", loaded.LawLevels, 9},
	} {
		if table.labels.PrintedMax() != table.printedMax {
			t.Errorf("%s describes 0 to %d, want 0 to %d", table.name, table.labels.PrintedMax(), table.printedMax)
		}

		for value := 0; value <= table.printedMax; value++ {
			label, ok := table.labels.Label(value)
			if !ok || label == "" {
				t.Errorf("%s has no label for the value %d", table.name, value)
			}
		}

		if _, ok := table.labels.Label(table.printedMax + 1); ok {
			t.Errorf("%s labels a value beyond its printed range", table.name)
		}

		if _, ok := table.labels.Label(-1); ok {
			t.Errorf("%s labels a negative value", table.name)
		}
	}
}

// TestStarportChartDescriptions checks the p. 5 chart carries a
// description for every type. The descriptions themselves are editorial
// transcriptions of the book's prose, so retyping them here would compare
// a transcription against itself; they are checked by re-reading the page.
func TestStarportChartDescriptions(t *testing.T) {
	t.Parallel()

	chart := load(t).StarportChart
	for _, port := range starmap.Starports() {
		row, err := chart.Row(port)
		if err != nil {
			t.Fatalf("starport %s: %v", port, err)
		}

		if row.Description == "" {
			t.Errorf("starport %s has no description", port)
		}
	}

	_, err := chart.Row(starmap.Starport('Q'))
	if err == nil {
		t.Error("the chart returned a row for a starport the book does not print")
	}

	if _, ok := chart.NavalBase(starmap.Starport('Q')); ok {
		t.Error("the chart returned a naval throw for a starport the book does not print")
	}

	if _, ok := chart.ScoutBase(starmap.Starport('Q')); ok {
		t.Error("the chart returned a scout throw for a starport the book does not print")
	}
}

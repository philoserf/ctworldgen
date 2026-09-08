package starmap_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/starmap"
)

func hex(t *testing.T, col, row int) starmap.Hex {
	t.Helper()

	h, err := starmap.NewHex(col, row)
	if err != nil {
		t.Fatal(err)
	}

	return h
}

// TestParseArea reads the form the referee types. The DM is cut off at the
// @ before the corners are split, because a negative DM opens with the
// same hyphen the corners are separated by.
func TestParseArea(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		text string
		from starmap.Hex
		to   starmap.Hex
		dm   int
	}{
		{"-1@0101-0410", starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 4, Row: 10}, -1},
		{"+1@0101-0410", starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 4, Row: 10}, 1},
		{"1@0101-0410", starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 4, Row: 10}, 1},
		{"0@0101-0410", starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 4, Row: 10}, 0},
		// Two opposite corners describe one rectangle whichever two they
		// are: a referee points at the area, not at its low corner.
		{"-1@0410-0101", starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 4, Row: 10}, -1},
		{"-1@0110-0401", starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 4, Row: 10}, -1},
		// One hex is an area, which p. 1 does not forbid and a referee
		// placing a single dead system wants.
		{"+1@0303-0303", starmap.Hex{Col: 3, Row: 3}, starmap.Hex{Col: 3, Row: 3}, 1},
	} {
		t.Run(testCase.text, func(t *testing.T) {
			t.Parallel()

			area, err := starmap.ParseArea(testCase.text)
			if err != nil {
				t.Fatalf("%q: %v", testCase.text, err)
			}

			want := starmap.Area{From: testCase.from, To: testCase.to, DM: testCase.dm}
			if area != want {
				t.Errorf("%q read as %+v; want %+v", testCase.text, area, want)
			}
		})
	}
}

// TestParseAreaRefusesWhatIsNotOne. An area that half-parses is worse than
// one that does not parse at all: the referee typed a geography and would
// be handed a different one without being told.
func TestParseAreaRefusesWhatIsNotOne(t *testing.T) {
	t.Parallel()

	for _, text := range []string{
		"",
		"0101-0410",       // no DM
		"-1",              // no rectangle
		"-1@0101",         // one corner
		"-1@0101-0410-08", // three
		"-1@0101 0410",    // the wrong separator
		"x@0101-0410",     // not a number
		"-1@0000-0410",    // not a hex
		"-1@abcd-0410",
		"-1@0101-3341", // off every grid there is
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()

			_, err := starmap.ParseArea(text)
			if err == nil {
				t.Errorf("%q was read as an area", text)
			}
		})
	}
}

// TestAnAreaContainsItsRectangle walks the p. 3 grid rather than sampling
// it, so an off-by-one at any of the four edges is named.
func TestAnAreaContainsItsRectangle(t *testing.T) {
	t.Parallel()

	area := starmap.NewArea(hex(t, 2, 3), hex(t, 5, 7), -1)

	for col := 1; col <= starmap.Columns; col++ {
		for row := 1; row <= starmap.Rows; row++ {
			want := col >= 2 && col <= 5 && row >= 3 && row <= 7
			if got := area.Contains(hex(t, col, row)); got != want {
				t.Errorf("%s: Contains is %v for the area %s; want %v", hex(t, col, row), got, area, want)
			}
		}
	}
}

// TestHoldAreasRefusesWhatErrataTwelveRefuses holds the reading against
// the grid, and names which of its parts each case offends.
func TestHoldAreasRefusesWhatErrataTwelveRefuses(t *testing.T) {
	t.Parallel()

	grid := starmap.PageThreeGrid()

	for name, testCase := range map[string]struct {
		areas []starmap.Area
		want  error
	}{
		"a DM the page does not offer": {
			[]starmap.Area{starmap.NewArea(starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 4, Row: 10}, 2)},
			starmap.ErrOccurrenceDM,
		},
		"corners the wrong way round": {
			[]starmap.Area{{From: starmap.Hex{Col: 4, Row: 10}, To: starmap.Hex{Col: 1, Row: 1}, DM: -1}},
			starmap.ErrAreaCorners,
		},
		"a corner off the grid": {
			[]starmap.Area{starmap.NewArea(starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 9, Row: 10}, -1)},
			starmap.ErrOffGrid,
		},
		"two areas sharing a hex": {
			[]starmap.Area{
				starmap.NewArea(starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 4, Row: 10}, -1),
				starmap.NewArea(starmap.Hex{Col: 4, Row: 1}, starmap.Hex{Col: 8, Row: 10}, 1),
			},
			starmap.ErrAreasOverlap,
		},
		"two areas sharing one row": {
			[]starmap.Area{
				starmap.NewArea(starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 8, Row: 5}, -1),
				starmap.NewArea(starmap.Hex{Col: 1, Row: 5}, starmap.Hex{Col: 8, Row: 10}, 1),
			},
			starmap.ErrAreasOverlap,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := grid.HoldAreas(testCase.areas)
			if !errors.Is(err, testCase.want) {
				t.Errorf("%s gave %v; want %v", name, err, testCase.want)
			}
		})
	}
}

// TestHoldAreasAcceptsWhatErrataTwelveAllows guards the test above: a rule
// that refused everything would pass every case in it.
func TestHoldAreasAcceptsWhatErrataTwelveAllows(t *testing.T) {
	t.Parallel()

	grid := starmap.PageThreeGrid()

	for name, areas := range map[string][]starmap.Area{
		"no areas at all": nil,
		"the whole grid": {
			starmap.NewArea(starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 8, Row: 10}, -1),
		},
		"two bands that touch and share no hex": {
			starmap.NewArea(starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 8, Row: 5}, -1),
			starmap.NewArea(starmap.Hex{Col: 1, Row: 6}, starmap.Hex{Col: 8, Row: 10}, 1),
		},
		"a corner and the opposite corner": {
			starmap.NewArea(starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 2, Row: 2}, -1),
			starmap.NewArea(starmap.Hex{Col: 7, Row: 9}, starmap.Hex{Col: 8, Row: 10}, 1),
		},
		"one hex": {
			starmap.NewArea(starmap.Hex{Col: 3, Row: 3}, starmap.Hex{Col: 3, Row: 3}, 1),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := grid.HoldAreas(areas)
			if err != nil {
				t.Errorf("%s was refused: %v", name, err)
			}
		})
	}
}

// TestASectorGridHoldsSectorAreas: HoldAreas is a method on the grid
// because which hexes exist is the grid's question. Nothing generates a
// sector with areas today, and the check must still be the grid's rather
// than the p. 3 grid's written into it.
func TestASectorGridHoldsSectorAreas(t *testing.T) {
	t.Parallel()

	area := []starmap.Area{starmap.NewArea(starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 16, Row: 20}, -1)}

	err := starmap.SectorGrid().HoldAreas(area)
	if err != nil {
		t.Errorf("the sector grid refused an area inside it: %v", err)
	}

	err = starmap.PageThreeGrid().HoldAreas(area)
	if err == nil {
		t.Error("the p. 3 grid accepted an area reaching to 1620")
	}
}

// TestARecordWithNoAreasIsUnchanged holds the omitempty guarantee that
// made this an additive change, as TestARecordWithNoNotesIsUnchanged does
// for the referee's notes: a record with no broad areas must serialise
// exactly as it did before the field existed, or every golden moves and
// the dice stream is indistinguishable from having moved with them.
func TestARecordWithNoAreasIsUnchanged(t *testing.T) {
	t.Parallel()

	for name, areas := range map[string][]starmap.Area{
		"none at all":   nil,
		"an empty list": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			encoded, err := starmap.Marshal(starmap.New(1977, "Aramis", -1, areas))
			if err != nil {
				t.Fatal(err)
			}

			if strings.Contains(string(encoded), "occurrence_areas") {
				t.Errorf("a record with %s wrote the key anyway:\n%s", name, encoded)
			}
		})
	}
}

// TestARecordWithAreasCarriesThem is the other half: the field is written
// where there is something to write, in the order the referee gave it.
func TestARecordWithAreasCarriesThem(t *testing.T) {
	t.Parallel()

	record := starmap.New(1, "Aramis", 0, []starmap.Area{
		starmap.NewArea(hex(t, 1, 6), hex(t, 8, 10), 1),
		starmap.NewArea(hex(t, 1, 1), hex(t, 8, 5), -1),
	})

	encoded, err := starmap.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}

	read, err := starmap.Decode(strings.NewReader(string(encoded)))
	if err != nil {
		t.Fatalf("a record carrying broad areas did not read back: %v", err)
	}

	if len(read.OccurrenceAreas) != len(record.OccurrenceAreas) {
		t.Fatalf("wrote %d areas and read back %d", len(record.OccurrenceAreas), len(read.OccurrenceAreas))
	}

	// Nothing sorts them, so the second area written is the second read.
	for index, area := range record.OccurrenceAreas {
		if read.OccurrenceAreas[index] != area {
			t.Errorf("area %d was written %+v and read back %+v", index, area, read.OccurrenceAreas[index])
		}
	}
}

// TestDecodeRefusesAreasTheReadingRefuses is the read path's half of the
// two obligations. A schema alone rejects nothing at read time, and two of
// these -- the corners and the overlap -- are things JSON Schema cannot
// state at all.
func TestDecodeRefusesAreasTheReadingRefuses(t *testing.T) {
	t.Parallel()

	for name, testCase := range map[string]struct {
		clause string
		want   error
	}{
		"a DM the page does not offer": {
			`"occurrence_areas":[{"from":"0101","to":"0410","dm":2}],`, starmap.ErrOccurrenceDM,
		},
		"corners the wrong way round": {
			`"occurrence_areas":[{"from":"0410","to":"0101","dm":-1}],`, starmap.ErrAreaCorners,
		},
		"a corner off the record's grid": {
			`"occurrence_areas":[{"from":"0101","to":"0910","dm":-1}],`, starmap.ErrOffGrid,
		},
		"two areas sharing a hex": {
			`"occurrence_areas":[{"from":"0101","to":"0410","dm":-1},` +
				`{"from":"0401","to":"0810","dm":1}],`, starmap.ErrAreasOverlap,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := starmap.Decode(strings.NewReader(recordWithAreas(testCase.clause)))
			if !errors.Is(err, testCase.want) {
				t.Errorf("%s gave %v; want %v", name, err, testCase.want)
			}
		})
	}
}

// TestDecodeAcceptsTheAreasTheEngineWrites guards the test above.
func TestDecodeAcceptsTheAreasTheEngineWrites(t *testing.T) {
	t.Parallel()

	clause := `"occurrence_areas":[{"from":"0101","to":"0805","dm":-1},{"from":"0106","to":"0810","dm":1}],`

	record, err := starmap.Decode(strings.NewReader(recordWithAreas(clause)))
	if err != nil {
		t.Fatalf("a record carrying the banded areas was refused: %v", err)
	}

	if len(record.OccurrenceAreas) != 2 {
		t.Errorf("read %d areas; want 2", len(record.OccurrenceAreas))
	}
}

// recordWithAreas builds a world-less record carrying whatever
// occurrence_areas clause is given, so a test can hand Decode a set of
// areas the engine would never have written.
func recordWithAreas(clause string) string {
	return `{"schema_version":1,"ruleset":"ct-1977-book3-pp1-12","engine_version":"1",` +
		`"rng_algorithm":"go-math-rand-v2-pcg","seed":1,"errata":[],"name":"Aramis","occurrence_dm":0,` +
		clause + `"grid":{"columns":8,"rows":10},"worlds":[],"routes":[]}`
}

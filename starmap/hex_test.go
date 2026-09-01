package starmap_test

import (
	"encoding/json"
	"testing"

	"github.com/philoserf/ctworldgen/starmap"
)

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
		{"0105", "0204", 1, "0204 sits between rows 4 and 5 of column 2; a flipped parity gives 2"},
		{"0201", "0301", 1, "even column to odd column"},
		{"0110", "0210", 1, "the same step at the bottom of the grid"},
		{"0101", "0301", 2, "two columns across, stepping through 0201"},
		{"0101", "0105", 4, "four rows down the same column"},
		{"0101", "0501", 4, "four columns across along the zig-zag"},
		{"0101", "0110", 9, "the full height of a column"},
		{"0101", "0810", 13, "corner to corner: seven column steps, six of them also downward, then six more down"},
	}

	for _, testCase := range cases {
		t.Run(testCase.a+"-"+testCase.b, func(t *testing.T) {
			t.Parallel()

			first, err := starmap.ParseHex(testCase.a)
			if err != nil {
				t.Fatal(err)
			}

			second, err := starmap.ParseHex(testCase.b)
			if err != nil {
				t.Fatal(err)
			}

			if got := first.Distance(second); got != testCase.want {
				t.Errorf("%s to %s = %d parsecs, want %d (%s)", testCase.a, testCase.b, got, testCase.want, testCase.note)
			}

			if got := second.Distance(first); got != testCase.want {
				t.Errorf("%s to %s = %d parsecs, want %d: distance is symmetric", testCase.b, testCase.a, got, testCase.want)
			}
		})
	}
}

// TestNewHexRejectsOffGrid bounds the identifier, which is the largest
// grid there is: a hex is four digits whether it names a subsector or a
// sector, and 0000 or 3341 is not one. Whether a hex is on a particular
// record's grid is Grid.Contains, checked below.
func TestNewHexRejectsOffGrid(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ col, row int }{{0, 1}, {1, 0}, {33, 1}, {1, 41}, {-1, -1}} {
		_, err := starmap.NewHex(c.col, c.row)
		if err == nil {
			t.Errorf("NewHex(%d, %d) succeeded; the sector grid is %d columns of %d rows",
				c.col, c.row, starmap.SectorColumns, starmap.SectorRows)
		}
	}
}

// TestAGridHoldsOnlyItsOwnHexes: the p. 3 grid stops at 0810 even though
// the identifier runs to 3240, so a subsector record carrying 0910 is a
// record on the wrong grid and Decode says so.
func TestAGridHoldsOnlyItsOwnHexes(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		col, row int
		onPage3  bool
	}{{1, 1, true}, {8, 10, true}, {9, 1, false}, {1, 11, false}, {32, 40, false}} {
		hex, err := starmap.NewHex(testCase.col, testCase.row)
		if err != nil {
			t.Fatal(err)
		}

		if got := starmap.PageThreeGrid().Contains(hex); got != testCase.onPage3 {
			t.Errorf("the p. 3 grid contains %s: %v, want %v", hex, got, testCase.onPage3)
		}

		if !starmap.SectorGrid().Contains(hex) {
			t.Errorf("the sector grid does not contain %s, and every hex is on it", hex)
		}
	}
}

func TestParseHex(t *testing.T) {
	t.Parallel()

	hex, err := starmap.ParseHex("0810")
	if err != nil {
		t.Fatal(err)
	}

	if hex.Col != 8 || hex.Row != 10 {
		t.Errorf("ParseHex(0810) = %d,%d, want 8,10", hex.Col, hex.Row)
	}

	// 3240 is the last hex of the sector grid, and 0111 -- off the p. 3
	// grid but on that one -- is a hex the identifier accepts and a
	// subsector record does not carry (TestAGridHoldsOnlyItsOwnHexes).
	corner, err := starmap.ParseHex("3240")
	if err != nil {
		t.Fatal(err)
	}

	if corner.Col != starmap.SectorColumns || corner.Row != starmap.SectorRows {
		t.Errorf("ParseHex(3240) = %d,%d, want %d,%d",
			corner.Col, corner.Row, starmap.SectorColumns, starmap.SectorRows)
	}

	for _, bad := range []string{"", "101", "01011", "0900", "3341", "0141", "abcd", "01 1", "0000"} {
		_, err := starmap.ParseHex(bad)
		if err == nil {
			t.Errorf("ParseHex(%q) succeeded", bad)
		}
	}
}

func TestHexRoundTripsThroughJSON(t *testing.T) {
	t.Parallel()

	for col := 1; col <= starmap.Columns; col++ {
		for row := 1; row <= starmap.Rows; row++ {
			hex, err := starmap.NewHex(col, row)
			if err != nil {
				t.Fatal(err)
			}

			encoded, err := json.Marshal(hex)
			if err != nil {
				t.Fatal(err)
			}

			var back starmap.Hex

			err = json.Unmarshal(encoded, &back)
			if err != nil {
				t.Fatalf("%s: %v", hex, err)
			}

			if back != hex {
				t.Errorf("%s round-tripped to %s", hex, back)
			}
		}
	}
}

func TestZeroHexIsNotAHex(t *testing.T) {
	t.Parallel()

	_, err := json.Marshal(starmap.Hex{Col: 0, Row: 0})
	if err == nil {
		t.Error("marshaling the zero Hex succeeded; it is outside the grid")
	}

	var hex starmap.Hex

	err = json.Unmarshal([]byte(`"0900"`), &hex)
	if err == nil {
		t.Error("unmarshaling an off-grid hex succeeded")
	}

	err = json.Unmarshal([]byte(`7`), &hex)
	if err == nil {
		t.Error("unmarshaling a number as a hex succeeded")
	}
}

// TestNumberOrdersHexesTheWayTheGridNumbersThem pins the scan order of
// ERRATA E002: column by column, row within column.
func TestNumberOrdersHexesTheWayTheGridNumbersThem(t *testing.T) {
	t.Parallel()

	ordered := []string{"0101", "0102", "0110", "0201", "0210", "0801", "0810"}
	for index := 1; index < len(ordered); index++ {
		earlier, err := starmap.ParseHex(ordered[index-1])
		if err != nil {
			t.Fatal(err)
		}

		later, err := starmap.ParseHex(ordered[index])
		if err != nil {
			t.Fatal(err)
		}

		if !earlier.Less(later) {
			t.Errorf("%s should sort before %s", earlier, later)
		}

		if later.Less(earlier) {
			t.Errorf("%s should not sort before %s", later, earlier)
		}
	}
}

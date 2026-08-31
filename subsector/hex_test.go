package subsector_test

import (
	"encoding/json"
	"testing"

	"github.com/philoserf/ctworldgen/subsector"
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
		want subsector.Parsecs
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

			first, err := subsector.ParseHex(testCase.a)
			if err != nil {
				t.Fatal(err)
			}

			second, err := subsector.ParseHex(testCase.b)
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

func TestNewHexRejectsOffGrid(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ col, row int }{{0, 1}, {1, 0}, {9, 1}, {1, 11}, {-1, -1}} {
		_, err := subsector.NewHex(c.col, c.row)
		if err == nil {
			t.Errorf("NewHex(%d, %d) succeeded; the p. 3 grid is 8 columns of 10 rows", c.col, c.row)
		}
	}
}

func TestParseHex(t *testing.T) {
	t.Parallel()

	hex, err := subsector.ParseHex("0810")
	if err != nil {
		t.Fatal(err)
	}

	if hex.Col != 8 || hex.Row != 10 {
		t.Errorf("ParseHex(0810) = %d,%d, want 8,10", hex.Col, hex.Row)
	}

	for _, bad := range []string{"", "101", "01011", "0900", "0111", "abcd", "01 1", "0000"} {
		_, err := subsector.ParseHex(bad)
		if err == nil {
			t.Errorf("ParseHex(%q) succeeded", bad)
		}
	}
}

func TestHexRoundTripsThroughJSON(t *testing.T) {
	t.Parallel()

	for col := 1; col <= subsector.Columns; col++ {
		for row := 1; row <= subsector.Rows; row++ {
			hex, err := subsector.NewHex(col, row)
			if err != nil {
				t.Fatal(err)
			}

			encoded, err := json.Marshal(hex)
			if err != nil {
				t.Fatal(err)
			}

			var back subsector.Hex

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

	_, err := json.Marshal(subsector.Hex{Col: 0, Row: 0})
	if err == nil {
		t.Error("marshaling the zero Hex succeeded; it is outside the grid")
	}

	var hex subsector.Hex

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
		earlier, err := subsector.ParseHex(ordered[index-1])
		if err != nil {
			t.Fatal(err)
		}

		later, err := subsector.ParseHex(ordered[index])
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

package worldgen

import "testing"

// TestDistanceMatchesThePrintedGrid measures against Book 3 p. 3 rather
// than against the implementation. Every pair below was traced hex by hex
// on the printed grid; a parity error in the offset-to-cube conversion
// leaves distances internally consistent and wrong for half the map, so
// this is the only test that can catch one.
func TestDistanceMatchesThePrintedGrid(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		// A hex is zero from itself.
		{"0101", "0101", 0},
		// The six neighbours of 0202, an interior hex of an even (lowered)
		// column: two in its own column, two in column 1, two in column 3.
		{"0202", "0201", 1},
		{"0202", "0203", 1},
		{"0202", "0102", 1},
		{"0202", "0103", 1},
		{"0202", "0302", 1},
		{"0202", "0303", 1},
		// The six neighbours of 0303, an interior hex of an odd (raised)
		// column. The column-2 and column-4 neighbours sit one row up from
		// the even-column case above — that asymmetry is the parity.
		{"0303", "0302", 1},
		{"0303", "0304", 1},
		{"0303", "0202", 1},
		{"0303", "0203", 1},
		{"0303", "0402", 1},
		{"0303", "0403", 1},
		// Straight down a column.
		{"0101", "0102", 1},
		{"0101", "0110", 9},
		// Along the top edge: every column step is one jump, because the
		// lowered even columns zig-zag between the raised odd ones.
		{"0101", "0201", 1},
		{"0101", "0301", 2},
		{"0101", "0801", 7},
		{"0110", "0810", 7},
		// The two long diagonals of the whole grid.
		{"0101", "0210", 10},
		{"0101", "0810", 13},
		// Distances at the jump-routes table's reach (p. 2, four columns).
		{"0101", "0105", 4},
		// 0101 to 0405 is six, not the four a glance at the row and column
		// numbers suggests: three column steps down-right reach 0402, and
		// the remaining three rows cost one jump each.
		{"0101", "0405", 6},
		{"0101", "0403", 4},
		{"0404", "0805", 4},
	}

	for _, c := range cases {
		a, err := ParseHex(c.a)
		if err != nil {
			t.Fatalf("ParseHex(%q): %v", c.a, err)
		}

		b, err := ParseHex(c.b)
		if err != nil {
			t.Fatalf("ParseHex(%q): %v", c.b, err)
		}

		if got := a.Distance(b); got != c.want {
			t.Errorf("Distance(%s, %s) = %d, want %d (measured on the p. 3 grid)", c.a, c.b, got, c.want)
		}

		// Distance is symmetric; p. 2 examines each pair once and must not
		// care which world it names first.
		if got := b.Distance(a); got != c.want {
			t.Errorf("Distance(%s, %s) = %d, want %d (symmetry)", c.b, c.a, got, c.want)
		}
	}
}

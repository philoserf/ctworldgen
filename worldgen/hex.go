package worldgen

import (
	"errors"
	"fmt"
	"strconv"
)

// The subsector grid printed on Book 3 p. 3: eight columns of ten rows,
// eighty hexes, each one parsec across (p. 1).
const (
	Columns = 8
	Rows    = 10
	Hexes   = Columns * Rows
)

// ErrBadHex reports a hex identifier that is not a hex of the p. 3 grid.
var ErrBadHex = errors.New("bad hex")

// Hex is one hex of the p. 3 grid, addressed the way the grid prints it:
// a column 1-8 and a row 1-10, written as the four digits 0101 through
// 0810. Even-numbered columns sit half a hex lower than odd-numbered ones,
// as printed, and Distance is the metric that layout gives.
type Hex struct {
	Col int
	Row int
}

// ParseHex reads the four-digit identifier the p. 3 grid prints.
func ParseHex(s string) (Hex, error) {
	if len(s) != 4 {
		return Hex{}, fmt.Errorf("%w: %q, want four digits like 0101", ErrBadHex, s)
	}

	col, colErr := strconv.Atoi(s[:2])
	row, rowErr := strconv.Atoi(s[2:])

	if colErr != nil || rowErr != nil {
		return Hex{}, fmt.Errorf("%w: %q, want four digits like 0101", ErrBadHex, s)
	}

	h := Hex{Col: col, Row: row}
	if !h.OnGrid() {
		return Hex{}, fmt.Errorf("%w: %q is outside the %dx%d grid of p. 3", ErrBadHex, s, Columns, Rows)
	}

	return h, nil
}

// OnGrid reports whether the hex is one of the eighty the p. 3 grid prints.
func (h Hex) OnGrid() bool {
	return h.Col >= 1 && h.Col <= Columns && h.Row >= 1 && h.Row <= Rows
}

// String renders the four-digit identifier: column then row, each padded
// to two digits (p. 3).
func (h Hex) String() string { return fmt.Sprintf("%02d%02d", h.Col, h.Row) }

// AllHexes is the eighty hexes in ascending identifier order — column by
// column, row within column, 0101 through 0810. That is the order the
// p. 3 grid numbers them and the order the occurrence scan runs in
// (docs/ERRATA.md E002). Because the identifier is fixed-width, this order
// is also plain string order, which is what makes the record's hexes sort
// the way the scan ran.
func AllHexes() []Hex {
	out := make([]Hex, 0, Hexes)

	for col := 1; col <= Columns; col++ {
		for row := 1; row <= Rows; row++ {
			out = append(out, Hex{Col: col, Row: row})
		}
	}

	return out
}

// Distance is the number of jumps of one parsec between two hexes: the
// hex-grid distance on the p. 3 layout (p. 1, "1 hex = 1 parsec").
func (h Hex) Distance(o Hex) int {
	x1, y1, z1 := h.cube()
	x2, y2, z2 := o.cube()

	return (abs(x1-x2) + abs(y1-y2) + abs(z1-z2)) / 2
}

// cube converts the offset address the grid prints into cube coordinates,
// in which hex distance is a subtraction.
//
// The conversion has to match the printed offset, and the printed offset
// is legible on p. 3: 0101 sits at the top-left corner and 0201 half a hex
// below it, so even-numbered columns are the ones pushed down. In the
// zero-based indices below that is the odd columns, which is the standard
// "odd-q" offset layout.
//
// Getting the parity backwards is the failure this whole comment exists to
// prevent: it leaves every distance internally consistent and silently
// wrong by one hex for half the grid, which no replay can catch, because
// replay compares a record against a record. TestDistanceMatchesThePrintedGrid
// measures against the page instead.
func (h Hex) cube() (int, int, int) {
	col := h.Col - 1
	row := h.Row - 1

	x := col
	z := row - (col-(col&1))/2

	return x, -x - z, z
}

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}

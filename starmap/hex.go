// Package starmap holds the domain types a record is written in and the
// record itself. Book 3 p. 1 heads the procedure STAR MAPPING and p. 2
// calls what it produces "the star map"; the package is named for that
// rather than for the sub-sector, because a Record is equally the sector
// of sixteen sub-sectors ERRATA E006 assembles.
//
// The rule dividing the types from the data is that types carry identity,
// never rule invariants. Hex enforces the eight columns and ten rows of
// the Book 3 p. 3 grid and Digit enforces the p. 2 alphabet, because an
// identifier and a notation are identity. A characteristic's value range
// is not: those stay int.
package starmap

import (
	"encoding/json"
	"fmt"
)

// The sub-sector hex grid printed on Book 3 p. 3: eight columns of ten
// rows, numbered 0101 through 0810. One hex is one parsec (p. 1).
const (
	Columns = 8
	Rows    = 10
)

// A sector is four sub-sectors by four laid on one grid, 0101 through
// 3240. The book maps "in convenient segments, called subsectors" (p. 1)
// and charts growth as "additional subsectors will have to be charted",
// so it prints no sector grid; this is the shape issue 1 asked for, and
// ERRATA E006 is the reading that assembles it.
const (
	SectorAcross  = 4
	SectorColumns = Columns * SectorAcross
	SectorRows    = Rows * SectorAcross
)

// Grid is the hex grid a record is drawn on. A record carries its own
// because the renderer cannot infer one -- a subsector whose eighth
// column drew no world is not a seven-column subsector.
type Grid struct {
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

// PageThreeGrid is the sub-sector grid p. 3 prints, which is what a
// record is on unless it says otherwise.
func PageThreeGrid() Grid { return Grid{Columns: Columns, Rows: Rows} }

// SectorGrid is sixteen of those.
func SectorGrid() Grid { return Grid{Columns: SectorColumns, Rows: SectorRows} }

// Contains reports whether a hex is on this grid.
func (g Grid) Contains(h Hex) bool {
	return h.Col >= 1 && h.Col <= g.Columns && h.Row >= 1 && h.Row <= g.Rows
}

// Zero reports a grid no one set, which Decode reads as the p. 3 grid so
// that a record written before grids were recorded still reads.
func (g Grid) Zero() bool { return g.Columns == 0 && g.Rows == 0 }

// hold returns the off-grid error for a hex this grid does not contain,
// and nil for one it does. Every hex a record carries is checked through
// here, because Hex bounds itself by the largest grid there is: 0910
// parses, and only the record's own grid refuses it.
func (g Grid) hold(h Hex) error {
	if g.Contains(h) {
		return nil
	}

	return fmt.Errorf("%w: hex %s, and the grid is %dx%d", ErrOffGrid, h, g.Columns, g.Rows)
}

// The shape of the four-digit grid number the p. 3 grid prints: two
// decimal digits of column followed by two of row.
const (
	hexDigits      = 4
	hexCoordDigits = 2
	decimal        = 10
)

// Hex identifies one hex of the subsector grid. It marshals to the
// four-digit column-and-row number the p. 3 grid prints -- "0101" -- which
// is the identifier a referee writes in a notebook (p. 4).
//
// The zero value is not a hex, and marshaling it is an error.
type Hex struct{ Col, Row int }

// NewHex returns the hex at a column and row of the p. 3 grid, both
// one-based.
func NewHex(col, row int) (Hex, error) {
	h := Hex{Col: col, Row: row}
	if !h.valid() {
		return Hex{}, fmt.Errorf("%w: hex %d,%d, and the grid is %dx%d", ErrOffGrid, col, row, SectorColumns, SectorRows)
	}

	return h, nil
}

// ParseHex reads the four-digit form the p. 3 grid prints.
func ParseHex(text string) (Hex, error) {
	if len(text) != hexDigits {
		return Hex{}, fmt.Errorf("%w: %q", ErrNotAHex, text)
	}

	var col, row int

	for index, char := range []byte(text) {
		if char < '0' || char > '9' {
			return Hex{}, fmt.Errorf("%w: %q", ErrNotAHex, text)
		}

		if index < hexCoordDigits {
			col = col*decimal + int(char-'0')
		} else {
			row = row*decimal + int(char-'0')
		}
	}

	return NewHex(col, row)
}

// String returns the four-digit grid number.
func (h Hex) String() string { return fmt.Sprintf("%02d%02d", h.Col, h.Row) }

// Number is the grid number as an integer, which orders hexes the way the
// p. 3 grid numbers them: column by column, row within column. That order
// is the reading of ERRATA E002 and it fixes the dice stream.
func (h Hex) Number() int { return h.Col*100 + h.Row }

// Less orders hexes by ascending grid number (ERRATA E002).
func (h Hex) Less(o Hex) bool { return h.Number() < o.Number() }

// Distance returns the jump distance between two hexes. One hex is one
// parsec (p. 1).
func (h Hex) Distance(o Hex) Parsecs {
	ax, ay, az := h.cube()
	bx, by, bz := o.cube()

	return Parsecs(max(abs(ax-bx), abs(ay-by), abs(az-bz)))
}

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}

// MarshalJSON writes the four-digit grid number as a string.
func (h Hex) MarshalJSON() ([]byte, error) {
	if !h.valid() {
		return nil, fmt.Errorf("%w: hex %d,%d, and the grid is %dx%d", ErrOffGrid, h.Col, h.Row, SectorColumns, SectorRows)
	}

	b, err := json.Marshal(h.String())
	if err != nil {
		return nil, fmt.Errorf("marshaling hex %s: %w", h, err)
	}

	return b, nil
}

// UnmarshalJSON reads the four-digit grid number.
func (h *Hex) UnmarshalJSON(b []byte) error {
	var text string

	err := json.Unmarshal(b, &text)
	if err != nil {
		return fmt.Errorf("reading a hex: %w", err)
	}

	parsed, err := ParseHex(text)
	if err != nil {
		return err
	}

	*h = parsed

	return nil
}

// cube converts the offset coordinates of the p. 3 grid to cube
// coordinates.
//
// The grid prints 0101 at the top left with 0201 half a hex below it, so
// the even-numbered printed columns are the ones pushed down. In
// zero-based indices that is the standard odd-q vertical layout. Getting
// this backwards leaves every distance internally consistent and wrong by
// one for half the map, which no record-against-record check can catch --
// hex_test.go measures against the printed page instead. Never change
// this without re-measuring there.
func (h Hex) cube() (int, int, int) {
	// Every second column is pushed down half a hex, so a column
	// contributes half its index to the row offset (p. 3).
	const columnsPerRowStep = 2

	q := h.Col - 1
	r := h.Row - 1
	x := q
	z := r - (q-(q&1))/columnsPerRowStep

	return x, -x - z, z
}

// valid bounds a hex by the largest grid there is, because a hex is an
// identifier and the identifier is four digits either way. Whether a hex
// is on *this* record's grid is Grid.Contains, and the record checks it.
func (h Hex) valid() bool {
	return h.Col >= 1 && h.Col <= SectorColumns && h.Row >= 1 && h.Row <= SectorRows
}

// Parsecs is a jump distance. One hex is one parsec (p. 1).
type Parsecs int

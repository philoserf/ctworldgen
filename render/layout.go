package render

import (
	"github.com/philoserf/ctworldgen/starmap"
)

// The geometry of the sub-sector hex grid printed on Book 3 p. 3, read
// off the page rather than remembered.
//
// The page draws flat-top hexagons in columns: a flat edge across the top
// and the bottom, a vertex at the left and at the right, and the
// even-numbered columns pushed half a hex down. The four-digit grid
// number is printed small inside the upper-left of every hex.
//
// That parity is the same one starmap.Hex.cube encodes, arrived at by a
// second road, and it is the trap CLAUDE.md names: draw the odd columns
// low instead of the even ones and every hex still lands in a tidy grid,
// while half the map disagrees with Hex.Distance. Never change this
// without re-measuring against the printed page --
// TestTheDrawnMapIsTheGridPrintedOnPageThree is where that measurement
// lives.
const (
	// columnStep is how far apart two columns sit, in hex sides. A
	// flat-top hexagon is two sides wide and columns interlock by a
	// quarter of that at each edge.
	columnStep = 1.5

	// halfColumn is the fraction of a side a column's rightmost vertex
	// reaches past the last column's centre, which is what makes a grid of
	// C columns (1.5C + 0.5) sides wide.
	halfColumn = 0.5
)

// root3 is the height of a flat-top hexagon in sides: a row step, and
// twice the half-step the even-numbered columns are pushed down by. It is
// written out rather than computed because a package-level math.Sqrt(3)
// is a variable, and this is a constant of the shape.
const root3 = 1.7320508075688772

// Point is a place on the page, in PostScript points -- 72 to the inch,
// which is the unit the booklet is laid out in throughout.
type Point struct{ X, Y float64 }

// box is a rectangle of the page, measured from the top-left corner
// downward, as PDF user space with fpdf's default origin does.
type box struct{ X, Y, Width, Height float64 }

// window is the rectangle of a grid a map draws, in grid columns and rows
// inclusive. A subsector's map draws its whole grid; a sector's member map
// draws that member's band and one hex of bleed around it, which is why a
// map is fitted to a window rather than to a Grid.
//
// A window may run past the grid it is read against -- a member on the
// edge of the sector has no hexes on one side of it. That is deliberate:
// the fit is taken from the window so that all sixteen members draw at one
// size, and the drawing skips the hexes the record's grid does not hold.
type window struct{ FromCol, ToCol, FromRow, ToRow int }

func (w window) columns() int { return w.ToCol - w.FromCol + 1 }
func (w window) rows() int    { return w.ToRow - w.FromRow + 1 }

// wholeGrid is the window that draws every hex of a grid, which is what a
// subsector's map and a sector's index map both draw.
func wholeGrid(grid starmap.Grid) window {
	return window{FromCol: 1, ToCol: grid.Columns, FromRow: 1, ToRow: grid.Rows}
}

// memberWindow is the hexes one of a sector's sixteen members draws: its
// own eighty, and the ring of one hex around them (ERRATA E008).
//
// The ring is what makes a lane that crosses a seam land on a hex a
// referee can look up. A fifth of a sector's lanes cross one.
func memberWindow(index int) window {
	first, last := starmap.MemberBounds(index)

	return window{FromCol: first.Col - 1, ToCol: last.Col + 1, FromRow: first.Row - 1, ToRow: last.Row + 1}
}

// onMemberMap reports whether a hex is drawn on one member's map: one of
// its own eighty, or one of the ring one parsec outside them (ERRATA E008
// part 2).
//
// The ring is measured in parsecs and not taken as the rectangle a member
// is fitted in, because those are not the same set: the rectangle's corner
// is two parsecs from the nearest hex of the band, and a hex drawn there
// would sit in the ring looking like a neighbour and not be one.
//
// The map is still fitted to the rectangle, so all sixteen draw at one
// size whatever their ring costs them.
func onMemberMap(index int, hex starmap.Hex) bool {
	first, last := starmap.MemberBounds(index)

	if starmap.MemberOf(hex) == index {
		return true
	}

	// The hex of the band nearest this one, which is the one to measure
	// against: a hex outside a rectangle is closest to the point clamped
	// into it.
	near := starmap.Hex{
		Col: min(max(hex.Col, first.Col), last.Col),
		Row: min(max(hex.Row, first.Row), last.Row),
	}

	return hex.Distance(near) == 1
}

// everywhere is what a map that draws its whole window shows: a
// subsector's map and a sector's index both do.
func everywhere(starmap.Hex) bool { return true }

// mapFit is a drawn grid: the size of one hex, where the window's first
// hex sits, and which hex that is.
//
// The size is fitted rather than fixed, because the same drawing serves
// every window a document draws -- eight columns by ten, the ten by twelve
// of a member and its bleed, or the sector's thirty-two by forty (ERRATA
// E006) -- and a size that suits one overruns the page on the others.
type mapFit struct {
	Side   float64
	Origin Point
	From   starmap.Hex
}

// fitMap returns the largest hexes that draw a whole window inside a box,
// with the window's first hex placed so that the drawing's top-left corner
// is the box's.
//
// A window of C columns is (1.5C + 0.5) sides across and one of R rows is
// root3*(R + 0.5) sides down -- the half a row being the step the
// even-numbered columns are pushed by, which hangs below the last odd
// column's bottom edge.
func fitMap(draw window, within box) mapFit {
	side := min(
		within.Width/(columnStep*float64(draw.columns())+halfColumn),
		within.Height/(root3*(float64(draw.rows())+halfColumn)),
	)

	// The origin is where the window's first hex sits if its column is an
	// odd one: its leftmost vertex is one side left of it and its flat top
	// edge is half a hex height above. An even first column is pushed down
	// by hexCenter like any other, into the half a row the height above
	// reserves.
	return mapFit{
		Side:   side,
		Origin: Point{X: within.X + side, Y: within.Y + root3*side/two},
		From:   starmap.Hex{Col: draw.FromCol, Row: draw.FromRow},
	}
}

// hexCenter is where a hex is drawn. This is the whole of the p. 3
// parity, in one expression, in one place.
//
// The window moves the origin and nothing else. Parity is read off the
// hex's own column, never off its offset within the window: a member whose
// first column is even would otherwise draw its whole map upside down and
// still land in a tidy grid, which is the trap CLAUDE.md names.
func (f mapFit) hexCenter(hex starmap.Hex) Point {
	center := Point{
		X: f.Origin.X + float64(hex.Col-f.From.Col)*columnStep*f.Side,
		Y: f.Origin.Y + float64(hex.Row-f.From.Row)*root3*f.Side,
	}

	// The even-numbered columns sit half a hex low. That is how p. 3
	// prints the grid, and it is the direction that must never be flipped.
	if hex.Col%2 == 0 {
		center.Y += root3 * f.Side / two
	}

	return center
}

// hexSides is the six corners of a hexagon, which is what makes it one.
const hexSides = 6

// hexOutline returns the corners of the flat-top hexagon drawn around a
// centre, beginning at the rightmost vertex and running clockwise down
// the page.
func hexOutline(center Point, side float64) [hexSides]Point {
	half := side / two
	rise := root3 * side / two

	return [hexSides]Point{
		{X: center.X + side, Y: center.Y},
		{X: center.X + half, Y: center.Y + rise},
		{X: center.X - half, Y: center.Y + rise},
		{X: center.X - side, Y: center.Y},
		{X: center.X - half, Y: center.Y - rise},
		{X: center.X + half, Y: center.Y - rise},
	}
}

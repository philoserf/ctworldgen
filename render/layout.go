package render

import (
	"github.com/philoserf/ctworldgen/subsector"
)

// The geometry of the sub-sector hex grid printed on Book 3 p. 3, read
// off the page rather than remembered.
//
// The page draws flat-top hexagons in columns: a flat edge across the top
// and the bottom, a vertex at the left and at the right, and the
// even-numbered columns pushed half a hex down. The four-digit grid
// number is printed small inside the upper-left of every hex.
//
// That parity is the same one subsector.Hex.cube encodes, arrived at by a
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

// mapFit is a drawn grid: the size of one hex and where hex 0101 sits.
//
// The size is fitted rather than fixed, because the same drawing serves
// both grids a record can be on -- eight columns by ten, or the sector's
// thirty-two by forty (ERRATA E006) -- and a size that suits one overruns
// the page on the other.
type mapFit struct {
	Side   float64
	Origin Point
}

// fitMap returns the largest hexes that draw the whole grid inside a box,
// with hex 0101 placed so that the drawing's top-left corner is the box's.
//
// A grid of C columns is (1.5C + 0.5) sides across and one of R rows is
// root3*(R + 0.5) sides down -- the half a row being the step the
// even-numbered columns are pushed by, which hangs below the last odd
// column's bottom edge.
func fitMap(grid subsector.Grid, within box) mapFit {
	side := min(
		within.Width/(columnStep*float64(grid.Columns)+halfColumn),
		within.Height/(root3*(float64(grid.Rows)+halfColumn)),
	)

	// The origin is the centre of hex 0101, whose leftmost vertex is one
	// side left of it and whose flat top edge is half a hex height above.
	return mapFit{
		Side:   side,
		Origin: Point{X: within.X + side, Y: within.Y + root3*side/two},
	}
}

// hexCenter is where a hex is drawn. This is the whole of the p. 3
// parity, in one expression, in one place.
func (f mapFit) hexCenter(hex subsector.Hex) Point {
	center := Point{
		X: f.Origin.X + float64(hex.Col-1)*columnStep*f.Side,
		Y: f.Origin.Y + float64(hex.Row-1)*root3*f.Side,
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

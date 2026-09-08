package starmap

import (
	"fmt"
	"strconv"
	"strings"
)

// The two separators an area is written with: the DM from the rectangle,
// and the rectangle's two corners from each other.
const (
	areaDMSeparator     = "@"
	areaCornerSeparator = "-"
	areaCorners         = 2
)

// The world occurrence DM Book 3 p. 1 offers, which is the same rule for
// the whole subsector and for a broad area within one (ERRATA E012).
const (
	minOccurrenceDM = -1
	maxOccurrenceDM = 1
)

// HoldOccurrenceDM refuses a world occurrence DM the page does not offer.
//
// It is one function because p. 1 states one rule. The referee may impose
// "a DM of +1 or -1 on the whole subsector, or on broad areas within a
// subsector", so the record's own DM and every area's are bounded by the
// same sentence, and a second check would be a second place for that
// sentence to be got wrong. Zero is the absence of a DM rather than a
// value the page offers, and it is accepted as such.
func HoldOccurrenceDM(dm int) error {
	if dm < minOccurrenceDM || dm > maxOccurrenceDM {
		return fmt.Errorf("%w: %+d", ErrOccurrenceDM, dm)
	}

	return nil
}

// Area is one of the broad areas Book 3 p. 1 offers the referee, carrying
// its own world occurrence DM (ERRATA E012).
//
// It is a rectangle of the grid's own numbering -- a range of columns by a
// range of rows -- and not a geometric rectangle: p. 3 sets the
// even-numbered columns half a hex low, so the drawn shape has a ragged
// edge. What the referee points at on the sheet is the numbering.
//
// From is the low corner and To the high, the convention [Route] already
// uses, so one area is never written two ways.
type Area struct {
	From Hex `json:"from"`
	To   Hex `json:"to"`
	DM   int `json:"dm"`
}

// NewArea returns the area two opposite corners describe, whichever two
// they are. The referee points at a rectangle rather than at its low
// corner, so a range typed bottom-right to top-left is the same area.
func NewArea(a, b Hex, dm int) Area {
	return Area{
		From: Hex{Col: min(a.Col, b.Col), Row: min(a.Row, b.Row)},
		To:   Hex{Col: max(a.Col, b.Col), Row: max(a.Row, b.Row)},
		DM:   dm,
	}
}

// ParseArea reads the form the referee types: a DM, an @, and the two
// corners of the rectangle it applies to -- "-1@0101-0410".
//
// The corners are separated by the same hyphen a negative DM opens with,
// which is why the DM is cut off first: what remains has exactly one
// hyphen in it.
func ParseArea(text string) (Area, error) {
	written, rectangle, found := strings.Cut(text, areaDMSeparator)
	if !found {
		return Area{}, fmt.Errorf("%w: %q", ErrNotAnArea, text)
	}

	modifier, err := strconv.Atoi(written)
	if err != nil {
		return Area{}, fmt.Errorf("%w: %q", ErrNotAnArea, text)
	}

	corners := strings.Split(rectangle, areaCornerSeparator)
	if len(corners) != areaCorners {
		return Area{}, fmt.Errorf("%w: %q", ErrNotAnArea, text)
	}

	first, err := ParseHex(corners[0])
	if err != nil {
		return Area{}, fmt.Errorf("the first corner of %q: %w", text, err)
	}

	second, err := ParseHex(corners[1])
	if err != nil {
		return Area{}, fmt.Errorf("the second corner of %q: %w", text, err)
	}

	return NewArea(first, second, modifier), nil
}

// Contains reports whether a hex lies in this area.
func (a Area) Contains(h Hex) bool {
	return h.Col >= a.From.Col && h.Col <= a.To.Col &&
		h.Row >= a.From.Row && h.Row <= a.To.Row
}

// String returns the rectangle as its two corners, which is how a heading
// or an error names it. The DM is not in it: both documents write a DM
// through one formatter, so an area that carried its own would be a second
// place for "+1" and "0" to be written differently.
func (a Area) String() string { return a.From.String() + areaCornerSeparator + a.To.String() }

// overlaps reports whether two areas share a hex. Two ranges overlap
// unless one ends before the other begins, in both directions at once.
func (a Area) overlaps(o Area) bool {
	return a.From.Col <= o.To.Col && o.From.Col <= a.To.Col &&
		a.From.Row <= o.To.Row && o.From.Row <= a.To.Row
}

// HoldAreas is the one definition of what a legal set of broad areas is,
// so that the engine's inputs and the read path cannot come to enforce two
// different rules -- the pattern [Record.Validate] already follows, and for
// the same reason: a schema alone rejects nothing at read time.
//
// Three things are held. Every DM is one p. 1 offers. Every corner is on
// this grid, which [Hex] cannot do for itself -- an identifier is four
// digits whether it names a sub-sector or a sector, so 0910 parses and
// only the record's own grid refuses it. And no two areas overlap, which
// is ERRATA E012's fourth part: p. 1 gives no basis for combining two DMs,
// so a set that would need one is refused before any die is thrown.
//
// The order of the list is deliberately not held. Areas cannot overlap, so
// every hex has exactly one DM whatever order they are written in, and a
// referee who hand-edits his record is not made to sort it.
func (g Grid) HoldAreas(areas []Area) error {
	for index, area := range areas {
		err := HoldOccurrenceDM(area.DM)
		if err != nil {
			return fmt.Errorf("the broad area %s: %w", area, err)
		}

		if area.From.Col > area.To.Col || area.From.Row > area.To.Row {
			return fmt.Errorf("%w: %s", ErrAreaCorners, area)
		}

		for _, corner := range []Hex{area.From, area.To} {
			err = g.hold(corner)
			if err != nil {
				return fmt.Errorf("the broad area %s: %w", area, err)
			}
		}

		for _, other := range areas[index+1:] {
			if area.overlaps(other) {
				return fmt.Errorf("%w: %s and %s", ErrAreasOverlap, area, other)
			}
		}
	}

	return nil
}

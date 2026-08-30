package tables

import (
	"fmt"
	"slices"

	"github.com/philoserf/ctworldgen/dice"
)

// StarportColumn names the technological index matrix's first column
// (p. 9), the one keyed by a starport letter rather than a characteristic
// digit.
const StarportColumn = "starport"

var techColumns = [...]string{StarportColumn, Size, Atmosphere, Hydrographics, Population, Government}

// TechColumns is the six columns of the p. 9 technological index matrix,
// in the order it prints them. Law level is absent because the matrix has
// no column for it: the index is derived from the other six basics and the
// starport.
//
// A fresh slice each call, like StarportTypes.
func TechColumns() []string { return slices.Clone(techColumns[:]) }

// StarportForThrow reads the p. 1 starports table: a two-dice total to a
// starport type.
func (c *Charts) StarportForThrow(total int) (string, error) {
	t, ok := c.starportByThrow[total]
	if !ok {
		return "", fmt.Errorf("%w: starports table, throw %d", ErrNoSuchValue, total)
	}

	return t, nil
}

// Starport reads the p. 5 starport chart.
//
// A fresh copy each call, for the reason StarportTypes and its siblings
// clone: what a caller is handed must not be the package's own chart. This
// one is a pointer into a map rather than a slice header, so without the
// copy a single write through it would edit the chart for every later
// lookup on the same Charts — the loaded charts are the book, and nothing
// downstream should be able to rewrite them.
func (c *Charts) Starport(t string) (*Starport, error) {
	sp, ok := c.starports[t]
	if !ok {
		return nil, fmt.Errorf("%w: starport chart, type %q", ErrNoSuchValue, t)
	}

	out := *sp

	return &out, nil
}

// NavalBaseTarget is the p. 5 chart's naval base throw for a starport
// type, and whether the chart prints one at all (C, D, E and X do not).
func (c *Charts) NavalBaseTarget(t string) (dice.Target, bool) {
	target, ok := c.navalTarget[t]

	return target, ok
}

// ScoutBaseTarget is the p. 5 chart's scout base throw for a starport
// type, and whether the chart prints one at all (E and X do not).
func (c *Charts) ScoutBaseTarget(t string) (dice.Target, bool) {
	target, ok := c.scoutTarget[t]

	return target, ok
}

// RouteTarget reads the p. 2 jump routes table: the one-die target for a
// pair of starport types at a jump distance. It reports false where no
// lane is possible — the printed em-dash, a distance outside the table's
// four columns, or a pair with an X starport, which the table has no row
// for (docs/ERRATA.md E003). Only a pair of types the chart does not know
// at all is an error.
func (c *Charts) RouteTarget(a, b string, distance int) (dice.Target, bool, error) {
	for _, t := range []string{a, b} {
		if !slices.Contains(starportTypes[:], t) {
			return dice.Target{}, false, fmt.Errorf("%w: jump routes table, starport %q", ErrNoSuchValue, t)
		}
	}

	targets, ok := c.routeTargets[pairKey(a, b)]
	if !ok || distance < 1 || distance > MaxJumpDistance {
		return dice.Target{}, false, nil
	}

	target := targets[distance-1]
	if target == 0 {
		return dice.Target{}, false, nil
	}

	return dice.Target{Value: target, Mode: dice.Plus}, true, nil
}

// TechDM reads one cell of the p. 9 technological index matrix: the DM the
// named column contributes when that characteristic's own value is the
// given digit.
func (c *Charts) TechDM(column string, value byte) (int, error) {
	row, ok := c.tech[string(value)]
	if !ok {
		return 0, fmt.Errorf("%w: technological index matrix, value %q", ErrNoSuchValue, string(value))
	}

	switch column {
	case StarportColumn:
		return row.Starport, nil
	case Size:
		return row.Size, nil
	case Atmosphere:
		return row.Atmosphere, nil
	case Hydrographics:
		return row.Hydrographics, nil
	case Population:
		return row.Population, nil
	case Government:
		return row.Government, nil
	}

	return 0, fmt.Errorf("%w: technological index matrix, column %q", ErrNoSuchValue, column)
}

// Label reads a descriptive table (pp. 5-7). It reports false for a value
// past the table's last printed row, which for law level is a real state:
// E004 leaves law level uncapped, so a world can hold a level the p. 7
// table does not print a row for.
func (c *Charts) Label(table string, value int) (string, bool) {
	rows, ok := c.labels[table]
	if !ok || value < 0 || value >= len(rows) {
		return "", false
	}

	return rows[value].Label, true
}

// MaxValue is the last value a descriptive table prints a row for: the
// ceiling a generated value is clamped to (docs/ERRATA.md E004). It is
// read off the table rather than transcribed a second time, so the clamp
// and the chart cannot come apart — checkLabelRows is what makes the
// length mean this.
func (c *Charts) MaxValue(table string) (int, bool) {
	rows, ok := c.labels[table]
	if !ok || len(rows) == 0 {
		return 0, false
	}

	return len(rows) - 1, true
}

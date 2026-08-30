package tables

import (
	"fmt"
	"slices"
	"strings"

	"github.com/philoserf/ctworldgen/dice"
)

// twoDiceRange is every total two dice can show (B1 p. 2). The p. 1
// starports table prints a row for each, and all eleven must load: the
// throw is unconditional, so a missing row would turn a legal throw into a
// runtime failure.
const (
	minTwoDice = 2
	maxTwoDice = 12
)

func (c *Charts) loadStarports() error {
	var data starportData
	if err := read("starports.json", &data); err != nil {
		return err
	}

	for _, t := range data.Types {
		if _, dup := c.starports[t.Type]; dup {
			return fmt.Errorf("%w: starports.json: duplicate starport type %q", ErrInvalidData, t.Type)
		}

		c.starports[t.Type] = &t
	}

	for _, want := range starportTypes {
		if _, ok := c.starports[want]; !ok {
			return fmt.Errorf("%w: starports.json: no chart row for starport %q", ErrInvalidData, want)
		}
	}

	if len(c.starports) != len(starportTypes) {
		return fmt.Errorf("%w: starports.json: %d chart rows, want the %d types of p. 5",
			ErrInvalidData, len(c.starports), len(starportTypes))
	}

	if err := c.loadBaseTargets(); err != nil {
		return err
	}

	return c.loadStarportThrows(data.Throws)
}

// loadBaseTargets parses the p. 5 chart's printed base throws. An empty
// string is the chart printing no base for the type (E and X have
// neither, and C and D have no naval base), which is absence rather than
// a parse failure.
func (c *Charts) loadBaseTargets() error {
	for _, t := range starportTypes {
		sp, ok := c.starports[t]
		if !ok {
			// loadStarports asserts every type is present before calling
			// this, so reaching here means that check was removed.
			return fmt.Errorf("%w: starports.json: no chart row for starport %q", ErrInvalidData, t)
		}

		for _, base := range []struct {
			label  string
			notate string
			into   map[string]dice.Target
		}{
			{"naval_base", sp.NavalBase, c.navalTarget},
			{"scout_base", sp.ScoutBase, c.scoutTarget},
		} {
			if base.notate == "" {
				continue
			}

			target, err := dice.ParseTarget(base.notate)
			if err != nil {
				return fmt.Errorf("%w: starports.json: starport %s %s: %w", ErrInvalidData, t, base.label, err)
			}

			base.into[t] = target
		}
	}

	return nil
}

func (c *Charts) loadStarportThrows(throws []starportThrow) error {
	for _, row := range throws {
		if _, ok := c.starports[row.Type]; !ok {
			return fmt.Errorf("%w: starports.json: throw %d names unknown starport %q", ErrInvalidData, row.Die, row.Type)
		}

		if row.Die < minTwoDice || row.Die > maxTwoDice {
			return fmt.Errorf("%w: starports.json: throw %d is outside the two-dice range %d-%d",
				ErrInvalidData, row.Die, minTwoDice, maxTwoDice)
		}

		if _, dup := c.starportByThrow[row.Die]; dup {
			return fmt.Errorf("%w: starports.json: duplicate throw %d", ErrInvalidData, row.Die)
		}

		c.starportByThrow[row.Die] = row.Type
	}

	for die := minTwoDice; die <= maxTwoDice; die++ {
		if _, ok := c.starportByThrow[die]; !ok {
			return fmt.Errorf("%w: starports.json: no starport for a throw of %d", ErrInvalidData, die)
		}
	}

	return nil
}

// pairKey names a jump routes table row. The table prints each unordered
// pair once (p. 2), so the two types are sorted before they are joined and
// a lookup for C-A finds the printed A-C row.
func pairKey(a, b string) string {
	if a > b {
		a, b = b, a
	}

	return a + "-" + b
}

const maxOneDie = 6

func (c *Charts) loadRoutes() error {
	var data routeData
	if err := read("routes.json", &data); err != nil {
		return err
	}

	want := make([]int, MaxJumpDistance)
	for i := range want {
		want[i] = i + 1
	}

	if !slices.Equal(data.Distances, want) {
		return fmt.Errorf("%w: routes.json: distance columns are %v, want %v (the p. 2 table's four columns)",
			ErrInvalidData, data.Distances, want)
	}

	for _, row := range data.Pairs {
		if err := c.addRoutePair(row); err != nil {
			return err
		}
	}

	return c.checkRoutesComplete()
}

func (c *Charts) addRoutePair(row routePair) error {
	a, b, ok := strings.Cut(row.Pair, "-")
	if !ok {
		return fmt.Errorf("%w: routes.json: pair %q is not two starport types joined by a hyphen", ErrInvalidData, row.Pair)
	}

	for _, t := range []string{a, b} {
		if !slices.Contains(laneStarportTypes[:], t) {
			return fmt.Errorf("%w: routes.json: pair %q names %q, which has no row in the p. 2 table",
				ErrInvalidData, row.Pair, t)
		}
	}

	if len(row.Targets) != MaxJumpDistance {
		return fmt.Errorf("%w: routes.json: pair %q has %d targets, want %d",
			ErrInvalidData, row.Pair, len(row.Targets), MaxJumpDistance)
	}

	for i, target := range row.Targets {
		// 0 is the printed em-dash: no lane possible at that distance.
		if target < 0 || target > maxOneDie {
			return fmt.Errorf("%w: routes.json: pair %q jump-%d target %d is not 0 (no lane) or a one-die target 1-%d",
				ErrInvalidData, row.Pair, i+1, target, maxOneDie)
		}
	}

	key := pairKey(a, b)
	if _, dup := c.routeTargets[key]; dup {
		return fmt.Errorf("%w: routes.json: pair %q appears twice", ErrInvalidData, key)
	}

	c.routeTargets[key] = row.Targets

	return nil
}

func (c *Charts) checkRoutesComplete() error {
	for i, a := range laneStarportTypes {
		for _, b := range laneStarportTypes[i:] {
			if _, ok := c.routeTargets[pairKey(a, b)]; !ok {
				return fmt.Errorf("%w: routes.json: no row for pair %s", ErrInvalidData, pairKey(a, b))
			}
		}
	}

	// Fifteen unordered pairs of five types; anything more is a row the
	// p. 2 table does not print.
	if want := len(laneStarportTypes) * (len(laneStarportTypes) + 1) / 2; len(c.routeTargets) != want {
		return fmt.Errorf("%w: routes.json: %d pairs, want the %d of p. 2", ErrInvalidData, len(c.routeTargets), want)
	}

	return nil
}

func (c *Charts) loadTech() error {
	var data techData
	if err := read("techmatrix.json", &data); err != nil {
		return err
	}

	for _, row := range data.Rows {
		if len(row.Value) != 1 {
			return fmt.Errorf("%w: techmatrix.json: row value %q is not a single character", ErrInvalidData, row.Value)
		}

		if _, dup := c.tech[row.Value]; dup {
			return fmt.Errorf("%w: techmatrix.json: duplicate row %q", ErrInvalidData, row.Value)
		}

		c.tech[row.Value] = row
	}

	return c.checkTechComplete()
}

// checkTechComplete asserts a row for every value any characteristic can
// present to the matrix: the digits 0-9 and A-E that the p. 9 matrix
// prints, plus X for the starport. The clamps of docs/ERRATA.md E004 hold
// every generated value inside that set, and this is the assertion that
// they do.
func (c *Charts) checkTechComplete() error {
	want := []string{}

	for v := range 15 {
		d, ok := Digit(v)
		if !ok {
			return fmt.Errorf("%w: techmatrix.json: value %d has no digit", ErrInvalidData, v)
		}

		want = append(want, string(d))
	}

	want = append(want, StarportX)

	for _, v := range want {
		if _, ok := c.tech[v]; !ok {
			return fmt.Errorf("%w: techmatrix.json: no row for value %q", ErrInvalidData, v)
		}
	}

	if len(c.tech) != len(want) {
		return fmt.Errorf("%w: techmatrix.json: %d rows, want the %d of p. 9", ErrInvalidData, len(c.tech), len(want))
	}

	return nil
}

func (c *Charts) loadCharacteristics() error {
	var data characteristicData
	if err := read("characteristics.json", &data); err != nil {
		return err
	}

	for name, rows := range map[string][]labelRow{
		Size:          data.Size,
		Atmosphere:    data.Atmosphere,
		Hydrographics: data.Hydrographics,
		Population:    data.Population,
		Government:    data.Government,
		LawLevel:      data.LawLevel,
	} {
		if err := checkLabelRows(name, rows); err != nil {
			return err
		}

		c.labels[name] = rows
	}

	for _, name := range characteristicTables {
		if _, ok := c.labels[name]; !ok {
			return fmt.Errorf("%w: characteristics.json: no table %q", ErrInvalidData, name)
		}
	}

	return nil
}

// checkLabelRows asserts a table is the book's: consecutive values from 0
// up, written in the p. 2 notation, each with a label. Consecutiveness is
// what lets the clamp ceiling be read off the table's length rather than
// transcribed a second time.
func checkLabelRows(name string, rows []labelRow) error {
	if len(rows) == 0 {
		return fmt.Errorf("%w: characteristics.json: table %q is empty", ErrInvalidData, name)
	}

	for i, row := range rows {
		want, ok := Digit(i)
		if !ok {
			return fmt.Errorf("%w: characteristics.json: table %q is longer than the digit alphabet", ErrInvalidData, name)
		}

		if row.Value != string(want) {
			return fmt.Errorf("%w: characteristics.json: table %q row %d is %q, want %q (rows run 0 up without gaps)",
				ErrInvalidData, name, i, row.Value, string(want))
		}

		if strings.TrimSpace(row.Label) == "" {
			return fmt.Errorf("%w: characteristics.json: table %q row %q has no label", ErrInvalidData, name, row.Value)
		}
	}

	return nil
}

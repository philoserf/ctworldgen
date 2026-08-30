package worldgen

import (
	"fmt"

	"github.com/philoserf/ctworldgen/tables"
)

// worldCreation is the World Creation section (pp. 2-9) and steps 2B-2H of
// the p. 12 checklist: the seven basic characteristics in the order the
// book generates them, then the technological index derived from them.
//
// The order is the book's, and it is load-bearing twice over. It fixes
// dice-stream consumption for replay, and it is a data dependency:
// atmosphere and hydrographics read planetary size, government reads
// population, law level reads government, and the technological index
// reads six of them.
func (g *generator) worldCreation() error {
	g.step(StepWorldCreation, "Generate the characteristics of each world (pp. 2-9)")

	for i := range g.sub.Worlds {
		if err := g.world(&g.sub.Worlds[i]); err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) world(w *World) error {
	for _, generate := range []func(*World) error{
		g.size, g.atmosphere, g.hydrographics,
		g.population, g.government, g.lawLevel, g.technologicalIndex,
	} {
		if err := generate(w); err != nil {
			return err
		}
	}

	return g.profile(w)
}

// size is 2D - 2 (pp. 4, 12). The formula cannot leave the p. 5 table's
// 0-10, so the clamp never binds here; it is applied anyway so that every
// characteristic reaches the record through the same gate.
func (g *generator) size(w *World) error {
	total, _, ref := g.roll(throwSpec{
		step: StepWorldCreation, label: "planetary size", hex: w.Hex, count: 2,
		dms: []EventDM{{Source: "automatic (p. 4)", Value: -2}},
	})

	v, err := g.clamped(w.Hex, tables.Size, total, ref)
	if err != nil {
		return err
	}

	w.Size = v
	g.describe(w.Hex, "planetary size", tables.Size, v, ref)

	return nil
}

// atmosphere is 2D - 7 + planetary size, except that "a planet of size
// zero automatically has an atmosphere of zero" (pp. 4, 12) — automatic,
// so no die is thrown for it.
func (g *generator) atmosphere(w *World) error {
	if w.Size == 0 {
		w.Atmosphere = 0
		g.outcome(StepWorldCreation, w.Hex,
			"planetary atmosphere 0 automatically: the world is an asteroid/planetoid complex (p. 4)", 0)
		g.describe(w.Hex, "planetary atmosphere", tables.Atmosphere, 0, 0)

		return nil
	}

	total, _, ref := g.roll(throwSpec{
		step: StepWorldCreation, label: "planetary atmosphere", hex: w.Hex, count: 2,
		dms: []EventDM{
			{Source: "automatic (p. 4)", Value: -7},
			{Source: "planetary size (p. 4)", Value: w.Size},
		},
	})

	v, err := g.clamped(w.Hex, tables.Atmosphere, total, ref)
	if err != nil {
		return err
	}

	w.Atmosphere = v
	g.describe(w.Hex, "planetary atmosphere", tables.Atmosphere, v, ref)

	return nil
}

// vacuumOrExoticAtmosphere is the -4 hydrographics condition: "if planetary
// atmosphere is 0, 1, or greater than 9" (p. 4), which the p. 12 checklist
// writes as "Atmosphere 0, 1 or A+". The two wordings agree — A is 10 — and
// are one condition here (docs/ERRATA.md, Noted discrepancies).
func vacuumOrExoticAtmosphere(atmosphere int) bool {
	return atmosphere == 0 || atmosphere == 1 || atmosphere > 9
}

// hydrographics is 2D - 7 + planetary size, less 4 for an atmosphere that
// cannot hold an ocean, except that "a planetary size of 0 or 1 indicates
// an automatic result of 0" (pp. 4, 12) — automatic, so no die is thrown.
func (g *generator) hydrographics(w *World) error {
	if w.Size <= 1 {
		w.Hydrographics = 0
		g.outcome(StepWorldCreation, w.Hex,
			fmt.Sprintf("hydrographic percentage 0 automatically: planetary size %d (p. 4)", w.Size), 0)
		g.describe(w.Hex, "hydrographic percentage", tables.Hydrographics, 0, 0)

		return nil
	}

	dms := []EventDM{
		{Source: "automatic (p. 4)", Value: -7},
		{Source: "planetary size (p. 4)", Value: w.Size},
	}

	if vacuumOrExoticAtmosphere(w.Atmosphere) {
		dms = append(dms, EventDM{Source: "atmosphere 0, 1, or greater than 9 (p. 4)", Value: -4})
	}

	total, _, ref := g.roll(throwSpec{
		step: StepWorldCreation, label: "hydrographic percentage", hex: w.Hex, count: 2, dms: dms,
	})

	v, err := g.clamped(w.Hex, tables.Hydrographics, total, ref)
	if err != nil {
		return err
	}

	w.Hydrographics = v
	g.describe(w.Hex, "hydrographic percentage", tables.Hydrographics, v, ref)

	return nil
}

// population is 2D - 2, an exponent of 10 (pp. 8, 12). Like size, the
// formula cannot leave its table.
func (g *generator) population(w *World) error {
	total, _, ref := g.roll(throwSpec{
		step: StepWorldCreation, label: "population", hex: w.Hex, count: 2,
		dms: []EventDM{{Source: "automatic (p. 8)", Value: -2}},
	})

	v, err := g.clamped(w.Hex, tables.Population, total, ref)
	if err != nil {
		return err
	}

	w.Population = v
	g.describe(w.Hex, "population", tables.Population, v, ref)

	return nil
}

// government is 2D - 7 + the population digit (pp. 8, 12).
func (g *generator) government(w *World) error {
	total, _, ref := g.roll(throwSpec{
		step: StepWorldCreation, label: "planetary government", hex: w.Hex, count: 2,
		dms: []EventDM{
			{Source: "automatic (p. 8)", Value: -7},
			{Source: "population (p. 8)", Value: w.Population},
		},
	})

	v, err := g.clamped(w.Hex, tables.Government, total, ref)
	if err != nil {
		return err
	}

	w.Government = v
	g.describe(w.Hex, "planetary government", tables.Government, v, ref)

	return nil
}

// lawLevel is 2D - 7 + the government type (pp. 8, 12). It is the one
// characteristic with no ceiling: the p. 7 table ends at 9, but its note
// is written for levels above that and law level feeds no matrix
// (docs/ERRATA.md E004).
func (g *generator) lawLevel(w *World) error {
	total, _, ref := g.roll(throwSpec{
		step: StepWorldCreation, label: "law level", hex: w.Hex, count: 2,
		dms: []EventDM{
			{Source: "automatic (p. 8)", Value: -7},
			{Source: "government type (p. 8)", Value: w.Government},
		},
	})

	w.LawLevel = g.floored(w.Hex, "law level", total, ref)
	g.describe(w.Hex, "law level", tables.LawLevel, w.LawLevel, ref)

	return nil
}

// technologicalIndex is one die modified by the sum of the p. 9 matrix's
// DMs: each of the world's six matrix characteristics selects a row by its
// own value and contributes that row's cell in its own column. Only the
// non-zero cells are recorded — the matrix's em-dashes are "no DM", and
// the page says to "note all DMs indicated".
func (g *generator) technologicalIndex(w *World) error {
	dms, err := g.techDMs(w)
	if err != nil {
		return err
	}

	total, _, ref := g.roll(throwSpec{
		step: StepWorldCreation, label: "technological index", hex: w.Hex, count: 1, dms: dms,
	})

	w.TechnologicalIndex = g.clampRange(w.Hex, "technological index",
		total, minTechnologicalIndex, maxTechnologicalIndex, ref)

	return nil
}

func (g *generator) techDMs(w *World) ([]EventDM, error) {
	values := map[string]string{
		tables.StarportColumn: w.Starport,
		tables.Size:           "",
		tables.Atmosphere:     "",
		tables.Hydrographics:  "",
		tables.Population:     "",
		tables.Government:     "",
	}

	for column, v := range map[string]int{
		tables.Size:          w.Size,
		tables.Atmosphere:    w.Atmosphere,
		tables.Hydrographics: w.Hydrographics,
		tables.Population:    w.Population,
		tables.Government:    w.Government,
	} {
		d, ok := tables.Digit(v)
		if !ok {
			return nil, fmt.Errorf("hex %s: %s %d: %w", w.Hex, column, v, ErrNoDigit)
		}

		values[column] = string(d)
	}

	var dms []EventDM

	for _, column := range tables.TechColumns() {
		value := values[column]

		dm, err := g.charts.TechDM(column, value[0])
		if err != nil {
			return nil, fmt.Errorf("hex %s: %w", w.Hex, err)
		}

		if dm != 0 {
			dms = append(dms, EventDM{Source: fmt.Sprintf("%s %s (p. 9)", column, value), Value: dm})
		}
	}

	return dms, nil
}

// profile writes the eight characteristics as the string of digits of
// p. 4: the p. 4 box's order, one character each in the p. 2 notation, and
// nothing between them (docs/ERRATA.md E005).
func (g *generator) profile(w *World) error {
	g.stampErratum(ErrataProfile)

	out := []byte(w.Starport)

	for _, v := range []int{
		w.Size, w.Atmosphere, w.Hydrographics,
		w.Population, w.Government, w.LawLevel, w.TechnologicalIndex,
	} {
		d, ok := tables.Digit(v)
		if !ok {
			return fmt.Errorf("hex %s: value %d: %w", w.Hex, v, ErrNoDigit)
		}

		out = append(out, d)
	}

	w.Profile = string(out)
	g.outcome(StepWorldCreation, w.Hex, "profile "+w.Profile, 0)

	return nil
}

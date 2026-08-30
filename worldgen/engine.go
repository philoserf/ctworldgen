package worldgen

import (
	"fmt"

	"github.com/philoserf/ctworldgen/dice"
	"github.com/philoserf/ctworldgen/tables"
)

// Config is everything a generation run takes. There is nothing else: the
// procedure has no choice points, so the seed and these inputs determine
// the record completely (docs/PRD.md).
type Config struct {
	Seed         uint64
	Name         string
	OccurrenceDM int
}

// MaxOccurrenceDM bounds the referee's subsector-wide modifier on the
// world occurrence throw. P. 1 offers "a DM of +1 or -1" and no others, so
// those two and the unmodified throw are the whole range.
const MaxOccurrenceDM = 1

// occurrenceTarget is the world occurrence throw: one die, "marking the
// hex with a circle if the result is a 4, 5, or 6" (p. 1).
var occurrenceTarget = dice.Target{Value: 4, Mode: dice.Plus}

// The technological index runs "from zero to 18" (p. 9), and the
// technological level tables print rows 0 through 18 and stop (pp. 10-11).
// The one-die throw plus the matrix's DMs can leave that range at both
// ends (docs/ERRATA.md E004).
const (
	minTechnologicalIndex = 0
	maxTechnologicalIndex = 18
)

type generator struct {
	charts *tables.Charts
	stream *dice.Stream
	sub    *Subsector
	seq    int
	errata map[string]bool
}

// Generate walks Book 3 pp. 1-12 and returns the subsector record.
func Generate(cfg Config) (*Subsector, error) {
	if cfg.OccurrenceDM < -MaxOccurrenceDM || cfg.OccurrenceDM > MaxOccurrenceDM {
		return nil, fmt.Errorf("%w: occurrence DM %+d: p. 1 offers +1 or -1 and nothing else",
			ErrBadInput, cfg.OccurrenceDM)
	}

	charts, err := tables.Load()
	if err != nil {
		return nil, fmt.Errorf("loading the Book 3 charts: %w", err)
	}

	g := &generator{
		charts: charts,
		stream: dice.New(cfg.Seed),
		errata: map[string]bool{},
		sub: &Subsector{
			SchemaVersion: SchemaVersion,
			Ruleset:       Ruleset,
			EngineVersion: EngineVersion,
			RNG:           RNG{Algorithm: dice.Algorithm, Seed: cfg.Seed},
			Inputs:        Inputs{Name: cfg.Name, OccurrenceDM: cfg.OccurrenceDM},
			Errata:        []string{},
			Name:          cfg.Name,
			Worlds:        []World{},
			Routes:        []Route{},
			Events:        []Event{},
		},
	}

	if err := g.run(); err != nil {
		return nil, err
	}

	g.sub.Errata = g.stamped()

	return g.sub, nil
}

// run is the p. 12 checklist in order: map the subsector, then generate
// the worlds on it. The passes are the book's own numbered steps — p. 1
// throws for starports "for each world in the sub-sector", which the
// occurrence scan must have finished naming — so the pass structure is not
// a reading.
func (g *generator) run() error {
	g.occurrence()

	for _, pass := range []func() error{g.starports, g.routes, g.worldCreation} {
		if err := pass(); err != nil {
			return err
		}
	}

	return nil
}

// occurrence is p. 1 step 1: one die per hex, a world on 4, 5, or 6, with
// the referee's optional subsector-wide DM. The scan order is
// docs/ERRATA.md E002.
func (g *generator) occurrence() {
	g.step(StepOccurrence, "Check each hex of the subsector for a world (p. 1)")
	g.stampErratum(ErrataScanOrder)

	var dms []EventDM
	if g.sub.Inputs.OccurrenceDM != 0 {
		dms = []EventDM{{Source: "referee's subsector-wide modifier (p. 1)", Value: g.sub.Inputs.OccurrenceDM}}
	}

	for _, hex := range AllHexes() {
		id := hex.String()

		_, present, ref := g.roll(throwSpec{
			step: StepOccurrence, label: "world occurrence", hex: id,
			count: 1, dms: dms, target: &occurrenceTarget,
		})

		if present {
			g.sub.Worlds = append(g.sub.Worlds, World{Hex: id})
			g.outcome(StepOccurrence, id, "a world and its attendant star system are present", ref)
		}
	}

	g.outcome(StepOccurrence, "", fmt.Sprintf("%s in %d hexes", Plural(len(g.sub.Worlds), "world"), Hexes), 0)
}

// starports is p. 1 step 2 and, with it, the p. 5 chart's base throws:
// two dice for the starport type, then a throw for each base the chart
// prints for that type (docs/ERRATA.md E001).
func (g *generator) starports() error {
	g.step(StepStarport, "Determine the starport type of each world, and its bases (pp. 1, 5)")

	for i := range g.sub.Worlds {
		world := &g.sub.Worlds[i]

		total, _, ref := g.roll(throwSpec{
			step: StepStarport, label: "starport type", hex: world.Hex, count: 2,
		})

		starport, err := g.charts.StarportForThrow(total)
		if err != nil {
			return fmt.Errorf("hex %s: %w", world.Hex, err)
		}

		world.Starport = starport
		g.outcome(StepStarport, world.Hex, "starport type "+starport, ref)

		g.bases(world)
	}

	return nil
}

// bases makes the p. 5 chart's naval and scout base throws, in the order
// the chart lists them, for the starport types it prints a throw for.
func (g *generator) bases(world *World) {
	for _, base := range []struct {
		label string
		into  *bool
		get   func(string) (dice.Target, bool)
	}{
		{"naval base", &world.NavalBase, g.charts.NavalBaseTarget},
		{"scout base", &world.ScoutBase, g.charts.ScoutBaseTarget},
	} {
		target, printed := base.get(world.Starport)
		if !printed {
			continue
		}

		g.stampErratum(ErrataBaseThrows)

		_, present, ref := g.roll(throwSpec{
			step: StepStarport, label: base.label, hex: world.Hex,
			count: 2, target: &target,
		})

		*base.into = present

		if present {
			g.outcome(StepStarport, world.Hex, base.label+" present", ref)
		}
	}
}

// routes is p. 1 step 3: one die per eligible pair of worlds against the
// p. 2 jump routes table. Which pairs are eligible, and in what order, is
// docs/ERRATA.md E003.
func (g *generator) routes() error {
	g.step(StepRoutes, "Determine the space lanes connecting the worlds (p. 2)")

	if len(g.sub.Worlds) >= 2 {
		g.stampErratum(ErrataLanePairs)
	}

	for i := range g.sub.Worlds {
		for j := i + 1; j < len(g.sub.Worlds); j++ {
			if err := g.lane(&g.sub.Worlds[i], &g.sub.Worlds[j]); err != nil {
				return err
			}
		}
	}

	g.outcome(StepRoutes, "", Plural(len(g.sub.Routes), "space lane")+" charted", 0)

	return nil
}

// lane examines one pair of worlds. A pair the p. 2 table has no cell for
// — further apart than four hexes, an X starport on either side, or a
// printed em-dash — consumes no die: there is no throw to make.
func (g *generator) lane(from, to *World) error {
	a, err := ParseHex(from.Hex)
	if err != nil {
		return fmt.Errorf("route from %s: %w", from.Hex, err)
	}

	b, err := ParseHex(to.Hex)
	if err != nil {
		return fmt.Errorf("route to %s: %w", to.Hex, err)
	}

	distance := a.Distance(b)

	target, possible, err := g.charts.RouteTarget(from.Starport, to.Starport, distance)
	if err != nil {
		return fmt.Errorf("route %s-%s: %w", from.Hex, to.Hex, err)
	}

	if !possible {
		return nil
	}

	label := fmt.Sprintf("space lane %s-%s at jump-%d", from.Starport, to.Starport, distance)

	_, exists, ref := g.roll(throwSpec{
		step: StepRoutes, label: label, hex: from.Hex + "-" + to.Hex,
		count: 1, target: &target,
	})

	if exists {
		g.sub.Routes = append(g.sub.Routes, Route{From: from.Hex, To: to.Hex, Distance: distance})
		g.outcome(StepRoutes, from.Hex+"-"+to.Hex, fmt.Sprintf("space lane at jump-%d", distance), ref)
	}

	return nil
}

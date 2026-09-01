// Package gen walks the star mapping and world creation procedure of Book
// 3 pp. 1-12 and fills a record.
//
// Dice-stream consumption order is load-bearing: it fixes what a seed
// means. The passes below follow the p. 12 checklist exactly, each a
// complete pass over the whole subsector before the next begins, and
// within a pass the hexes run in ascending grid number (ERRATA E002).
// Reordering any of it changes every seed's meaning.
package gen

import (
	"errors"
	"fmt"
	"math"

	"github.com/philoserf/ctworldgen/dice"
	"github.com/philoserf/ctworldgen/starmap"
	"github.com/philoserf/ctworldgen/tables"
)

// occurrenceTarget is the world occurrence throw of p. 1. The page marks
// the hex "if the result is a 4, 5, or 6", which is a target and not a set
// of three faces: the same paragraph lets the referee make worlds more or
// less frequent by a DM, and only a target reading makes the DM do that.
const occurrenceTarget dice.Target = 4

// worldsInAPair is the point at which ERRATA E003 has something to govern.
const worldsInAPair = 2

// The automatic DMs the world creation throws carry: "a two dice throw,
// minus an automatic DM of -2" for size and population, and "an automatic
// DM of -7" for the four centred on another characteristic (pp. 4, 8, 12).
// The -4 is R8's own, applied when the atmosphere is 0, 1 or above 9.
const (
	automaticMinusTwo   = 2
	automaticMinusSeven = 7
	dryAtmosphereDM     = 4
)

// ErrOccurrenceDM is the referee's world occurrence DM, which p. 1 offers
// as +1, 0 or -1 and nothing else.
var ErrOccurrenceDM = errors.New("occurrence DM is not one of -1, 0 or +1 (Book 3 p. 1)")

// Inputs are the referee's choices, made before the run rather than
// during it. They are recorded, and with the seed they reproduce the
// subsector exactly.
type Inputs struct {
	Seed         uint64
	Name         string
	OccurrenceDM int
}

// Validate reports whether the inputs are ones the book offers.
func (in Inputs) Validate() error {
	if in.OccurrenceDM < -1 || in.OccurrenceDM > 1 {
		return fmt.Errorf("%w: %+d", ErrOccurrenceDM, in.OccurrenceDM)
	}

	return nil
}

// Engine walks the procedure. It holds the charts of pp. 1-12, which are
// read and validated once when the engine is built rather than once per
// subsector -- the charts are the same charts for every seed.
type Engine struct {
	charts *tables.Tables
}

// New loads and validates the charts, and returns the engine that walks
// them. A chart that does not describe its whole printed range fails here
// rather than at the throw that needed it.
func New() (*Engine, error) {
	charts, err := tables.Load()
	if err != nil {
		return nil, fmt.Errorf("loading the Book 3 charts: %w", err)
	}

	return &Engine{charts: charts}, nil
}

// Generate produces a subsector.
func (e *Engine) Generate(inputs Inputs) (*starmap.Record, error) {
	err := inputs.Validate()
	if err != nil {
		return nil, err
	}

	record := starmap.New(inputs.Seed, inputs.Name, inputs.OccurrenceDM)
	stream := dice.NewStream(inputs.Seed)

	// The order of the passes, and of the hexes within them, is a reading.
	record.Stamp("E002")

	// 1.A. Throw for each hex; 4, 5, or 6 indicates a world is present.
	// The referee's DM applies to the whole subsector (p. 1).
	hexes, err := scan(stream, inputs.OccurrenceDM)
	if err != nil {
		return nil, err
	}

	// 1.B. Determine starport type; two dice throw and consult the
	// starports table (pp. 1, 12). The base throws follow immediately,
	// naval then scout, because a base is a property of the starport and
	// the starport chart is where the throw is printed (ERRATA E001).
	err = e.starports(stream, record, hexes)
	if err != nil {
		return nil, err
	}

	// 1.C. "Determine space lanes; check all possible jump routes"
	// (pp. 2, 12). Quoted from the p. 12 checklist, which is why this one
	// says "space lanes" where the rest of the code says routes: the page
	// prints both words and the code picked one, but a quotation keeps the
	// page's.
	record.Routes = e.routes(stream, record.Worlds)

	// A pair is the thing the reading governs, so it takes two worlds to
	// have governed anything.
	if len(record.Worlds) >= worldsInAPair {
		record.Stamp("E003")
	}

	// 2. Generate specific worlds.
	err = e.createWorlds(stream, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// createWorlds is pass 2: each world finished before the next is begun
// (p. 12, ERRATA E002).
func (e *Engine) createWorlds(stream *dice.Stream, record *starmap.Record) error {
	for index := range record.Worlds {
		err := e.detail(stream, &record.Worlds[index])
		if err != nil {
			return err
		}

		if len(record.Worlds[index].Clamps) > 0 {
			record.Stamp("E004")
		}
	}

	if len(record.Worlds) > 0 {
		record.Stamp("E005")
	}

	return nil
}

// maxTechIndex is the one cap the book prints for a value rather than a
// table that happens to end: "Technological index may vary from zero to
// 18" (p. 9). It binds -- the matrix's DMs reach +14, which with a die of
// 6 gives 20 (ERRATA E004).
const maxTechIndex = 18

// uncapped is the high bound for every characteristic but the
// technological index. Nothing is capped because a table stops describing
// it: a value no table names is a gap in the descriptive table, and p. 8's
// remedy for one is the referee's own description rather than a clamp.
const uncapped = math.MaxInt

// detail is pass 2: the six characteristics of pp. 4-8 and then the
// technological index of p. 9, in the order the p. 12 checklist lists them
// at 2.B through 2.H.
func (e *Engine) detail(stream *dice.Stream, world *starmap.World) error {
	// 2.B. Planetary size. 2D-2 (pp. 4, 12).
	world.Size = clamp(world, starmap.Size, stream.D2()-automaticMinusTwo, uncapped)

	// 2.C. Planetary atmosphere. 2D-7 + planetary size. A planet of size
	// zero automatically has an atmosphere of zero, and no die is thrown:
	// rolling and dropping one would shift every later world (R13).
	if world.Size > 0 {
		world.Atmosphere = clamp(world, starmap.Atmosphere, stream.D2()-automaticMinusSeven+world.Size, uncapped)
	}

	// 2.D. Hydrographic percentage. 2D-7 + planetary size, with a further
	// DM of -4 if the atmosphere is 0, 1, or greater than 9. A size of 0 or
	// 1 gives an automatic 0, and again no die is thrown (R13).
	if world.Size > 1 {
		raw := stream.D2() - automaticMinusSeven + world.Size
		if world.Atmosphere <= 1 || world.Atmosphere > 9 {
			raw -= dryAtmosphereDM
		}

		world.Hydrographics = clamp(world, starmap.Hydrographics, raw, uncapped)
	}

	// 2.E. Population. 2D-2, an exponent of 10 (pp. 8, 12).
	world.Population = clamp(world, starmap.Population, stream.D2()-automaticMinusTwo, uncapped)

	// 2.F. Planetary government. 2D-7 + the population digit (pp. 8, 12).
	world.Government = clamp(world, starmap.Government, stream.D2()-automaticMinusSeven+world.Population, uncapped)

	// 2.G. Law level. 2D-7 + the government type (pp. 8, 12). Government
	// feeds this already floored: a clamped value is the value.
	world.LawLevel = clamp(world, starmap.LawLevel, stream.D2()-automaticMinusSeven+world.Government, uncapped)

	// 2.H. Technological index. One die, modified by the sum of the DMs the
	// matrix gives for the starport, size, atmosphere, hydrographics,
	// population and government (p. 9).
	matrix := e.charts.TechIndexMatrix
	modifier := dice.Sum(
		matrix.StarportDM(world.Starport),
		matrix.DM(tables.ColSize, world.Size),
		matrix.DM(tables.ColAtmosphere, world.Atmosphere),
		matrix.DM(tables.ColHydrographics, world.Hydrographics),
		matrix.DM(tables.ColPopulation, world.Population),
		matrix.DM(tables.ColGovernment, world.Government),
	)

	world.TechIndex = clamp(world, starmap.TechIndex, stream.Die()+modifier, maxTechIndex)

	digits, err := world.DigitString()
	if err != nil {
		return fmt.Errorf("the string of digits: %w", err)
	}

	world.Digits = digits

	return nil
}

// clamp holds a generated value inside what the notation and the page
// allow, and records on the world when it bound.
//
// The floor at 0 is forced rather than chosen: R15 requires one character
// per characteristic, and neither Book 1 p. 8's hexadecimal nor Book 3
// p. 2's letters has a character for a negative number (ERRATA E004).
func clamp(world *starmap.World, which starmap.Characteristic, raw, high int) int {
	value := min(max(raw, 0), high)

	if value != raw {
		world.Clamps = append(world.Clamps, starmap.Clamp{Characteristic: which, Raw: raw, Value: value})
	}

	return value
}

// starports is pass 1.B: a starport type for every world found, each
// followed at once by its base throws, naval then scout (ERRATA E001).
func (e *Engine) starports(stream *dice.Stream, record *starmap.Record, hexes []starmap.Hex) error {
	record.Worlds = make([]starmap.World, 0, len(hexes))

	basesThrown := false

	for _, hex := range hexes {
		port, err := e.charts.Starports.Type(stream.D2())
		if err != nil {
			return fmt.Errorf("starport for the world at %s: %w", hex, err)
		}

		world := starmap.World{
			Hex: hex, Name: "", Notes: "", Starport: port,
			NavalBase: false, ScoutBase: false,
			Size: 0, Atmosphere: 0, Hydrographics: 0, Population: 0,
			Government: 0, LawLevel: 0, TechIndex: 0,
			Digits: "", Clamps: nil,
		}

		if target, printed := e.charts.StarportChart.NavalBase(port); printed {
			world.NavalBase = target.Met(stream.D2())
			basesThrown = true
		}

		if target, printed := e.charts.StarportChart.ScoutBase(port); printed {
			world.ScoutBase = target.Met(stream.D2())
			basesThrown = true
		}

		record.Worlds = append(record.Worlds, world)
	}

	if basesThrown {
		record.Stamp("E001")
	}

	return nil
}

// routes is pass 1.C: every pair of worlds, once, in ascending hex number
// of the first world and then of the second (ERRATA E003).
//
// A die is thrown only where the jump routes table states a number. A pair
// with an X starport has no row, a dash cell states nothing, and neither
// is examined -- p. 2 describes the throw as being made against a stated
// number, and neither states one. Consuming a die there would shift every
// throw after it.
func (e *Engine) routes(stream *dice.Stream, worlds []starmap.World) []starmap.Route {
	routes := []starmap.Route{}

	for i, first := range worlds {
		for _, second := range worlds[i+1:] {
			distance := first.Hex.Distance(second.Hex)

			target, stated := e.charts.JumpRoutes.Target(first.Starport, second.Starport, distance)
			if !stated {
				continue
			}

			if target.Met(stream.Die()) {
				routes = append(routes, starmap.Route{From: first.Hex, To: second.Hex, Distance: distance})
			}
		}
	}

	return routes
}

// scan is pass 1.A: every hex of the grid in ascending grid number, one
// die each. A hex that fails the throw is left blank and consumes its die
// like any other.
func scan(stream *dice.Stream, occurrenceDM int) ([]starmap.Hex, error) {
	var found []starmap.Hex

	for col := 1; col <= starmap.Columns; col++ {
		for row := 1; row <= starmap.Rows; row++ {
			hex, err := starmap.NewHex(col, row)
			if err != nil {
				return nil, fmt.Errorf("hex %d,%d: %w", col, row, err)
			}

			if occurrenceTarget.Met(stream.Die() + occurrenceDM) {
				found = append(found, hex)
			}
		}
	}

	return found, nil
}

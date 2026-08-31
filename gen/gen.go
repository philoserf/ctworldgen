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

	"github.com/philoserf/ctworldgen/dice"
	"github.com/philoserf/ctworldgen/subsector"
	"github.com/philoserf/ctworldgen/tables"
)

// occurrenceTarget is the world occurrence throw of p. 1. The page marks
// the hex "if the result is a 4, 5, or 6", which is a target and not a set
// of three faces: the same paragraph lets the referee make worlds more or
// less frequent by a DM, and only a target reading makes the DM do that.
const occurrenceTarget dice.Target = 4

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
func (e *Engine) Generate(inputs Inputs) (*subsector.Subsector, error) {
	err := inputs.Validate()
	if err != nil {
		return nil, err
	}

	record := subsector.New(inputs.Seed, inputs.Name, inputs.OccurrenceDM)
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
	// starports table (pp. 1, 12).
	record.Worlds = make([]subsector.World, 0, len(hexes))
	for _, hex := range hexes {
		port, err := e.charts.Starports.Type(stream.D2())
		if err != nil {
			return nil, fmt.Errorf("starport for the world at %s: %w", hex, err)
		}

		record.Worlds = append(record.Worlds, subsector.World{Hex: hex, Name: "", Starport: port})
	}

	return record, nil
}

// scan is pass 1.A: every hex of the grid in ascending grid number, one
// die each. A hex that fails the throw is left blank and consumes its die
// like any other.
func scan(stream *dice.Stream, occurrenceDM int) ([]subsector.Hex, error) {
	var found []subsector.Hex

	for col := 1; col <= subsector.Columns; col++ {
		for row := 1; row <= subsector.Rows; row++ {
			hex, err := subsector.NewHex(col, row)
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

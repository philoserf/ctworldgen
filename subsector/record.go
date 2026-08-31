package subsector

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/philoserf/ctworldgen/dice"
)

// The provenance constants a record stamps.
const (
	// SchemaVersion tracks the shape of the records the engine writes.
	SchemaVersion = 1

	// Ruleset names the held pages that govern: Book 3 Worlds and
	// Adventures pp. 1-12 of the FFE reprint of the (c) 1977 text.
	Ruleset = "ct-1977-book3-pp1-12"

	// EngineVersion bumps for three things and no others: a rule change, a
	// change to dice-stream consumption order, or a change to the RNG
	// construction. Nothing else can alter what a seed produces, and it is
	// deliberately not the build's own version -- a tag cut for a
	// documentation fix must not move it.
	//
	// It is "0" until the engine walks the whole of pp. 1-12.
	EngineVersion = "0"
)

// Subsector is the record. The JSON record is the source of truth; the
// Markdown listing is a render of it.
//
// Name, OccurrenceDM and Seed are the inputs a run is reproducible from,
// and the regeneration test in gen reads them back to reproduce each
// golden.
type Subsector struct {
	SchemaVersion int      `json:"schema_version"`
	Ruleset       string   `json:"ruleset"`
	EngineVersion string   `json:"engine_version"`
	RNGAlgorithm  string   `json:"rng_algorithm"`
	Seed          uint64   `json:"seed"`
	Errata        []string `json:"errata"`
	Name          string   `json:"name"`
	OccurrenceDM  int      `json:"occurrence_dm"`
	Worlds        []World  `json:"worlds"`
}

// World is one generated world.
//
// Name is always empty: Book 3 p. 12 step 3 says to name each world and
// prints no table, so the hex identifies it. The field is carried because
// the record is the referee's notebook page and he writes in it.
type World struct {
	Hex      Hex      `json:"hex"`
	Name     string   `json:"name"`
	Starport Starport `json:"starport"`
}

// New returns a record stamped with the provenance of the run that is
// about to fill it.
func New(seed uint64, name string, occurrenceDM int) *Subsector {
	return &Subsector{
		SchemaVersion: SchemaVersion,
		Ruleset:       Ruleset,
		EngineVersion: EngineVersion,
		RNGAlgorithm:  dice.Algorithm,
		Seed:          seed,
		Errata:        []string{},
		Name:          name,
		OccurrenceDM:  occurrenceDM,
		Worlds:        []World{},
	}
}

// Stamp records that a reading governed this record, keeping the errata
// in document order and never repeating one.
func (s *Subsector) Stamp(id string) {
	if slices.Contains(s.Errata, id) {
		return
	}

	s.Errata = append(s.Errata, id)
	for i := len(s.Errata) - 1; i > 0 && s.Errata[i] < s.Errata[i-1]; i-- {
		s.Errata[i], s.Errata[i-1] = s.Errata[i-1], s.Errata[i]
	}
}

// Decode reads a record, refusing any field the current schema does not
// define.
//
// Rejecting unknown fields is two obligations: "additionalProperties":
// false at every level of subsector.schema.json, and DisallowUnknownFields
// here. A schema alone rejects nothing at read time. Both are required, so
// that a record from a newer schema fails loudly rather than silently
// dropping data.
func Decode(r io.Reader) (*Subsector, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var record Subsector

	err := dec.Decode(&record)
	if err != nil {
		return nil, fmt.Errorf("decoding the record: %w", err)
	}

	return &record, nil
}

// Marshal renders a record as the JSON a golden holds: indented, with a
// trailing newline, so that a fixture is readable and diffs line by line.
//
// It lives here, beside Decode, because the goldens are compared byte for
// byte against what it writes: a second definition of this shape anywhere
// would let a fixture and the command drift apart, and the diff would read
// as a moved dice stream when the stream had not moved.
func Marshal(s *Subsector) ([]byte, error) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling the record: %w", err)
	}

	return append(b, '\n'), nil
}

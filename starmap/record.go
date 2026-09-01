package starmap

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

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
	// It was "0" while the engine was being built, because during the
	// milestones every increment would have been a rule change and the
	// number would have been noise. It became "1" at the milestone that
	// finished the walk of pp. 1-12, and the three conditions above are
	// the whole rule from here.
	EngineVersion = "1"
)

// Record is one generated map: the p. 3 sub-sector grid, or the sector of
// sixteen of them (ERRATA E006). Its Grid says which. The JSON record is
// the source of truth; the Markdown listing is a render of it.
//
// Name, OccurrenceDM and Seed are the inputs a run is reproducible from,
// and the regeneration test in gen reads them back to reproduce each
// golden.
type Record struct {
	SchemaVersion int      `json:"schema_version"`
	Ruleset       string   `json:"ruleset"`
	EngineVersion string   `json:"engine_version"`
	RNGAlgorithm  string   `json:"rng_algorithm"`
	Seed          uint64   `json:"seed"`
	Errata        []string `json:"errata"`
	Name          string   `json:"name"`
	OccurrenceDM  int      `json:"occurrence_dm"`

	// Grid is what the hexes below are numbered on: the p. 3 sub-sector
	// grid, or the sector grid of sixteen of them (ERRATA E006).
	Grid Grid `json:"grid"`

	Worlds []World `json:"worlds"`
	Routes []Route `json:"routes"`
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

	// NavalBase and ScoutBase are the throws the p. 5 starport chart
	// prints and the p. 12 checklist omits (ERRATA E001). Which of them
	// was thrown for at all is recomputable from the starport type: the
	// chart prints a naval throw only at A and B, a scout throw at A
	// through D, and neither at E or X.
	NavalBase bool `json:"naval_base"`
	ScoutBase bool `json:"scout_base"`

	// The six characteristics thrown for on pp. 4-8, then the
	// technological index of p. 9, in the order the p. 4 Planetary
	// Characteristics box lists them.
	//
	// These stay int. A range in a type would be a rules claim, and a
	// rules claim belongs on a page with a cite: R14 floors every one of
	// them at 0 because the notation has no character for a negative
	// number, and caps only the technological index, because p. 9 prints
	// a range for that value rather than a table that happens to end.
	Size          int `json:"size"`
	Atmosphere    int `json:"atmosphere"`
	Hydrographics int `json:"hydrographics"`
	Population    int `json:"population"`
	Government    int `json:"government"`
	LawLevel      int `json:"law_level"`
	TechIndex     int `json:"tech_index"`

	// Digits is the string of digits of p. 4: eight characters, the
	// starport type and then the seven above, with nothing between them
	// (ERRATA E005). It is derived from the values beside it and stored
	// because it is what a referee reads.
	Digits string `json:"digits"`

	// Clamps records a floor or the cap that actually bound, and is absent
	// where nothing did, which is most worlds. The raw value is the one
	// thing here that cannot be recomputed from a value, which is why it
	// is kept.
	Clamps []Clamp `json:"clamps,omitempty"`
}

// Clamp is a floor or the cap that bound a generated value (R14, ERRATA
// E004). Both are called clamps because they are the one mechanism.
//
// A clamped value is the value: it is what the record carries and what
// every later step consumes, so government feeds law level already
// floored. Raw is used for nothing but the record.
type Clamp struct {
	Characteristic Characteristic `json:"characteristic"`
	Raw            int            `json:"raw"`
	Value          int            `json:"value"`
}

// Route is a commercial route between two worlds (p. 2).
//
// From is the lower hex identifier: a route is not directed, and ordering
// the pair is what keeps one route from being written two ways.
type Route struct {
	From     Hex     `json:"from"`
	To       Hex     `json:"to"`
	Distance Parsecs `json:"distance"`
}

// New returns a record stamped with the provenance of the run that is
// about to fill it.
func New(seed uint64, name string, occurrenceDM int) *Record {
	return &Record{
		SchemaVersion: SchemaVersion,
		Ruleset:       Ruleset,
		EngineVersion: EngineVersion,
		RNGAlgorithm:  dice.Algorithm,
		Seed:          seed,
		Errata:        []string{},
		Name:          name,
		OccurrenceDM:  occurrenceDM,
		Grid:          PageThreeGrid(),
		Worlds:        []World{},
		Routes:        []Route{},
	}
}

// Stamp records that a reading governed this record, keeping the errata
// in document order and never repeating one.
func (s *Record) Stamp(id string) {
	if slices.Contains(s.Errata, id) {
		return
	}

	s.Errata = append(s.Errata, id)
	for i := len(s.Errata) - 1; i > 0 && s.Errata[i] < s.Errata[i-1]; i-- {
		s.Errata[i], s.Errata[i-1] = s.Errata[i-1], s.Errata[i]
	}
}

// Decode reads a record and holds it to what record.schema.json states:
// the three provenance constants, the fields it marks required, the two
// grids it names, and no field it does not define.
//
// Every one of those is two obligations -- the schema, and a check here --
// because a schema alone rejects nothing at read time. Rejecting an
// unknown field is DisallowUnknownFields below and
// "additionalProperties": false in the schema; the rest are the checks
// beneath this function.
//
// Unknown-field rejection was once the whole of it, on the reasoning that
// a record from a newer schema would fail loudly. That holds only when the
// newer schema *added* a field. A record claiming a different schema
// version, a different ruleset, or a different generator parsed cleanly
// and rendered, which is the one thing this record cannot afford: it would
// report a subsector under provenance stamps that are not true of it.
//
// A record is one document, and content after it is refused for the same
// reason: a file holding two records -- concatenated by hand, or written
// by any tool that emits a stream of them -- would otherwise decode the
// first and discard the rest in silence.
func Decode(r io.Reader) (*Record, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var record Record

	err := dec.Decode(&record)
	if err != nil {
		return nil, fmt.Errorf("decoding the record: %w", err)
	}

	// A record written before grids were recorded is a subsector, which is
	// the only shape the tool wrote then. The reporter of issue 1 has
	// sixteen such files; they still read.
	if record.Grid.Zero() {
		record.Grid = PageThreeGrid()
	}

	// The schema names these two grids and nothing else, and unknown-shape
	// rejection is two obligations: the schema, and this.
	if record.Grid != PageThreeGrid() && record.Grid != SectorGrid() {
		return nil, fmt.Errorf("%w: %dx%d", ErrNotAGrid, record.Grid.Columns, record.Grid.Rows)
	}

	err = record.carriesThisToolsProvenance()
	if err != nil {
		return nil, err
	}

	err = record.carriesTheFieldsTheSchemaRequires()
	if err != nil {
		return nil, err
	}

	err = record.onItsOwnGrid()
	if err != nil {
		return nil, err
	}

	_, err = dec.Token()
	if !errors.Is(err, io.EOF) {
		return nil, ErrTrailingContent
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
func Marshal(s *Record) ([]byte, error) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling the record: %w", err)
	}

	return append(b, '\n'), nil
}

// DigitString returns the string of digits of Book 3 p. 4: the starport
// followed by the seven characteristics, in the order the p. 4 Planetary
// Characteristics box lists them, with nothing between them.
//
// Eight characters. The familiar hyphen before the technological index --
// A867A69-8 -- is not in the held text and is not added here (ERRATA
// E005). Every characteristic is stored numerically beside this, so the
// format loses nothing.
func (w World) DigitString() (string, error) {
	var built strings.Builder

	built.WriteString(w.Starport.String())

	for _, value := range []int{
		w.Size, w.Atmosphere, w.Hydrographics,
		w.Population, w.Government, w.LawLevel, w.TechIndex,
	} {
		digit, err := NewDigit(value)
		if err != nil {
			return "", fmt.Errorf("the world at %s: %w", w.Hex, err)
		}

		built.WriteString(digit.String())
	}

	return built.String(), nil
}

// carriesThisToolsProvenance holds the three stamps record.schema.json
// states as constants against the constants this package defines.
//
// These are the fields a referee trusts without checking: they say which
// pages govern and which generator drew the dice. A record carrying any
// other value did not come from this engine, and nothing later in the read
// would notice -- every remaining field would parse and the listing would
// render.
//
// The string of digits is deliberately not checked against the values
// beside it. A referee adjusts a world to suit his campaign, and the
// record is his notebook page; the digits are what he reads, so they are
// carried as written rather than recomputed under him.
func (s *Record) carriesThisToolsProvenance() error {
	if s.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: the record says %d and this is schema %d",
			ErrNotThisSchema, s.SchemaVersion, SchemaVersion)
	}

	if s.Ruleset != Ruleset {
		return fmt.Errorf("%w: the record says %q and this tool implements %q",
			ErrNotThisRuleset, s.Ruleset, Ruleset)
	}

	if s.RNGAlgorithm != dice.Algorithm {
		return fmt.Errorf("%w: the record says %q and this tool draws from %q",
			ErrNotThisRNG, s.RNGAlgorithm, dice.Algorithm)
	}

	return nil
}

// carriesTheFieldsTheSchemaRequires holds the record to the ten fields
// record.schema.json lists as required.
//
// Six of them are already held: schema_version, ruleset and rng_algorithm
// are the constants above, and an absent one reads as a zero that cannot
// match. The three arrays are checked here, and encoding/json makes that
// exact rather than approximate -- it leaves an absent array nil and an
// explicit [] an empty non-nil slice, so the two are distinguishable
// without a second pass over the document.
//
// name, occurrence_dm and seed need no check and get none: "", 0 and 0 are
// all values the engine legitimately writes. The seed-zero golden is a
// record with a seed of 0 and no name.
func (s *Record) carriesTheFieldsTheSchemaRequires() error {
	if s.EngineVersion == "" {
		return fmt.Errorf("%w: engine_version, which the schema gives a minimum length of 1", ErrFieldMissing)
	}

	for _, field := range []struct {
		name    string
		present bool
	}{
		{"errata", s.Errata != nil},
		{"worlds", s.Worlds != nil},
		{"routes", s.Routes != nil},
	} {
		if !field.present {
			return fmt.Errorf("%w: %s", ErrFieldMissing, field.name)
		}
	}

	// Two of a world's thirteen fields have an absence the zero value does
	// not stand for. The seven characteristics are legitimately 0, both
	// base flags are legitimately false, and a name is legitimately empty;
	// an absent hex is off every grid and onItsOwnGrid refuses it.
	//
	// An absent starport is Starport(0), which is nothing the book prints
	// -- and the map draws it, at a width no column allows, so the p. 3
	// grid stops being the p. 3 grid from that hex rightward.
	//
	// An absent string of digits is "", and DigitString always writes
	// eight characters, so the engine never writes one. The roster prints
	// an empty cell for it and the detail heading trails off after the
	// hex: a world reported with no characteristics at all, on a record
	// that carries every one of them in the fields beside it.
	for _, world := range s.Worlds {
		if !world.Starport.Valid() {
			return fmt.Errorf("%w: starport, for the world at %s", ErrFieldMissing, world.Hex)
		}

		if world.Digits == "" {
			return fmt.Errorf("%w: digits, for the world at %s", ErrFieldMissing, world.Hex)
		}
	}

	// A route's distance is the one field of its three the zero value does
	// not refuse on its own: an absent end is the zero Hex, which is off
	// every grid. The schema gives distance a minimum of 1 because one hex
	// is one parsec (p. 1) and a route joins two worlds, so a distance of 0
	// would be a world joined to itself.
	for _, route := range s.Routes {
		if route.Distance < 1 {
			return fmt.Errorf("%w: distance, for the route %s to %s", ErrFieldMissing, route.From, route.To)
		}
	}

	return nil
}

// onItsOwnGrid holds every hex the record carries against the grid the
// record says it is on. Hex bounds itself by the largest grid there is --
// an identifier is four digits whether it names a subsector or a sector --
// so 0910 parses, and this is the only thing that refuses it on a p. 3
// record.
//
// The routes are checked as well as the worlds. A route's two ends are
// hexes of the same grid, and a record reaching off its own grid is wrong
// about itself whether the hex it names carries a world or not.
func (s *Record) onItsOwnGrid() error {
	for _, world := range s.Worlds {
		err := s.Grid.hold(world.Hex)
		if err != nil {
			return err
		}
	}

	for _, route := range s.Routes {
		err := s.Grid.hold(route.From)
		if err != nil {
			return err
		}

		err = s.Grid.hold(route.To)
		if err != nil {
			return err
		}
	}

	return nil
}

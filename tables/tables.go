// Package tables holds the charts of Book 3 pp. 1-12 as embedded JSON,
// with the load-time validation and the typed lookups that make them
// usable.
//
// Every table here was transcribed from a visual read of the page, and is
// transcribed a second time inside tables_test.go. The two must agree.
// That second transcription is not redundancy for its own sake: the held
// PDFs' embedded font maps the em-dash to the glyph 4 and the minus sign
// to 3, so a text extraction of the jump routes table renders its empty
// cells as the digit 4 and reads the size formula "2D - 2" as "2D32".
// Both are wrong and both look like data.
package tables

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/philoserf/ctworldgen/dice"
	"github.com/philoserf/ctworldgen/starmap"
)

// The charts state these rules, so each error carries its cite here
// rather than in the message built where the load failed. Callers wrap
// with %w and add the row that offended.
var (
	// ErrNoStarportForThrow and ErrNoChartRow are returned by the lookups,
	// so they are the two a caller can reasonably match on.
	ErrNoStarportForThrow = errors.New("no starport type for that throw (the starports table of Book 3 p. 1 runs 2 to 12)")
	ErrNoChartRow         = errors.New("no chart row for that starport (Book 3 p. 5)")

	errDuplicateRow      = errors.New("row appears twice")
	errMissingRow        = errors.New("no row")
	errRowCount          = errors.New("wrong number of rows")
	errNotAOneDieTarget  = errors.New("not a one-die target")
	errNotATwoDiceTarget = errors.New("not a two-dice target")
	errNoDescription     = errors.New("no description")
	errNoPrintedMax      = errors.New("table does not state the last value it describes")
	errLabelCount        = errors.New("table describes the wrong number of values")
	errNoLabel           = errors.New("no label for a value in the printed range")
)

//go:embed data/*.json
var files embed.FS

// Tables is every chart of pp. 1-12 that generation consults.
type Tables struct {
	Starports       Starports
	JumpRoutes      JumpRoutes
	StarportChart   StarportChart
	TechIndexMatrix TechIndexMatrix

	// The descriptive tables of pp. 5-7. A generated value may exceed a
	// table's printed range -- an atmosphere of 13, a government of 14 --
	// and the book prints no label for one. That is a gap in the page, not
	// an error to correct: the listing prints the digit and no description.
	Size          Labels
	Atmosphere    Labels
	Hydrographics Labels
	Population    Labels
	Government    Labels
	LawLevels     Labels

	// TechLevels is pp. 10-11 and the chart E011 borrows to gloss them.
	// It generates nothing: the index is thrown from the p. 9 matrix.
	TechLevels TechLevels
}

// Load reads and validates every embedded table. It is the only way to
// get a Tables: a chart that does not describe its whole printed range,
// or that is missing a row, fails here rather than at the throw that
// needed it.
func Load() (*Tables, error) {
	var loaded Tables

	loaders := []struct {
		file string
		load func([]byte) error
	}{
		{"starports.json", func(b []byte) error { return loaded.Starports.load(b) }},
		{"jump_routes.json", func(b []byte) error { return loaded.JumpRoutes.load(b) }},
		{"starport_chart.json", func(b []byte) error { return loaded.StarportChart.load(b) }},
		{"technological_index_matrix.json", func(b []byte) error { return loaded.TechIndexMatrix.load(b) }},
		{"planetary_size.json", func(b []byte) error { return loaded.Size.load(b) }},
		{"planetary_atmosphere.json", func(b []byte) error { return loaded.Atmosphere.load(b) }},
		{"hydrographic_percentage.json", func(b []byte) error { return loaded.Hydrographics.load(b) }},
		{"population.json", func(b []byte) error { return loaded.Population.load(b) }},
		{"governmental_type.json", func(b []byte) error { return loaded.Government.load(b) }},
		{"law_levels.json", func(b []byte) error { return loaded.LawLevels.load(b) }},
		{"technological_eras.json", func(b []byte) error { return loaded.TechLevels.loadBorrowed(b) }},
		{"technological_levels.json", func(b []byte) error { return loaded.TechLevels.loadHeld(b) }},
	}
	for _, loader := range loaders {
		b, err := files.ReadFile("data/" + loader.file)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", loader.file, err)
		}

		err = loader.load(b)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", loader.file, err)
		}
	}

	return &loaded, nil
}

// The range of a two-dice throw (B1 pp. 2-3), which is the starports
// table's own range of rows.
const (
	minThrow = 2
	maxThrow = 12
)

// Starports is the starports table of p. 1: two dice for each world, read
// against a distribution of starport types.
type Starports struct{ types map[int]starmap.Starport }

// Type returns the starport for a two-dice throw.
func (s *Starports) Type(throw int) (starmap.Starport, error) {
	p, ok := s.types[throw]
	if !ok {
		return 0, fmt.Errorf("%w: %d", ErrNoStarportForThrow, throw)
	}

	return p, nil
}

func (s *Starports) load(data []byte) error {
	var doc struct {
		Rows []struct {
			Die  int    `json:"die"`
			Type string `json:"type"`
		} `json:"rows"`
	}

	err := json.Unmarshal(data, &doc)
	if err != nil {
		return fmt.Errorf("reading the starports table: %w", err)
	}

	s.types = make(map[int]starmap.Starport, len(doc.Rows))
	for _, row := range doc.Rows {
		port, err := starmap.ParseStarport(row.Type)
		if err != nil {
			return fmt.Errorf("die %d: %w", row.Die, err)
		}

		if _, dup := s.types[row.Die]; dup {
			return fmt.Errorf("%w: die %d", errDuplicateRow, row.Die)
		}

		s.types[row.Die] = port
	}

	for die := minThrow; die <= maxThrow; die++ {
		if _, ok := s.types[die]; !ok {
			return fmt.Errorf("%w: a throw of %d", errMissingRow, die)
		}
	}

	if want := maxThrow - minThrow + 1; len(s.types) != want {
		return fmt.Errorf("%w: the starports table has %d, want %d", errRowCount, len(s.types), want)
	}

	return nil
}

// JumpRoutes is the jump routes table of p. 2. Its rows run A-A through
// E-E and there is none for X; twenty-nine of its sixty cells print an
// em-dash. Neither an absent row nor a dash cell states a number, so
// neither is thrown against and neither consumes a die (ERRATA E003).
type JumpRoutes struct{ targets map[string][4]*int }

func pairKey(a, b starmap.Starport) string {
	if b < a {
		a, b = b, a
	}

	return a.String() + "-" + b.String()
}

// Target returns the one-die target for a pair of starports at a
// distance, and whether the table states one at all. It states none for a
// pair involving X, which has no row, and none at a dash cell; in both
// cases no die is thrown.
func (j *JumpRoutes) Target(a, b starmap.Starport, distance starmap.Parsecs) (dice.Target, bool) {
	if distance < 1 || distance > starmap.MaxJump {
		return 0, false
	}

	row, ok := j.targets[pairKey(a, b)]
	if !ok {
		return 0, false
	}

	target := row[distance-1]
	if target == nil {
		return 0, false
	}

	return dice.Target(*target), true
}

func (j *JumpRoutes) load(data []byte) error {
	var doc struct {
		Rows []struct {
			Pair    string  `json:"pair"`
			Targets [4]*int `json:"targets"`
		} `json:"rows"`
	}

	err := json.Unmarshal(data, &doc)
	if err != nil {
		return fmt.Errorf("reading the jump routes table: %w", err)
	}

	j.targets = make(map[string][4]*int, len(doc.Rows))
	for _, row := range doc.Rows {
		for i, target := range row.Targets {
			if target != nil && (*target < 1 || *target > 6) {
				return fmt.Errorf("%w: pair %s at jump-%d states %d", errNotAOneDieTarget, row.Pair, i+1, *target)
			}
		}

		if _, dup := j.targets[row.Pair]; dup {
			return fmt.Errorf("%w: pair %s", errDuplicateRow, row.Pair)
		}

		j.targets[row.Pair] = row.Targets
	}

	return j.verify()
}

// verify checks that the table has a row for every pair of starports the
// page prints one for, and no others. There is deliberately no row for X.
func (j *JumpRoutes) verify() error {
	pairs := 0
	ports := []starmap.Starport{
		starmap.StarportA, starmap.StarportB, starmap.StarportC,
		starmap.StarportD, starmap.StarportE,
	}

	for i, a := range ports {
		for _, b := range ports[i:] {
			key := pairKey(a, b)
			if _, ok := j.targets[key]; !ok {
				return fmt.Errorf("%w: for the pair %s", errMissingRow, key)
			}

			pairs++
		}
	}

	if len(j.targets) != pairs {
		return fmt.Errorf("%w: the jump routes table has %d, want %d", errRowCount, len(j.targets), pairs)
	}

	return nil
}

// StarportChart is the starport chart of p. 5: the description of each
// starport type, and the base throws the p. 12 checklist omits.
type StarportChart struct {
	rows map[starmap.Starport]ChartRow
}

// ChartRow is one starport type's line of the p. 5 chart. NavalBase and
// ScoutBase are nil where the chart prints no throw: starports E and X
// have neither, and C and D have no naval base.
type ChartRow struct {
	Description string
	NavalBase   *int
	ScoutBase   *int
}

// Row returns a starport type's line of the chart.
func (s *StarportChart) Row(p starmap.Starport) (ChartRow, error) {
	row, ok := s.rows[p]
	if !ok {
		return ChartRow{}, fmt.Errorf("%w: %s", ErrNoChartRow, p)
	}

	return row, nil
}

// NavalBase returns the throw a naval base is present on, and whether the
// chart prints one for this starport type at all.
func (s *StarportChart) NavalBase(p starmap.Starport) (dice.Target, bool) {
	row, ok := s.rows[p]
	if !ok || row.NavalBase == nil {
		return 0, false
	}

	return dice.Target(*row.NavalBase), true
}

// ScoutBase returns the throw a scout base is present on, and whether the
// chart prints one for this starport type at all.
func (s *StarportChart) ScoutBase(p starmap.Starport) (dice.Target, bool) {
	row, ok := s.rows[p]
	if !ok || row.ScoutBase == nil {
		return 0, false
	}

	return dice.Target(*row.ScoutBase), true
}

func (s *StarportChart) load(data []byte) error {
	var doc struct {
		Rows []struct {
			Type        string `json:"type"`
			Description string `json:"description"`
			NavalBase   *int   `json:"naval_base"`
			ScoutBase   *int   `json:"scout_base"`
		} `json:"rows"`
	}

	err := json.Unmarshal(data, &doc)
	if err != nil {
		return fmt.Errorf("reading the starport chart: %w", err)
	}

	s.rows = make(map[starmap.Starport]ChartRow, len(doc.Rows))

	for _, row := range doc.Rows {
		port, err := starmap.ParseStarport(row.Type)
		if err != nil {
			return fmt.Errorf("starport chart row %q: %w", row.Type, err)
		}

		err = checkChartRow(row.Type, row.Description, row.NavalBase, row.ScoutBase)
		if err != nil {
			return err
		}

		if _, dup := s.rows[port]; dup {
			return fmt.Errorf("%w: starport %s", errDuplicateRow, row.Type)
		}

		s.rows[port] = ChartRow{Description: row.Description, NavalBase: row.NavalBase, ScoutBase: row.ScoutBase}
	}

	for _, p := range starmap.Starports() {
		if _, ok := s.rows[p]; !ok {
			return fmt.Errorf("%w: for starport %s", errMissingRow, p)
		}
	}

	return nil
}

// checkChartRow holds one line of the p. 5 chart to what the page prints:
// a description, and base throws that a two-dice throw can reach.
func checkChartRow(typ, description string, naval, scout *int) error {
	if description == "" {
		return fmt.Errorf("%w: starport %s", errNoDescription, typ)
	}

	for name, throw := range map[string]*int{"naval base": naval, "scout base": scout} {
		if throw != nil && (*throw < minThrow || *throw > maxThrow) {
			return fmt.Errorf("%w: starport %s, %s throw of %d", errNotATwoDiceTarget, typ, name, *throw)
		}
	}

	return nil
}

// Column names one of the five value-indexed columns of the technological
// index matrix. The starport column is not among them: it is indexed by
// starport type, not by a number, and StarportDM reads it.
//
// The type exists so that the matrix cannot be asked for a column that
// does not exist -- law level and the technological index have none.
type Column int

// The value-indexed columns of the p. 9 matrix.
const (
	ColSize Column = iota
	ColAtmosphere
	ColHydrographics
	ColPopulation
	ColGovernment
)

// TechIndexMatrix is the technological index matrix of p. 9: the DMs a
// world's characteristics contribute to its one-die technological index
// throw.
//
// The matrix's Value column runs 0 through 9, A through E, and X. A
// generated value of 15 has no row at all, and an absent row contributes
// nothing -- which is what the printed dashes already mean (ERRATA E004).
type TechIndexMatrix struct {
	byValue    map[int][5]*int
	byStarport map[starmap.Starport]*int
}

// DM returns the modifier a characteristic's value contributes. A value
// with no row contributes nothing.
func (m *TechIndexMatrix) DM(col Column, value int) int {
	row, ok := m.byValue[value]
	if !ok {
		return 0
	}

	if dm := row[col]; dm != nil {
		return *dm
	}

	return 0
}

// StarportDM returns the modifier a starport type contributes.
func (m *TechIndexMatrix) StarportDM(p starmap.Starport) int {
	if dm, ok := m.byStarport[p]; ok && dm != nil {
		return *dm
	}

	return 0
}

func (m *TechIndexMatrix) load(data []byte) error {
	var doc struct {
		Rows []struct {
			Value         string `json:"value"`
			Starport      *int   `json:"starport"`
			Size          *int   `json:"size"`
			Atmosphere    *int   `json:"atmosphere"`
			Hydrographics *int   `json:"hydrographics"`
			Population    *int   `json:"population"`
			Government    *int   `json:"government"`
		} `json:"rows"`
	}

	err := json.Unmarshal(data, &doc)
	if err != nil {
		return fmt.Errorf("reading the technological index matrix: %w", err)
	}

	m.byValue = make(map[int][5]*int, len(doc.Rows))
	m.byStarport = make(map[starmap.Starport]*int, len(doc.Rows))

	seen := make(map[string]bool, len(doc.Rows))
	for _, row := range doc.Rows {
		if seen[row.Value] {
			return fmt.Errorf("%w: value %s", errDuplicateRow, row.Value)
		}

		seen[row.Value] = true
		if row.Value == starmap.StarportX.String() {
			m.byStarport[starmap.StarportX] = row.Starport

			continue
		}

		d, err := starmap.ParseDigit(row.Value)
		if err != nil {
			return fmt.Errorf("matrix row %q: %w", row.Value, err)
		}

		m.byValue[d.Value()] = [5]*int{row.Size, row.Atmosphere, row.Hydrographics, row.Population, row.Government}

		p, err := starmap.ParseStarport(row.Value)
		if err == nil {
			m.byStarport[p] = row.Starport
		}
	}

	return m.verify()
}

// matrixValues is the count of numbered rows the p. 9 Value column prints:
// 0 through 9 and A through E. X is a starport, not a value.
const matrixValues = 15

// verify checks the matrix describes every value and starport the page
// gives a row to.
func (m *TechIndexMatrix) verify() error {
	for value := range matrixValues {
		if _, ok := m.byValue[value]; !ok {
			return fmt.Errorf("%w: for the value %d", errMissingRow, value)
		}
	}

	for _, p := range starmap.Starports() {
		if _, ok := m.byStarport[p]; !ok {
			return fmt.Errorf("%w: for starport %s", errMissingRow, p)
		}
	}

	return nil
}

// Labels is a descriptive table of pp. 5-7, indexed by value from 0.
type Labels struct{ labels []string }

// PrintedMax is the last value the table describes.
func (l *Labels) PrintedMax() int { return len(l.labels) - 1 }

// Label returns a value's description and whether the table prints one. A
// generated value beyond the printed range has none, and the listing
// prints its digit alone.
func (l *Labels) Label(value int) (string, bool) {
	if value < 0 || value >= len(l.labels) {
		return "", false
	}

	return l.labels[value], true
}

func (l *Labels) load(data []byte) error {
	var doc struct {
		PrintedMax *int     `json:"printed_max"`
		Labels     []string `json:"labels"`
	}

	err := json.Unmarshal(data, &doc)
	if err != nil {
		return fmt.Errorf("reading a descriptive table: %w", err)
	}

	if doc.PrintedMax == nil {
		return errNoPrintedMax
	}

	if len(doc.Labels) != *doc.PrintedMax+1 {
		return fmt.Errorf("%w: %d, want %d (0 to %d)", errLabelCount, len(doc.Labels), *doc.PrintedMax+1, *doc.PrintedMax)
	}

	for value, label := range doc.Labels {
		if label == "" {
			return fmt.Errorf("%w: the value %d", errNoLabel, value)
		}
	}

	l.labels = doc.Labels

	return nil
}

// A technological index is not thrown against these two tables -- it is
// thrown from the p. 9 matrix (R12) -- so they generate nothing and are
// read only to say what an index means at the table.
//
// Both are printed sparse, and both are read the same way: the last entry
// at or below a level is what a world at that level has (ERRATA E009).
// A hole is the page's invitation to the referee, not an absence (E010).

// rung is one entry of one column, at the level the page prints it.
type rung struct {
	level float64
	entry string
}

// ladder is one column of a sparse table, in the order its levels
// ascend. Reading it is E009: the last entry at or below a level.
//
// The level is a float because T5 prints rows at 1.3, 1.6, 3.3 and 3.6. No
// index ever equals one, but 1.6 is below 2, so a world at index 2 has
// the cities that row prints rather than the villages of the row before
// it (E011).
type ladder []rung

func (l ladder) at(level float64) string {
	best := ""

	for _, printed := range l {
		if printed.level > level {
			break
		}

		best = printed.entry
	}

	return best
}

// Borrowed is what T5 says a level means: Core Book 2 pp. 230-232, cited
// for description alone (ERRATA E011).
//
// It is a type of its own so that the boundary is visible in the code as
// well as in the document. Nothing here is Book 3's, and a document that
// prints one of these fields says whose statement it is.
type Borrowed struct {
	Band      string
	Era       string
	Energy    string
	Society   string
	Environ   string
	Transport string
	Computers string
}

// Held is what Book 3 pp. 10-11 print a world at a level can build.
type Held struct {
	Personal      string
	Armor         string
	Special       string
	Computers     string
	Communication string

	Water string
	Land  string
	Air   string
	Space string
	Fuels string

	// MatterTransport is p. 11's row 16, which is printed across the
	// water, land and air columns rather than in one of them and so
	// belongs to none (E010 part 2).
	MatterTransport bool
}

// TechLevel is one technological index read across both tables.
type TechLevel struct {
	Borrowed Borrowed
	Held     Held
}

// TechLevels is the two tables, each column held as its own ladder.
type TechLevels struct {
	band      []bandRow
	spanLevel float64

	era, energy, society, environ, transport, borrowedComputers ladder

	personal, armor, special, computers, communication ladder
	water, land, air, space, fuels                     ladder
}

// bandRow is one of T5's bands and the levels it covers, inclusive.
type bandRow struct {
	name     string
	from, to int
}

// Level reads both tables at an index.
func (t *TechLevels) Level(index int) TechLevel {
	asLevel := float64(index)

	return TechLevel{
		Borrowed: Borrowed{
			Band:      t.bandAt(index),
			Era:       t.era.at(asLevel),
			Energy:    t.energy.at(asLevel),
			Society:   t.society.at(asLevel),
			Environ:   t.environ.at(asLevel),
			Transport: t.transport.at(asLevel),
			Computers: t.borrowedComputers.at(asLevel),
		},
		Held: Held{
			Personal:        t.personal.at(asLevel),
			Armor:           t.armor.at(asLevel),
			Special:         t.special.at(asLevel),
			Computers:       t.computers.at(asLevel),
			Communication:   t.communication.at(asLevel),
			Water:           t.water.at(asLevel),
			Land:            t.land.at(asLevel),
			Air:             t.air.at(asLevel),
			Space:           t.space.at(asLevel),
			Fuels:           t.fuels.at(asLevel),
			MatterTransport: asLevel >= t.spanLevel,
		},
	}
}

func (t *TechLevels) bandAt(index int) string {
	for _, row := range t.band {
		if index >= row.from && index <= row.to {
			return row.name
		}
	}

	return ""
}

// printedLevels is the range both tables print, which is also the range
// the index is capped to (E004 part 3).
const (
	minTechLevel = 0
	maxTechLevel = 18
)

func (t *TechLevels) loadHeld(data []byte) error {
	var doc struct {
		Rows []struct {
			Level         int     `json:"level"`
			Personal      *string `json:"personal"`
			Armor         *string `json:"armor"`
			Special       *string `json:"special"`
			Computers     *string `json:"computers"`
			Communication *string `json:"communication"`
			Water         *string `json:"water"`
			Land          *string `json:"land"`
			Air           *string `json:"air"`
			Space         *string `json:"space"`
			Fuels         *string `json:"fuels"`
		} `json:"rows"`
		Spanning struct {
			Level *int `json:"level"`
		} `json:"spanning"`
	}

	err := json.Unmarshal(data, &doc)
	if err != nil {
		return fmt.Errorf("reading the technological levels tables: %w", err)
	}

	if len(doc.Rows) != maxTechLevel-minTechLevel+1 {
		return fmt.Errorf("%w: %d, want %d (%d to %d)",
			errRowCount, len(doc.Rows), maxTechLevel-minTechLevel+1, minTechLevel, maxTechLevel)
	}

	if doc.Spanning.Level == nil {
		return fmt.Errorf("%w: p. 11's matter transport", errMissingRow)
	}

	t.spanLevel = float64(*doc.Spanning.Level)

	for want, row := range doc.Rows {
		if row.Level != want {
			return fmt.Errorf("%w: level %d where %d was expected; the rows ascend",
				errMissingRow, row.Level, want)
		}

		asLevel := float64(row.Level)

		t.personal = climb(t.personal, asLevel, row.Personal)
		t.armor = climb(t.armor, asLevel, row.Armor)
		t.special = climb(t.special, asLevel, row.Special)
		t.computers = climb(t.computers, asLevel, row.Computers)
		t.communication = climb(t.communication, asLevel, row.Communication)
		t.water = climb(t.water, asLevel, row.Water)
		t.land = climb(t.land, asLevel, row.Land)
		t.air = climb(t.air, asLevel, row.Air)
		t.space = climb(t.space, asLevel, row.Space)
		t.fuels = climb(t.fuels, asLevel, row.Fuels)
	}

	return nil
}

func (t *TechLevels) loadBorrowed(data []byte) error {
	var doc struct {
		Bands map[string]json.RawMessage `json:"bands"`
		Rows  []struct {
			Level     float64 `json:"level"`
			Band      string  `json:"band"`
			Era       *string `json:"era"`
			Energy    *string `json:"energy"`
			Society   *string `json:"society"`
			Environ   *string `json:"environ"`
			Transport *string `json:"transport"`
			Computers *string `json:"computers"`
		} `json:"rows"`
	}

	err := json.Unmarshal(data, &doc)
	if err != nil {
		return fmt.Errorf("reading the borrowed technology chart: %w", err)
	}

	err = t.loadBands(doc.Bands)
	if err != nil {
		return err
	}

	previous := math.Inf(-1)

	for _, row := range doc.Rows {
		if row.Level <= previous {
			return fmt.Errorf("%w: level %v does not ascend", errMissingRow, row.Level)
		}

		previous = row.Level

		if row.Band != t.bandAt(int(row.Level)) {
			return fmt.Errorf("%w: level %v is in band %q and the bands put it in %q",
				errMissingRow, row.Level, row.Band, t.bandAt(int(row.Level)))
		}

		t.era = climb(t.era, row.Level, row.Era)
		t.energy = climb(t.energy, row.Level, row.Energy)
		t.society = climb(t.society, row.Level, row.Society)
		t.environ = climb(t.environ, row.Level, row.Environ)
		t.transport = climb(t.transport, row.Level, row.Transport)
		t.borrowedComputers = climb(t.borrowedComputers, row.Level, row.Computers)
	}

	// Every index the tool can produce must sit in a band, or a world
	// would be glossed with no answer to the first thing the gloss says.
	for index := minTechLevel; index <= maxTechLevel; index++ {
		if t.bandAt(index) == "" {
			return fmt.Errorf("%w: no band covers level %d", errMissingRow, index)
		}
	}

	return nil
}

func (t *TechLevels) loadBands(bands map[string]json.RawMessage) error {
	for name, raw := range bands {
		var span []int

		if json.Unmarshal(raw, &span) != nil {
			// The bands object carries a comment beside its entries.
			continue
		}

		const fromAndTo = 2

		if len(span) != fromAndTo {
			return fmt.Errorf("%w: band %s covers %d levels, want a first and a last",
				errRowCount, name, len(span))
		}

		t.band = append(t.band, bandRow{name: name, from: span[0], to: span[1]})
	}

	if len(t.band) == 0 {
		return fmt.Errorf("%w: the chart names no bands", errMissingRow)
	}

	return nil
}

// climb appends an entry to a column's ladder where the page prints one.
func climb(column ladder, level float64, entry *string) ladder {
	if entry == nil {
		return column
	}

	return append(column, rung{level: level, entry: *entry})
}

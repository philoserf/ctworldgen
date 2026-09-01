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

// MaxJump is the greatest distance the jump routes table states a target
// for, and so the greatest distance at which a route is possible.
const MaxJump starmap.Parsecs = 4

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
	if distance < 1 || distance > MaxJump {
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

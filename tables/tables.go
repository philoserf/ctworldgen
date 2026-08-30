// Package tables holds the data-driven Book 3 charts: the starports table
// and starport chart (pp. 1, 5), the jump routes table (p. 2), the
// technological index matrix (p. 9), and the descriptive tables of the
// seven basic characteristics (pp. 5-7). The tables are embedded JSON
// validated at load time; the procedure's orchestration and its
// exceptional mechanics stay in the worldgen package (docs/PRD.md,
// Architecture notes).
package tables

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/philoserf/ctworldgen/dice"
)

//go:embed data/*.json
var dataFS embed.FS

// ErrInvalidData is a broken embedded chart: a build defect, surfaced at
// load rather than at some later roll.
var ErrInvalidData = errors.New("invalid chart data")

// ErrNoSuchValue is a lookup for a value no chart row describes. After the
// clamps of docs/ERRATA.md E004 the engine cannot raise it, so reaching it
// means a chart and the clamp bounds have come apart.
var ErrNoSuchValue = errors.New("no chart row for value")

// digitAlphabet is the notation a characteristic value is written in: the
// UPP's hexadecimal digits (B1 p. 8, 10-15 as A-F) extended by Book 3
// p. 2's letter set, "letters (A through Z, omitting O and I as they may
// be confused with numbers)". So 16 is G, 17 is H, and 18 is J
// (docs/ERRATA.md E005).
const digitAlphabet = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// Digit renders a characteristic value in the p. 2 notation. It reports
// false for a value the alphabet cannot write, which after E004's clamps
// the engine cannot produce.
func Digit(v int) (byte, bool) {
	if v < 0 || v >= len(digitAlphabet) {
		return 0, false
	}

	return digitAlphabet[v], true
}

// Value reads a digit back to its numeric value: the inverse of Digit,
// used to key the chart rows the data files write as characters.
func Value(d byte) (int, bool) {
	i := strings.IndexByte(digitAlphabet, d)

	return i, i >= 0
}

// Starport types, in the order the p. 5 starport chart lists them. Type X
// is "no starport" rather than a grade of one, which is why it has no row
// in the jump routes table (docs/ERRATA.md E003).
const (
	StarportA = "A"
	StarportB = "B"
	StarportC = "C"
	StarportD = "D"
	StarportE = "E"
	StarportX = "X"
)

var starportTypes = [...]string{StarportA, StarportB, StarportC, StarportD, StarportE, StarportX}

// StarportTypes is the six starport codes in the p. 5 chart's order.
//
// A fresh slice each call: a caller that writes to what it is handed must
// not reach the package's own copy.
func StarportTypes() []string { return slices.Clone(starportTypes[:]) }

// laneStarportTypes is the five types the jump routes table has rows for
// (p. 2): everything but X.
var laneStarportTypes = [...]string{StarportA, StarportB, StarportC, StarportD, StarportE}

// The characteristic tables, named as the data file keys them. These are
// the tables a generated value is clamped against (docs/ERRATA.md E004)
// and the tables render reads its labels from.
const (
	Size          = "size"
	Atmosphere    = "atmosphere"
	Hydrographics = "hydrographics"
	Population    = "population"
	Government    = "government"
	LawLevel      = "law_level"
)

var characteristicTables = [...]string{Size, Atmosphere, Hydrographics, Population, Government, LawLevel}

// CharacteristicTables is the six basic characteristics that have a
// descriptive table, in the order the p. 4 Planetary Characteristics box
// lists them (starport type is the seventh basic and has its own chart;
// the technological index is derived, not thrown against a table).
//
// A fresh slice each call, like StarportTypes.
func CharacteristicTables() []string { return slices.Clone(characteristicTables[:]) }

// MaxJumpDistance is the reach of the jump routes table's four columns
// (p. 2): pairs of worlds further apart than this are not examined.
const MaxJumpDistance = 4

// Starport is one row of the p. 5 starport chart. NavalBase and ScoutBase
// hold the chart's printed throws in the book's notation, empty where the
// chart prints no base for the type.
type Starport struct {
	Type      string `json:"type"`
	Quality   string `json:"quality"`
	Fuel      string `json:"fuel"`
	Overhaul  bool   `json:"overhaul"`
	Shipyard  string `json:"shipyard"`
	NavalBase string `json:"naval_base"`
	ScoutBase string `json:"scout_base"`
}

// starportThrow is one row of the p. 1 two-dice starports table.
type starportThrow struct {
	Die  int    `json:"die"`
	Type string `json:"type"`
}

type starportData struct {
	Note   string          `json:"note"`
	Throws []starportThrow `json:"throws"`
	Types  []Starport      `json:"types"`
}

// routePair is one row of the p. 2 jump routes table: the one-die target
// at each of the four jump distances, with 0 for the printed em-dash.
type routePair struct {
	Pair    string `json:"pair"`
	Targets []int  `json:"targets"`
}

type routeData struct {
	Note      string      `json:"note"`
	Distances []int       `json:"distances"`
	Pairs     []routePair `json:"pairs"`
}

// techRow is one row of the p. 9 technological index matrix: the DM each
// characteristic contributes when its own value is this row's value.
type techRow struct {
	Value         string `json:"value"`
	Starport      int    `json:"starport"`
	Size          int    `json:"size"`
	Atmosphere    int    `json:"atmosphere"`
	Hydrographics int    `json:"hydrographics"`
	Population    int    `json:"population"`
	Government    int    `json:"government"`
}

type techData struct {
	Note string    `json:"note"`
	Rows []techRow `json:"rows"`
}

// labelRow is one row of a descriptive table (pp. 5-7).
type labelRow struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type characteristicData struct {
	Note          string     `json:"note"`
	Size          []labelRow `json:"size"`
	Atmosphere    []labelRow `json:"atmosphere"`
	Hydrographics []labelRow `json:"hydrographics"`
	Population    []labelRow `json:"population"`
	Government    []labelRow `json:"government"`
	LawLevel      []labelRow `json:"law_level"`
}

// Charts is every Book 3 chart the procedure reads, loaded and validated.
type Charts struct {
	starportByThrow map[int]string
	starports       map[string]*Starport
	navalTarget     map[string]dice.Target
	scoutTarget     map[string]dice.Target
	routeTargets    map[string][]int
	tech            map[string]techRow
	labels          map[string][]labelRow
}

// Load reads and validates the embedded charts. A failure here is a build
// defect, not a runtime condition.
func Load() (*Charts, error) {
	c := &Charts{
		starportByThrow: map[int]string{},
		starports:       map[string]*Starport{},
		navalTarget:     map[string]dice.Target{},
		scoutTarget:     map[string]dice.Target{},
		routeTargets:    map[string][]int{},
		tech:            map[string]techRow{},
		labels:          map[string][]labelRow{},
	}

	for _, load := range []func() error{c.loadStarports, c.loadRoutes, c.loadTech, c.loadCharacteristics} {
		if err := load(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// decodeStrict rejects unknown fields, so a data file that has drifted
// from the structs fails at load rather than losing a column silently.
func decodeStrict(raw []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidData, err)
	}

	return nil
}

func read(file string, dst any) error {
	raw, err := dataFS.ReadFile("data/" + file)
	if err != nil {
		return fmt.Errorf("%w: reading %s: %w", ErrInvalidData, file, err)
	}

	if err := decodeStrict(raw, dst); err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	return nil
}

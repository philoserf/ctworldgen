// Package worldgen is the Worlds chapter of Book 3 pp. 1-12: the subsector
// record, the generation event log, the engine that walks star mapping and
// world creation, and replay. The JSON record is the source of truth;
// everything else renders or verifies it (docs/PRD.md).
package worldgen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Version stamps carried by every record (docs/PRD.md, Replay and
// provenance contract).
const (
	// SchemaVersion tracks the shape of the records the engine writes
	// (docs/subsector.schema.json).
	SchemaVersion = "1"
	// EngineVersion changes when generation behaviour changes: rules,
	// dice-stream consumption order, the RNG construction — or the text of
	// any event the log carries. That last one is easy to miss and is not
	// cosmetic: Replay compares whole Event values, Text included, so an
	// edited outcome string diverges every record written before it, and
	// --ignore-provenance does not help (it waives the stamp check, not the
	// event comparison). Bumping here turns that into an honest provenance
	// mismatch naming both versions instead of a bare event divergence.
	EngineVersion = "0.1.0"
	// Ruleset pins the pages every rule was read from. Book 1 pp. 2-3 and
	// p. 8 are named because Book 3 uses the die roll conventions and the
	// hexadecimal digit notation without restating them.
	Ruleset = "Classic Traveller Book 3 pp. 1-12 (with Book 1 pp. 2-3, 8), © 1977 text, FFE reprints"
)

// There is no PolicyVersion. World generation has no choice points, so
// there is no policy to version — see docs/PRD.md, "The architectural
// delta from ctchargen".

// The docs/ERRATA.md readings a record can stamp.
const (
	// ErrataBaseThrows is where the p. 5 base throws sit in the stream.
	ErrataBaseThrows = "E001"
	// ErrataScanOrder is the hex scan order of the occurrence throw.
	ErrataScanOrder = "E002"
	// ErrataLanePairs is which pairs are examined for lanes, and in what order.
	ErrataLanePairs = "E003"
	// ErrataClamp is the clamping of a generated value to its table's range.
	ErrataClamp = "E004"
	// ErrataProfile is the string of digits' order, alphabet, and lack of a separator.
	ErrataProfile = "E005"
)

var errataIDs = [...]string{ErrataBaseThrows, ErrataScanOrder, ErrataLanePairs, ErrataClamp, ErrataProfile}

// ErrataIDs is every reading the engine can stamp, in document order.
// docs/ERRATA.md is held to this list by TestErrataIDsMatchTheDocument.
//
// A fresh slice each call: a caller that appends to what it is handed must
// not reach the package's own copy.
func ErrataIDs() []string { return append([]string(nil), errataIDs[:]...) }

// Errors this package reports.
var (
	// ErrBadInput is a caller-supplied input the book does not allow.
	ErrBadInput = errors.New("invalid input")
	// ErrProvenance is a record whose stamps do not match this build
	// (replay --ignore-provenance waives the match).
	ErrProvenance = errors.New("provenance mismatch")
	// ErrDiverged is a replay that did not reproduce the record.
	ErrDiverged = errors.New("replay diverged")
	// ErrTrailingData is a record file holding more than the one record
	// (UnmarshalRecord).
	ErrTrailingData = errors.New("trailing data after the record")
	// ErrNoDigit is a value the p. 2 notation cannot write. After the
	// clamps of docs/ERRATA.md E004 the engine cannot produce one, so
	// reaching it means a clamp bound and a chart have come apart.
	ErrNoDigit = errors.New("value has no digit in the p. 2 notation")
)

// RNG names the algorithm and seed the dice stream was built from.
type RNG struct {
	Algorithm string `json:"algorithm"`
	Seed      uint64 `json:"seed"`
}

// Inputs is everything the caller supplied; with the seed, it is all a
// replay needs. There is no choice-event channel to reapply: the engine
// takes no decisions (docs/PRD.md).
type Inputs struct {
	Name string `json:"name"`
	// OccurrenceDM is the referee's subsector-wide modifier on the world
	// occurrence throw, "a DM of +1 or -1" (p. 1). Zero is the normal
	// one-half chance.
	OccurrenceDM int `json:"occurrence_dm"`
}

// World is one generated world: the hex it occupies, its starport and
// bases from star mapping (p. 1, p. 5), the seven basic characteristics
// and the technological index from world creation (pp. 2-9), and the
// string of digits they are written as (p. 4; docs/ERRATA.md E005).
type World struct {
	Hex  string `json:"hex"`
	Name string `json:"name"` // p. 12 step 3; the book prints no name table, so this is the referee's
	// Profile is the eight characteristics as a string of digits, in the
	// order the p. 4 Planetary Characteristics box lists them.
	Profile   string `json:"profile"`
	Starport  string `json:"starport"`
	NavalBase bool   `json:"naval_base"`
	ScoutBase bool   `json:"scout_base"`

	Size               int `json:"size"`
	Atmosphere         int `json:"atmosphere"`
	Hydrographics      int `json:"hydrographics"`
	Population         int `json:"population"`
	Government         int `json:"government"`
	LawLevel           int `json:"law_level"`
	TechnologicalIndex int `json:"technological_index"`
}

// Route is one commercial space lane (p. 2): the two hexes it connects and
// the jump distance between them. From is always the lower hex identifier,
// which is the order the pairs were examined in (docs/ERRATA.md E003).
type Route struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Distance int    `json:"distance"`
}

// EventDM is one applied die modification, with the rule that granted it.
type EventDM struct {
	Source string `json:"source"`
	Value  int    `json:"value"`
}

// Event is one entry of the generation record (FR17): a procedure step
// entered, a throw, or an outcome. Fields are populated per kind; Seq
// starts at 1 so a Ref of 0 means "none". There is no choice kind: the
// procedure has no choice points.
type Event struct {
	Seq   int    `json:"seq"`
	Kind  string `json:"kind"` // "step", "throw", "outcome"
	Step  string `json:"step"`
	Label string `json:"label,omitempty"`
	Hex   string `json:"hex,omitempty"` // the world or route the event concerns

	// throw
	Dice    []int     `json:"dice,omitempty"`
	DMs     []EventDM `json:"dms,omitempty"`
	Target  string    `json:"target,omitempty"`
	Total   int       `json:"total,omitempty"`
	Success *bool     `json:"success,omitempty"`

	// outcome
	Text string `json:"text,omitempty"`
	Ref  int    `json:"ref,omitempty"` // seq of the causing throw
}

// Event kinds.
const (
	KindStep    = "step"
	KindThrow   = "throw"
	KindOutcome = "outcome"
)

// Procedure steps, in the order the p. 12 checklist gives them.
const (
	StepOccurrence    = "world-occurrence"
	StepStarport      = "starport-type"
	StepRoutes        = "route-determination"
	StepWorldCreation = "world-creation"
)

// Subsector is the record (FR15). JSON is canonical; the Markdown listing
// is a render of it.
type Subsector struct {
	SchemaVersion string   `json:"schema_version"`
	Ruleset       string   `json:"ruleset"`
	EngineVersion string   `json:"engine_version"`
	RNG           RNG      `json:"rng"`
	Inputs        Inputs   `json:"inputs"`
	Errata        []string `json:"errata"`

	Name   string  `json:"name"`
	Worlds []World `json:"worlds"`
	Routes []Route `json:"routes"`

	Events []Event `json:"events"`
}

// MarshalRecord renders the canonical JSON bytes: two-space indent with a
// trailing newline. Golden fixtures and CLI output both use exactly this.
func (s *Subsector) MarshalRecord() ([]byte, error) {
	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling subsector record: %w", err)
	}

	return append(out, '\n'), nil
}

// UnmarshalRecord parses a subsector record, rejecting unknown fields so a
// record from a newer schema fails loudly rather than silently dropping
// data, and rejecting anything after the record so a file holding two is
// not read as one. This decoder is the strict one — tables.decodeStrict
// reads go:embed data this repository writes, where a second value cannot
// appear — because its input is a file the user names, and because
// replay's verdict is computed from the parsed value and a re-marshal of
// it, never from the bytes on disk: whatever this drops, nothing
// downstream ever sees.
func UnmarshalRecord(data []byte) (*Subsector, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	sub := &Subsector{}
	if err := dec.Decode(sub); err != nil {
		return nil, fmt.Errorf("parsing subsector record: %w", err)
	}

	// More skips whitespace, so MarshalRecord's trailing newline passes.
	if dec.More() {
		return nil, fmt.Errorf("parsing subsector record: %w", ErrTrailingData)
	}

	return sub, nil
}

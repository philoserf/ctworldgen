package starmap_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/philoserf/ctworldgen/starmap"
)

func TestNewRecordCarriesItsProvenance(t *testing.T) {
	t.Parallel()

	record := starmap.New(42, "Aramis", -1, nil)
	if record.SchemaVersion != starmap.SchemaVersion ||
		record.Ruleset != starmap.Ruleset ||
		record.EngineVersion != starmap.EngineVersion {
		t.Errorf("record does not carry its stamps: %+v", record)
	}

	if record.Seed != 42 || record.Name != "Aramis" || record.OccurrenceDM != -1 {
		t.Errorf("record does not carry its seed and inputs: %+v", record)
	}

	if record.RNGAlgorithm != "go-math-rand-v2-pcg" {
		t.Errorf("rng_algorithm is %q", record.RNGAlgorithm)
	}

	if record.Errata == nil || record.Worlds == nil {
		t.Error("errata and worlds should be empty arrays, not null")
	}
}

// TestStampKeepsDocumentOrderAndDoesNotRepeat covers the record's errata
// array: the identifiers of the readings that actually governed it, in
// document order.
func TestStampKeepsDocumentOrderAndDoesNotRepeat(t *testing.T) {
	t.Parallel()

	const governsEveryRecord = "E002"

	record := starmap.New(0, "", 0, nil)
	for _, id := range []string{"E003", governsEveryRecord, "E005", governsEveryRecord, "E001"} {
		record.Stamp(id)
	}

	want := []string{"E001", governsEveryRecord, "E003", "E005"}
	if len(record.Errata) != len(want) {
		t.Fatalf("errata = %v, want %v", record.Errata, want)
	}

	for i, id := range want {
		if record.Errata[i] != id {
			t.Fatalf("errata = %v, want %v", record.Errata, want)
		}
	}
}

// completeRecord is a complete record in the shape the engine writes, as text, so
// that a field can be added to it that no Go type would let through.
const completeRecord = `{
  "schema_version": 1,
  "ruleset": "ct-1977-book3-pp1-12",
  "engine_version": "0",
  "rng_algorithm": "go-math-rand-v2-pcg",
  "seed": 1977,
  "errata": ["E002"],
  "name": "Aramis",
  "occurrence_dm": -1,
  "worlds": [{
    "hex": "0105", "name": "", "starport": "X",
    "naval_base": false, "scout_base": false,
    "size": 0, "atmosphere": 0, "hydrographics": 0,
    "population": 0, "government": 0, "law_level": 0, "tech_index": 0,
    "digits": "X0000000"
  }],
  "routes": []%s
}`

// TestDecodeRejectsUnknownFields is the Go half of the two obligations:
// "additionalProperties": false at every level of record.schema.json,
// and DisallowUnknownFields in Decode. A schema alone rejects nothing at
// read time, so without this half a record from a newer schema is read
// with the field it carries silently dropped.
func TestDecodeRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	decoded, err := starmap.Decode(strings.NewReader(fmt.Sprintf(completeRecord, "")))
	if err != nil {
		t.Fatalf("a record in the shape the engine writes did not decode: %v", err)
	}

	if decoded.Seed != 1977 || len(decoded.Worlds) != 1 || decoded.Worlds[0].Starport != starmap.StarportX {
		t.Errorf("the record decoded to %+v", decoded)
	}

	_, err = starmap.Decode(strings.NewReader(fmt.Sprintf(completeRecord, `,
  "surprise": 1`)))
	if err == nil {
		t.Error("a field the current schema does not define was accepted; it must fail loudly, not be dropped")
	}

	_, err = starmap.Decode(strings.NewReader(`{"schema_version":`))
	if err == nil {
		t.Error("a truncated record was accepted")
	}
}

// TestDecodeRejectsMoreThanOneDocument: a record is one JSON document, so
// a file holding two of them fails loudly rather than decoding the first
// and discarding the rest.
//
// The second document is named in the message. What follows the record is
// the referee's own file, hand-edited or concatenated, and the offending
// token is the thing he has to go and find.
func TestDecodeRejectsMoreThanOneDocument(t *testing.T) {
	t.Parallel()

	one := fmt.Sprintf(completeRecord, "")

	_, err := starmap.Decode(strings.NewReader(one + "\n" + one))
	if !errors.Is(err, starmap.ErrTrailingContent) {
		t.Errorf("two records in one read gave %v; the second was dropped in silence", err)
	}

	if !strings.Contains(fmt.Sprint(err), "{") {
		t.Errorf("the message does not say what was found after the record: %v", err)
	}

	// Trailing whitespace is not content: Marshal writes a newline.
	_, err = starmap.Decode(strings.NewReader(one + "\n"))
	if err != nil {
		t.Errorf("a record with the newline Marshal writes did not decode: %v", err)
	}
}

// TestDecodeDoesNotCallAMalformedTailASecondDocument: reading past the
// record has three outcomes, and collapsing them into one is the failure
// this holds against. Treating anything that is not io.EOF as
// ErrTrailingContent reports `{...}]` -- a stray bracket, a hand-edit
// gone wrong -- as "more than one document in the record read; a record
// is one JSON document", a specific and confident claim about a file
// holding no second document at all, and it drops the decoder's own error
// and the byte offset in it.
//
// So the two that are not a second document are asserted here, and each
// by what it actually is: the syntax error the decoder raised, and the
// reader's own failure, both reachable through errors.As and errors.Is
// rather than flattened into a sentinel that means something else.
func TestDecodeDoesNotCallAMalformedTailASecondDocument(t *testing.T) {
	t.Parallel()

	one := fmt.Sprintf(completeRecord, "")

	_, err := starmap.Decode(strings.NewReader(one + "]"))

	syntax, isSyntax := errors.AsType[*json.SyntaxError](err)
	if !isSyntax {
		t.Fatalf("a malformed tail gave %v; want the decoder's own syntax error", err)
	}

	// The offset is the thing the decoder knows and the sentinel could
	// not carry: where in his file the referee should look.
	if syntax.Offset <= 0 {
		t.Errorf("the syntax error carries no byte offset: %v", syntax)
	}

	if errors.Is(err, starmap.ErrTrailingContent) {
		t.Errorf("a malformed tail was reported as a second document: %v", err)
	}

	// A reader that fails after the record is neither a second document
	// nor malformed JSON, and the caller cannot tell which it was unless
	// the error it raised is the one that comes back.
	_, err = starmap.Decode(io.MultiReader(strings.NewReader(one), iotest.ErrReader(io.ErrUnexpectedEOF)))
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("a reader that failed past the record gave %v; want the reader's own error", err)
	}

	if errors.Is(err, starmap.ErrTrailingContent) {
		t.Errorf("a reader failure was reported as a second document: %v", err)
	}
}

// TestDecodeRejectsAHexOffTheRecordsGrid: a hex is four digits whether it
// names a subsector or a sector, so 0910 parses. It is not on the p. 3
// grid, and a record that says it is on the p. 3 grid is wrong about
// itself.
func TestDecodeRejectsAHexOffTheRecordsGrid(t *testing.T) {
	t.Parallel()

	record := `{"schema_version":1,"ruleset":"ct-1977-book3-pp1-12","engine_version":"1",` +
		`"rng_algorithm":"go-math-rand-v2-pcg","seed":1,"errata":[],"name":"Aramis","occurrence_dm":0,` +
		`"grid":{"columns":8,"rows":10},"worlds":[{"hex":"0910","name":"","starport":"A",` +
		`"naval_base":false,"scout_base":false,"size":0,"atmosphere":0,"hydrographics":0,` +
		`"population":0,"government":0,"law_level":0,"tech_index":0,"digits":"A0000000"}],"routes":[]}`

	_, err := starmap.Decode(strings.NewReader(record))
	if err == nil {
		t.Fatal("Decode accepted a world at 0910 on a record whose grid is 8 columns of 10 rows")
	}

	if !strings.Contains(err.Error(), "0910") {
		t.Errorf("the error does not name the hex that is off the grid: %v", err)
	}
}

// TestDecodeRejectsAnotherToolsProvenance: the three constants
// record.schema.json states are what a referee trusts without checking --
// which pages govern, and which generator drew the dice. Nothing else
// catches a wrong one: every other field still parses and the listing
// still renders, so a record stamped by another tool reads as this
// tool's.
func TestDecodeRejectsAnotherToolsProvenance(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name  string
		field string
		was   string
		now   string
		want  error
	}{
		{"a newer schema", "schema_version", `"schema_version":1`, `"schema_version":99`, starmap.ErrNotThisSchema},
		{
			"another ruleset", "ruleset",
			`"ruleset":"ct-1977-book3-pp1-12"`, `"ruleset":"ct-1981-book3"`, starmap.ErrNotThisRuleset,
		},
		{
			"another generator", "rng_algorithm",
			`"rng_algorithm":"go-math-rand-v2-pcg"`, `"rng_algorithm":"mersenne"`, starmap.ErrNotThisRNG,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			record := strings.Replace(recordWith("", "0101"), testCase.was, testCase.now, 1)
			if record == recordWith("", "0101") {
				t.Fatalf("the test did not change %s, so it proves nothing", testCase.field)
			}

			_, err := starmap.Decode(strings.NewReader(record))
			if !errors.Is(err, testCase.want) {
				t.Fatalf("Decode(%s) = %v, want %v", testCase.field, err, testCase.want)
			}
		})
	}
}

// TestDecodeRejectsARecordMissingARequiredField: the schema lists ten
// required fields and rejects nothing at read time. An empty document once
// decoded to a nameless, seedless, worldless subsector and rendered as
// one.
func TestDecodeRejectsARecordMissingARequiredField(t *testing.T) {
	t.Parallel()

	full := recordWith("", "0101")

	// Each case names the error it must raise. An empty document is the
	// one that does not reach here: it fails on its schema version, which
	// is 0 and not 1, so asserting only that it errored would let this
	// whole table pass while the required-field check did nothing.
	for _, testCase := range []struct {
		name   string
		record string
		want   error
	}{
		{"an empty document", `{}`, starmap.ErrNotThisSchema},
		{
			"no engine_version",
			strings.Replace(full, `"engine_version":"1",`, "", 1), starmap.ErrFieldMissing,
		},
		{"no errata", strings.Replace(full, `"errata":[],`, "", 1), starmap.ErrFieldMissing},
		{"no routes", strings.Replace(full, `,"routes":[]`, "", 1), starmap.ErrFieldMissing},
		{
			"no digits, on a world carrying everything else",
			strings.Replace(full, `,"digits":"A0000000"`, "", 1), starmap.ErrFieldMissing,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.record == full {
				t.Fatal("the test removed nothing, so it proves nothing")
			}

			_, err := starmap.Decode(strings.NewReader(testCase.record))
			if !errors.Is(err, testCase.want) {
				t.Fatalf("Decode(%s) = %v, want %v", testCase.name, err, testCase.want)
			}
		})
	}
}

// TestDecodeRejectsARouteWithNoDistance: a route's distance is the one
// field of its three that the zero value does not refuse on its own -- an
// absent end is the zero Hex and is off every grid. A route with no
// distance printed "| 0105 | 0106 | 0 |", a world joined to itself at a
// range the jump routes table states no target for.
func TestDecodeRejectsARouteWithNoDistance(t *testing.T) {
	t.Parallel()

	record := `{"schema_version":1,"ruleset":"ct-1977-book3-pp1-12","engine_version":"1",` +
		`"rng_algorithm":"go-math-rand-v2-pcg","seed":1,"errata":[],"name":"Aramis","occurrence_dm":0,` +
		`"worlds":[],"routes":[{"from":"0105","to":"0106"}]}`

	_, err := starmap.Decode(strings.NewReader(record))
	if !errors.Is(err, starmap.ErrFieldMissing) {
		t.Fatalf("Decode(a route with no distance) = %v, want %v", err, starmap.ErrFieldMissing)
	}

	if !strings.Contains(err.Error(), "0105") {
		t.Errorf("the error does not name the route that has no distance: %v", err)
	}
}

// TestDecodeRejectsARouteBeyondTheJumpRoutesTable is the maximum twin of
// the test above. record.schema.json gives distance a minimum of 1 and a
// maximum of 4, and both halves have to be checked at read time: a bound
// the schema states binds only the records something validates.
//
// It is not merely tidy. legible walks distance 1 to MaxJump, so a route
// recorded at five parsecs never reached the drawn set, and the route
// table then printed lanesNote -- telling the referee the lane "joins two
// worlds already joined by shorter lanes, which p. 2 says may be ignored
// in the drawing". That is false of a route no rule could have drawn, and
// the document said it about a record it had accepted.
//
// The error is its own, not ErrFieldMissing: five parsecs is present and
// out of range, not absent.
func TestDecodeRejectsARouteBeyondTheJumpRoutesTable(t *testing.T) {
	t.Parallel()

	record := `{"schema_version":1,"ruleset":"ct-1977-book3-pp1-12","engine_version":"1",` +
		`"rng_algorithm":"go-math-rand-v2-pcg","seed":1,"errata":[],"name":"Aramis","occurrence_dm":0,` +
		`"worlds":[],"routes":[{"from":"0105","to":"0605","distance":5}]}`

	_, err := starmap.Decode(strings.NewReader(record))
	if !errors.Is(err, starmap.ErrRouteTooFar) {
		t.Fatalf("Decode(a route of five parsecs) = %v, want %v", err, starmap.ErrRouteTooFar)
	}

	if errors.Is(err, starmap.ErrFieldMissing) {
		t.Errorf("a distance that is present and out of range was reported as a missing field: %v", err)
	}

	// The referee has to find the route in his own file, so the message
	// names the distance and the ends rather than only the rule.
	for _, want := range []string{"5", "0105", "0605"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not name %q: %v", want, err)
		}
	}

	// The last distance the table does state is accepted, so the check
	// bounds rather than forbids.
	atTheLimit := strings.Replace(record, `"distance":5`, `"distance":4`, 1)

	_, err = starmap.Decode(strings.NewReader(atTheLimit))
	if err != nil {
		t.Errorf("a route at the table's own maximum was refused: %v", err)
	}
}

// TestDecodeRejectsAWorldWithNoStarport: the map marks a world's hex with
// the letter of its starport (p. 1). A world with no starport key decodes
// to Starport(0), which the map draws as the eleven characters
// "Starport(0)" -- wider than a hex's slot, so every hex to the right of
// it on that line shifts and the drawn grid is no longer the p. 3 grid.
// gridLine's own comment reasons about "a Starport the schema would have
// rejected", and this is what does the rejecting.
func TestDecodeRejectsAWorldWithNoStarport(t *testing.T) {
	t.Parallel()

	full := recordWith("", "0101")

	record := strings.Replace(full, `"starport":"A",`, "", 1)
	if record == full {
		t.Fatal("the test removed no starport, so it proves nothing")
	}

	_, err := starmap.Decode(strings.NewReader(record))
	if !errors.Is(err, starmap.ErrFieldMissing) {
		t.Fatalf("Decode(a world with no starport) = %v, want %v", err, starmap.ErrFieldMissing)
	}

	if !strings.Contains(err.Error(), "0101") {
		t.Errorf("the error does not name the world that has no starport: %v", err)
	}
}

// TestAnEmptyArrayIsNotAMissingOne: the distinction the required-field
// check rests on. An empty subsector is a legal result and writes
// "worlds": [], which must decode; a record with no worlds key at all is
// missing a field the schema requires.
func TestAnEmptyArrayIsNotAMissingOne(t *testing.T) {
	t.Parallel()

	empty := `{"schema_version":1,"ruleset":"ct-1977-book3-pp1-12","engine_version":"1",` +
		`"rng_algorithm":"go-math-rand-v2-pcg","seed":0,"errata":[],"name":"","occurrence_dm":0,` +
		`"worlds":[],"routes":[]}`

	record, err := starmap.Decode(strings.NewReader(empty))
	if err != nil {
		t.Fatalf("an empty subsector did not decode: %v", err)
	}

	if len(record.Worlds) != 0 {
		t.Errorf("an empty subsector decoded with %d worlds", len(record.Worlds))
	}

	// The same record with the key taken away is a different thing.
	_, err = starmap.Decode(strings.NewReader(strings.Replace(empty, `"worlds":[],`, "", 1)))
	if !errors.Is(err, starmap.ErrFieldMissing) {
		t.Errorf("Decode(no worlds key) = %v, want %v", err, starmap.ErrFieldMissing)
	}
}

// TestNotesRoundTripAndDoNotLoosenTheRecord: the record is the referee's
// notebook page, so it has a place for him to write (issue 1 #6). Both
// halves are asserted together, because the second is the whole reason the
// first is shaped as a named field rather than an escape hatch: `notes` is
// now a field the record defines, and every key it does not define is
// refused exactly as before.
func TestNotesRoundTripAndDoNotLoosenTheRecord(t *testing.T) {
	t.Parallel()

	const (
		onTheMap   = "The rift campaign. Players start at 0602."
		onTheWorld = "dust storms; ask about the yard"
	)

	written := strings.Replace(recordWith("", "0101"),
		`"name":"Aramis",`, `"name":"Aramis","notes":"`+onTheMap+`",`, 1)

	written = strings.Replace(written,
		`"name":"","starport":"A"`, `"name":"","notes":"`+onTheWorld+`","starport":"A"`, 1)

	record, err := starmap.Decode(strings.NewReader(written))
	if err != nil {
		t.Fatalf("a record carrying the referee's notes did not decode: %v", err)
	}

	if record.Notes != onTheMap {
		t.Errorf("the map's note decoded as %q", record.Notes)
	}

	if record.Worlds[0].Notes != onTheWorld {
		t.Errorf("the world's note decoded as %q", record.Worlds[0].Notes)
	}

	// Re-marshalled, both survive: what he writes is still there when the
	// record is written back out.
	encoded, err := starmap.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{onTheMap, onTheWorld} {
		if !strings.Contains(string(encoded), want) {
			t.Errorf("re-marshalling lost %q", want)
		}
	}

	// And the strictness that makes the record worth trusting is untouched.
	unknown := strings.Replace(recordWith("", "0101"), `"name":"","starport":"A"`,
		`"name":"","nonesuch":"x","starport":"A"`, 1)

	_, err = starmap.Decode(strings.NewReader(unknown))
	if err == nil {
		t.Error("a key the record does not define was accepted; adding notes must not have loosened this")
	}
}

// TestARecordWithNoNotesIsUnchanged holds the omitempty guarantee that
// made this an additive change: a record with nothing written in it must
// serialise exactly as it did before the field existed, or every golden
// moves and the field is not additive at all.
func TestARecordWithNoNotesIsUnchanged(t *testing.T) {
	t.Parallel()

	record := starmap.New(1977, "Aramis", -1, nil)

	encoded, err := starmap.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(encoded), "notes") {
		t.Errorf("a record with no notes wrote the key anyway:\n%s", encoded)
	}
}

// recordWith builds a one-world record with whatever grid clause is given,
// so a test can hand Decode a record that has no grid at all.
func recordWith(gridClause, hex string) string {
	return `{"schema_version":1,"ruleset":"ct-1977-book3-pp1-12","engine_version":"1",` +
		`"rng_algorithm":"go-math-rand-v2-pcg","seed":1,"errata":[],"name":"Aramis","occurrence_dm":0,` +
		gridClause + `"worlds":[{"hex":"` + hex + `","name":"","starport":"A",` +
		`"naval_base":false,"scout_base":false,"size":0,"atmosphere":0,"hydrographics":0,` +
		`"population":0,"government":0,"law_level":0,"tech_index":0,"digits":"A0000000"}],"routes":[]}`
}

// TestARecordWithNoGridIsASubsector: grids were added to the record when
// sectors were (ERRATA E006), and every record written before that is a
// subsector on the p. 3 grid. Those files still read, which is why the
// schema leaves `grid` optional -- this is the half of that promise the
// schema cannot keep on its own.
func TestARecordWithNoGridIsASubsector(t *testing.T) {
	t.Parallel()

	record, err := starmap.Decode(strings.NewReader(recordWith("", "0810")))
	if err != nil {
		t.Fatal(err)
	}

	if record.Grid != starmap.PageThreeGrid() {
		t.Errorf("a record with no grid decoded onto a %dx%d grid, want the p. 3 grid",
			record.Grid.Columns, record.Grid.Rows)
	}
}

// TestDecodeRejectsAGridTheSchemaDoesNotName: the schema names two grids,
// and rejecting a third is two obligations -- the schema, and Decode. A
// schema alone rejects nothing at read time.
func TestDecodeRejectsAGridTheSchemaDoesNotName(t *testing.T) {
	t.Parallel()

	for _, clause := range []string{
		`"grid":{"columns":9,"rows":10},`,
		`"grid":{"columns":8,"rows":11},`,
		`"grid":{"columns":32,"rows":10},`,
	} {
		_, err := starmap.Decode(strings.NewReader(recordWith(clause, "0101")))
		if err == nil {
			t.Errorf("Decode accepted %s", clause)
		}
	}
}

// TestDecodeRejectsARouteEndOffTheRecordsGrid: a route's two ends are hexes
// of the record's grid like any other. Bounding Hex by the sector grid
// made ParseHex accept 0910, so nothing but this refuses a p. 3 record
// whose route reaches one -- checking the worlds alone leaves the ends
// unchecked entirely.
func TestDecodeRejectsARouteEndOffTheRecordsGrid(t *testing.T) {
	t.Parallel()

	for _, route := range []string{
		`{"from":"0910","to":"0810","distance":1}`,
		`{"from":"0810","to":"0910","distance":1}`,
	} {
		record := `{"schema_version":1,"ruleset":"ct-1977-book3-pp1-12","engine_version":"1",` +
			`"rng_algorithm":"go-math-rand-v2-pcg","seed":1,"errata":[],"name":"Aramis","occurrence_dm":0,` +
			`"grid":{"columns":8,"rows":10},"worlds":[],"routes":[` + route + `]}`

		_, err := starmap.Decode(strings.NewReader(record))
		if err == nil {
			t.Errorf("Decode accepted the route %s on a record whose grid is 8 columns of 10 rows", route)

			continue
		}

		if !strings.Contains(err.Error(), "0910") {
			t.Errorf("the error does not name the hex that is off the grid: %v", err)
		}
	}
}

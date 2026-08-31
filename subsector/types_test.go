package subsector_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/subsector"
)

func TestStarportParsesAndMarshals(t *testing.T) {
	t.Parallel()

	for _, printed := range []string{"A", "B", "C", "D", "E", "X"} {
		port, err := subsector.ParseStarport(printed)
		if err != nil {
			t.Fatal(err)
		}

		if port.String() != printed {
			t.Errorf("%q parsed to %q", printed, port)
		}

		encoded, err := json.Marshal(port)
		if err != nil {
			t.Fatal(err)
		}

		if string(encoded) != `"`+printed+`"` {
			t.Errorf("%q marshaled to %s", printed, encoded)
		}

		var back subsector.Starport

		err = json.Unmarshal(encoded, &back)
		if err != nil {
			t.Fatal(err)
		}

		if back != port {
			t.Errorf("%q round-tripped to %q", printed, back)
		}
	}
}

func TestStarportRejectsWhatTheBookDoesNotPrint(t *testing.T) {
	t.Parallel()

	for _, bad := range []string{"", "a", "F", "AB", "0", " "} {
		_, err := subsector.ParseStarport(bad)
		if err == nil {
			t.Errorf("ParseStarport(%q) succeeded", bad)
		}
	}

	_, err := json.Marshal(subsector.Starport(0))
	if err == nil {
		t.Error("marshaling the zero Starport succeeded")
	}

	var port subsector.Starport

	err = json.Unmarshal([]byte(`"a"`), &port)
	if err == nil {
		t.Error(`unmarshaling "a" as a starport succeeded`)
	}

	err = json.Unmarshal([]byte(`3`), &port)
	if err == nil {
		t.Error("unmarshaling a number as a starport succeeded")
	}

	if got := subsector.Starport('Q').String(); got != "Starport(81)" {
		t.Errorf("Starport('Q').String() = %q", got)
	}
}

// TestDigitAlphabet transcribes the notation a second time: Book 1 p. 8's
// hexadecimal extended by Book 3 p. 2's letter set, omitting O and I.
func TestDigitAlphabet(t *testing.T) {
	t.Parallel()

	// The values ERRATA E005 names explicitly, plus the ends.
	cases := map[int]string{
		0: "0", 9: "9",
		10: "A", 15: "F", 16: "G", 17: "H", 18: "J", 19: "K", 20: "L",
		33: "Z",
	}
	for value, want := range cases {
		digit, err := subsector.NewDigit(value)
		if err != nil {
			t.Fatal(err)
		}

		if digit.String() != want {
			t.Errorf("value %d is written %q, want %q", value, digit, want)
		}

		if digit.Value() != value {
			t.Errorf("digit %q reads back as %d, want %d", digit, digit.Value(), value)
		}
	}

	for _, omitted := range []string{"I", "O"} {
		_, err := subsector.ParseDigit(omitted)
		if err == nil {
			t.Errorf("%q is in the alphabet; p. 2 omits it as confusable with a number", omitted)
		}
	}
}

// TestDigitReaches20 is R15's requirement: law level's maximum under R14
// is 20, and the notation must be able to write it.
func TestDigitReaches20(t *testing.T) {
	t.Parallel()

	if subsector.MaxDigit < 20 {
		t.Fatalf("the alphabet stops at %d; law level reaches 20", subsector.MaxDigit)
	}

	_, err := subsector.NewDigit(subsector.MaxDigit + 1)
	if err == nil {
		t.Error("a value beyond the alphabet was given a digit")
	}

	_, err = subsector.NewDigit(-1)
	if err == nil {
		t.Error("a negative value was given a digit; the notation has no character for one")
	}
}

func TestDigitMarshals(t *testing.T) {
	t.Parallel()

	for value := 0; value <= subsector.MaxDigit; value++ {
		digit, err := subsector.NewDigit(value)
		if err != nil {
			t.Fatal(err)
		}

		encoded, err := json.Marshal(digit)
		if err != nil {
			t.Fatal(err)
		}

		var back subsector.Digit

		err = json.Unmarshal(encoded, &back)
		if err != nil {
			t.Fatal(err)
		}

		if back != digit {
			t.Errorf("value %d round-tripped from %q to %q", value, digit, back)
		}
	}

	_, err := json.Marshal(subsector.Digit('I'))
	if err == nil {
		t.Error("marshaling an out-of-alphabet digit succeeded")
	}

	var digit subsector.Digit

	err = json.Unmarshal([]byte(`"II"`), &digit)
	if err == nil {
		t.Error("unmarshaling a two-character digit succeeded")
	}

	err = json.Unmarshal([]byte(`5`), &digit)
	if err == nil {
		t.Error("unmarshaling a number as a digit succeeded")
	}

	if got := subsector.Digit('I').String(); got != "Digit(73)" {
		t.Errorf("Digit('I').String() = %q", got)
	}
}

// TestCharacteristicsAreTheSevenOfPage4 pins the order of the p. 4
// Planetary Characteristics box, which is also the order of the string of
// digits (ERRATA E005) and of the p. 12 checklist steps 2.B through 2.H.
func TestCharacteristicsAreTheSevenOfPage4(t *testing.T) {
	t.Parallel()

	want := []string{
		"size", "atmosphere", "hydrographics", "population",
		"government", "law_level", "tech_index",
	}

	got := subsector.Characteristics()
	if len(got) != len(want) {
		t.Fatalf("there are %d characteristics, want %d", len(got), len(want))
	}

	for i, c := range got {
		if c.String() != want[i] {
			t.Errorf("characteristic %d is %q, want %q", i, c, want[i])
		}
	}
}

// TestCharacteristicRoundTripsThroughJSON covers the marshalers.
func TestCharacteristicRoundTripsThroughJSON(t *testing.T) {
	t.Parallel()

	for _, characteristic := range subsector.Characteristics() {
		encoded, err := json.Marshal(characteristic)
		if err != nil {
			t.Fatal(err)
		}

		var back subsector.Characteristic

		err = json.Unmarshal(encoded, &back)
		if err != nil {
			t.Fatal(err)
		}

		if back != characteristic {
			t.Errorf("%q round-tripped to %q", characteristic, back)
		}
	}
}

// TestCharacteristicRejectsWhatIsNotOne: starport is a table lookup and
// never arithmetic, so it is not among the seven.
func TestCharacteristicRejectsWhatIsNotOne(t *testing.T) {
	t.Parallel()

	if got := subsector.Characteristic(99).String(); got != "Characteristic(99)" {
		t.Errorf("Characteristic(99).String() = %q", got)
	}

	_, err := json.Marshal(subsector.Characteristic(99))
	if err == nil {
		t.Error("marshaling an unknown characteristic succeeded")
	}

	var characteristic subsector.Characteristic

	err = json.Unmarshal([]byte(`"starport"`), &characteristic)
	if err == nil {
		t.Error("starport unmarshaled as a characteristic; it is a lookup, never arithmetic")
	}

	err = json.Unmarshal([]byte(`1`), &characteristic)
	if err == nil {
		t.Error("unmarshaling a number as a characteristic succeeded")
	}
}

func TestNewRecordCarriesItsProvenance(t *testing.T) {
	t.Parallel()

	record := subsector.New(42, "Aramis", -1)
	if record.SchemaVersion != subsector.SchemaVersion ||
		record.Ruleset != subsector.Ruleset ||
		record.EngineVersion != subsector.EngineVersion {
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

	record := subsector.New(0, "", 0)
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
  "worlds": [{"hex": "0105", "name": "", "starport": "X"}]%s
}`

// TestDecodeRejectsUnknownFields is the Go half of the two obligations:
// "additionalProperties": false at every level of subsector.schema.json,
// and DisallowUnknownFields in Decode. A schema alone rejects nothing at
// read time, so without this half a record from a newer schema is read
// with the field it carries silently dropped.
func TestDecodeRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	decoded, err := subsector.Decode(strings.NewReader(fmt.Sprintf(completeRecord, "")))
	if err != nil {
		t.Fatalf("a record in the shape the engine writes did not decode: %v", err)
	}

	if decoded.Seed != 1977 || len(decoded.Worlds) != 1 || decoded.Worlds[0].Starport != subsector.StarportX {
		t.Errorf("the record decoded to %+v", decoded)
	}

	_, err = subsector.Decode(strings.NewReader(fmt.Sprintf(completeRecord, `,
  "surprise": 1`)))
	if err == nil {
		t.Error("a field the current schema does not define was accepted; it must fail loudly, not be dropped")
	}

	_, err = subsector.Decode(strings.NewReader(`{"schema_version":`))
	if err == nil {
		t.Error("a truncated record was accepted")
	}
}

// TestDecodeRejectsMoreThanOneDocument: a record is one JSON document, so
// a JSONL batch handed to a reader of records fails loudly rather than
// decoding its first member and discarding the rest.
func TestDecodeRejectsMoreThanOneDocument(t *testing.T) {
	t.Parallel()

	one := fmt.Sprintf(completeRecord, "")

	_, err := subsector.Decode(strings.NewReader(one + "\n" + one))
	if !errors.Is(err, subsector.ErrTrailingContent) {
		t.Errorf("two records in one read gave %v; the second was dropped in silence", err)
	}

	_, err = subsector.Decode(strings.NewReader(one + "\nsurprise"))
	if !errors.Is(err, subsector.ErrTrailingContent) {
		t.Errorf("content after the record gave %v", err)
	}

	// Trailing whitespace is not content: Marshal writes a newline.
	_, err = subsector.Decode(strings.NewReader(one + "\n"))
	if err != nil {
		t.Errorf("a record with the newline Marshal writes did not decode: %v", err)
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

	_, err := subsector.Decode(strings.NewReader(record))
	if err == nil {
		t.Fatal("Decode accepted a world at 0910 on a record whose grid is 8 columns of 10 rows")
	}

	if !strings.Contains(err.Error(), "0910") {
		t.Errorf("the error does not name the hex that is off the grid: %v", err)
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

	record, err := subsector.Decode(strings.NewReader(recordWith("", "0810")))
	if err != nil {
		t.Fatal(err)
	}

	if record.Grid != subsector.PageThreeGrid() {
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
		_, err := subsector.Decode(strings.NewReader(recordWith(clause, "0101")))
		if err == nil {
			t.Errorf("Decode accepted %s", clause)
		}
	}
}

// TestDecodeRejectsALaneEndOffTheRecordsGrid: a lane's two ends are hexes
// of the record's grid like any other. Bounding Hex by the sector grid
// made ParseHex accept 0910, so nothing but this refuses a p. 3 record
// whose lane reaches one -- checking the worlds alone leaves the ends
// unchecked entirely.
func TestDecodeRejectsALaneEndOffTheRecordsGrid(t *testing.T) {
	t.Parallel()

	for _, lane := range []string{
		`{"from":"0910","to":"0810","distance":1}`,
		`{"from":"0810","to":"0910","distance":1}`,
	} {
		record := `{"schema_version":1,"ruleset":"ct-1977-book3-pp1-12","engine_version":"1",` +
			`"rng_algorithm":"go-math-rand-v2-pcg","seed":1,"errata":[],"name":"Aramis","occurrence_dm":0,` +
			`"grid":{"columns":8,"rows":10},"worlds":[],"routes":[` + lane + `]}`

		_, err := subsector.Decode(strings.NewReader(record))
		if err == nil {
			t.Errorf("Decode accepted the lane %s on a record whose grid is 8 columns of 10 rows", lane)

			continue
		}

		if !strings.Contains(err.Error(), "0910") {
			t.Errorf("the error does not name the hex that is off the grid: %v", err)
		}
	}
}

package subsector_test

import (
	"encoding/json"
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

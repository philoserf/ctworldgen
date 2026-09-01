package starmap_test

import (
	"encoding/json"
	"testing"

	"github.com/philoserf/ctworldgen/starmap"
)

// TestCharacteristicsAreTheSevenOfPage4 pins the order of the p. 4
// Planetary Characteristics box, which is also the order of the string of
// digits (ERRATA E005) and of the p. 12 checklist steps 2.B through 2.H.
func TestCharacteristicsAreTheSevenOfPage4(t *testing.T) {
	t.Parallel()

	want := []string{
		"size", "atmosphere", "hydrographics", "population",
		"government", "law_level", "tech_index",
	}

	got := starmap.Characteristics()
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

	for _, characteristic := range starmap.Characteristics() {
		encoded, err := json.Marshal(characteristic)
		if err != nil {
			t.Fatal(err)
		}

		var back starmap.Characteristic

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

	if got := starmap.Characteristic(99).String(); got != "Characteristic(99)" {
		t.Errorf("Characteristic(99).String() = %q", got)
	}

	_, err := json.Marshal(starmap.Characteristic(99))
	if err == nil {
		t.Error("marshaling an unknown characteristic succeeded")
	}

	var characteristic starmap.Characteristic

	err = json.Unmarshal([]byte(`"starport"`), &characteristic)
	if err == nil {
		t.Error("starport unmarshaled as a characteristic; it is a lookup, never arithmetic")
	}

	err = json.Unmarshal([]byte(`1`), &characteristic)
	if err == nil {
		t.Error("unmarshaling a number as a characteristic succeeded")
	}
}

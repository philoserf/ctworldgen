package starmap_test

import (
	"encoding/json"
	"testing"

	"github.com/philoserf/ctworldgen/starmap"
)

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
		digit, err := starmap.NewDigit(value)
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
		_, err := starmap.ParseDigit(omitted)
		if err == nil {
			t.Errorf("%q is in the alphabet; p. 2 omits it as confusable with a number", omitted)
		}
	}
}

// TestDigitReaches20 is R15's requirement: law level's maximum under R14
// is 20, and the notation must be able to write it.
func TestDigitReaches20(t *testing.T) {
	t.Parallel()

	if starmap.MaxDigit < 20 {
		t.Fatalf("the alphabet stops at %d; law level reaches 20", starmap.MaxDigit)
	}

	_, err := starmap.NewDigit(starmap.MaxDigit + 1)
	if err == nil {
		t.Error("a value beyond the alphabet was given a digit")
	}

	_, err = starmap.NewDigit(-1)
	if err == nil {
		t.Error("a negative value was given a digit; the notation has no character for one")
	}
}

func TestDigitMarshals(t *testing.T) {
	t.Parallel()

	for value := 0; value <= starmap.MaxDigit; value++ {
		digit, err := starmap.NewDigit(value)
		if err != nil {
			t.Fatal(err)
		}

		encoded, err := json.Marshal(digit)
		if err != nil {
			t.Fatal(err)
		}

		var back starmap.Digit

		err = json.Unmarshal(encoded, &back)
		if err != nil {
			t.Fatal(err)
		}

		if back != digit {
			t.Errorf("value %d round-tripped from %q to %q", value, digit, back)
		}
	}

	_, err := json.Marshal(starmap.Digit('I'))
	if err == nil {
		t.Error("marshaling an out-of-alphabet digit succeeded")
	}

	var digit starmap.Digit

	err = json.Unmarshal([]byte(`"II"`), &digit)
	if err == nil {
		t.Error("unmarshaling a two-character digit succeeded")
	}

	err = json.Unmarshal([]byte(`5`), &digit)
	if err == nil {
		t.Error("unmarshaling a number as a digit succeeded")
	}

	if got := starmap.Digit('I').String(); got != "Digit(73)" {
		t.Errorf("Digit('I').String() = %q", got)
	}
}

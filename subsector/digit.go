package subsector

import (
	"encoding/json"
	"fmt"
	"strings"
)

// alphabet is the notation a characteristic is written in: Book 1 p. 8's
// hexadecimal -- "the digits 0 through 9 are represented by common arabic
// numbers; the digits 10 through 15 are represented by the letters A
// through F" -- extended by Book 3 p. 2's letter set, "single digits (the
// numbers 0 through 9) and letters (A through Z, omitting O and I as they
// may be confused with numbers)".
//
// So 10 is A, 15 is F, 16 is G, 17 is H, 18 is J -- I having been skipped
// -- 19 is K and 20 is L (ERRATA E005).
const alphabet = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// MaxDigit is the largest value the alphabet can write.
//
// The alphabet is not bounded at law level's maximum of 20. Its range is
// the notation's range, and a bound on a characteristic's value is a rule
// that belongs on a page with a cite, not in a type: R15 requires only
// that the notation *reach* 20, which is a floor on it and not a cap.
const MaxDigit = len(alphabet) - 1

// Digit is one character of that alphabet, the notation a characteristic
// is recorded in.
type Digit byte

// NewDigit returns the digit that writes a value.
func NewDigit(value int) (Digit, error) {
	if value < 0 || value > MaxDigit {
		return 0, fmt.Errorf("%w: %d, and the alphabet runs 0 to %d", ErrNoDigitForValue, value, MaxDigit)
	}

	return Digit(alphabet[value]), nil
}

// ParseDigit reads a single character of the alphabet.
func ParseDigit(text string) (Digit, error) {
	if len(text) != 1 {
		return 0, fmt.Errorf("digit %w: %q", ErrNotOneCharacter, text)
	}

	d := Digit(text[0])
	if !d.Valid() {
		return 0, fmt.Errorf("%w: %q", ErrNotADigit, text)
	}

	return d, nil
}

// Valid reports whether the character is in the alphabet.
func (d Digit) Valid() bool { return strings.IndexByte(alphabet, byte(d)) >= 0 }

// Value returns the number the digit writes, or -1 if it writes none.
func (d Digit) Value() int { return strings.IndexByte(alphabet, byte(d)) }

// String returns the single character.
func (d Digit) String() string {
	if !d.Valid() {
		return fmt.Sprintf("Digit(%d)", byte(d))
	}

	return string(rune(d))
}

// MarshalJSON writes the single character.
func (d Digit) MarshalJSON() ([]byte, error) {
	if !d.Valid() {
		return nil, fmt.Errorf("%w: %d", ErrNotADigit, byte(d))
	}

	b, err := json.Marshal(string(rune(d)))
	if err != nil {
		return nil, fmt.Errorf("marshaling digit %s: %w", d, err)
	}

	return b, nil
}

// UnmarshalJSON reads the single character.
func (d *Digit) UnmarshalJSON(b []byte) error {
	var text string

	err := json.Unmarshal(b, &text)
	if err != nil {
		return fmt.Errorf("reading a digit: %w", err)
	}

	parsed, err := ParseDigit(text)
	if err != nil {
		return err
	}

	*d = parsed

	return nil
}

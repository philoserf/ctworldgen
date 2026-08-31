package subsector

import (
	"encoding/json"
	"fmt"
)

// Starport is a starport type. It marshals to the single character the
// book prints: A through E, the best through the worst, plus X for no
// starport (Book 3 p. 4).
//
// The zero value is not a starport, and marshaling it is an error.
type Starport byte

// The starport types of Book 3 p. 4.
const (
	StarportA Starport = 'A'
	StarportB Starport = 'B'
	StarportC Starport = 'C'
	StarportD Starport = 'D'
	StarportE Starport = 'E'
	StarportX Starport = 'X'
)

// Starports lists the types in the order the p. 5 starport chart prints
// them.
func Starports() []Starport {
	return []Starport{StarportA, StarportB, StarportC, StarportD, StarportE, StarportX}
}

// ParseStarport reads a single starport character. It is deliberately
// strict: "a" and "" are not starports.
func ParseStarport(text string) (Starport, error) {
	if len(text) != 1 {
		return 0, fmt.Errorf("starport %w: %q", ErrNotOneCharacter, text)
	}

	p := Starport(text[0])
	if !p.Valid() {
		return 0, fmt.Errorf("%w: %q", ErrNotAStarport, text)
	}

	return p, nil
}

// Valid reports whether the type is one the book prints.
func (p Starport) Valid() bool {
	switch p {
	case StarportA, StarportB, StarportC, StarportD, StarportE, StarportX:
		return true
	default:
		return false
	}
}

// String returns the single character the book prints.
func (p Starport) String() string {
	if !p.Valid() {
		return fmt.Sprintf("Starport(%d)", byte(p))
	}

	return string(rune(p))
}

// MarshalJSON writes the single character the book prints.
func (p Starport) MarshalJSON() ([]byte, error) {
	if !p.Valid() {
		return nil, fmt.Errorf("%w: %d", ErrNotAStarport, byte(p))
	}

	b, err := json.Marshal(string(rune(p)))
	if err != nil {
		return nil, fmt.Errorf("marshaling starport %s: %w", p, err)
	}

	return b, nil
}

// UnmarshalJSON reads the single character the book prints.
func (p *Starport) UnmarshalJSON(b []byte) error {
	var text string

	err := json.Unmarshal(b, &text)
	if err != nil {
		return fmt.Errorf("reading a starport: %w", err)
	}

	parsed, err := ParseStarport(text)
	if err != nil {
		return err
	}

	*p = parsed

	return nil
}

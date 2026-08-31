package subsector

import "errors"

// The rules these errors state are the rules of the held pages, so each
// carries its cite here rather than in the message built at the point of
// failure. Callers wrap with %w and add the value that offended.
var (
	// ErrNotACharacteristic is the p. 4 Planetary Characteristics box.
	ErrNotACharacteristic = errors.New("not one of the seven characteristics of Book 3 p. 4")

	// ErrNotADigit and ErrNoDigitForValue are the notation of Book 1 p. 8
	// extended by Book 3 p. 2 (ERRATA E005).
	ErrNotADigit       = errors.New("not in the alphabet of Book 1 p. 8 and Book 3 p. 2")
	ErrNoDigitForValue = errors.New("no digit writes this value")

	// ErrOffGrid and ErrNotAHex are the sub-sector hex grid of Book 3 p. 3.
	ErrOffGrid = errors.New("outside the subsector grid of Book 3 p. 3")
	ErrNotAHex = errors.New("a hex is four digits")

	// ErrNotAStarport is the starport types of Book 3 p. 4.
	ErrNotAStarport = errors.New("not one of A B C D E X (Book 3 p. 4)")

	// ErrTrailingContent is the record's own shape: one JSON document.
	ErrTrailingContent = errors.New("more than one document in the record read; a record is one JSON document")

	// ErrNotOneCharacter is the shape the record prints these in.
	ErrNotOneCharacter = errors.New("not a single character")
)

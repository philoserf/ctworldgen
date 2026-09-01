package render

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestClipCutsWholeCharacters is an internal test because clip's rule
// cannot be reached from the booklet at a width the test chooses, and the
// width is the whole of the question. That is also why this file is not
// part of pdf_test.go, where the rest of pdf.go's tests live: a package
// clause is per file, and that one is package render_test.
//
// The rule: a name trimmed to its column is trimmed a character at a
// time. Cutting inside a multi-byte character leaves invalid UTF-8, which
// encode reads as the replacement rune and draws as a question mark --
// silently destroying the character the trim stopped on, which is the one
// thing encode promises it will not do.
//
// It is swept across every width rather than checked at one, because at
// one width the bug is usually invisible. The first attempt at this used
// a run of e-acute, and it passed against the byte-slicing version: in
// Helvetica an e-acute and a question mark are both 556 units wide, so
// the corrupted intermediate is never the first candidate to fit. The
// sweep does not depend on that arithmetic working out.
func TestClipCutsWholeCharacters(t *testing.T) {
	t.Parallel()

	book := &booklet{
		pdf: newPDF(), charts: nil, record: nil, drawn: nil, names: nil,
		latin: windows1252(), y: pageMargin,
	}

	book.pdf.AddPage()
	book.pdf.SetFont("Helvetica", "", rosterSize)

	// Every character two bytes wide, so every cut that is not on a
	// character boundary is a cut inside one. AE-ligature is a thousand
	// units to a question mark's five hundred and fifty-six, so the two
	// candidates a byte-wise trim offers are never the same width and the
	// wrong one is reachable.
	name := strings.Repeat("Æ", 20)

	for width := 1; width <= int(book.width(name)); width++ {
		trimmed := book.clip(name, float64(width))

		if !utf8.ValidString(trimmed) {
			t.Fatalf("clipping to %dpt cut inside a character: %q", width, trimmed)
		}

		if strings.ContainsRune(trimmed, utf8.RuneError) {
			t.Fatalf("clipping to %dpt produced a replacement rune: %q", width, trimmed)
		}
	}
}

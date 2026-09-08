package render

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/philoserf/ctworldgen/starmap"
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
		alphabet: drawnAlphabet(), y: pageMargin,
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

// TestMemberSeedRefusesAnIndexThatIsNotAMembers holds the bound the
// conversion in memberSeed relies on. A sector has sixteen members and
// nothing asks for a seventeenth, but the guard is what makes converting
// the index to the seed's width safe rather than merely true today.
func TestMemberSeedRefusesAnIndexThatIsNotAMembers(t *testing.T) {
	t.Parallel()

	for _, index := range []int{-1, starmap.Members, starmap.Members + 1} {
		if got := memberSeed(100, index); got != 100 {
			t.Errorf("memberSeed(100, %d) = %d; an index that is not a member's returns the base", index, got)
		}
	}

	if got := memberSeed(100, starmap.Members-1); got != 100+starmap.Members-1 {
		t.Errorf("memberSeed(100, %d) = %d; want %d", starmap.Members-1, got, 100+starmap.Members-1)
	}
}

// TestSplitMeasuresTheCharacterItWillDraw is an internal test for the same
// reason clip's is: the width is the whole of the question, and it cannot
// be chosen from outside the package.
//
// split crosses into fpdf by a different door than every other string
// does. Drawing hands fpdf the Windows-1252 bytes encode makes; SplitText
// indexes the same 256-entry width table by rune. The two disagreed, and
// a rune above 255 -- the curly apostrophe macOS types for every
// apostrophe -- indexed off the end of that table and panicked (issue
// #17).
//
// So two things are asserted, and the second is what a guard that merely
// stopped the panic would fail. Helvetica draws an em-dash a thousand
// units wide and a question mark five hundred and fifty-six, so a
// paragraph of em-dashes measured as question marks wraps into too few
// lines: it would be drawn running off the column at its right.
func TestSplitMeasuresTheCharacterItWillDraw(t *testing.T) {
	t.Parallel()

	book := &booklet{
		pdf: newPDF(), charts: nil, record: nil, drawn: nil, names: nil,
		alphabet: drawnAlphabet(), y: pageMargin,
	}

	book.pdf.AddPage()
	book.pdf.SetFont("Helvetica", "", bodySize)

	// No spaces in either, so both wrap on width alone and the two counts
	// are comparable. Long enough that the wider character costs whole
	// lines: sixty of either still fits on one.
	const runLength = 200

	dashes := book.split(strings.Repeat("—", runLength), contentWidth)
	marks := book.split(strings.Repeat("?", runLength), contentWidth)

	if strings.Join(dashes, "") != strings.Repeat("—", runLength) {
		t.Errorf("wrapping did not return the paragraph it was given: %q", dashes)
	}

	if len(dashes) <= len(marks) {
		t.Errorf("%d em-dashes wrapped into %d lines and %d question marks into %d; "+
			"the em-dash is the wider character, so it was measured as something else",
			runLength, len(dashes), runLength, len(marks))
	}

	// The apostrophe of the crash itself. A line of them must come back
	// as they went in, rather than as a panic or a row of question marks.
	quoted := book.split(strings.Repeat("’", runLength), contentWidth)

	if strings.Join(quoted, "") != strings.Repeat("’", runLength) {
		t.Errorf("wrapping did not carry the curly apostrophe: %q", quoted)
	}
}

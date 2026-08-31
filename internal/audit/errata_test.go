package audit_test

import (
	"testing"

	"github.com/philoserf/ctworldgen/internal/audit"
)

func root(t *testing.T) string {
	t.Helper()

	r, err := audit.Root()
	if err != nil {
		t.Fatal(err)
	}

	return r
}

// TestEveryCitationResolvesToAReading is one direction: an E00N written
// in the code or the documents must name an entry in ERRATA.md. A reading
// applied without an entry is a reading applied silently, which is the
// thing this repository forbids.
func TestEveryCitationResolvesToAReading(t *testing.T) {
	t.Parallel()

	repoRoot := root(t)

	headings, err := audit.Headings(repoRoot)
	if err != nil {
		t.Fatal(err)
	}

	recorded := map[string]bool{}
	for _, id := range headings {
		recorded[id] = true
	}

	cited, err := audit.Citations(repoRoot)
	if err != nil {
		t.Fatal(err)
	}

	for id, files := range cited {
		if !recorded[id] {
			t.Errorf("%s is cited in %v and has no heading in docs/ERRATA.md", id, files)
		}
	}
}

// TestEveryReadingIsCited is the other direction: an entry nothing cites
// is either a reading the code forgot to apply or an entry that should not
// be there.
func TestEveryReadingIsCited(t *testing.T) {
	t.Parallel()

	repoRoot := root(t)

	headings, err := audit.Headings(repoRoot)
	if err != nil {
		t.Fatal(err)
	}

	if len(headings) == 0 {
		t.Fatal("docs/ERRATA.md records no readings")
	}

	cited, err := audit.Citations(repoRoot)
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range headings {
		if len(cited[id]) == 0 {
			t.Errorf("%s is recorded in docs/ERRATA.md and cited nowhere", id)
		}
	}
}

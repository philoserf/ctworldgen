package render_test

import (
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/starmap"
)

// broadAreasRecord is the roster's banded fixture, generated. Both
// documents are checked against the one record, because they are required
// to report it in the same words.
func broadAreasRecord(t *testing.T) *starmap.Record {
	t.Helper()

	engine, err := gen.New()
	if err != nil {
		t.Fatalf("building the engine: %v", err)
	}

	golden := fixture.BroadAreasGolden()

	record, err := engine.Generate(gen.Inputs{
		Seed: golden.Seed, Name: golden.Name,
		OccurrenceDM: golden.OccurrenceDM, OccurrenceAreas: golden.OccurrenceAreas,
	})
	if err != nil {
		t.Fatalf("generating the broad-areas fixture: %v", err)
	}

	if len(record.OccurrenceAreas) == 0 {
		t.Fatal("the broad-areas fixture carries no areas, so this test would hold nothing")
	}

	return record
}

// TestBroadAreasReachBothDocuments: a referee reading either document has
// to be told what geography the run was made under, or the map in front of
// him is thinner in one half for no stated reason.
//
// Both are asserted in one test because they share the sentence -- summary
// is one function -- and checking only one is how the two documents came
// to carry two bullet lists that agreed by convention.
func TestBroadAreasReachBothDocuments(t *testing.T) {
	t.Parallel()

	record := broadAreasRecord(t)

	// One rectangle and its DM, written as the summary writes them.
	const (
		rift    = "0101-0805 at -1"
		cluster = "0106-0810 at +1"
	)

	written := listing(t, record)
	for _, want := range []string{rift, cluster, "E012"} {
		if !strings.Contains(written, want) {
			t.Errorf("the listing does not say %q:\n%s", want, written)
		}
	}

	// The booklet sets the same sentence, wrapped, so the text of the
	// whole first page is what carries it.
	var set strings.Builder

	for _, stamp := range stamps(t, pages(t, drawn(t, record))[0]) {
		set.WriteString(stamp.Text)
		set.WriteString(" ")
	}

	// A minus sign is the one character here the booklet re-encodes, so
	// the DM is looked for with the rectangle rather than alone.
	for _, want := range []string{"0101-0805", "0106-0810", "E012"} {
		if !strings.Contains(set.String(), want) {
			t.Errorf("the booklet's first page does not say %q:\n%s", want, set.String())
		}
	}
}

// TestARecordWithNoBroadAreasSaysNothingAboutThem holds the other half of
// the omitempty guarantee, on the documents rather than the record: the
// four goldens that carry no areas must open with the sentence they always
// opened with, and a listing that grew "Broad areas: none" would move all
// four of them.
func TestARecordWithNoBroadAreasSaysNothingAboutThem(t *testing.T) {
	t.Parallel()

	written := listing(t, starmap.New(1, aramis, 0, nil))

	if strings.Contains(written, "Broad area") {
		t.Errorf("a listing of a record with no broad areas mentions them:\n%s", written)
	}

	if strings.Contains(written, "E012") {
		t.Errorf("a listing of a record with no broad areas cites E012:\n%s", written)
	}
}

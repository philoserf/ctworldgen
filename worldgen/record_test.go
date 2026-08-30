package worldgen_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/worldgen"
)

// TestRecordRoundTripsThroughJSON: the JSON record is the source of truth,
// so marshaling and parsing it must not change it.
func TestRecordRoundTripsThroughJSON(t *testing.T) {
	sub := generate(t, worldgen.Config{Seed: 99, Name: "Round"})

	data, err := sub.MarshalRecord()
	if err != nil {
		t.Fatalf("MarshalRecord: %v", err)
	}

	if !strings.HasSuffix(string(data), "}\n") {
		t.Error("the canonical record does not end with a newline")
	}

	back, err := worldgen.UnmarshalRecord(data)
	if err != nil {
		t.Fatalf("UnmarshalRecord: %v", err)
	}

	again, err := back.MarshalRecord()
	if err != nil {
		t.Fatalf("MarshalRecord after the round trip: %v", err)
	}

	if string(again) != string(data) {
		t.Error("the record changed across a JSON round trip")
	}
}

// TestUnmarshalRejectsAnUnknownField: a record from a newer schema must
// fail loudly rather than silently drop what this build cannot read.
func TestUnmarshalRejectsAnUnknownField(t *testing.T) {
	_, err := worldgen.UnmarshalRecord([]byte(`{"schema_version":"1","gas_giants":3}`))
	if err == nil || !strings.Contains(err.Error(), "gas_giants") {
		t.Errorf("UnmarshalRecord with an unknown field: err = %v", err)
	}
}

// TestUnmarshalRejectsTwoRecordsInOneFile: a JSONL batch file handed to
// `render` would otherwise be read as its first line alone.
func TestUnmarshalRejectsTwoRecordsInOneFile(t *testing.T) {
	_, err := worldgen.UnmarshalRecord([]byte("{\"schema_version\":\"1\"}\n{\"schema_version\":\"1\"}\n"))
	if !errors.Is(err, worldgen.ErrTrailingData) {
		t.Errorf("UnmarshalRecord of two records: err = %v, want ErrTrailingData", err)
	}
}

// TestUnmarshalRejectsMalformedJSON.
func TestUnmarshalRejectsMalformedJSON(t *testing.T) {
	if _, err := worldgen.UnmarshalRecord([]byte("not json")); err == nil {
		t.Error("UnmarshalRecord accepted text that is not JSON")
	}
}

// TestErrataIDsAreDistinctAndWellFormed: the ids are stamped into records
// and looked up in docs/ERRATA.md by hand, so a duplicate or a typo would
// mislead.
func TestErrataIDsAreDistinctAndWellFormed(t *testing.T) {
	seen := map[string]bool{}

	for _, id := range worldgen.ErrataIDs() {
		if seen[id] {
			t.Errorf("duplicate erratum id %q", id)
		}

		seen[id] = true

		if len(id) != 4 || id[0] != 'E' {
			t.Errorf("erratum id %q is not of the form E001", id)
		}
	}
}

// TestParseHexRejectsWhatIsNotOnTheGrid (p. 3: eight columns of ten rows).
func TestParseHexRejectsWhatIsNotOnTheGrid(t *testing.T) {
	for _, bad := range []string{"", "01", "010101", "0000", "0900", "0101x", "0111", "0011", "ab01", "01ab"} {
		if _, err := worldgen.ParseHex(bad); !errors.Is(err, worldgen.ErrBadHex) {
			t.Errorf("ParseHex(%q): err = %v, want ErrBadHex", bad, err)
		}
	}

	for _, good := range []string{"0101", "0810", "0110", "0801"} {
		h, err := worldgen.ParseHex(good)
		if err != nil {
			t.Errorf("ParseHex(%q): %v", good, err)

			continue
		}

		if h.String() != good {
			t.Errorf("ParseHex(%q).String() = %q", good, h.String())
		}
	}
}

// TestAllHexesIsTheWholeGridInScanOrder (docs/ERRATA.md E002).
func TestAllHexesIsTheWholeGridInScanOrder(t *testing.T) {
	all := worldgen.AllHexes()
	if len(all) != worldgen.Hexes {
		t.Fatalf("AllHexes has %d hexes, the p. 3 grid has %d", len(all), worldgen.Hexes)
	}

	previous := ""

	for _, h := range all {
		if !h.OnGrid() {
			t.Errorf("AllHexes yielded %s, which is off the grid", h)
		}

		if h.String() <= previous {
			t.Errorf("AllHexes yielded %s after %s, out of scan order", h, previous)
		}

		previous = h.String()
	}

	if all[0].String() != "0101" || all[len(all)-1].String() != "0810" {
		t.Errorf("AllHexes runs %s to %s, want 0101 to 0810", all[0], all[len(all)-1])
	}
}

package worldgen_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/worldgen"
)

// update rewrites the golden fixtures instead of comparing against them.
// Run it through `task goldens`, never by hand: the task runs the full
// gate afterwards, and the fixtures are the engine's own output, so a
// regeneration is a deliberate act whose diff is meant to be read.
var update = flag.Bool("update", false, "rewrite the golden fixtures")

func goldenPath(name string) string { return filepath.Join("testdata", name+".json") }

// TestGoldenSubsectors compares each fixture's record byte for byte
// against the file on disk. This is the test that notices when a rule, a
// table, or the dice-stream consumption order moves.
func TestGoldenSubsectors(t *testing.T) {
	for _, f := range fixture.All() {
		t.Run(f.Name, func(t *testing.T) {
			sub, err := worldgen.Generate(f.Config)
			if err != nil {
				t.Fatalf("Generate(%+v): %v", f.Config, err)
			}

			got, err := sub.MarshalRecord()
			if err != nil {
				t.Fatalf("MarshalRecord: %v", err)
			}

			path := goldenPath(f.Name)

			if *update {
				if err := os.WriteFile(path, got, 0o644); err != nil { //nolint:gosec // test fixture
					t.Fatalf("writing %s: %v", path, err)
				}

				return
			}

			want, err := os.ReadFile(path) //nolint:gosec // a testdata path this test built
			if err != nil {
				t.Fatalf("reading %s (run `task goldens` to create it): %v", path, err)
			}

			if string(got) != string(want) {
				t.Errorf("record for %s does not match %s; run `task goldens` if the change was intended", f.Name, path)
			}
		})
	}
}

// TestGoldenRecordsReplay proves the recorded fixtures still reproduce.
// The golden test above would pass on a record and a build that were both
// wrong in the same way; this one re-runs the engine from the record's own
// seed and compares, which is the contract goal 2 states.
func TestGoldenRecordsReplay(t *testing.T) {
	if *update {
		t.Skip("fixtures are being rewritten")
	}

	for _, f := range fixture.All() {
		t.Run(f.Name, func(t *testing.T) {
			data, err := os.ReadFile(goldenPath(f.Name))
			if err != nil {
				t.Fatalf("reading the fixture: %v", err)
			}

			sub, err := worldgen.UnmarshalRecord(data)
			if err != nil {
				t.Fatalf("UnmarshalRecord: %v", err)
			}

			if err := worldgen.Replay(sub, false); err != nil {
				t.Errorf("Replay: %v", err)
			}
		})
	}
}

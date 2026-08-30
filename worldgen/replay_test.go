package worldgen_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/worldgen"
)

func generate(t *testing.T, cfg worldgen.Config) *worldgen.Subsector {
	t.Helper()

	sub, err := worldgen.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate(%+v): %v", cfg, err)
	}

	return sub
}

// TestReplayRoundTrip is goal 2: the record replays from its own seed and
// inputs, at every occurrence DM.
func TestReplayRoundTrip(t *testing.T) {
	for _, dm := range []int{-1, 0, 1} {
		sub := generate(t, worldgen.Config{Seed: 12345, Name: "Round Trip", OccurrenceDM: dm})
		if err := worldgen.Replay(sub, false); err != nil {
			t.Errorf("Replay at dm %+d: %v", dm, err)
		}
	}
}

// TestReplayAcceptsSpelledOutEmptySlices: Dice and DMs carry omitempty, so
// a throw with no DMs is written with no dms key. A record that spells that
// out as an empty array is the same record — it marshals back to the same
// bytes — and must replay, not be reported as diverged. The event loop
// compares parsed values rather than the marshaled form, so this is the one
// place the two halves of compare could disagree.
func TestReplayAcceptsSpelledOutEmptySlices(t *testing.T) {
	sub := generate(t, worldgen.Config{Seed: 7, Name: "Vega"})

	data, err := sub.MarshalRecord()
	if err != nil {
		t.Fatalf("MarshalRecord: %v", err)
	}

	rec, err := worldgen.UnmarshalRecord(data)
	if err != nil {
		t.Fatalf("UnmarshalRecord: %v", err)
	}

	spelled := 0

	for i, ev := range rec.Events {
		if ev.Kind == worldgen.KindThrow && len(ev.DMs) == 0 {
			rec.Events[i].DMs = []worldgen.EventDM{}
			spelled++
		}
	}

	if spelled == 0 {
		t.Fatal("no throw in the record has an empty DM list; the case is untested")
	}

	if err := worldgen.Replay(rec, false); err != nil {
		t.Errorf("Replay of a record spelling out %d empty DM lists: %v", spelled, err)
	}
}

// TestReplayDetectsATamperedThrow: the event log is verification data, so
// a record whose log has been edited must fail rather than be believed.
func TestReplayDetectsATamperedThrow(t *testing.T) {
	sub := generate(t, worldgen.Config{Seed: 5})

	for i, ev := range sub.Events {
		if ev.Kind == worldgen.KindThrow {
			sub.Events[i].Total++

			break
		}
	}

	err := worldgen.Replay(sub, false)
	if err == nil || !errors.Is(err, worldgen.ErrDiverged) {
		t.Fatalf("Replay of a tampered record: err = %v, want ErrDiverged", err)
	}

	// The message has to name the diverging event, which is the whole
	// point of comparing the logs before the records.
	if !strings.Contains(err.Error(), "at event ") {
		t.Errorf("divergence does not name the event: %v", err)
	}
}

// TestReplayDetectsATamperedWorld covers the other half of compare: a
// record whose event log is intact but whose worlds have been edited.
// The starport is edited rather than the name, because the name is the
// referee's annotation and replay ignores it by design.
func TestReplayDetectsATamperedWorld(t *testing.T) {
	sub := generate(t, worldgen.Config{Seed: 5})
	if len(sub.Worlds) == 0 {
		t.Fatal("seed 5 placed no worlds")
	}

	sub.Worlds[0].Starport = "X"

	if err := worldgen.Replay(sub, false); !errors.Is(err, worldgen.ErrDiverged) {
		t.Errorf("Replay of an edited world: err = %v, want ErrDiverged", err)
	}
}

// TestReplayToleratesRefereeNames: naming each world is step 3 of the
// p. 12 checklist and the book prints no table for it, so the name is
// written into the record after generation and appears in neither the
// inputs nor the event log. A record whose referee has done what the book
// asks must still verify, or the documented workflow would invalidate the
// documented verification.
func TestReplayToleratesRefereeNames(t *testing.T) {
	sub := generate(t, worldgen.Config{Seed: 5, Name: "Vega"})
	if len(sub.Worlds) < 2 {
		t.Fatal("seed 5 placed too few worlds")
	}

	sub.Worlds[0].Name = "Regina"
	sub.Worlds[1].Name = "Efate"

	if err := worldgen.Replay(sub, false); err != nil {
		t.Errorf("Replay of a record the referee has named: %v", err)
	}

	// Replay must not have edited the caller's record to achieve that.
	if sub.Worlds[0].Name != "Regina" {
		t.Errorf("Replay cleared the caller's world name")
	}
}

// TestReplayDetectsATamperedInput: inputs feed the engine, so changing one
// changes the subsector it regenerates.
func TestReplayDetectsATamperedInput(t *testing.T) {
	sub := generate(t, worldgen.Config{Seed: 5, OccurrenceDM: 0})
	sub.Inputs.OccurrenceDM = 1

	if err := worldgen.Replay(sub, false); !errors.Is(err, worldgen.ErrDiverged) {
		t.Errorf("Replay with an edited occurrence DM: err = %v, want ErrDiverged", err)
	}
}

// TestReplayRejectsAnInputThePageForbids: a record can be edited to hold
// an occurrence DM the book does not offer, and the regeneration must
// refuse it rather than replay it.
func TestReplayRejectsAnInputThePageForbids(t *testing.T) {
	sub := generate(t, worldgen.Config{Seed: 5})
	sub.Inputs.OccurrenceDM = 3

	if err := worldgen.Replay(sub, false); !errors.Is(err, worldgen.ErrBadInput) {
		t.Errorf("Replay with an out-of-range occurrence DM: err = %v, want ErrBadInput", err)
	}
}

// TestReplayChecksProvenance and TestReplayIgnoreProvenanceWaivesOnlyTheStamps:
// --ignore-provenance waives the version match and nothing else.
func TestReplayChecksProvenance(t *testing.T) {
	for _, tamper := range []struct {
		name string
		edit func(*worldgen.Subsector)
	}{
		{"schema_version", func(s *worldgen.Subsector) { s.SchemaVersion = "0" }},
		{"engine_version", func(s *worldgen.Subsector) { s.EngineVersion = "0.0.0-not-this-build" }},
		{"ruleset", func(s *worldgen.Subsector) { s.Ruleset = "some other edition" }},
		{"rng.algorithm", func(s *worldgen.Subsector) { s.RNG.Algorithm = "mt19937" }},
	} {
		t.Run(tamper.name, func(t *testing.T) {
			sub := generate(t, worldgen.Config{Seed: 5})
			tamper.edit(sub)

			if err := worldgen.Replay(sub, false); !errors.Is(err, worldgen.ErrProvenance) {
				t.Errorf("Replay: err = %v, want ErrProvenance", err)
			}

			// The stamp is the only thing that changed, so the waiver must
			// let this record through.
			if err := worldgen.Replay(sub, true); err != nil {
				t.Errorf("Replay --ignore-provenance: %v", err)
			}
		})
	}
}

// TestIgnoreProvenanceDoesNotWaiveDivergence: the waiver is for the
// stamps, not for the record.
func TestIgnoreProvenanceDoesNotWaiveDivergence(t *testing.T) {
	sub := generate(t, worldgen.Config{Seed: 5})
	sub.EngineVersion = "0.0.0-not-this-build"
	sub.Events[0].Text = "edited"

	if err := worldgen.Replay(sub, true); !errors.Is(err, worldgen.ErrDiverged) {
		t.Errorf("Replay --ignore-provenance of an edited log: err = %v, want ErrDiverged", err)
	}
}

// TestReplayDetectsATruncatedLog and its opposite: the comparison walks
// both logs, so a record that stops early or runs long is caught.
func TestReplayDetectsALogOfTheWrongLength(t *testing.T) {
	t.Run("truncated", func(t *testing.T) {
		sub := generate(t, worldgen.Config{Seed: 5})
		sub.Events = sub.Events[:len(sub.Events)-1]

		if err := worldgen.Replay(sub, false); !errors.Is(err, worldgen.ErrDiverged) {
			t.Errorf("Replay of a truncated log: err = %v, want ErrDiverged", err)
		}
	})

	t.Run("extended", func(t *testing.T) {
		sub := generate(t, worldgen.Config{Seed: 5})
		sub.Events = append(sub.Events, worldgen.Event{
			Seq: len(sub.Events) + 1, Kind: worldgen.KindOutcome, Step: "made up", Text: "made up",
		})

		if err := worldgen.Replay(sub, false); !errors.Is(err, worldgen.ErrDiverged) {
			t.Errorf("Replay of an extended log: err = %v, want ErrDiverged", err)
		}
	})
}

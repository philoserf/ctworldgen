package worldgen

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/philoserf/ctworldgen/dice"
)

// Replay re-runs the engine from the record's own seed and inputs and
// checks that it reproduces the record. The stored event log is
// verification data, not input: nothing in it is fed back to the engine.
// That is the whole of replay here, because the procedure takes no
// decisions — there is no choice channel to reapply, as there is in the
// sibling (docs/PRD.md, "The architectural delta from ctchargen").
//
// ignoreProvenance waives the version-stamp match and nothing else: a
// record whose events diverge still fails.
func Replay(rec *Subsector, ignoreProvenance bool) error {
	if !ignoreProvenance {
		if err := checkProvenance(rec); err != nil {
			return err
		}
	}

	regen, err := Generate(Config{
		Seed:         rec.RNG.Seed,
		Name:         rec.Inputs.Name,
		OccurrenceDM: rec.Inputs.OccurrenceDM,
	})
	if err != nil {
		return fmt.Errorf("regenerating from the record: %w", err)
	}

	return compare(rec, regen, ignoreProvenance)
}

// checkProvenance holds a record to the build that is about to replay it.
// The algorithm is checked as well as the versions: a record made under a
// different RNG cannot reproduce, and saying so by name beats reporting a
// divergence at event 1.
func checkProvenance(rec *Subsector) error {
	for _, stamp := range []struct {
		field string
		got   string
		want  string
	}{
		{"schema_version", rec.SchemaVersion, SchemaVersion},
		{"engine_version", rec.EngineVersion, EngineVersion},
		{"ruleset", rec.Ruleset, Ruleset},
		{"rng.algorithm", rec.RNG.Algorithm, dice.Algorithm},
	} {
		if stamp.got != stamp.want {
			return fmt.Errorf("%w: record %s is %q, this build writes %q (replay --ignore-provenance waives this)",
				ErrProvenance, stamp.field, stamp.got, stamp.want)
		}
	}

	return nil
}

// compare reports the first divergence. Events come first and are compared
// one at a time, because their sequence number is the most useful thing a
// mismatch can name: it points at the throw that went a different way.
// The whole record is compared afterwards, which catches a divergence in
// something the log does not carry.
func compare(rec, regen *Subsector, ignoreProvenance bool) error {
	for i, want := range regen.Events {
		if i >= len(rec.Events) {
			return fmt.Errorf("%w: the record ends at event %d, the replay continues to %d (%s)",
				ErrDiverged, len(rec.Events), len(regen.Events), describe(want))
		}

		if got := rec.Events[i]; !sameEvent(got, want) {
			return fmt.Errorf("%w at event %d: the record has %s, the replay has %s",
				ErrDiverged, want.Seq, describe(got), describe(want))
		}
	}

	if len(rec.Events) > len(regen.Events) {
		return fmt.Errorf("%w: the record continues to event %d, the replay ends at %d (%s)",
			ErrDiverged, len(rec.Events), len(regen.Events), describe(rec.Events[len(regen.Events)]))
	}

	return compareRecords(rec, regen, ignoreProvenance)
}

// sameEvent reports whether a recorded event and the regenerated one are
// the same event, comparing copies with their empty slices normalised
// away.
//
// The normalisation is what keeps this check as strict as, and no stricter
// than, the record comparison below. Dice and DMs carry omitempty, so a
// throw with no DMs is written with no dms key at all; a record that spells
// that out as "dms": [] parses to an empty non-nil slice where the engine
// holds nil, and reflect.DeepEqual calls those different. The two records
// are byte-identical once marshaled, so compareRecords would pass the very
// pair this loop rejected — and it rejected it with a message that reads
// the same on both sides, because no field describe prints had changed.
//
// Only the slices are degenerate this way. Success is a *bool and marshals
// as false rather than being omitted, so a nil-versus-false difference
// there is a real divergence and is left alone.
func sameEvent(got, want Event) bool {
	return reflect.DeepEqual(withoutEmptySlices(got), withoutEmptySlices(want))
}

// withoutEmptySlices returns the event with empty Dice and DMs replaced by
// nil. Event is taken by value, so the caller's copy is untouched.
func withoutEmptySlices(ev Event) Event {
	if len(ev.Dice) == 0 {
		ev.Dice = nil
	}

	if len(ev.DMs) == 0 {
		ev.DMs = nil
	}

	return ev
}

// compareRecords compares the two records byte for byte, in canonical
// form. Marshaling both sides means the comparison is of the records and
// not of the bytes the file happened to hold, so whitespace and key order
// in a hand-formatted file cannot fail a replay that reproduced.
//
// Under ignoreProvenance the four stamps checkProvenance would have
// checked are blanked on both sides first. Without that the waiver would
// waive nothing: the stamps are fields of the record, so a record carrying
// a version this build has left behind would clear the stamp check and
// then fail here on the very same four values — which is the one case the
// flag exists for.
func compareRecords(rec, regen *Subsector, ignoreProvenance bool) error {
	// Both sides are copied before anything is blanked: Replay must not
	// edit the record its caller handed it.
	rec, regen = withoutWorldNames(rec), withoutWorldNames(regen)

	if ignoreProvenance {
		for _, s := range []*Subsector{rec, regen} {
			s.SchemaVersion = ""
			s.EngineVersion = ""
			s.Ruleset = ""
			s.RNG.Algorithm = ""
		}
	}

	got, err := rec.MarshalRecord()
	if err != nil {
		return fmt.Errorf("re-marshaling the record: %w", err)
	}

	want, err := regen.MarshalRecord()
	if err != nil {
		return fmt.Errorf("marshaling the replay: %w", err)
	}

	if !bytes.Equal(got, want) {
		return fmt.Errorf("%w: the event logs agree but the records do not", ErrDiverged)
	}

	return nil
}

// withoutWorldNames copies a record with every world's name cleared.
//
// Naming each world is step 3 of the p. 12 checklist, and the book prints
// no table for it: the name is the referee's, written into the record
// afterwards. It is therefore in neither the inputs nor the event log,
// so the engine can only ever regenerate it as empty — and a record whose
// referee has done what p. 12 asks would otherwise be reported as
// diverged. Replay verifies what the dice produced; the names are an
// annotation layer over it.
//
// The Worlds slice is copied, not just the struct around it, because the
// blanking writes through to its elements — unlike the provenance stamps,
// which are scalars on the record itself.
func withoutWorldNames(sub *Subsector) *Subsector {
	out := *sub
	out.Worlds = make([]World, len(sub.Worlds))
	copy(out.Worlds, sub.Worlds)

	for i := range out.Worlds {
		out.Worlds[i].Name = ""
	}

	return &out
}

// describe renders an event for a divergence message.
func describe(ev Event) string {
	switch ev.Kind {
	case KindThrow:
		// The DMs are printed, not just the total they produced: a
		// divergence message whose two halves read identically tells the
		// reader nothing, and the DMs are the field most likely to differ
		// while every other printed field matches.
		var out strings.Builder
		fmt.Fprintf(&out, "throw %q %v", ev.Label, ev.Dice)

		for _, dm := range ev.DMs {
			fmt.Fprintf(&out, " %+d %s", dm.Value, dm.Source)
		}

		fmt.Fprintf(&out, " total %d", ev.Total)

		if ev.Target != "" {
			out.WriteString(" vs " + ev.Target)
		}

		if ev.Hex != "" {
			out.WriteString(" at " + ev.Hex)
		}

		return out.String()
	case KindOutcome:
		return fmt.Sprintf("outcome %q", ev.Text)
	case KindStep:
		return fmt.Sprintf("step %q", ev.Step)
	}

	return ev.Kind + " event"
}

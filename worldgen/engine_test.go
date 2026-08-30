package worldgen_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/tables"
	"github.com/philoserf/ctworldgen/worldgen"
)

// seeds is the sweep the property tests run over. Every assertion below is
// a rule of Book 3 that must hold for every subsector the engine can
// produce, so the way to test them is to produce a great many.
func seeds() []uint64 {
	out := make([]uint64, 0, 60)
	for i := range uint64(60) {
		out = append(out, i*7919+1)
	}

	return out
}

func generateAll(t *testing.T) []*worldgen.Subsector {
	t.Helper()

	subs := make([]*worldgen.Subsector, 0, len(seeds())*3)

	for _, seed := range seeds() {
		for _, dm := range []int{-1, 0, 1} {
			sub, err := worldgen.Generate(worldgen.Config{Seed: seed, OccurrenceDM: dm})
			if err != nil {
				t.Fatalf("Generate(seed %d, dm %+d): %v", seed, dm, err)
			}

			subs = append(subs, sub)
		}
	}

	return subs
}

// TestOccurrenceDMIsBoundedByThePage: "a DM of +1 or -1" (p. 1) is the
// whole of the referee's latitude here, so anything else is refused rather
// than quietly applied.
func TestOccurrenceDMIsBoundedByThePage(t *testing.T) {
	for _, dm := range []int{-2, 2, 7, -100} {
		if _, err := worldgen.Generate(worldgen.Config{Seed: 1, OccurrenceDM: dm}); !errors.Is(err, worldgen.ErrBadInput) {
			t.Errorf("Generate with occurrence DM %+d: err = %v, want ErrBadInput", dm, err)
		}
	}

	for _, dm := range []int{-1, 0, 1} {
		if _, err := worldgen.Generate(worldgen.Config{Seed: 1, OccurrenceDM: dm}); err != nil {
			t.Errorf("Generate with occurrence DM %+d: %v", dm, err)
		}
	}
}

// TestCharacteristicsStayInTheirPrintedRanges is docs/ERRATA.md E004's
// contract, and the precondition the p. 9 technological index matrix
// depends on: every value it is asked for must have a row.
func TestCharacteristicsStayInTheirPrintedRanges(t *testing.T) {
	charts, err := tables.Load()
	if err != nil {
		t.Fatal(err)
	}

	for _, sub := range generateAll(t) {
		for _, w := range sub.Worlds {
			checkTabledRanges(t, charts, sub, w)
			checkUntabledRanges(t, sub, w)
		}
	}
}

// checkTabledRanges holds the five characteristics that are clamped to
// their own table's printed range. This is docs/ERRATA.md E004's contract
// and the precondition the p. 9 technological index matrix depends on:
// every value it is asked for must have a row.
func checkTabledRanges(t *testing.T, charts *tables.Charts, sub *worldgen.Subsector, w worldgen.World) {
	t.Helper()

	for table, value := range map[string]int{
		tables.Size:          w.Size,
		tables.Atmosphere:    w.Atmosphere,
		tables.Hydrographics: w.Hydrographics,
		tables.Population:    w.Population,
		tables.Government:    w.Government,
	} {
		maximum, ok := charts.MaxValue(table)
		if !ok {
			t.Fatalf("no %s table", table)
		}

		if value < 0 || value > maximum {
			t.Errorf("seed %d hex %s: %s is %d, outside the printed 0-%d",
				sub.RNG.Seed, w.Hex, table, value, maximum)
		}
	}
}

// checkUntabledRanges holds the two that are not clamped to a descriptive
// table: law level is floored and never capped (E004), and the
// technological index has the book's own zero-to-18 (p. 9).
func checkUntabledRanges(t *testing.T, sub *worldgen.Subsector, w worldgen.World) {
	t.Helper()

	if w.LawLevel < 0 {
		t.Errorf("seed %d hex %s: law level %d is below zero", sub.RNG.Seed, w.Hex, w.LawLevel)
	}

	if w.TechnologicalIndex < 0 || w.TechnologicalIndex > 18 {
		t.Errorf("seed %d hex %s: technological index %d is outside the p. 9 zero-to-18",
			sub.RNG.Seed, w.Hex, w.TechnologicalIndex)
	}
}

// TestClampsAreStampedAndLogged: E004 must never be applied silently. If
// any world holds a value the raw arithmetic could not have produced
// unclamped, the record has to say so.
func TestClampsAreStampedAndLogged(t *testing.T) {
	clampsSeen := 0

	for _, sub := range generateAll(t) {
		logged := false

		for _, ev := range sub.Events {
			if strings.Contains(ev.Text, "clamped to") {
				logged = true
				clampsSeen++
			}
		}

		stamped := slices.Contains(sub.Errata, worldgen.ErrataClamp)
		if logged != stamped {
			t.Errorf("seed %d dm %+d: clamp logged = %v but E004 stamped = %v",
				sub.RNG.Seed, sub.Inputs.OccurrenceDM, logged, stamped)
		}
	}

	// A sweep this size that never clamps would mean the assertion above is
	// vacuous.
	if clampsSeen == 0 {
		t.Error("no clamp fired across the whole sweep; E004's branches are untested")
	}
}

// TestAutomaticValuesConsumeNoDie. "A planet of size zero automatically
// has an atmosphere of zero" and "a planetary size of 0 or 1 indicates an
// automatic result of 0" for hydrographics (pp. 4, 12). Automatic means
// determined without a throw, and because the dice stream is consumed in
// procedure order, a throw made and discarded would shift every later
// world's characteristics.
func TestAutomaticValuesConsumeNoDie(t *testing.T) {
	sizeZero, sizeOne := 0, 0

	for _, sub := range generateAll(t) {
		thrown := throwsByHex(sub)

		for _, w := range sub.Worlds {
			if w.Size == 0 {
				sizeZero++

				checkAutomatic(t, sub, w, "planetary atmosphere", w.Atmosphere, thrown[w.Hex])
			}

			if w.Size == 1 {
				sizeOne++
			}

			if w.Size <= 1 {
				checkAutomatic(t, sub, w, "hydrographic percentage", w.Hydrographics, thrown[w.Hex])
			}
		}
	}

	if sizeZero == 0 || sizeOne == 0 {
		t.Errorf("the sweep produced %d size-0 and %d size-1 worlds; both automatic rules need one",
			sizeZero, sizeOne)
	}
}

// throwsByHex indexes the log by the world each throw was made for, so a
// test can ask whether a particular throw happened at all.
func throwsByHex(sub *worldgen.Subsector) map[string]map[string]bool {
	out := map[string]map[string]bool{}

	for _, ev := range sub.Events {
		if ev.Kind != worldgen.KindThrow || ev.Hex == "" {
			continue
		}

		if out[ev.Hex] == nil {
			out[ev.Hex] = map[string]bool{}
		}

		out[ev.Hex][ev.Label] = true
	}

	return out
}

func checkAutomatic(
	t *testing.T, sub *worldgen.Subsector, w worldgen.World,
	label string, got int, thrown map[string]bool,
) {
	t.Helper()

	if got != 0 {
		t.Errorf("seed %d hex %s: size %d but %s %d", sub.RNG.Seed, w.Hex, w.Size, label, got)
	}

	if thrown[label] {
		t.Errorf("seed %d hex %s: size %d threw for %s anyway", sub.RNG.Seed, w.Hex, w.Size, label)
	}
}

// TestRoutesFollowThePrintedTable is docs/ERRATA.md E003's contract: pairs
// within four hexes, no X starport at either end, each pair at most once,
// and the lower hex first.
func TestRoutesFollowThePrintedTable(t *testing.T) {
	for _, sub := range generateAll(t) {
		starport := map[string]string{}
		for _, w := range sub.Worlds {
			starport[w.Hex] = w.Starport
		}

		seen := map[string]bool{}

		for _, r := range sub.Routes {
			key := r.From + "-" + r.To
			if seen[key] {
				t.Errorf("seed %d: pair %s charted twice", sub.RNG.Seed, key)
			}

			seen[key] = true

			checkRoute(t, sub, r, starport)
		}
	}
}

func checkRoute(t *testing.T, sub *worldgen.Subsector, r worldgen.Route, starport map[string]string) {
	t.Helper()

	key := r.From + "-" + r.To

	if r.From >= r.To {
		t.Errorf("seed %d: route %s is not in ascending hex order", sub.RNG.Seed, key)
	}

	if r.Distance < 1 || r.Distance > tables.MaxJumpDistance {
		t.Errorf("seed %d: route %s has jump distance %d, outside the p. 2 table's four columns",
			sub.RNG.Seed, key, r.Distance)
	}

	from, err := worldgen.ParseHex(r.From)
	if err != nil {
		t.Fatalf("seed %d: route %s: %v", sub.RNG.Seed, key, err)
	}

	to, err := worldgen.ParseHex(r.To)
	if err != nil {
		t.Fatalf("seed %d: route %s: %v", sub.RNG.Seed, key, err)
	}

	if got := from.Distance(to); got != r.Distance {
		t.Errorf("seed %d: route %s records jump-%d, the grid says %d", sub.RNG.Seed, key, r.Distance, got)
	}

	for _, hex := range []string{r.From, r.To} {
		if starport[hex] == tables.StarportX {
			t.Errorf("seed %d: route %s runs to %s, which has no starport (p. 2 prints no X row)",
				sub.RNG.Seed, key, hex)
		}
	}
}

// TestBasesOnlyWhereTheChartPrintsAThrow (p. 5): E and X have neither
// base, and C and D have no naval base.
func TestBasesOnlyWhereTheChartPrintsAThrow(t *testing.T) {
	for _, sub := range generateAll(t) {
		for _, w := range sub.Worlds {
			naval := w.Starport == tables.StarportA || w.Starport == tables.StarportB
			scout := naval || w.Starport == tables.StarportC || w.Starport == tables.StarportD

			if w.NavalBase && !naval {
				t.Errorf("seed %d hex %s: naval base at a type %s starport", sub.RNG.Seed, w.Hex, w.Starport)
			}

			if w.ScoutBase && !scout {
				t.Errorf("seed %d hex %s: scout base at a type %s starport", sub.RNG.Seed, w.Hex, w.Starport)
			}
		}
	}
}

// TestWorldsAreInScanOrderAndOnTheGrid is docs/ERRATA.md E002: the record
// lists the worlds in the order the occurrence scan found them, which for
// the p. 3 grid's fixed-width identifiers is plain ascending order.
func TestWorldsAreInScanOrderAndOnTheGrid(t *testing.T) {
	for _, sub := range generateAll(t) {
		if len(sub.Worlds) > worldgen.Hexes {
			t.Errorf("seed %d: %d worlds in a %d-hex subsector", sub.RNG.Seed, len(sub.Worlds), worldgen.Hexes)
		}

		previous := ""

		for _, w := range sub.Worlds {
			if _, err := worldgen.ParseHex(w.Hex); err != nil {
				t.Errorf("seed %d: %v", sub.RNG.Seed, err)
			}

			if w.Hex <= previous {
				t.Errorf("seed %d: hex %s follows %s, out of scan order", sub.RNG.Seed, w.Hex, previous)
			}

			previous = w.Hex
		}
	}
}

// TestProfileRestatesTheCharacteristics is docs/ERRATA.md E005: eight
// characters, the p. 4 box's order, no separator, and every digit the one
// the value is written as.
func TestProfileRestatesTheCharacteristics(t *testing.T) {
	for _, sub := range generateAll(t) {
		for _, w := range sub.Worlds {
			if len(w.Profile) != 8 {
				t.Errorf("seed %d hex %s: profile %q is %d characters, want 8",
					sub.RNG.Seed, w.Hex, w.Profile, len(w.Profile))

				continue
			}

			if got := string(w.Profile[0]); got != w.Starport {
				t.Errorf("seed %d hex %s: profile starts %q, starport is %q", sub.RNG.Seed, w.Hex, got, w.Starport)
			}

			for i, value := range []int{
				w.Size, w.Atmosphere, w.Hydrographics,
				w.Population, w.Government, w.LawLevel, w.TechnologicalIndex,
			} {
				want, ok := tables.Digit(value)
				if !ok {
					t.Fatalf("seed %d hex %s: value %d has no digit", sub.RNG.Seed, w.Hex, value)
				}

				if w.Profile[i+1] != want {
					t.Errorf("seed %d hex %s: profile %q character %d is %q, value %d is %q",
						sub.RNG.Seed, w.Hex, w.Profile, i+1, string(w.Profile[i+1]), value, string(want))
				}
			}
		}
	}
}

// TestStampedErrataAreRealReadings: a stamp naming a reading the document
// does not carry would send a reader looking for something that is not
// there.
func TestStampedErrataAreRealReadings(t *testing.T) {
	known := worldgen.ErrataIDs()

	for _, sub := range generateAll(t) {
		for _, id := range sub.Errata {
			if !slices.Contains(known, id) {
				t.Errorf("seed %d stamps %q, which this engine does not apply", sub.RNG.Seed, id)
			}
		}

		if !slices.IsSorted(sub.Errata) {
			t.Errorf("seed %d stamps %v, not in document order", sub.RNG.Seed, sub.Errata)
		}
	}
}

// TestEventSequenceIsMonotonic: an outcome references its causing throw by
// sequence number (FR17), so the numbers must be dense from 1 and a
// reference must point backwards at a throw.
func TestEventSequenceIsMonotonic(t *testing.T) {
	for _, sub := range generateAll(t) {
		throws := map[int]bool{}

		for i, ev := range sub.Events {
			if ev.Seq != i+1 {
				t.Fatalf("seed %d: event %d has seq %d", sub.RNG.Seed, i, ev.Seq)
			}

			if ev.Kind == worldgen.KindThrow {
				throws[ev.Seq] = true
			}

			if ev.Ref != 0 && !throws[ev.Ref] {
				t.Errorf("seed %d: event %d references %d, which is not an earlier throw",
					sub.RNG.Seed, ev.Seq, ev.Ref)
			}
		}
	}
}

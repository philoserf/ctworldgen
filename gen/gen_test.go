package gen_test

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/philoserf/ctworldgen/dice"
	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/subsector"
	"github.com/philoserf/ctworldgen/tables"
)

// newEngine builds the engine once per test. The charts are the same for
// every seed, so loading them per subsector would re-parse and re-validate
// ten embedded documents several hundred times across this suite.
func newEngine(t *testing.T) *gen.Engine {
	t.Helper()

	engine, err := gen.New()
	if err != nil {
		t.Fatalf("building the engine: %v", err)
	}

	return engine
}

func generate(t *testing.T, engine *gen.Engine, in gen.Inputs) *subsector.Subsector {
	t.Helper()

	s, err := engine.Generate(in)
	if err != nil {
		t.Fatalf("generating %+v: %v", in, err)
	}

	return s
}

func marshal(t *testing.T, s *subsector.Subsector) []byte {
	t.Helper()

	b, err := subsector.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func goldenPath(g fixture.Golden) string {
	return filepath.Join("testdata", g.File+".json")
}

// TestGoldens pins the dice stream. A golden moves only by regeneration:
// run `task regenerate` and read the diff. A change here that was not
// intended is a change to what every seed means.
func TestGoldens(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			want, err := os.ReadFile(goldenPath(golden))
			if err != nil {
				t.Fatalf("%v (run `task regenerate` to create it)", err)
			}

			in := gen.Inputs{Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM}

			got := marshal(t, generate(t, engine, in))
			if string(got) != string(want) {
				t.Errorf("%s does not match the golden.\n"+
					"If this change was intended, run `task regenerate` and read the diff.", golden.File)
			}
		})
	}
}

// TestRegenerationRoundTrip is the guarantee that replaces a `verify`
// subcommand: every golden is reproduced from its own recorded seed and
// inputs, read back out of the record itself.
func TestRegenerationRoundTrip(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			encoded, err := os.ReadFile(goldenPath(golden))
			if err != nil {
				t.Fatal(err)
			}

			recorded, err := subsector.Decode(bytes.NewReader(encoded))
			if err != nil {
				t.Fatalf("decoding %s: %v", golden.File, err)
			}

			again := generate(t, engine, gen.Inputs{
				Seed:         recorded.Seed,
				Name:         recorded.Name,
				OccurrenceDM: recorded.OccurrenceDM,
			})
			if string(marshal(t, again)) != string(encoded) {
				t.Errorf("%s did not reproduce from its own recorded seed and inputs", golden.File)
			}
		})
	}
}

// TestSameSeedSameSubsector is determinism stated directly.
func TestSameSeedSameSubsector(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)

	in := gen.Inputs{Seed: 12345, Name: "Aramis", OccurrenceDM: 0}
	first := marshal(t, generate(t, engine, in))

	second := marshal(t, generate(t, engine, in))
	if string(first) != string(second) {
		t.Error("the same seed and inputs produced two different subsectors")
	}
}

// TestOccurrenceDMChangesTheStream is the point of ERRATA E002's noted
// discrepancy: the world occurrence throw is a target, so the referee's DM
// makes worlds more or less frequent, which under a set-membership reading
// a +1 could not do.
func TestOccurrenceDMChangesTheStream(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)

	counts := map[int]int{}
	for _, dm := range []int{-1, 0, 1} {
		counts[dm] = len(generate(t, engine, gen.Inputs{Seed: 7, Name: "", OccurrenceDM: dm}).Worlds)
	}

	if counts[-1] >= counts[0] || counts[0] >= counts[1] {
		t.Errorf("worlds by DM: -1 => %d, 0 => %d, +1 => %d; a DM should make worlds more or less frequent",
			counts[-1], counts[0], counts[1])
	}
}

func TestRejectsDMsTheBookDoesNotOffer(t *testing.T) {
	t.Parallel()

	for _, dm := range []int{-2, 2, 7, -100} {
		_, err := newEngine(t).Generate(gen.Inputs{Seed: 0, Name: "", OccurrenceDM: dm})
		if err == nil {
			t.Errorf("an occurrence DM of %d was accepted; p. 1 offers -1, 0 and +1", dm)
		}
	}
}

// TestInvariantsOverManySeeds sweeps the book's invariants for everything
// milestone 1 generates.
func TestInvariantsOverManySeeds(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)

	charts, err := tables.Load()
	if err != nil {
		t.Fatal(err)
	}

	certain := 0

	for i := range 200 {
		seed := uint64(i)
		for _, dm := range []int{-1, 0, 1} {
			record := generate(t, engine, gen.Inputs{Seed: seed, Name: "", OccurrenceDM: dm})

			assertWorldsWellFormed(t, record, seed)
			assertRecordCarriesItsInputs(t, record, seed, dm)
			assertBasesFollowTheChart(t, record, seed)
			assertLanesFollowTheTable(t, charts, record, seed)

			certain += assertLanesTheTableMakesCertain(t, charts, record, seed)
		}
	}

	// The certain-lane check is the only one of these that can pass by
	// finding nothing to look at, so say how much it looked at.
	if certain == 0 {
		t.Error("no pair in the sweep sat on a cell stating 1, so the certain-lane direction proved nothing")
	}
}

// assertWorldsWellFormed: every world sits on the p. 3 grid, once, in
// ascending hex number (ERRATA E002), with a starport the book prints and
// no name (p. 12 prints no naming table).
func assertWorldsWellFormed(t *testing.T, record *subsector.Subsector, seed uint64) {
	t.Helper()

	if len(record.Worlds) > subsector.Columns*subsector.Rows {
		t.Fatalf("seed %d: %d worlds in an eighty-hex subsector", seed, len(record.Worlds))
	}

	seen := map[subsector.Hex]bool{}

	var previous subsector.Hex

	for index, world := range record.Worlds {
		assertWorldWellFormed(t, world, seed)

		if seen[world.Hex] {
			t.Fatalf("seed %d: two worlds at %s", seed, world.Hex)
		}

		seen[world.Hex] = true

		if index > 0 && !previous.Less(world.Hex) {
			t.Fatalf("seed %d: worlds are not in ascending hex number (%s then %s)", seed, previous, world.Hex)
		}

		previous = world.Hex
	}
}

// assertWorldWellFormed holds one world to what the pages allow.
func assertWorldWellFormed(t *testing.T, world subsector.World, seed uint64) {
	t.Helper()

	if !onTheGrid(world.Hex) {
		t.Fatalf("seed %d: world at %s is off the p. 3 grid", seed, world.Hex)
	}

	if !world.Starport.Valid() {
		t.Fatalf("seed %d: world at %s has starport %q", seed, world.Hex, world.Starport)
	}

	if world.Name != "" {
		t.Fatalf("seed %d: world at %s is named %q; p. 12 prints no naming table", seed, world.Hex, world.Name)
	}
}

// onTheGrid reports whether a hex is one of the eighty the p. 3 grid
// prints.
func onTheGrid(h subsector.Hex) bool {
	return h.Col >= 1 && h.Col <= subsector.Columns && h.Row >= 1 && h.Row <= subsector.Rows
}

// assertRecordCarriesItsInputs: the seed and inputs a run is reproducible
// from, and the readings that governed it. E002 governs every record and
// is the only reading milestone 1 implements a step for.
func assertRecordCarriesItsInputs(t *testing.T, record *subsector.Subsector, seed uint64, occurrenceDM int) {
	t.Helper()

	// E002 governs every record. E001 is stamped where any base throw was
	// made, which is any world whose starport the p. 5 chart prints a
	// throw for. E003 is stamped once there are two worlds, the point at
	// which there is a pair for the reading to govern.
	want := []string{"E002"}

	for _, world := range record.Worlds {
		if world.Starport != subsector.StarportE && world.Starport != subsector.StarportX {
			want = append(want, "E001")

			break
		}
	}

	if len(record.Worlds) >= 2 {
		want = append(want, "E003")
	}

	slices.Sort(want)

	if !slices.Equal(record.Errata, want) {
		t.Fatalf("seed %d: errata = %v, want %v", seed, record.Errata, want)
	}

	if record.Seed != seed || record.OccurrenceDM != occurrenceDM {
		t.Fatalf("seed %d, DM %+d: record carries seed %d, DM %+d", seed, occurrenceDM, record.Seed, record.OccurrenceDM)
	}
}

// TestStarportDistributionFollowsThePage checks the p. 1 distribution
// shows up in bulk. The table gives each type a span of the two-dice
// range, so a type's frequency is the frequency of its throws, not the
// count of them: C takes 7 and 8 and is the most likely at 11 of 36,
// A takes 2, 3 and 4 for 6 of 36, D takes 9 alone for 4 of 36, and X
// takes 12 alone for 1 of 36.
func TestStarportDistributionFollowsThePage(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)

	counts := map[subsector.Starport]int{}

	for i := range 300 {
		seed := uint64(i)
		for _, w := range generate(t, engine, gen.Inputs{Seed: seed, Name: "", OccurrenceDM: 1}).Worlds {
			counts[w.Starport]++
		}
	}

	if counts[subsector.StarportX] == 0 {
		t.Error("no starport X in 300 subsectors; a throw of 12 gives one")
	}

	if counts[subsector.StarportC] <= counts[subsector.StarportD] {
		t.Errorf("starport C (%d) is no more common than D (%d); the page gives C two throws and D one",
			counts[subsector.StarportC], counts[subsector.StarportD])
	}

	if counts[subsector.StarportD] <= counts[subsector.StarportX] {
		t.Errorf("starport D (%d) is no more common than X (%d)", counts[subsector.StarportD], counts[subsector.StarportX])
	}
}

// assertBasesFollowTheChart: the p. 5 chart prints a naval base throw at
// starports A and B only and a scout base throw at A through D, so a base
// anywhere else was never thrown for and must be absent (ERRATA E001).
func assertBasesFollowTheChart(t *testing.T, record *subsector.Subsector, seed uint64) {
	t.Helper()

	for _, world := range record.Worlds {
		navalPossible := world.Starport == subsector.StarportA || world.Starport == subsector.StarportB
		if world.NavalBase && !navalPossible {
			t.Fatalf("seed %d: naval base at %s, a starport %s, and the chart prints no throw for one",
				seed, world.Hex, world.Starport)
		}

		scoutPossible := world.Starport != subsector.StarportE && world.Starport != subsector.StarportX
		if world.ScoutBase && !scoutPossible {
			t.Fatalf("seed %d: scout base at %s, a starport %s, and the chart prints no throw for one",
				seed, world.Hex, world.Starport)
		}
	}
}

// assertLanesFollowTheTable: every lane is four parsecs or fewer, between
// two worlds whose starport pair has a row and whose cell states a number,
// written lower hex first, and examined only once (R5, ERRATA E003).
func assertLanesFollowTheTable(t *testing.T, charts *tables.Tables, record *subsector.Subsector, seed uint64) {
	t.Helper()

	worlds := map[subsector.Hex]subsector.Starport{}
	for _, world := range record.Worlds {
		worlds[world.Hex] = world.Starport
	}

	seen := map[string]bool{}

	for _, route := range record.Routes {
		if !route.From.Less(route.To) {
			t.Fatalf("seed %d: lane %s-%s is not written lower hex first", seed, route.From, route.To)
		}

		key := route.From.String() + route.To.String()
		if seen[key] {
			t.Fatalf("seed %d: the pair %s-%s was examined twice", seed, route.From, route.To)
		}

		seen[key] = true

		fromPort, okFrom := worlds[route.From]
		toPort, okTo := worlds[route.To]

		if !okFrom || !okTo {
			t.Fatalf("seed %d: lane %s-%s reaches a hex with no world", seed, route.From, route.To)
		}

		if fromPort == subsector.StarportX || toPort == subsector.StarportX {
			t.Fatalf("seed %d: lane %s-%s touches a starport X, which p. 5 gives no starship landings",
				seed, route.From, route.To)
		}

		assertLaneHasACell(t, charts, route, fromPort, toPort, seed)
	}
}

// assertLaneHasACell: the distance is the one the p. 3 grid gives, within
// the four columns the jump routes table prints, at a cell that states a
// number rather than an em-dash.
func assertLaneHasACell(
	t *testing.T, charts *tables.Tables, route subsector.Route,
	fromPort, toPort subsector.Starport, seed uint64,
) {
	t.Helper()

	if route.Distance != route.From.Distance(route.To) {
		t.Fatalf("seed %d: lane %s-%s records %d parsecs, and the grid gives %d",
			seed, route.From, route.To, route.Distance, route.From.Distance(route.To))
	}

	if route.Distance < 1 || route.Distance > tables.MaxJump {
		t.Fatalf("seed %d: lane %s-%s is %d parsecs, beyond the four the table states targets for",
			seed, route.From, route.To, route.Distance)
	}

	if _, stated := charts.JumpRoutes.Target(fromPort, toPort, route.Distance); !stated {
		t.Fatalf("seed %d: lane %s-%s sits at a cell the page prints an em-dash in", seed, route.From, route.To)
	}
}

// certainTarget is the one-die target a throw cannot miss: every face of
// one die is equal to or greater than 1, so a jump routes cell stating it
// draws a lane without fail (p. 2).
const certainTarget dice.Target = 1

// assertLanesTheTableMakesCertain is R5 in the direction the other checks
// cannot see. assertLanesFollowTheTable holds the lanes that exist to the
// page, so it passes for a record with no lanes at all -- and a target
// read backwards produces exactly that kind of record: every lane it
// draws still sits at a stated cell, four parsecs or fewer, away from an
// X. The cells stating 1 are the ones that settle it, because a lane
// there is not a matter of the throw.
//
// It returns the number of certain pairs it examined, so that the sweep
// can say the check had something to check.
func assertLanesTheTableMakesCertain(
	t *testing.T, charts *tables.Tables, record *subsector.Subsector, seed uint64,
) int {
	t.Helper()

	drawn := map[string]bool{}
	for _, route := range record.Routes {
		drawn[route.From.String()+route.To.String()] = true
	}

	examined := 0

	for i, first := range record.Worlds {
		for _, second := range record.Worlds[i+1:] {
			distance := first.Hex.Distance(second.Hex)

			target, stated := charts.JumpRoutes.Target(first.Starport, second.Starport, distance)
			if !stated || target != certainTarget {
				continue
			}

			examined++

			if !drawn[first.Hex.String()+second.Hex.String()] {
				t.Fatalf("seed %d: no lane %s-%s, and the table states 1 for %s-%s at jump-%d, which one die always meets",
					seed, first.Hex, second.Hex, first.Starport, second.Starport, distance)
			}
		}
	}

	return examined
}

// The two-dice throw the p. 5 base targets are read against (B1 p. 2).
const (
	faces           = 6
	twoDiceOutcomes = faces * faces
)

// baseSweepSubsectors is enough subsectors for a base rate to settle. At
// the +1 DM a subsector averages about fifty worlds, so even starport A --
// six of the starports table's thirty-six throws -- lands some thousands
// of times.
const baseSweepSubsectors = 300

// baseRateTolerance is how far an observed rate may sit from the one the
// chart's throw predicts. The seeds are fixed, so this does not flake; it
// is wide enough to absorb the sampling in a few thousand worlds and
// narrower than the gap a misread target opens, which is twice the
// target's own distance from an even chance -- a sixth at the narrowest.
const baseRateTolerance = 0.10

// TestBaseThrowsFollowTheChart is R4 in the direction assertBasesFollowTheChart
// cannot see. That check holds a base to the starport it was thrown for,
// so it passes whatever the throw decided -- a target read backwards puts
// scout bases at A, where the chart asks 10+, far more often than at D,
// where it asks 7+, and every base still sits at a starport the chart
// prints a throw for. The rate is what tells the two apart.
func TestBaseThrowsFollowTheChart(t *testing.T) {
	t.Parallel()

	charts, err := tables.Load()
	if err != nil {
		t.Fatal(err)
	}

	engine := newEngine(t)

	worlds := map[subsector.Starport]int{}
	naval := map[subsector.Starport]int{}
	scout := map[subsector.Starport]int{}

	for i := range baseSweepSubsectors {
		for _, world := range generate(t, engine, gen.Inputs{Seed: uint64(i), Name: "", OccurrenceDM: 1}).Worlds {
			worlds[world.Starport]++

			if world.NavalBase {
				naval[world.Starport]++
			}

			if world.ScoutBase {
				scout[world.Starport]++
			}
		}
	}

	for _, port := range subsector.Starports() {
		if worlds[port] == 0 {
			t.Errorf("no starport %s in %d subsectors, so its base throws were never sampled", port, baseSweepSubsectors)

			continue
		}

		if target, printed := charts.StarportChart.NavalBase(port); printed {
			assertRateFollowsTheThrow(t, "naval base", port, target, naval[port], worlds[port])
		}

		if target, printed := charts.StarportChart.ScoutBase(port); printed {
			assertRateFollowsTheThrow(t, "scout base", port, target, scout[port], worlds[port])
		}
	}
}

// assertRateFollowsTheThrow holds an observed base rate to the frequency
// the chart's own target predicts, counted over the thirty-six two-dice
// outcomes rather than written down as a number.
func assertRateFollowsTheThrow(
	t *testing.T, what string, port subsector.Starport, target dice.Target, got, worlds int,
) {
	t.Helper()

	want := float64(twoDiceOdds(target)) / float64(twoDiceOutcomes)

	rate := float64(got) / float64(worlds)
	if math.Abs(rate-want) > baseRateTolerance {
		t.Errorf("%s at starport %s: %d of %d worlds, a rate of %.3f, and the chart's throw of %d+ predicts %.3f",
			what, port, got, worlds, rate, int(target), want)
	}
}

// twoDiceOdds counts the two-dice outcomes that meet a target, which is
// the frequency the chart's throw states.
func twoDiceOdds(target dice.Target) int {
	met := 0

	for first := 1; first <= faces; first++ {
		for second := 1; second <= faces; second++ {
			if target.Met(first + second) {
				met++
			}
		}
	}

	return met
}

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
	"github.com/philoserf/ctworldgen/starmap"
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

func generate(t *testing.T, engine *gen.Engine, in gen.Inputs) *starmap.Record {
	t.Helper()

	s, err := engine.Generate(in)
	if err != nil {
		t.Fatalf("generating %+v: %v", in, err)
	}

	return s
}

func marshal(t *testing.T, s *starmap.Record) []byte {
	t.Helper()

	b, err := starmap.Marshal(s)
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

			recorded, err := starmap.Decode(bytes.NewReader(encoded))
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
// the engine generates: the occurrence scan, starports, bases, routes, and
// the eight characteristics of every world.
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
			assertRoutesFollowTheTable(t, charts, record, seed)

			certain += assertRoutesTheTableMakesCertain(t, charts, record, seed)

			for _, world := range record.Worlds {
				assertAutomaticZeros(t, world, seed)
				assertWithinFormula(t, world, seed)
				assertTechIndexIsTheMatrix(t, charts, world, seed)
				assertDigitsSpellTheWorld(t, world, seed)
				assertClampsAreHonest(t, world, seed)
			}
		}
	}

	// The certain-route check is the only one of these that can pass by
	// finding nothing to look at, so say how much it looked at.
	if certain == 0 {
		t.Error("no pair in the sweep sat on a cell stating 1, so the certain-route direction proved nothing")
	}
}

// assertWorldsWellFormed: every world sits on the p. 3 grid, once, in
// ascending hex number (ERRATA E002), with a starport the book prints and
// no name (p. 12 prints no naming table).
func assertWorldsWellFormed(t *testing.T, record *starmap.Record, seed uint64) {
	t.Helper()

	if len(record.Worlds) > starmap.Columns*starmap.Rows {
		t.Fatalf("seed %d: %d worlds in an eighty-hex subsector", seed, len(record.Worlds))
	}

	seen := map[starmap.Hex]bool{}

	var previous starmap.Hex

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
func assertWorldWellFormed(t *testing.T, world starmap.World, seed uint64) {
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
func onTheGrid(h starmap.Hex) bool {
	return h.Col >= 1 && h.Col <= starmap.Columns && h.Row >= 1 && h.Row <= starmap.Rows
}

// assertRecordCarriesItsInputs: the seed and inputs a run is reproducible
// from, and the readings that governed it. E002 governs every record and
// is the only reading milestone 1 implements a step for.
func assertRecordCarriesItsInputs(t *testing.T, record *starmap.Record, seed uint64, occurrenceDM int) {
	t.Helper()

	assertErrataStamped(t, record, seed)

	if record.Seed != seed || record.OccurrenceDM != occurrenceDM {
		t.Fatalf("seed %d, DM %+d: record carries seed %d, DM %+d", seed, occurrenceDM, record.Seed, record.OccurrenceDM)
	}
}

// assertErrataStamped: each entry of ERRATA.md states the condition under
// which a record stamps it, and those statements are the specification.
func assertErrataStamped(t *testing.T, record *starmap.Record, seed uint64) {
	t.Helper()

	// E002 governs every record. E001 is stamped where any base throw was
	// made, which is any world whose starport the p. 5 chart prints a
	// throw for. E003 is stamped once there are two worlds, the point at
	// which there is a pair for the reading to govern.
	want := []string{"E002"}

	for _, world := range record.Worlds {
		if world.Starport != starmap.StarportE && world.Starport != starmap.StarportX {
			want = append(want, "E001")

			break
		}
	}

	if len(record.Worlds) >= 2 {
		want = append(want, "E003")
	}

	// E004 is stamped where a floor or the cap actually bound a value.
	for _, world := range record.Worlds {
		if len(world.Clamps) > 0 {
			want = append(want, "E004")

			break
		}
	}

	// E005 governs the string of digits, and there is one per world.
	if len(record.Worlds) > 0 {
		want = append(want, "E005")
	}

	slices.Sort(want)

	if !slices.Equal(record.Errata, want) {
		t.Fatalf("seed %d: errata = %v, want %v", seed, record.Errata, want)
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

	counts := map[starmap.Starport]int{}

	for i := range 300 {
		seed := uint64(i)
		for _, w := range generate(t, engine, gen.Inputs{Seed: seed, Name: "", OccurrenceDM: 1}).Worlds {
			counts[w.Starport]++
		}
	}

	if counts[starmap.StarportX] == 0 {
		t.Error("no starport X in 300 subsectors; a throw of 12 gives one")
	}

	if counts[starmap.StarportC] <= counts[starmap.StarportD] {
		t.Errorf("starport C (%d) is no more common than D (%d); the page gives C two throws and D one",
			counts[starmap.StarportC], counts[starmap.StarportD])
	}

	if counts[starmap.StarportD] <= counts[starmap.StarportX] {
		t.Errorf("starport D (%d) is no more common than X (%d)", counts[starmap.StarportD], counts[starmap.StarportX])
	}
}

// assertBasesFollowTheChart: the p. 5 chart prints a naval base throw at
// starports A and B only and a scout base throw at A through D, so a base
// anywhere else was never thrown for and must be absent (ERRATA E001).
func assertBasesFollowTheChart(t *testing.T, record *starmap.Record, seed uint64) {
	t.Helper()

	for _, world := range record.Worlds {
		navalPossible := world.Starport == starmap.StarportA || world.Starport == starmap.StarportB
		if world.NavalBase && !navalPossible {
			t.Fatalf("seed %d: naval base at %s, a starport %s, and the chart prints no throw for one",
				seed, world.Hex, world.Starport)
		}

		scoutPossible := world.Starport != starmap.StarportE && world.Starport != starmap.StarportX
		if world.ScoutBase && !scoutPossible {
			t.Fatalf("seed %d: scout base at %s, a starport %s, and the chart prints no throw for one",
				seed, world.Hex, world.Starport)
		}
	}
}

// assertRoutesFollowTheTable: every route is four parsecs or fewer, between
// two worlds whose starport pair has a row and whose cell states a number,
// written lower hex first, and examined only once (R5, ERRATA E003).
func assertRoutesFollowTheTable(t *testing.T, charts *tables.Tables, record *starmap.Record, seed uint64) {
	t.Helper()

	worlds := map[starmap.Hex]starmap.Starport{}
	for _, world := range record.Worlds {
		worlds[world.Hex] = world.Starport
	}

	seen := map[string]bool{}

	for _, route := range record.Routes {
		if !route.From.Less(route.To) {
			t.Fatalf("seed %d: route %s-%s is not written lower hex first", seed, route.From, route.To)
		}

		key := route.From.String() + route.To.String()
		if seen[key] {
			t.Fatalf("seed %d: the pair %s-%s was examined twice", seed, route.From, route.To)
		}

		seen[key] = true

		fromPort, okFrom := worlds[route.From]
		toPort, okTo := worlds[route.To]

		if !okFrom || !okTo {
			t.Fatalf("seed %d: route %s-%s reaches a hex with no world", seed, route.From, route.To)
		}

		if fromPort == starmap.StarportX || toPort == starmap.StarportX {
			t.Fatalf("seed %d: route %s-%s touches a starport X, which p. 5 gives no starship landings",
				seed, route.From, route.To)
		}

		assertRouteHasACell(t, charts, route, fromPort, toPort, seed)
	}
}

// assertRouteHasACell: the distance is the one the p. 3 grid gives, within
// the four columns the jump routes table prints, at a cell that states a
// number rather than an em-dash.
func assertRouteHasACell(
	t *testing.T, charts *tables.Tables, route starmap.Route,
	fromPort, toPort starmap.Starport, seed uint64,
) {
	t.Helper()

	if route.Distance != route.From.Distance(route.To) {
		t.Fatalf("seed %d: route %s-%s records %d parsecs, and the grid gives %d",
			seed, route.From, route.To, route.Distance, route.From.Distance(route.To))
	}

	if route.Distance < 1 || route.Distance > tables.MaxJump {
		t.Fatalf("seed %d: route %s-%s is %d parsecs, beyond the four the table states targets for",
			seed, route.From, route.To, route.Distance)
	}

	if _, stated := charts.JumpRoutes.Target(fromPort, toPort, route.Distance); !stated {
		t.Fatalf("seed %d: route %s-%s sits at a cell the page prints an em-dash in", seed, route.From, route.To)
	}
}

// certainTarget is the one-die target a throw cannot miss: every face of
// one die is equal to or greater than 1, so a jump routes cell stating it
// draws a route without fail (p. 2).
const certainTarget dice.Target = 1

// assertRoutesTheTableMakesCertain is R5 in the direction the other checks
// cannot see. assertRoutesFollowTheTable holds the routes that exist to the
// page, so it passes for a record with no routes at all -- and a target
// read backwards produces exactly that kind of record: every route it
// draws still sits at a stated cell, four parsecs or fewer, away from an
// X. The cells stating 1 are the ones that settle it, because a route
// there is not a matter of the throw.
//
// It returns the number of certain pairs it examined, so that the sweep
// can say the check had something to check.
func assertRoutesTheTableMakesCertain(
	t *testing.T, charts *tables.Tables, record *starmap.Record, seed uint64,
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
				t.Fatalf("seed %d: no route %s-%s, and the table states 1 for %s-%s at jump-%d, which one die always meets",
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

	worlds := map[starmap.Starport]int{}
	naval := map[starmap.Starport]int{}
	scout := map[starmap.Starport]int{}

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

	for _, port := range starmap.Starports() {
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
	t *testing.T, what string, port starmap.Starport, target dice.Target, got, worlds int,
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

// assertAutomaticZeros is R13, the two throws the book does not make. A
// size-0 world's atmosphere and a size-0-or-1 world's hydrographics are
// automatic, so they are zero -- and nothing clamped them, because there
// was no throw to clamp. A clamp on one is proof a die was thrown that
// should not have been.
func assertAutomaticZeros(t *testing.T, world starmap.World, seed uint64) {
	t.Helper()

	if world.Size == 0 && world.Atmosphere != 0 {
		t.Fatalf("seed %d: %s is size 0 with atmosphere %d; p. 4 makes it 0", seed, world.Hex, world.Atmosphere)
	}

	if world.Size <= 1 && world.Hydrographics != 0 {
		t.Fatalf("seed %d: %s is size %d with hydrographics %d; p. 4 makes it 0",
			seed, world.Hex, world.Size, world.Hydrographics)
	}

	for _, clamp := range world.Clamps {
		automatic := (world.Size == 0 && clamp.Characteristic == starmap.Atmosphere) ||
			(world.Size <= 1 && clamp.Characteristic == starmap.Hydrographics)
		if automatic {
			t.Fatalf("seed %d: %s records a %s clamp, so a die was thrown for a value p. 4 makes automatic",
				seed, world.Hex, clamp.Characteristic)
		}
	}
}

// assertWithinFormula: 2D-2 spans 0 to 10, and 2D-7+X spans X-5 to X+5,
// floored at 0 (R6-R11, R14). The hydrographics DM is not a maybe -- the
// page applies it exactly when the atmosphere is 0, 1 or greater than 9 --
// so its span is exact too.
//
// Nothing here forbids an airless world with water, and nothing should:
// p. 4 ties the automatic zero to planetary size and gives atmosphere a
// DM of -4 instead, so the combination is legal and about one world in
// sixty has it (ERRATA.md, Noted discrepancies).
func assertWithinFormula(t *testing.T, world starmap.World, seed uint64) {
	t.Helper()

	inRange(t, seed, world, "size", world.Size, 0, 10)
	inRange(t, seed, world, "population", world.Population, 0, 10)

	if world.Size > 0 {
		inRange(t, seed, world, "atmosphere", world.Atmosphere, max(world.Size-5, 0), world.Size+5)
	}

	if world.Size > 1 {
		low, high := world.Size-5, world.Size+5
		if world.Atmosphere <= 1 || world.Atmosphere > 9 {
			low, high = low-4, high-4
		}

		inRange(t, seed, world, "hydrographics", world.Hydrographics, max(low, 0), high)
	}

	inRange(t, seed, world, "government", world.Government, max(world.Population-5, 0), world.Population+5)
	inRange(t, seed, world, "law level", world.LawLevel, max(world.Government-5, 0), world.Government+5)
}

func inRange(t *testing.T, seed uint64, world starmap.World, name string, value, low, high int) {
	t.Helper()

	if value < low || value > high {
		t.Fatalf("seed %d: %s has %s %d, and the formula allows %d to %d", seed, world.Hex, name, value, low, high)
	}
}

// assertTechIndexIsTheMatrix recomputes the DM total from the p. 9 matrix
// and asserts the recorded index is one a single die could have produced.
// This is exact: six outcomes, and the record must be one of them.
func assertTechIndexIsTheMatrix(t *testing.T, charts *tables.Tables, world starmap.World, seed uint64) {
	t.Helper()

	matrix := charts.TechIndexMatrix
	modifier := matrix.StarportDM(world.Starport) +
		matrix.DM(tables.ColSize, world.Size) +
		matrix.DM(tables.ColAtmosphere, world.Atmosphere) +
		matrix.DM(tables.ColHydrographics, world.Hydrographics) +
		matrix.DM(tables.ColPopulation, world.Population) +
		matrix.DM(tables.ColGovernment, world.Government)

	reachable := map[int]bool{}
	for die := 1; die <= 6; die++ {
		reachable[min(max(die+modifier, 0), maxTechIndexUnderTest)] = true
	}

	if !reachable[world.TechIndex] {
		t.Fatalf("seed %d: %s has technological index %d, and the matrix total of %+d puts it in %v",
			seed, world.Hex, world.TechIndex, modifier, reachable)
	}
}

// assertDigitsSpellTheWorld decodes the string of digits rather than
// re-encoding it, so it cannot agree with the writer by sharing its code.
// Eight characters, starport first, nothing between them (ERRATA E005).
func assertDigitsSpellTheWorld(t *testing.T, world starmap.World, seed uint64) {
	t.Helper()

	const digitsInTheString = 8

	if len(world.Digits) != digitsInTheString {
		t.Fatalf("seed %d: %s has the string %q, and p. 4 wants eight characters", seed, world.Hex, world.Digits)
	}

	if world.Digits[:1] != world.Starport.String() {
		t.Fatalf("seed %d: %s spells starport %q and carries %s", seed, world.Hex, world.Digits[:1], world.Starport)
	}

	values := []int{
		world.Size, world.Atmosphere, world.Hydrographics,
		world.Population, world.Government, world.LawLevel, world.TechIndex,
	}

	for position, value := range values {
		digit, err := starmap.ParseDigit(world.Digits[position+1 : position+2])
		if err != nil {
			t.Fatalf("seed %d: %s spells %q: %v", seed, world.Hex, world.Digits, err)
		}

		if digit.Value() != value {
			t.Fatalf("seed %d: %s spells %q, whose character %d reads %d and the record carries %d",
				seed, world.Hex, world.Digits, position+1, digit.Value(), value)
		}
	}
}

// assertClampsAreHonest: a clamp is recorded only where one bound, its
// kept value is the one the world carries, and only the technological
// index is ever capped from above (R14, ERRATA E004).
func assertClampsAreHonest(t *testing.T, world starmap.World, seed uint64) {
	t.Helper()

	carried := map[starmap.Characteristic]int{
		starmap.Size: world.Size, starmap.Atmosphere: world.Atmosphere,
		starmap.Hydrographics: world.Hydrographics, starmap.Population: world.Population,
		starmap.Government: world.Government, starmap.LawLevel: world.LawLevel,
		starmap.TechIndex: world.TechIndex,
	}

	for _, clamp := range world.Clamps {
		if clamp.Raw == clamp.Value {
			t.Fatalf("seed %d: %s records a %s clamp that did not bind", seed, world.Hex, clamp.Characteristic)
		}

		if clamp.Raw > clamp.Value && clamp.Characteristic != starmap.TechIndex {
			t.Fatalf("seed %d: %s caps %s at %d; only the technological index has a printed cap",
				seed, world.Hex, clamp.Characteristic, clamp.Value)
		}

		if clamp.Raw < clamp.Value && clamp.Value != 0 {
			t.Fatalf("seed %d: %s floors %s to %d rather than 0", seed, world.Hex, clamp.Characteristic, clamp.Value)
		}

		if got, ok := carried[clamp.Characteristic]; !ok || got != clamp.Value {
			t.Fatalf("seed %d: %s clamps %s to %d and carries %d",
				seed, world.Hex, clamp.Characteristic, clamp.Value, got)
		}
	}
}

// TestTheClampsThatBindAreTheOnesR14Names is the empirical half of ERRATA
// E004. The reading says which characteristics can fall below zero and
// that the technological index cap, though rare, is reachable; this walks
// enough subsectors to show both, on fixed seeds so it cannot flake.
//
// Size and population are 2D-2 and have a floor of 0 already, so a clamp
// on either would mean the arithmetic had changed.
func TestTheClampsThatBindAreTheOnesR14Names(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)
	floors := map[starmap.Characteristic]int{}
	caps := map[starmap.Characteristic]int{}
	highest := 0

	for index := range 3000 {
		record := generate(t, engine, gen.Inputs{Seed: uint64(index), Name: "", OccurrenceDM: 1})
		for _, world := range record.Worlds {
			highest = max(highest, world.TechIndex)

			for _, clamp := range world.Clamps {
				if clamp.Raw < clamp.Value {
					floors[clamp.Characteristic]++
				} else {
					caps[clamp.Characteristic]++
				}
			}
		}
	}

	assertFloorsMatchE004(t, floors, caps)
	assertOnlyTheIndexIsCapped(t, caps, highest)
}

// assertFloorsMatchE004: "Atmosphere, hydrographics, government, law level
// and the technological index can all go negative; planetary size and
// population cannot.".
func assertFloorsMatchE004(t *testing.T, floors, caps map[starmap.Characteristic]int) {
	t.Helper()

	for _, characteristic := range []starmap.Characteristic{
		starmap.Atmosphere, starmap.Hydrographics,
		starmap.Government, starmap.LawLevel, starmap.TechIndex,
	} {
		if floors[characteristic] == 0 {
			t.Errorf("%s never floored, and E004 says it can fall below zero", characteristic)
		}
	}

	for _, characteristic := range []starmap.Characteristic{starmap.Size, starmap.Population} {
		if floors[characteristic] != 0 || caps[characteristic] != 0 {
			t.Errorf("%s was clamped, and 2D-2 has a floor of 0 already", characteristic)
		}
	}
}

// assertOnlyTheIndexIsCapped: the cap is rare -- an asteroid belt with a
// first class starport and ten billion inhabitants -- and still reachable.
func assertOnlyTheIndexIsCapped(t *testing.T, caps map[starmap.Characteristic]int, highest int) {
	t.Helper()

	if caps[starmap.TechIndex] == 0 {
		t.Error("the technological index cap never bound, and E004 says it is reachable")
	}

	for characteristic, count := range caps {
		if characteristic != starmap.TechIndex {
			t.Errorf("%s was capped %d times; p. 9 prints a range for the technological index alone",
				characteristic, count)
		}
	}

	if highest != maxTechIndexUnderTest {
		t.Errorf("the highest technological index in the sweep is %d, and p. 9 caps it at %d",
			highest, maxTechIndexUnderTest)
	}
}

// maxTechIndexUnderTest is p. 9's printed cap, retyped here rather than
// read from the engine so the two must agree.
const maxTechIndexUnderTest = 18

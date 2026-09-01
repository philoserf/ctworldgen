package gen_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/starmap"
)

func sector(t *testing.T, golden fixture.Golden) *starmap.Record {
	t.Helper()

	engine, err := gen.New()
	if err != nil {
		t.Fatal(err)
	}

	record, err := engine.Sector(gen.Inputs{
		Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM,
	})
	if err != nil {
		t.Fatal(err)
	}

	return record
}

// placed is the engine's own translation, not a second copy of the
// arithmetic. A test that re-derived the translation would assert only
// that its own copy is self-consistent, and would pass while the engine
// laid the members out any way at all.
func placed(t *testing.T, index int, hex starmap.Hex) starmap.Hex {
	t.Helper()

	on := gen.Place(index, hex)
	if !starmap.SectorGrid().Contains(on) {
		t.Fatalf("member %d puts %s at %s, off the sector grid", index, hex, on)
	}

	return on
}

// TestASectorsMembersAreTheSubsectorsNewWrites is the property the whole
// composite reading exists to keep, and the one the alpha report checked
// by hand: "tessarane-05.json, seed 3141597 -- identical to `ctworldgen
// new --seed 3141597`."
//
// Every world of every member is compared whole against the subsector
// that member's seed produces on its own. A sector that re-threw anything
// -- or laid a member in the wrong band -- fails here.
func TestASectorsMembersAreTheSubsectorsNewWrites(t *testing.T) {
	t.Parallel()

	golden := fixture.SectorGolden()
	record := sector(t, golden)

	engine, err := gen.New()
	if err != nil {
		t.Fatal(err)
	}

	inSector := make(map[starmap.Hex]starmap.World, len(record.Worlds))
	for _, world := range record.Worlds {
		inSector[world.Hex] = world
	}

	counted := 0

	for index := range gen.SectorMembers {
		alone, genErr := engine.Generate(gen.Inputs{
			Seed: golden.Seed + uint64(index), Name: golden.Name, OccurrenceDM: golden.OccurrenceDM,
		})
		if genErr != nil {
			t.Fatal(genErr)
		}

		for _, want := range alone.Worlds {
			want.Hex = placed(t, index, want.Hex)
			counted++

			got, ok := inSector[want.Hex]
			if !ok {
				t.Errorf("member %d has a world at %s and the sector has none", index, want.Hex)

				continue
			}

			// World carries a Clamps slice, so compare the marshalled
			// form: it is the whole world, and it is what a referee reads.
			if !reflect.DeepEqual(got, want) {
				t.Errorf("member %d at %s: the sector has %+v, `new --seed %d` writes %+v",
					index, want.Hex, got, golden.Seed+uint64(index), want)
			}
		}
	}

	if counted != len(record.Worlds) {
		t.Errorf("the members hold %d worlds between them and the sector holds %d", counted, len(record.Worlds))
	}
}

// TestRoutesCrossTheSeams is the invariant a sector exists for. Sixteen
// independent subsectors have hard borders -- the alpha report's words --
// and the seam pass is the whole of the difference.
func TestRoutesCrossTheSeams(t *testing.T) {
	t.Parallel()

	record := sector(t, fixture.SectorGolden())

	crossing := gen.CrossingRoutes(record)
	if len(crossing) == 0 {
		t.Fatal("no route crosses a member border, so the sector is sixteen subsectors in a trench coat")
	}

	for _, route := range crossing {
		if gen.MemberOf(route.From) == gen.MemberOf(route.To) {
			t.Errorf("%s-%s is counted as crossing and both ends are in member %d",
				route.From, route.To, gen.MemberOf(route.From))
		}

		if route.Distance < 1 || route.Distance > 4 {
			t.Errorf("the seam route %s-%s is %d parsecs, and p. 2 examines pairs within four hexes",
				route.From, route.To, route.Distance)
		}

		if got := route.From.Distance(route.To); got != route.Distance {
			t.Errorf("%s-%s records %d parsecs and measures %d", route.From, route.To, route.Distance, got)
		}
	}
}

// TestEveryPairIsExaminedOnce: p. 2 says "each specific pair of worlds
// should be examined for jump routes only once", so an interior pair its
// own member already examined must not be examined again at the seams.
func TestEveryPairIsExaminedOnce(t *testing.T) {
	t.Parallel()

	record := sector(t, fixture.SectorGolden())

	seen := make(map[[2]starmap.Hex]bool, len(record.Routes))

	for _, route := range record.Routes {
		pair := [2]starmap.Hex{route.From, route.To}
		if seen[pair] {
			t.Errorf("the pair %s-%s carries two routes", route.From, route.To)
		}

		seen[pair] = true

		if reversed := ([2]starmap.Hex{route.To, route.From}); seen[reversed] {
			t.Errorf("the pair %s-%s is recorded in both directions", route.From, route.To)
		}
	}
}

// TestTranslationKeepsThePageThreeParity is the sector's version of the
// parity trap. A subsector is eight columns wide and eight is even, so a
// column's odd-or-even parity survives translation and an interior pair
// measures the same distance in sector coordinates as it did at home
// (ERRATA E006 part 2). An odd band width -- or an off-by-one in the
// offset -- flips the parity of every second band and quietly changes
// interior distances, which nothing else here would catch.
func TestTranslationKeepsThePageThreeParity(t *testing.T) {
	t.Parallel()

	for index := range gen.SectorMembers {
		for aCol := 1; aCol <= starmap.Columns; aCol++ {
			for aRow := 1; aRow <= starmap.Rows; aRow++ {
				assertOneHexKeepsItsDistances(t, index, aCol, aRow)
			}
		}
	}
}

func assertOneHexKeepsItsDistances(t *testing.T, index, col, row int) {
	t.Helper()

	from, err := starmap.NewHex(col, row)
	if err != nil {
		t.Fatal(err)
	}

	for bCol := 1; bCol <= starmap.Columns; bCol++ {
		for bRow := 1; bRow <= starmap.Rows; bRow++ {
			other, hexErr := starmap.NewHex(bCol, bRow)
			if hexErr != nil {
				t.Fatal(hexErr)
			}

			home := from.Distance(other)

			away := placed(t, index, from).Distance(placed(t, index, other))
			if home != away {
				t.Fatalf("member %d: %s-%s is %d parsecs at home and %s-%s is %d on the sector grid",
					index, from, other, home, placed(t, index, from), placed(t, index, other), away)
			}
		}
	}
}

// TestTheSeamsGolden pins the one dice stream a sector adds.
func TestTheSeamsGolden(t *testing.T) {
	t.Parallel()

	encoded, err := os.ReadFile(filepath.Join("testdata", fixture.SectorGolden().File+".json"))
	if err != nil {
		t.Fatalf("%v (run `task regenerate` to create it)", err)
	}

	var want []starmap.Route

	err = json.Unmarshal(encoded, &want)
	if err != nil {
		t.Fatal(err)
	}

	got := gen.CrossingRoutes(sector(t, fixture.SectorGolden()))

	if len(got) != len(want) {
		t.Fatalf("the seam pass drew %d routes and the golden has %d.\n"+
			"If this change was intended, run `task regenerate` and read the diff.", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("seam route %d is %s-%s and the golden has %s-%s.\n"+
				"If this change was intended, run `task regenerate` and read the diff.",
				i, got[i].From, got[i].To, want[i].From, want[i].To)
		}
	}
}

// TestSameSeedSameSector: a sector reproduces from its one recorded seed
// like everything else this tool writes.
func TestSameSeedSameSector(t *testing.T) {
	t.Parallel()

	golden := fixture.SectorGolden()
	first, second := sector(t, golden), sector(t, golden)

	if len(first.Worlds) != len(second.Worlds) || len(first.Routes) != len(second.Routes) {
		t.Fatalf("two runs of seed %d gave %d/%d worlds and %d/%d routes",
			golden.Seed, len(first.Worlds), len(second.Worlds), len(first.Routes), len(second.Routes))
	}

	for i := range first.Routes {
		if first.Routes[i] != second.Routes[i] {
			t.Fatalf("two runs of seed %d differ at route %d", golden.Seed, i)
		}
	}
}

// TestSectorRejectsADMTheBookDoesNotOffer: a sector validates its inputs
// before it generates sixteen subsectors, not after.
func TestSectorRejectsADMTheBookDoesNotOffer(t *testing.T) {
	t.Parallel()

	engine, err := gen.New()
	if err != nil {
		t.Fatal(err)
	}

	for _, dm := range []int{-2, 2} {
		_, err := engine.Sector(gen.Inputs{Seed: 1, Name: "Aramis", OccurrenceDM: dm})
		if err == nil {
			t.Errorf("Sector accepted an occurrence DM of %+d, and p. 1 offers -1, 0 and +1", dm)
		}
	}
}

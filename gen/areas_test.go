package gen_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/philoserf/ctworldgen/dice"
	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/starmap"
)

// wholeGrid is the p. 3 grid as one broad area, which is the whole
// subsector written the long way (ERRATA E012 part 5).
func wholeGrid(dm int) []starmap.Area {
	return []starmap.Area{starmap.NewArea(
		starmap.Hex{Col: 1, Row: 1},
		starmap.Hex{Col: starmap.Columns, Row: starmap.Rows},
		dm,
	)}
}

// TestAWholeGridAreaIsTheWholeGridDM is the property this milestone turns
// on, stated directly: a broad area varies the number the occurrence throw
// is read against and nothing else, so covering the whole grid at a DM
// produces exactly what that DM produces on the whole subsector.
//
// It is the regression the record shape had to be built around. The
// alternative implementation -- scanning the areas in turn -- passes this
// one, because a single area covering the grid is visited in grid order
// either way; TestTheOccurrenceThrowIsOneDiePerHexInGridOrder is what
// refuses that.
func TestAWholeGridAreaIsTheWholeGridDM(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)

	for seed := range uint64(40) {
		for _, modifier := range []int{-1, 0, 1} {
			plain := generate(t, engine, gen.Inputs{
				Seed: seed, Name: aramis, OccurrenceDM: modifier, OccurrenceAreas: nil,
			})
			byArea := generate(t, engine, gen.Inputs{
				Seed: seed, Name: aramis, OccurrenceDM: 0, OccurrenceAreas: wholeGrid(modifier),
			})

			// One reading governed the second record and not the first,
			// and it is the only stamp that may differ between them.
			if !slices.Equal(byArea.Errata, insert(plain.Errata, "E012")) {
				t.Fatalf("seed %d at DM %+d: whole-subsector stamped %v and one whole-grid area stamped %v",
					seed, modifier, plain.Errata, byArea.Errata)
			}

			// Everything else the two carry is compared as written, which
			// is the whole record rather than whichever fields a hand-written
			// comparison remembered.
			byArea.OccurrenceDM, byArea.OccurrenceAreas, byArea.Errata = modifier, nil, plain.Errata

			if string(marshal(t, plain)) != string(marshal(t, byArea)) {
				t.Fatalf("seed %d at DM %+d: one whole-grid area did not write the whole-subsector record",
					seed, modifier)
			}
		}
	}
}

// TestTheOccurrenceThrowIsOneDiePerHexInGridOrder pins ERRATA E012 part 5
// against the dice stream itself.
//
// It draws the eighty dice the scan draws, in the ascending grid number of
// E002, and reads each against the DM of the area that hex falls in --
// reconstructing the DM from starmap.Area.Contains rather than from the
// engine's own lookup, which is unexported and would in any case be the
// code under test checking itself.
//
// This is the assertion that refuses scanning the areas in turn. Under
// that implementation the hexes are visited 0101, 0201 ... 0801, 0102 ...
// for the first area and then the second, so the die each hex draws is a
// different die, and the record it produces disagrees with this
// reconstruction at the first hex where the two orders part.
func TestTheOccurrenceThrowIsOneDiePerHexInGridOrder(t *testing.T) {
	t.Parallel()

	golden := fixture.BroadAreasGolden()
	if len(golden.OccurrenceAreas) == 0 {
		t.Fatal("the broad-areas fixture carries no areas, so this test would hold nothing")
	}

	record := generate(t, newEngine(t), gen.Inputs{
		Seed: golden.Seed, Name: golden.Name,
		OccurrenceDM: golden.OccurrenceDM, OccurrenceAreas: golden.OccurrenceAreas,
	})

	placed := map[starmap.Hex]bool{}
	for _, world := range record.Worlds {
		placed[world.Hex] = true
	}

	// The same stream the engine drew, read from the top: the occurrence
	// scan is the first thing that consumes it.
	stream := dice.NewStream(golden.Seed)

	// The occurrence throw of p. 1: mark the hex on a 4, 5 or 6, which is a
	// target of 4+ (ERRATA E002's noted discrepancy).
	const target = 4

	for col := 1; col <= starmap.Columns; col++ {
		for row := 1; row <= starmap.Rows; row++ {
			hex := starmap.Hex{Col: col, Row: row}

			modifier := golden.OccurrenceDM

			for _, area := range golden.OccurrenceAreas {
				if area.Contains(hex) {
					modifier = area.DM
				}
			}

			want := stream.Die()+modifier >= target
			if placed[hex] != want {
				t.Fatalf("hex %s: the record %s a world there, and one die at DM %+d %s one",
					hex, was(placed[hex]), modifier, was(want))
			}
		}
	}
}

// insert puts an erratum into a list already in document order, the way
// Record.Stamp does, so the expected list is built rather than typed.
func insert(errata []string, id string) []string {
	with := append(slices.Clone(errata), id)
	slices.Sort(with)

	return with
}

func was(placed bool) string {
	if placed {
		return "has"
	}

	return "has no"
}

// TestBroadAreasChangeWhereTheWorldsAre is the fixture's own point: the
// banded areas make the top half of the grid poorer than the bottom, which
// is what "a rift in one corner and a cluster in the other" means.
//
// It is stated as strict nesting rather than as a count, and the
// difference matters. The occurrence scan draws the same eighty dice
// whatever the DM (ERRATA E012 part 5), so a hex is placed at -1 only if
// it is placed at 0, and at 0 only if it is placed at +1 -- the star
// fields nest. A count comparison would say "fewer worlds up here than
// down there", which for any one seed is very nearly a coin toss and
// passes just as happily when the areas are ignored altogether.
//
// Strictly nested is the property the areas actually have, and it is
// exactly equality when they are ignored.
func TestBroadAreasChangeWhereTheWorldsAre(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)
	golden := fixture.BroadAreasGolden()

	plain := generate(t, engine, gen.Inputs{
		Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM, OccurrenceAreas: nil,
	})
	banded := generate(t, engine, gen.Inputs{
		Seed: golden.Seed, Name: golden.Name,
		OccurrenceDM: golden.OccurrenceDM, OccurrenceAreas: golden.OccurrenceAreas,
	})

	// The rift band is at -1, so its worlds are strictly fewer; the cluster
	// band is at +1, so strictly more. Which band a hex is in is read off
	// the fixture's own areas rather than off a row number written here.
	rift, cluster := golden.OccurrenceAreas[0], golden.OccurrenceAreas[1]

	assertStrictlyInside(t, "the rift band at -1", in(banded, rift), in(plain, rift))
	assertStrictlyInside(t, "the cluster band at +1", in(plain, cluster), in(banded, cluster))
}

// in is the set of hexes a record placed a world on inside one area.
func in(record *starmap.Record, area starmap.Area) map[starmap.Hex]bool {
	found := map[starmap.Hex]bool{}

	for _, world := range record.Worlds {
		if area.Contains(world.Hex) {
			found[world.Hex] = true
		}
	}

	return found
}

// assertStrictlyInside holds one set inside another and refuses equality,
// which is what a DM that governed nothing would produce.
func assertStrictlyInside(t *testing.T, what string, fewer, more map[starmap.Hex]bool) {
	t.Helper()

	for hex := range fewer {
		if !more[hex] {
			t.Errorf("%s: %s carries a world the run at the higher DM does not; the star fields do not nest", what, hex)
		}
	}

	if len(fewer) >= len(more) {
		t.Errorf("%s holds %d worlds against %d at the DM one step higher; a broad area's DM told nothing",
			what, len(fewer), len(more))
	}
}

// TestRejectsAreasTheBookAndTheReadingDoNotOffer holds Inputs.Validate to
// ERRATA E012: a DM the page does not offer, corners that are not the low
// and the high hex, a corner off the p. 3 grid, and two areas sharing a
// hex.
//
// The off-grid case is the one that looks redundant and is not: a hex is
// four digits whether it names a sub-sector or a sector, so 0910 parses
// and only a grid refuses it.
func TestRejectsAreasTheBookAndTheReadingDoNotOffer(t *testing.T) {
	t.Parallel()

	hex := func(col, row int) starmap.Hex { return starmap.Hex{Col: col, Row: row} }

	for name, areas := range map[string][]starmap.Area{
		"a DM of +2": {starmap.NewArea(hex(1, 1), hex(4, 10), 2)},
		"a DM of -2": {starmap.NewArea(hex(1, 1), hex(4, 10), -2)},
		"corners the wrong way round": {
			{From: hex(4, 10), To: hex(1, 1), DM: -1},
		},
		"a corner off the p. 3 grid": {starmap.NewArea(hex(1, 1), hex(9, 10), -1)},
		"a row off the p. 3 grid":    {starmap.NewArea(hex(1, 1), hex(8, 11), -1)},
		"two areas sharing a hex": {
			starmap.NewArea(hex(1, 1), hex(4, 10), -1),
			starmap.NewArea(hex(4, 1), hex(8, 10), 1),
		},
		"an area inside another": {
			starmap.NewArea(hex(1, 1), hex(8, 10), -1),
			starmap.NewArea(hex(3, 3), hex(4, 4), 1),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := newEngine(t).Generate(gen.Inputs{
				Seed: 1, Name: "", OccurrenceDM: 0, OccurrenceAreas: areas,
			})
			if err == nil {
				t.Errorf("%s was accepted", name)
			}
		})
	}
}

// TestAdjacentAreasDoNotOverlap is the guard on the test above: a rule
// that refused everything would pass every case in it.
func TestAdjacentAreasDoNotOverlap(t *testing.T) {
	t.Parallel()

	_, err := newEngine(t).Generate(gen.Inputs{
		Seed: 1, Name: "", OccurrenceDM: 0,
		OccurrenceAreas: []starmap.Area{
			starmap.NewArea(starmap.Hex{Col: 1, Row: 1}, starmap.Hex{Col: 4, Row: 10}, -1),
			starmap.NewArea(starmap.Hex{Col: 5, Row: 1}, starmap.Hex{Col: 8, Row: 10}, 1),
		},
	})
	if err != nil {
		t.Errorf("two areas that touch but share no hex were refused: %v", err)
	}
}

// TestASectorTakesNoBroadAreas holds the scope of ERRATA E012: an area is
// a rectangle of one grid's numbering, and a sector's sixteen members are
// each generated on their own p. 3 grid (ERRATA E006 part 1).
//
// Refused rather than dropped. A sector that ignored them would carry a
// DM in its record that governed no throw in it.
func TestASectorTakesNoBroadAreas(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)

	// A p. 3 rectangle, which Validate would accept.
	_, err := engine.Sector(gen.Inputs{
		Seed: 1, Name: aramis, OccurrenceDM: 0, OccurrenceAreas: wholeGrid(-1),
	})
	if !errors.Is(err, gen.ErrSectorTakesNoAreas) {
		t.Errorf("a p. 3 area on a sector gave %v; want ErrSectorTakesNoAreas", err)
	}

	// And a sector-grid rectangle, which it would not. The refusal has to
	// come first, or a referee asking for something the tool does not build
	// is told his area is off an 8x10 grid -- a true sentence about the
	// wrong thing, on a record that is 32x40.
	_, err = engine.Sector(gen.Inputs{
		Seed: 1, Name: aramis, OccurrenceDM: 0,
		OccurrenceAreas: []starmap.Area{starmap.NewArea(
			starmap.Hex{Col: 1, Row: 1},
			starmap.Hex{Col: starmap.SectorColumns, Row: starmap.SectorRows}, -1)},
	})
	if !errors.Is(err, gen.ErrSectorTakesNoAreas) {
		t.Errorf("a sector-grid area on a sector gave %v; want ErrSectorTakesNoAreas", err)
	}
}

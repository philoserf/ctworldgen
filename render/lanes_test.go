package render_test

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/render"
	"github.com/philoserf/ctworldgen/starmap"
)

// routeRow identifies one lane by its two ends, which is the key these
// checks compare on. The two cells rather than the whole row: a referee's
// name for a world lands in them too, and the distance cell is not what
// identifies a lane.
func routeRow(route starmap.Route) string {
	return route.From.String() + "->" + route.To.String()
}

// drawnRows is the set of lanes a listing's route table actually carries,
// read back out of the document in the same form routeRow writes.
func drawnRows(t *testing.T, record *starmap.Record, lanes render.Lanes) map[string]bool {
	t.Helper()

	table := section(t, listingWith(t, record, lanes), "Routes")
	rows := map[string]bool{}

	for line := range strings.SplitSeq(table, "\n") {
		if !strings.HasPrefix(line, "| ") {
			continue
		}

		cells := strings.Split(strings.Trim(line, "| "), " | ")
		if len(cells) != 3 {
			continue
		}

		// The hex is the first token of each end, whether or not the
		// referee wrote a name after it.
		from, _, _ := strings.Cut(cells[0], " ")
		until, _, _ := strings.Cut(cells[1], " ")

		_, err := starmap.ParseHex(from)
		if err != nil {
			continue // the header row, or the separator
		}

		rows[from+"->"+until] = true
	}

	return rows
}

// TestTheDefaultListingDrawsLegibleLanes holds all four halves of the
// filtering, because three of them can pass while the thing is broken.
//
// That a drawn lane has a row says nothing about whether anything was
// suppressed -- a pass that suppresses nothing satisfies it. That a
// suppressed lane has no row is the half that proves the filtering
// happened, and the counts are what stop a lane going missing from both
// sides at once.
func TestTheDefaultListingDrawsLegibleLanes(t *testing.T) {
	t.Parallel()

	suppressedSomewhere := 0

	// One loop rather than subtests: the tally is read after it, and a
	// parallel subtest does not run until the parent returns -- so the
	// "nothing was suppressed anywhere" guard would read a zero that means
	// nothing and pass. Each message names its fixture instead.
	for _, golden := range fixture.Goldens() {
		{
			record := generated(t, golden)
			drawn := drawnRows(t, record, render.LegibleLanes)
			all := drawnRows(t, record, render.AllLanes)

			if len(all) != len(record.Routes) {
				t.Fatalf("%s: --lanes=all drew %d of the record's %d lanes", golden.File, len(all), len(record.Routes))
			}

			suppressed := 0

			for _, route := range record.Routes {
				row := routeRow(route)
				if !all[row] {
					t.Fatalf("%s: --lanes=all has no row for %s-%s", golden.File, route.From, route.To)
				}

				if !drawn[row] {
					suppressed++
				}
			}

			// Nothing vanished: every lane is either drawn or suppressed.
			if len(drawn)+suppressed != len(record.Routes) {
				t.Errorf("%s: %d drawn plus %d suppressed is not the record's %d lanes",
					golden.File, len(drawn), suppressed, len(record.Routes))
			}

			// And nothing was invented: the drawn set is a subset.
			for row := range drawn {
				if !all[row] {
					t.Errorf("%s: the listing drew a lane the record does not carry: %s", golden.File, row)
				}
			}

			suppressedSomewhere += suppressed
		}
	}

	if suppressedSomewhere == 0 {
		t.Error("no lane was suppressed in any fixture, so the filtering proved nothing")
	}
}

// TestLegibleLanesDoNotDependOnRouteOrder is the check that tells this
// reading apart from the greedy one, and the only one that does.
//
// Both rules keep a lane whose ends are not yet joined. The difference is
// when they join: a whole distance at a time, or as each lane is examined.
// Greedy draws fewer lanes and looks better, and which of several equal
// lanes it keeps depends on the order it met them -- so the drawn map stops
// being a function of the record and starts being a function of a sort
// order nobody promised (ERRATA E007).
func TestLegibleLanesDoNotDependOnRouteOrder(t *testing.T) {
	t.Parallel()

	// The densest fixture: the more equal-length lanes there are, the more
	// chances greedy has to answer differently.
	record := generated(t, fixture.Goldens()[2])
	want := drawnRows(t, record, render.LegibleLanes)

	if len(want) == len(record.Routes) {
		t.Fatal("this fixture suppresses nothing, so shuffling it proves nothing")
	}

	shuffled := *record

	shuffled.Routes = append([]starmap.Route(nil), record.Routes...)

	source := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // a shuffle in a test, not a secret

	for range 25 {
		source.Shuffle(len(shuffled.Routes), func(i, j int) {
			shuffled.Routes[i], shuffled.Routes[j] = shuffled.Routes[j], shuffled.Routes[i]
		})

		got := drawnRows(t, &shuffled, render.LegibleLanes)
		if len(got) != len(want) {
			t.Fatalf("a shuffled record drew %d lanes, want %d", len(got), len(want))
		}

		for row := range want {
			if !got[row] {
				t.Fatalf("a shuffled record did not draw %s", row)
			}
		}
	}
}

// TestLegibleLanesKeepEveryWorldReachable is the property a referee is
// trusting: suppressing a lane must never cut a world off from one it could
// reach. It holds by construction -- a lane is dropped only when its ends
// are already joined -- and is asserted because "by construction" is what
// the greedy version would also have claimed.
func TestLegibleLanesKeepEveryWorldReachable(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			record := generated(t, golden)

			all := components(record.Routes, record.Worlds)
			drawn := components(drawnRoutes(t, record), record.Worlds)

			if all != drawn {
				t.Errorf("the record's lanes leave %d groups of worlds and the drawn ones leave %d",
					all, drawn)
			}
		})
	}
}

// TestEveryJumpOneLaneIsDrawn pins the reading against a later
// simplification to the greedy form, which would drop a jump-1 lane whose
// ends another jump-1 pair had already joined. Nothing is shorter than one
// parsec, so nothing can make a jump-1 lane redundant (ERRATA E007).
func TestEveryJumpOneLaneIsDrawn(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		record := generated(t, golden)
		drawn := drawnRows(t, record, render.LegibleLanes)

		for _, route := range record.Routes {
			if route.Distance == 1 && !drawn[routeRow(route)] {
				t.Errorf("%s: the jump-1 lane %s-%s was suppressed", golden.File, route.From, route.To)
			}
		}
	}
}

// TestTheSummaryReportsWhatWasNotDrawn: a document that showed less than
// the record holds without saying so would be worse than an unreadable one.
func TestTheSummaryReportsWhatWasNotDrawn(t *testing.T) {
	t.Parallel()

	record := generated(t, fixture.Goldens()[2])
	drawn := len(drawnRows(t, record, render.LegibleLanes))

	if drawn == len(record.Routes) {
		t.Fatal("this fixture suppresses nothing, so the summary has nothing to report")
	}

	written := listing(t, record)
	if !strings.Contains(written, "routes, "+strconv.Itoa(drawn)+" drawn.") {
		t.Errorf("the listing's summary does not say how many lanes it drew:\n%s", head(written))
	}

	if !strings.Contains(written, "ERRATA E007") {
		t.Error("the listing suppresses lanes and does not cite the reading that lets it")
	}

	// And when nothing is suppressed, the sentence is the one it always was.
	plain := listingWith(t, record, render.AllLanes)
	if strings.Contains(plain, " drawn.") {
		t.Errorf("--lanes=all reported a drawn count:\n%s", head(plain))
	}
}

// drawnRoutes reads the record's own routes back out of the default
// listing, so the reachability check measures what the document drew rather
// than what a second copy of the rule says it should have.
func drawnRoutes(t *testing.T, record *starmap.Record) []starmap.Route {
	t.Helper()

	rows := drawnRows(t, record, render.LegibleLanes)
	drawn := make([]starmap.Route, 0, len(rows))

	for _, route := range record.Routes {
		if rows[routeRow(route)] {
			drawn = append(drawn, route)
		}
	}

	return drawn
}

// components counts the groups of worlds the lanes leave, an isolated world
// counting as its own.
func components(routes []starmap.Route, worlds []starmap.World) int {
	group := map[starmap.Hex]starmap.Hex{}

	var root func(starmap.Hex) starmap.Hex

	root = func(hex starmap.Hex) starmap.Hex {
		parent, known := group[hex]
		if !known || parent == hex {
			group[hex] = hex

			return hex
		}

		found := root(parent)

		group[hex] = found

		return found
	}

	for _, world := range worlds {
		group[world.Hex] = world.Hex
	}

	for _, route := range routes {
		group[root(route.From)] = root(route.To)
	}

	seen := map[starmap.Hex]bool{}
	for _, world := range worlds {
		seen[root(world.Hex)] = true
	}

	return len(seen)
}

// head is the first few lines of a document, for a failure message.
func head(written string) string {
	lines := strings.SplitN(written, "\n", 8)
	if len(lines) > 7 {
		lines = lines[:7]
	}

	return strings.Join(lines, "\n")
}

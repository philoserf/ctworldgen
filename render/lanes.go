package render

import (
	"github.com/philoserf/ctworldgen/starmap"
	"github.com/philoserf/ctworldgen/tables"
)

// Lanes says which of a record's commercial routes the documents draw.
//
// The record carries every one of them either way. This is ink, not dice:
// the engine examined every pair and consumed every die it should have
// (ERRATA E003), and nothing here changes what a seed produces.
type Lanes int

const (
	// LegibleLanes draws only the lanes that join something, which is what
	// p. 2 offers the map-drawer (ERRATA E007). It is the default because a
	// dense subsector draws a hundred and sixty lanes over forty-six
	// worlds, and that is not a map a referee can read.
	LegibleLanes Lanes = iota

	// AllLanes draws every lane the record carries.
	AllLanes
)

// legible returns the lanes a map draws when it is drawn to be read: those
// whose two worlds are not already joined by shorter ones (p. 2, ERRATA
// E007).
//
// The returned slice keeps the record's own order, which is E003's, so the
// route table reads the same way it always did with fewer rows in it.
//
// The layering is the whole rule and it is not an optimisation. A distance
// is examined in full against what the shorter distances joined, and only
// then does it join anything itself. Union each lane as it is examined
// instead -- the ordinary greedy spanning forest -- and the result stops
// being a function of the record: which of several equal-length lanes
// survives becomes whichever the loop reached first. It would look right,
// it would draw fewer lanes, and the drawn map would quietly depend on the
// order the routes happen to sit in.
func legible(routes []starmap.Route) []starmap.Route {
	joined := newGroups()
	keep := make(map[starmap.Route]bool, len(routes))

	for distance := starmap.Parsecs(1); distance <= tables.MaxJump; distance++ {
		layer := make([]starmap.Route, 0, len(routes))

		for _, route := range routes {
			if route.Distance == distance {
				layer = append(layer, route)
			}
		}

		// Examined against what the shorter lanes joined ...
		for _, route := range layer {
			if !joined.same(route.From, route.To) {
				keep[route] = true
			}
		}

		// ... and only then joining anything itself.
		for _, route := range layer {
			joined.merge(route.From, route.To)
		}
	}

	drawn := make([]starmap.Route, 0, len(keep))

	for _, route := range routes {
		if keep[route] {
			drawn = append(drawn, route)
		}
	}

	return drawn
}

// groups is the disjoint-set forest legible reads "already joined" from.
// Small enough to write out: the standard library has none, and a sector's
// six hundred and seventy worlds do not want anything cleverer.
type groups struct{ parent map[starmap.Hex]starmap.Hex }

func newGroups() *groups { return &groups{parent: map[starmap.Hex]starmap.Hex{}} }

// root returns the representative of a hex's group, flattening the path it
// walked so the next walk is shorter.
func (g *groups) root(hex starmap.Hex) starmap.Hex {
	_, known := g.parent[hex]
	if !known {
		g.parent[hex] = hex
	}

	for g.parent[hex] != hex {
		g.parent[hex] = g.parent[g.parent[hex]]
		hex = g.parent[hex]
	}

	return hex
}

// same reports whether two hexes are already joined.
func (g *groups) same(a, b starmap.Hex) bool { return g.root(a) == g.root(b) }

// merge joins two hexes' groups.
func (g *groups) merge(a, b starmap.Hex) {
	rootA, rootB := g.root(a), g.root(b)
	if rootA != rootB {
		g.parent[rootA] = rootB
	}
}

package render

import (
	"fmt"

	"github.com/philoserf/ctworldgen/starmap"
)

// member is one of a sector's sixteen sub-sectors, gathered out of the
// record the sector wrote.
//
// Nothing here is decided. Which member a world or a lane belongs to is
// read straight off the hex, because starmap.Place put member i's hexes in
// band i and starmap.MemberOf reads the band back (ERRATA E006 part 1).
// The record is not changed and carries no member field: a sector that had
// to remember its own decomposition could disagree with its own grid.
type member struct {
	Index int
	First starmap.Hex
	Last  starmap.Hex

	Worlds []starmap.World

	// Carried is every lane of the record with an end in this member, and
	// Lanes is the subset the documents draw (ERRATA E007). A lane that
	// crosses a seam is in both its members' lists, because a referee
	// reading one sub-sector needs to see the road out of it (ERRATA E008
	// part 3) -- so the sixteen tables together list more rows than the
	// record has lanes, and the section that carries them says so.
	Carried []starmap.Route
	Lanes   []starmap.Route

	// Crossing counts the lanes of Carried with an end in another member.
	Crossing int

	// OffMap counts the drawn lanes this member's map cannot draw. A lane
	// runs up to four parsecs (p. 2) and the map reaches one hex past the
	// member's border, so a long lane out of the sub-sector has no far end
	// on the sheet. It is in the lane table with the sub-sector it leaves
	// for named, and the map's note says how many (ERRATA E008 part 2).
	OffMap int
}

// members gathers the sixteen out of the record, in index order.
func members(record *starmap.Record, drawn []starmap.Route) []member {
	gathered := make([]member, starmap.Members)

	for index := range gathered {
		first, last := starmap.MemberBounds(index)

		gathered[index] = member{
			Index: index, First: first, Last: last,
			Worlds: nil, Carried: nil, Lanes: nil, Crossing: 0, OffMap: 0,
		}
	}

	for _, world := range record.Worlds {
		at := starmap.MemberOf(world.Hex)

		gathered[at].Worlds = append(gathered[at].Worlds, world)
	}

	for _, route := range record.Routes {
		home, other, crosses := ends(route)

		gathered[home].Carried = append(gathered[home].Carried, route)

		if crosses {
			gathered[home].Crossing++

			gathered[other].Carried = append(gathered[other].Carried, route)
			gathered[other].Crossing++
		}
	}

	for _, route := range drawn {
		home, other, crosses := ends(route)

		gathered[home].hold(route)

		if crosses {
			gathered[other].hold(route)
		}
	}

	return gathered
}

// hold adds a drawn lane to a member, counting it as one its map cannot
// draw when either end falls outside the hexes that map shows.
func (m *member) hold(route starmap.Route) {
	m.Lanes = append(m.Lanes, route)

	if !m.shows(route.From) || !m.shows(route.To) {
		m.OffMap++
	}
}

// shows reports whether a hex is on this member's map (ERRATA E008 part 2).
func (m *member) shows(hex starmap.Hex) bool { return onMemberMap(m.Index, hex) }

// ends returns the members a lane's two ends sit in, and whether they
// differ. A lane that crosses is carried by both (ERRATA E008 part 3).
func ends(route starmap.Route) (int, int, bool) {
	from := starmap.MemberOf(route.From)
	to := starmap.MemberOf(route.To)

	return from, to, from != to
}

// memberSeed is the seed that writes a member on its own: the sector's
// base plus the member's index (ERRATA E006 part 1). The index is one of
// sixteen, so the conversion cannot overflow, and the bound is asserted in
// code rather than in a comment.
//
// The guard cannot fire: every call site is an exhaustive loop over the
// sixteen. It stays because it is the bounds proof gosec's G115 accepts
// for the int-to-uint64 conversion, and a version without it fails the
// gate on a finding answerable only by a lint disable. Its fallback
// returns the sector's base, which is member 0's seed and so is itself a
// wrong seed -- but a renderer has no error path out, and no uint64 reads
// as "no seed". If the call sites ever stop being exhaustive, the answer
// is an error, not a better constant.
func memberSeed(base uint64, index int) uint64 {
	if index < 0 || index >= starmap.Members {
		return base
	}

	return base + uint64(index)
}

// provenance is the line that makes a member section usable on its own:
// the seed that writes this sub-sector standalone, which is the identity
// E006 part 1 keeps and the README already promises. A referee who wants
// this one sub-sector and nothing else can generate it from here.
func (m *member) provenance(seed uint64) string {
	return fmt.Sprintf(
		"Member %d of this sector. `ctworldgen new --seed %d` writes this sub-sector on its own "+
			"p. 3 grid, where its first hex is 0101; here it is laid on the sector's, and that hex "+
			"is %s (ERRATA E006).",
		m.Index, memberSeed(seed, m.Index), m.First,
	)
}

// summary counts this member the way the document's own heading counts the
// sector, and separates the lanes that stay at home from the ones that
// leave, because the second kind is what a sector adds.
func (m *member) summary() string {
	return fmt.Sprintf("%d worlds, %d lanes within it, %d crossing into its neighbours.",
		len(m.Worlds), len(m.Carried)-m.Crossing, m.Crossing)
}

// mapNote heads a member's own map. It says what the ring of neighbouring
// hexes is for and, where there are any, how many drawn lanes reach past
// it -- because a map that quietly stopped drawing some of its lanes would
// be worse than one that drew none (ERRATA E008 part 2).
func (m *member) mapNote() string {
	note := "This sub-sector's eighty hexes, numbered where the sector lays them, with one hex of " +
		"each neighbour drawn around the edge, so that a lane crossing a seam to an adjacent hex " +
		"has a far end that can be looked up (ERRATA E008 part 2)."

	if m.OffMap > 0 {
		note += fmt.Sprintf(" A lane runs up to four parsecs (p. 2) and this map reaches one, so %d "+
			"of its lanes end past the edge and are drawn here as no line; the table below carries "+
			"them, and names the sub-sector each one leaves for.", m.OffMap)
	}

	return note + mapNoteTail
}

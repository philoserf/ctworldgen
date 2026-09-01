package gen

import (
	"fmt"
	"slices"

	"github.com/philoserf/ctworldgen/dice"
	"github.com/philoserf/ctworldgen/starmap"
)

// SectorMembers is how many subsectors a sector is: four across and four
// down (ERRATA E006).
const SectorMembers = starmap.SectorAcross * starmap.SectorAcross

// seamSeedOffset is the seventeenth stream, after the sixteen the members
// consume (ERRATA E006 part 4).
const seamSeedOffset = SectorMembers

// Sector assembles sixteen subsectors on one grid and throws for the
// routes at their seams.
//
// The book charts a subsector and stops, but its route rule speaks of a
// world's "neighbors" rather than of a subsector, so the border is an
// artifact of generating one subsector at a time (ERRATA E006). Every
// member is generated whole and unchanged -- member i of a sector is the
// subsector `new --seed base+i` writes -- and the only throws made here
// are for pairs that straddle two members.
func (e *Engine) Sector(inputs Inputs) (*starmap.Record, error) {
	err := inputs.Validate()
	if err != nil {
		return nil, err
	}

	record := starmap.New(inputs.Seed, inputs.Name, inputs.OccurrenceDM)

	record.Grid = starmap.SectorGrid()

	// The members consume the streams of seeds base through base+15, in
	// order (ERRATA E006 part 4). Counting the seed alongside the index
	// keeps the derivation in one place and out of a conversion.
	seed := inputs.Seed

	for index := range SectorMembers {
		err := e.member(record, inputs, index, seed)
		if err != nil {
			return nil, err
		}

		seed++
	}

	// E002's order, now read across the whole grid.
	slices.SortFunc(record.Worlds, func(a, b starmap.World) int { return a.Hex.Number() - b.Hex.Number() })

	record.Routes = append(record.Routes,
		e.seams(dice.NewStream(inputs.Seed+seamSeedOffset), record.Worlds)...)
	slices.SortFunc(record.Routes, routeOrder)

	record.Stamp("E006")

	return record, nil
}

// member generates one subsector whole and lays it on the sector grid.
func (e *Engine) member(record *starmap.Record, inputs Inputs, index int, seed uint64) error {
	part, err := e.Generate(Inputs{Seed: seed, Name: inputs.Name, OccurrenceDM: inputs.OccurrenceDM})
	if err != nil {
		return fmt.Errorf("member %d of the sector: %w", index, err)
	}

	// A sector carries every reading its members carried: the members are
	// the record's worlds (ERRATA E006).
	for _, id := range part.Errata {
		record.Stamp(id)
	}

	for _, world := range part.Worlds {
		world.Hex = Place(index, world.Hex)
		if !record.Grid.Contains(world.Hex) {
			return fmt.Errorf("%w: member %d puts a world at %s", starmap.ErrOffGrid, index, world.Hex)
		}

		record.Worlds = append(record.Worlds, world)
	}

	for _, route := range part.Routes {
		record.Routes = append(record.Routes, starmap.Route{
			From: Place(index, route.From), To: Place(index, route.To), Distance: route.Distance,
		})
	}

	return nil
}

// seams throws for the pairs the members could not examine: those whose
// two worlds sit in different members (ERRATA E006 part 3). An interior
// pair was examined inside its member, and p. 2 examines each pair once.
//
// Everything else is E003 unchanged -- the same table, one die against a
// stated number, no row for an X starport and no throw at a dash cell,
// and the same order, read now in sector coordinates.
//
// Which member a world came from is not remembered: Place puts member i's
// hexes in band i, so MemberOf reads the band straight back off the hex
// (ERRATA E006 part 1). That is the same fact the translation states, held
// in one place rather than two that must agree.
func (e *Engine) seams(stream *dice.Stream, worlds []starmap.World) []starmap.Route {
	routes := []starmap.Route{}

	for index, first := range worlds {
		// Hoisted: the first world's band is fixed for the whole inner
		// loop, and deriving it per pair is a division per pair over the
		// two hundred thousand a sector has.
		firstMember := MemberOf(first.Hex)

		for _, second := range worlds[index+1:] {
			if firstMember == MemberOf(second.Hex) {
				continue
			}

			distance := first.Hex.Distance(second.Hex)

			target, stated := e.charts.JumpRoutes.Target(first.Starport, second.Starport, distance)
			if !stated {
				continue
			}

			if target.Met(stream.Die()) {
				routes = append(routes, starmap.Route{From: first.Hex, To: second.Hex, Distance: distance})
			}
		}
	}

	return routes
}

// routeOrder puts the routes in E003's order over the whole grid, so the
// listing reads the same way whether a route was thrown inside a member or
// at a seam.
func routeOrder(a, b starmap.Route) int {
	if a.From != b.From {
		return a.From.Number() - b.From.Number()
	}

	return a.To.Number() - b.To.Number()
}

// Place translates a member's local hex onto the sector grid: member
// index sits at column band index mod 4 and row band index div 4, and a
// local hex moves by whole bands (ERRATA E006 parts 1 and 2).
//
// A sub-sector is eight columns wide and eight is even, so a column's
// odd-or-even parity survives this -- which is what makes an interior
// pair measure the same distance on the sector grid as it did at home.
// It is exported because that property is worth asserting against the
// translation the engine actually uses, rather than against a second copy
// of the arithmetic written in a test.
func Place(index int, hex starmap.Hex) starmap.Hex {
	across := index % starmap.SectorAcross
	down := index / starmap.SectorAcross

	return starmap.Hex{Col: across*starmap.Columns + hex.Col, Row: down*starmap.Rows + hex.Row}
}

// MemberOf returns which of a sector's sixteen subsectors a hex fell in,
// numbered left to right and then down (ERRATA E006 part 1).
func MemberOf(hex starmap.Hex) int {
	across := (hex.Col - 1) / starmap.Columns
	down := (hex.Row - 1) / starmap.Rows

	return down*starmap.SectorAcross + across
}

// CrossingRoutes returns the routes whose two ends sit in different
// members: the ones a referee generating sixteen subsectors one at a time
// could never have found. It is exported because it is what the sector
// golden pins and what the listing would highlight.
func CrossingRoutes(record *starmap.Record) []starmap.Route {
	crossing := []starmap.Route{}

	for _, route := range record.Routes {
		if MemberOf(route.From) != MemberOf(route.To) {
			crossing = append(crossing, route)
		}
	}

	return crossing
}

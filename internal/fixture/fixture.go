// Package fixture holds the one definition of the golden-fixture roster,
// so that two golden trees cannot come to describe different subsectors
// under the same name. It is test support, not architecture.
package fixture

import (
	"fmt"
	"path/filepath"
	"strings"
)

// aramis is the subsector name the roster uses throughout.
const aramis = "Aramis"

// Golden names one fixture and the inputs that reproduce it.
type Golden struct {
	// File is the golden's base name, without extension.
	File string

	Seed         uint64
	Name         string
	OccurrenceDM int
}

// Goldens is the roster: the three occurrence DMs the book offers, plus
// the seed 0 case, which is an explicit and distinct choice rather than a
// request for a random seed.
//
// There is deliberately no empty-subsector golden. An empty subsector is
// a valid result, but eighty throws at the worst DM the book offers make
// one about a hundred-trillion-to-one; the case is covered instead by the
// minimal example record, which schema validation checks.
func Goldens() []Golden {
	return []Golden{
		{File: "dm-minus-one", Seed: 1, Name: aramis, OccurrenceDM: -1},
		{File: "dm-zero", Seed: 1, Name: aramis, OccurrenceDM: 0},
		{File: "dm-plus-one", Seed: 1, Name: aramis, OccurrenceDM: 1},
		{File: "seed-zero", Seed: 0, Name: "", OccurrenceDM: 0},
	}
}

// SectorGolden is the sector fixture. There is deliberately no golden of
// the whole 670-world record: at 351K it would be larger than every other
// fixture together, and it would pin almost nothing new. A sector's
// members are the subsectors `new --seed base+i` already writes, which a
// test compares directly, so the only thing a sector adds is the route
// pass at the seams -- and that is what SeamsPath pins.
func SectorGolden() Golden {
	return Golden{File: "sector-seams", Seed: 1, Name: aramis, OccurrenceDM: 0}
}

// SeamsPath is the golden of the routes that cross a member border,
// relative to the repository root. The name comes from SectorGolden so
// that the writer and the test that reads it cannot come to name two
// different files.
func SeamsPath() string {
	return filepath.Join("gen", "testdata", SectorGolden().File+".json")
}

// SectorSliceMember is which of a sector's sixteen the listing golden
// pins. An interior member, so its ring of neighbours is populated on all
// four sides and its lanes cross at every seam: a corner member would
// leave half the thing being pinned unexercised.
const SectorSliceMember = 5

// SectorSlicePath is the golden of one member's section of a sector's
// listing, relative to the repository root.
//
// One member and not the whole listing. A sector's listing is half a
// megabyte, and a golden exists so that a human reads its diff -- which
// nobody does at that size, and a golden nobody reads is a test that
// cannot fail wearing the costume of one that passes. A member's section
// is the size of the subsector goldens beside it, and it is where the
// decomposition, the ring and the doubled crossing lanes can be read.
//
// It catches drift and nothing else: it is regenerated from the code
// under test, so what holds the decomposition itself is the live
// assertions in render's tests.
func SectorSlicePath() string {
	return filepath.Join("render", "testdata", fmt.Sprintf("sector-member-%02d.md", SectorSliceMember))
}

// SectorSlice cuts the member section SectorSlicePath pins out of a whole
// sector listing. The writer and the reader of that golden both call it,
// so the two cannot come to cut different things.
func SectorSlice(listing string) string {
	heading := fmt.Sprintf("## Subsector %d ", SectorSliceMember)

	start := strings.Index(listing, heading)
	if start < 0 {
		return ""
	}

	section := listing[start:]
	if end := strings.Index(section[len(heading):], "\n## "); end >= 0 {
		section = section[:len(heading)+end+1]
	}

	return section
}

// CompleteExamplePath is the example record shipped beside the schema,
// relative to the repository root.
func CompleteExamplePath() string { return filepath.Join("docs", "examples", "complete.json") }

// exampleSeed is the year of the (c) 1977 text the ruleset names, chosen
// so the example's provenance is legible rather than arbitrary.
const exampleSeed = 1977

// CompleteExample is the inputs that produce that record. It is shipped as
// documentation but written by the engine, so it is regenerated and pinned
// exactly like a golden: an example that drifted from what `ctworldgen
// new` writes would document a record shape the tool does not produce.
func CompleteExample() Golden {
	return Golden{File: "complete", Seed: exampleSeed, Name: aramis, OccurrenceDM: -1}
}

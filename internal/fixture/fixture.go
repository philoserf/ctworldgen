// Package fixture holds the one definition of the golden-fixture roster,
// so that two golden trees cannot come to describe different subsectors
// under the same name. It is test support, not architecture.
package fixture

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

// Package fixture is the one definition of the golden-fixture set: the
// subsectors the worldgen and render packages both compare their output
// against, byte for byte.
//
// It exists so the set cannot be transcribed by hand in each of those two
// test packages. Nothing would detect drift between such copies: change a
// seed in one and both suites still pass — each matching its own testdata
// — while the two trees describe different subsectors under the same
// fixture name, and `task goldens` writes both.
//
// Test-only, but a regular package rather than a _test.go file: an
// external test package cannot import identifiers from another package's
// tests. Nothing in the command imports it, so it is not linked into the
// binary.
package fixture

import "github.com/philoserf/ctworldgen/worldgen"

// Fixture is one golden subsector: the inputs that generate it.
type Fixture struct {
	Name   string
	Config worldgen.Config
}

// All is the golden set: one subsector at each of the three world
// occurrence densities p. 1 allows. The DM is the only knob the procedure
// has, and it moves more than the world count — a denser subsector has
// quadratically more pairs to examine for space lanes, so the sparse and
// dense fixtures exercise very different amounts of the p. 2 table.
//
// Each uses its own seed rather than one shared seed at three DMs. Sharing
// would make the fixtures illustrate the DM nicely and test the dice
// poorly: the three runs would diverge after the first modified throw
// anyway, so the shared seed buys nothing and costs the extra coverage of
// three independent streams.
//
// A fresh slice each call: callers are tests, and one filtering or
// reordering its copy must not reach the others.
func All() []Fixture {
	return []Fixture{
		{Name: "standard", Config: worldgen.Config{Seed: 7, Name: "Vega", OccurrenceDM: 0}},
		{Name: "sparse", Config: worldgen.Config{Seed: 19, Name: "Outreach", OccurrenceDM: -1}},
		{Name: "dense", Config: worldgen.Config{Seed: 3, Name: "Core", OccurrenceDM: 1}},
	}
}

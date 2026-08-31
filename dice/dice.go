// Package dice implements the die roll conventions of Book 1 Characters
// and Combat pp. 2-3, drawn from a single seeded stream.
//
// Book 3's Worlds chapter uses one-die targets that the character
// generation procedure never does -- world occurrence is 4+ on one die
// (p. 1) and every cell of the jump routes table is a one-die target from
// 1 to 6 (p. 2) -- so a Target admits values below 2.
package dice

import "math/rand/v2"

// Algorithm names the random number generator a record stamps. It is the
// exact string the record carries, and it lives here because this file
// owns the construction it describes.
const Algorithm = "go-math-rand-v2-pcg"

// faces is the number of sides on a die (B1 pp. 2-3).
const faces = 6

// Stream is the single seeded source that every throw in a run is drawn
// from.
//
// Consumption order is load-bearing: it fixes what a seed means. Adding,
// removing or reordering a throw shifts every throw after it, which is
// why the two throws Book 3 p. 4 does not make are not made rather than
// made and discarded.
type Stream struct{ r *rand.Rand }

// NewStream seeds a PCG generator with the recorded seed in both words,
// so that the single seed a record carries reproduces the whole stream.
func NewStream(seed uint64) *Stream {
	// A deterministic generator is the point: a record must reproduce from
	// its recorded seed, which a cryptographic source cannot do.
	return &Stream{r: rand.New(rand.NewPCG(seed, seed))} //nolint:gosec // reproducibility, not secrecy
}

// Die throws one die.
//
// One die is exactly one IntN(6) draw plus one. This is as load-bearing
// as the consumption order it feeds: IntN(36), or a masked Uint64, would
// be the same generator under the same seed and would produce an entirely
// different subsector.
func (s *Stream) Die() int { return s.r.IntN(faces) + 1 }

// D2 throws two dice, the first die and then the second. B1 p. 2 makes
// two dice the unqualified throw.
func (s *Stream) D2() int { return s.Die() + s.Die() }

// Sum returns the cumulative DM of its arguments. Book 3 applies several
// DMs to a single throw -- the technological index matrix contributes six
// (p. 9) -- and they are summed into one total DM before the throw is
// read.
func Sum(dms ...int) int {
	total := 0
	for _, dm := range dms {
		total += dm
	}

	return total
}

// Target is a throw target of the form N+: the throw succeeds when it is
// equal to or greater than N.
//
// N+ is the only target kind Book 3 pp. 1-12 uses. World occurrence is
// 4+, the base throws are 7+ through 10+ (p. 5), and a jump routes cell
// states a number the one die throw must equal or exceed (p. 2).
type Target int

// Met reports whether a throw meets the target.
func (t Target) Met(throw int) bool { return throw >= int(t) }

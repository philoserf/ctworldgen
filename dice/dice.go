// Package dice is the rules' die-roll mechanics (Book 1 pp. 2-3): a seeded
// stream of six-sided dice, one- and two-die throws with cumulative DMs
// against N+/N-/exact targets. The stream is the only randomness in the
// program; every roll is consumed from it in procedure order, which makes
// that order load-bearing for replay (docs/PRD.md, Replay and provenance
// contract).
package dice

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
)

// Algorithm names the RNG in every subsector record's provenance block.
// Changing the algorithm (or how the seed expands into PCG state) is an
// engine version bump.
const Algorithm = "go-math-rand-v2-pcg"

// Stream is a seeded source of six-sided dice. It does not keep the seed:
// the record carries it (worldgen.RNG), and one copy of a value replay
// depends on is better than two that could disagree.
type Stream struct {
	rng *rand.Rand
}

// New returns a stream seeded with the given value. The single recorded
// seed fills both words of the PCG state, so the record's one seed field
// reproduces the stream exactly.
func New(seed uint64) *Stream {
	// The deterministic seeded stream is the contract (FR16; replay depends
	// on it), so the "weak" generator is the required one, not an oversight.
	return &Stream{rng: rand.New(rand.NewPCG(seed, seed))} // #nosec G404
}

// One rolls a single die (1-6).
func (s *Stream) One() int { return s.rng.IntN(6) + 1 }

// Two rolls two dice in order and reports them individually; throws
// record the dice, not just the sum (FR17).
func (s *Stream) Two() (int, int) { return s.One(), s.One() }

// Mode is how a target reads: 8+ means the total must equal or exceed 8,
// 8- means equal or less, and an exact target must match (B1 pp. 2-3).
type Mode int

// Target modes, in the notation of B1 pp. 2-3.
const (
	Plus Mode = iota
	Minus
	Exact
)

// Target is a throw's required result.
type Target struct {
	Value int
	Mode  Mode
}

// Met reports whether a throw total (dice plus DMs) satisfies the target.
func (t Target) Met(total int) bool {
	switch t.Mode {
	case Plus:
		return total >= t.Value
	case Minus:
		return total <= t.Value
	case Exact:
		return total == t.Value
	}

	return false
}

// ErrBadTarget reports throw-target notation that does not parse.
var ErrBadTarget = errors.New("bad throw target")

// Target values run 1-12, not the two-die 2-12 the character procedure
// needs. Book 3 throws one die as often as two: world occurrence is 4+ on
// one die (p. 1) and every cell of the jump routes table is a one-die
// target from 1 to 6 (p. 2). A 1 is a legal target for a one-die throw —
// the table's A-A jump-1 cell is exactly that, a lane that always exists —
// so the floor is 1 rather than 2.
const (
	minTarget = 1
	maxTarget = 12
)

// ParseTarget reads the book's notation: "8+", "8-", or an exact "12".
// It is the inverse of String, used by the data files' load-time
// validation.
func ParseTarget(s string) (Target, error) {
	if s == "" {
		return Target{}, fmt.Errorf("%w: empty", ErrBadTarget)
	}

	mode := Exact
	digits := s

	switch s[len(s)-1] {
	case '+':
		mode, digits = Plus, s[:len(s)-1]
	case '-':
		mode, digits = Minus, s[:len(s)-1]
	}

	value, err := strconv.Atoi(digits)
	if err != nil || value < minTarget || value > maxTarget {
		return Target{}, fmt.Errorf("%w: %q, want a die value %d-%d with optional +/-", ErrBadTarget, s, minTarget, maxTarget)
	}

	return Target{Value: value, Mode: mode}, nil
}

// String renders the target in the book's notation: "8+", "8-", "8".
func (t Target) String() string {
	switch t.Mode {
	case Plus:
		return strconv.Itoa(t.Value) + "+"
	case Minus:
		return strconv.Itoa(t.Value) + "-"
	case Exact:
		return strconv.Itoa(t.Value)
	default:
		return strconv.Itoa(t.Value)
	}
}

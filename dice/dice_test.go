package dice_test

import (
	"testing"

	"github.com/philoserf/ctworldgen/dice"
)

// TestSameSeedSameStream is what a recorded seed buys.
func TestSameSeedSameStream(t *testing.T) {
	t.Parallel()

	a, b := dice.NewStream(1977), dice.NewStream(1977)
	for i := range 500 {
		if x, y := a.Die(), b.Die(); x != y {
			t.Fatalf("throw %d: %d then %d from the same seed", i, x, y)
		}
	}
}

func TestDifferentSeedsDiverge(t *testing.T) {
	t.Parallel()

	a, b := dice.NewStream(1), dice.NewStream(2)
	same := 0

	for range 100 {
		if a.Die() == b.Die() {
			same++
		}
	}

	if same == 100 {
		t.Error("two seeds produced identical streams")
	}
}

// TestDieIsOneDie checks the face range and that every face appears.
func TestDieIsOneDie(t *testing.T) {
	t.Parallel()

	seen := map[int]int{}

	s := dice.NewStream(42)
	for range 6000 {
		throw := s.Die()
		if throw < 1 || throw > 6 {
			t.Fatalf("a die threw %d", throw)
		}

		seen[throw]++
	}

	for face := 1; face <= 6; face++ {
		if seen[face] == 0 {
			t.Errorf("the face %d never came up in 6000 throws", face)
		}
	}
}

// TestD2ConsumesTwoDiceInOrder pins the die primitive. A two-dice throw is
// the first die and then the second, drawn from the same stream: this is
// what makes the consumption order mean anything.
func TestD2ConsumesTwoDiceInOrder(t *testing.T) {
	t.Parallel()

	singles := dice.NewStream(99)
	first, second := singles.Die(), singles.Die()

	pairs := dice.NewStream(99)
	if got := pairs.D2(); got != first+second {
		t.Errorf("D2 = %d; the first two dice of the stream are %d and %d", got, first, second)
	}

	// And it leaves the stream where two dice would.
	third := singles.Die()
	if got := pairs.Die(); got != third {
		t.Errorf("after D2 the next die is %d, want %d: D2 consumed the wrong number of draws", got, third)
	}
}

func TestD2Range(t *testing.T) {
	t.Parallel()

	s := dice.NewStream(3)
	low, high := 12, 2

	for range 5000 {
		throw := s.D2()
		if throw < 2 || throw > 12 {
			t.Fatalf("two dice threw %d", throw)
		}

		low = min(low, throw)
		high = max(high, throw)
	}

	if low != 2 || high != 12 {
		t.Errorf("two dice ranged %d to %d over 5000 throws, want 2 to 12", low, high)
	}
}

func TestSumIsCumulative(t *testing.T) {
	t.Parallel()

	if got := dice.Sum(); got != 0 {
		t.Errorf("Sum() = %d, want 0", got)
	}

	// The greatest total the technological index matrix reaches (p. 9).
	if got := dice.Sum(6, 2, 1, 0, 4, 1); got != 14 {
		t.Errorf("Sum of the matrix maximum = %d, want 14", got)
	}

	if got := dice.Sum(-4, -2, 1); got != -5 {
		t.Errorf("Sum(-4, -2, 1) = %d, want -5", got)
	}
}

// TestTargetIsNPlus covers the only target kind Book 3 pp. 1-12 uses,
// including the one-die targets the character procedure never has: world
// occurrence at 4+ and the jump routes cells, which run from 1 to 6.
func TestTargetIsNPlus(t *testing.T) {
	t.Parallel()

	occurrence := dice.Target(4)
	for throw, want := range map[int]bool{1: false, 2: false, 3: false, 4: true, 5: true, 6: true} {
		if got := occurrence.Met(throw); got != want {
			t.Errorf("a throw of %d against 4+ = %v, want %v", throw, got, want)
		}
	}

	// A jump routes cell of 1 is met by any die; the DM'd occurrence throw
	// can reach 0 and 7.
	if !dice.Target(1).Met(1) {
		t.Error("a target of 1 was not met by a throw of 1")
	}

	if dice.Target(4).Met(3) {
		t.Error("a throw of 3 met a target of 4+")
	}

	if !dice.Target(4).Met(7) {
		t.Error("a throw of 7 did not meet a target of 4+")
	}

	if dice.Target(4).Met(0) {
		t.Error("a throw of 0 met a target of 4+")
	}

	// Scout base at starport A is 10+, which one die cannot reach; that is
	// why the base throws are two dice (ERRATA E001).
	if dice.Target(10).Met(6) {
		t.Error("a single die met the 10+ scout base target of starport A")
	}
}

func TestAlgorithmIsTheStringTheRecordStamps(t *testing.T) {
	t.Parallel()

	if dice.Algorithm != "go-math-rand-v2-pcg" {
		t.Errorf("the algorithm stamp is %q", dice.Algorithm)
	}
}

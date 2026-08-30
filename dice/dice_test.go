package dice_test

import (
	"errors"
	"testing"

	"github.com/philoserf/ctworldgen/dice"
)

// TestStreamStaysOnSixSidedDice: every throw of Book 3 reads a d6, so a
// stream that ever yielded anything else would corrupt a chart lookup
// rather than fail.
func TestStreamStaysOnSixSidedDice(t *testing.T) {
	s := dice.New(1)
	seen := map[int]bool{}

	for range 20000 {
		d := s.One()
		if d < 1 || d > 6 {
			t.Fatalf("One() = %d, want 1-6", d)
		}

		seen[d] = true
	}

	for face := 1; face <= 6; face++ {
		if !seen[face] {
			t.Errorf("20000 throws never showed a %d", face)
		}
	}
}

// TestSameSeedSameStream is the whole of replay's foundation.
func TestSameSeedSameStream(t *testing.T) {
	a, b := dice.New(4242), dice.New(4242)
	for i := range 500 {
		if x, y := a.One(), b.One(); x != y {
			t.Fatalf("roll %d: streams diverged, %d vs %d", i, x, y)
		}
	}
}

// TestDifferentSeedsDiverge: a seed that did not change the stream would
// make --seed decorative.
func TestDifferentSeedsDiverge(t *testing.T) {
	a, b := dice.New(1), dice.New(2)

	same := 0

	for range 200 {
		if a.One() == b.One() {
			same++
		}
	}

	if same == 200 {
		t.Error("two different seeds produced identical streams")
	}
}

// TestTwoReportsTheDiceInOrder: the log records the dice, not just the
// total (FR17), and Two must report them in the order they were drawn.
func TestTwoReportsTheDiceInOrder(t *testing.T) {
	want := dice.New(7)
	got := dice.New(7)

	for range 100 {
		a, b := got.Two()
		if a != want.One() || b != want.One() {
			t.Fatal("Two() did not draw two dice in order from the stream")
		}
	}
}

// TestTargetMet covers the three notations of B1 pp. 2-3.
func TestTargetMet(t *testing.T) {
	cases := []struct {
		target dice.Target
		total  int
		want   bool
	}{
		{dice.Target{Value: 8, Mode: dice.Plus}, 7, false},
		{dice.Target{Value: 8, Mode: dice.Plus}, 8, true},
		{dice.Target{Value: 8, Mode: dice.Plus}, 12, true},
		{dice.Target{Value: 8, Mode: dice.Minus}, 8, true},
		{dice.Target{Value: 8, Mode: dice.Minus}, 9, false},
		{dice.Target{Value: 8, Mode: dice.Exact}, 8, true},
		{dice.Target{Value: 8, Mode: dice.Exact}, 7, false},
		// The world occurrence throw of p. 1: one die, 4 or better.
		{dice.Target{Value: 4, Mode: dice.Plus}, 3, false},
		{dice.Target{Value: 4, Mode: dice.Plus}, 4, true},
		// The p. 2 jump routes table's A-A jump-1 cell: a lane that always
		// exists, because one die is never below 1.
		{dice.Target{Value: 1, Mode: dice.Plus}, 1, true},
	}

	for _, c := range cases {
		if got := c.target.Met(c.total); got != c.want {
			t.Errorf("Target%v.Met(%d) = %v, want %v", c.target, c.total, got, c.want)
		}
	}

	// A mode no constant names satisfies nothing, rather than everything.
	if (dice.Target{Value: 8, Mode: dice.Mode(99)}).Met(12) {
		t.Error("an unknown target mode was satisfied")
	}
}

// TestParseTargetRoundTrips: ParseTarget is the inverse of String, and the
// data files' load-time validation depends on it.
func TestParseTargetRoundTrips(t *testing.T) {
	for _, notation := range []string{"1+", "4+", "8+", "12+", "8-", "2-", "8", "12"} {
		target, err := dice.ParseTarget(notation)
		if err != nil {
			t.Errorf("ParseTarget(%q): %v", notation, err)

			continue
		}

		if got := target.String(); got != notation {
			t.Errorf("ParseTarget(%q).String() = %q", notation, got)
		}
	}
}

// TestParseTargetRejectsWhatIsNotNotation. The floor is 1, not the
// character procedure's 2: Book 3 throws one die as often as two.
//
// "+8" and "-8" are the cases worth naming: the sign goes after the number
// in the book's notation, and strconv.Atoi would happily read a leading one
// and hand back the exact target 8 — a different rule than the "8+" the
// writer meant. This function validates the p. 5 chart's printed throws at
// load time, so that has to be a build failure.
func TestParseTargetRejectsWhatIsNotNotation(t *testing.T) {
	for _, bad := range []string{"", "0+", "13+", "0", "13", "-1", "x+", "+", "8++", "eight", "+8", "-8", "+8+"} {
		if _, err := dice.ParseTarget(bad); !errors.Is(err, dice.ErrBadTarget) {
			t.Errorf("ParseTarget(%q): err = %v, want ErrBadTarget", bad, err)
		}
	}
}

// TestStringOfAnUnknownMode still renders the value, so a diagnostic is
// never blank.
func TestStringOfAnUnknownMode(t *testing.T) {
	if got := (dice.Target{Value: 8, Mode: dice.Mode(99)}).String(); got != "8" {
		t.Errorf("String() of an unknown mode = %q, want %q", got, "8")
	}
}

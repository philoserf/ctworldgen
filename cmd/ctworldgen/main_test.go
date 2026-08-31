package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/subsector"
)

func exec(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var out, errs bytes.Buffer

	err := run(args, &out, &errs)

	return out.String(), errs.String(), err
}

func decode(t *testing.T, out string) *subsector.Subsector {
	t.Helper()

	s, err := subsector.Decode(strings.NewReader(out))
	if err != nil {
		t.Fatalf("decoding the output: %v", err)
	}

	return s
}

func TestNewWritesARecordToStdout(t *testing.T) {
	t.Parallel()

	out, _, err := exec(t, "new", "--seed", "1", "--name", "Aramis", "--occurrence-dm", "0")
	if err != nil {
		t.Fatal(err)
	}

	s := decode(t, out)
	if s.Seed != 1 || s.Name != "Aramis" || s.OccurrenceDM != 0 {
		t.Errorf("record does not carry its inputs: seed %d, name %q, DM %+d", s.Seed, s.Name, s.OccurrenceDM)
	}

	if len(s.Worlds) == 0 {
		t.Error("no worlds; eighty throws at 4+ should place some")
	}

	if !strings.HasSuffix(out, "\n") {
		t.Error("the record does not end in a newline")
	}
}

// TestSeedZeroIsAChoice: --seed 0 is explicit and distinct, not a request
// for a random seed.
func TestSeedZeroIsAChoice(t *testing.T) {
	t.Parallel()

	first, _, err := exec(t, "new", "--seed", "0")
	if err != nil {
		t.Fatal(err)
	}

	second, _, err := exec(t, "new", "--seed", "0")
	if err != nil {
		t.Fatal(err)
	}

	if first != second {
		t.Error("--seed 0 produced two different subsectors")
	}

	if s := decode(t, first); s.Seed != 0 {
		t.Errorf("--seed 0 recorded a seed of %d", s.Seed)
	}
}

// TestASeedIsAlwaysRecorded: without --seed one is drawn from OS entropy
// and written into the record, so a run is reproducible after the fact.
func TestASeedIsAlwaysRecorded(t *testing.T) {
	t.Parallel()

	seeds := map[uint64]bool{}

	for range 5 {
		out, _, err := exec(t, "new")
		if err != nil {
			t.Fatal(err)
		}

		seeds[decode(t, out).Seed] = true
	}

	if len(seeds) < 2 {
		t.Errorf("five unseeded runs drew %d distinct seeds", len(seeds))
	}
}

func TestRejectsAnOccurrenceDMTheBookDoesNotOffer(t *testing.T) {
	t.Parallel()

	for _, dm := range []string{"2", "-2", "10"} {
		_, _, err := exec(t, "new", "--occurrence-dm", dm)
		if err == nil {
			t.Errorf("--occurrence-dm %s was accepted", dm)
		}
	}
}

func TestFlagsPrecedeAnyFilename(t *testing.T) {
	t.Parallel()

	_, _, err := exec(t, "new", "out.json")
	if err == nil {
		t.Error("new accepted a positional argument")
	}
}

func TestExistingFilesAreNeverOverwrittenWithoutForce(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "subsector.json")

	_, _, err := exec(t, "new", "--seed", "1", "-o", path)
	if err != nil {
		t.Fatal(err)
	}

	first, err := os.ReadFile(path) //nolint:gosec // a path this test created
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = exec(t, "new", "--seed", "2", "-o", path)
	if err == nil {
		t.Fatal("an existing file was overwritten without --force")
	}

	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("the error does not mention --force: %v", err)
	}

	after, readErr := os.ReadFile(path) //nolint:gosec // a path this test created
	if readErr != nil {
		t.Fatal(readErr)
	}

	if !bytes.Equal(first, after) {
		t.Error("the file changed despite the refusal")
	}

	_, _, forceErr := exec(t, "new", "--seed", "2", "-o", path, "--force")
	if forceErr != nil {
		t.Fatal(forceErr)
	}

	forced, err := os.ReadFile(path) //nolint:gosec // a path this test created
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(first, forced) {
		t.Error("--force did not overwrite the file")
	}
}

func TestWriteReportsAnUnusablePath(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "no-such-directory", "subsector.json")

	_, _, err := exec(t, "new", "--seed", "1", "-o", path)
	if err == nil {
		t.Error("writing into a directory that does not exist succeeded")
	}
}

func TestVersionReportsTheBuildAndTheStamps(t *testing.T) {
	t.Parallel()

	out, _, err := exec(t, "version")
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"ctworldgen",
		"engine     " + subsector.EngineVersion,
		"ruleset    " + subsector.Ruleset,
	}
	for _, want := range want {
		if !strings.Contains(out, want) {
			t.Errorf("version output has no %q:\n%s", want, out)
		}
	}
}

func TestUsage(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{{}, {"render"}, {"batch"}} {
		_, stderr, err := exec(t, args...)
		if err == nil {
			t.Errorf("%v was accepted", args)
		}

		if !strings.Contains(stderr, "usage:") {
			t.Errorf("%v printed no usage", args)
		}
	}

	_, _, err := exec(t, "new", "--nonesuch")
	if err == nil {
		t.Error("an unknown flag was accepted")
	}
}

func TestDirtySuffix(t *testing.T) {
	t.Parallel()

	if got := dirtySuffix(true); got != " (dirty)" {
		t.Errorf("dirtySuffix(true) = %q", got)
	}

	if got := dirtySuffix(false); got != "" {
		t.Errorf("dirtySuffix(false) = %q", got)
	}
}

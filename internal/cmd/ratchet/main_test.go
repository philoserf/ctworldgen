package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const profile = `mode: atomic
github.com/x/pkg/a.go:10.2,12.3 2 1
github.com/x/pkg/a.go:14.2,15.3 3 0
github.com/x/other/b.go:1.1,2.2 1 1
`

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)

	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	return path
}

func TestUncoveredCountsStatementsNotLines(t *testing.T) {
	t.Parallel()

	counts, err := uncovered(writeTemp(t, "coverage.out", profile))
	if err != nil {
		t.Fatal(err)
	}

	if got := counts["github.com/x/pkg"]; got != 3 {
		t.Errorf("pkg has %d uncovered statements, want 3", got)
	}

	// A fully covered package must still appear, or the gate cannot tell it
	// from one that vanished.
	if got, ok := counts["github.com/x/other"]; !ok || got != 0 {
		t.Errorf("other has %d uncovered statements (present: %v), want 0 and present", got, ok)
	}
}

func TestUncoveredRejectsAMalformedProfile(t *testing.T) {
	t.Parallel()

	bad := []string{
		"mode: atomic\nnonsense\n",
		"mode: atomic\na.go:1.1,2.2 x 0\n",
		"mode: atomic\na.go:1.1,2.2 1 x\n",
	}
	for _, bad := range bad {
		_, err := uncovered(writeTemp(t, "coverage.out", bad))
		if err == nil {
			t.Errorf("a malformed profile was accepted: %q", bad)
		}
	}

	_, err := uncovered(filepath.Join(t.TempDir(), "absent.out"))
	if err == nil {
		t.Error("a missing profile was accepted")
	}
}

func TestCompareFailsInBothDirections(t *testing.T) {
	t.Parallel()

	base := map[string]int{"a": 5}

	err := compare(base, map[string]int{"a": 5})
	if err != nil {
		t.Errorf("an unchanged count failed: %v", err)
	}

	err = compare(base, map[string]int{"a": 6})
	if err == nil || !strings.Contains(err.Error(), "up from 5") {
		t.Errorf("gaining an uncovered statement gave %v", err)
	}

	err = compare(base, map[string]int{"a": 4})
	if err == nil || !strings.Contains(err.Error(), "down from 5") {
		t.Errorf("losing an uncovered statement gave %v", err)
	}

	err = compare(base, map[string]int{})
	if err == nil {
		t.Error("a package that vanished from the profile was accepted")
	}

	err = compare(base, map[string]int{"a": 5, "b": 0})
	if err == nil {
		t.Error("a package missing from the baseline was accepted")
	}
}

func TestBaselineRoundTrips(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "coverage.baseline")

	want := map[string]int{"github.com/x/pkg": 3, "github.com/x/other": 0}

	err := writeBaseline(path, want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := readBaseline(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != len(want) {
		t.Fatalf("baseline round-tripped to %v, want %v", got, want)
	}

	for pkg, n := range want {
		if got[pkg] != n {
			t.Errorf("%s round-tripped to %d, want %d", pkg, got[pkg], n)
		}
	}

	_, err = readBaseline(writeTemp(t, "bad", "a b c\n"))
	if err == nil {
		t.Error("a malformed baseline was accepted")
	}

	_, err = readBaseline(writeTemp(t, "bad2", "a x\n"))
	if err == nil {
		t.Error("a non-numeric baseline count was accepted")
	}

	_, err = readBaseline(filepath.Join(t.TempDir(), "absent"))
	if err == nil {
		t.Error("a missing baseline was accepted")
	}
}

func TestSortedIsDeterministic(t *testing.T) {
	t.Parallel()

	got := sorted(map[string]int{"c": 0, "a": 0, "b": 0})
	if strings.Join(got, ",") != "a,b,c" {
		t.Errorf("sorted = %v", got)
	}
}

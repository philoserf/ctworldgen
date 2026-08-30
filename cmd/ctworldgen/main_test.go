package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/worldgen"
)

// fixedSeed stands in for the OS entropy source so that a test never
// depends on a random draw.
func fixedSeed() (uint64, error) { return 4242, nil }

type result struct {
	code   int
	stdout string
	stderr string
}

func invoke(t *testing.T, args ...string) result {
	t.Helper()

	var out, errOut bytes.Buffer

	code := run(args, fixedSeed, &out, &errOut)

	return result{code: code, stdout: out.String(), stderr: errOut.String()}
}

func TestNewWritesARecordToStdout(t *testing.T) {
	got := invoke(t, "new", "--seed", "7", "--name", "Vega")
	if got.code != exitOK {
		t.Fatalf("new: code %d, stderr %s", got.code, got.stderr)
	}

	sub, err := worldgen.UnmarshalRecord([]byte(got.stdout))
	if err != nil {
		t.Fatalf("the record on stdout does not parse: %v", err)
	}

	if sub.RNG.Seed != 7 || sub.Name != "Vega" {
		t.Errorf("record has seed %d and name %q", sub.RNG.Seed, sub.Name)
	}

	if err := worldgen.Replay(sub, false); err != nil {
		t.Errorf("the record on stdout does not replay: %v", err)
	}
}

// TestNewDrawsASeedOnlyWhenNoneIsGiven: --seed 0 is an explicit and
// reproducible choice, not a request for a random one.
func TestNewDrawsASeedOnlyWhenNoneIsGiven(t *testing.T) {
	drawn := invoke(t, "new")
	if seedOf(t, drawn.stdout) != 4242 {
		t.Errorf("without --seed the recorded seed is %d, want the drawn 4242", seedOf(t, drawn.stdout))
	}

	explicit := invoke(t, "new", "--seed", "0")
	if seedOf(t, explicit.stdout) != 0 {
		t.Errorf("--seed 0 recorded seed %d, want 0", seedOf(t, explicit.stdout))
	}
}

func seedOf(t *testing.T, record string) uint64 {
	t.Helper()

	sub, err := worldgen.UnmarshalRecord([]byte(record))
	if err != nil {
		t.Fatalf("parsing the record: %v", err)
	}

	return sub.RNG.Seed
}

func TestNewRefusesAnOccurrenceDMThePageDoesNotOffer(t *testing.T) {
	got := invoke(t, "new", "--seed", "1", "--occurrence-dm", "2")
	if got.code != exitError {
		t.Errorf("code %d, want %d; stderr %s", got.code, exitError, got.stderr)
	}

	if !strings.Contains(got.stderr, "p. 1") {
		t.Errorf("the refusal does not cite the page: %s", got.stderr)
	}
}

func TestNewWritesAFileAndRefusesToOverwriteIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub.json")

	if got := invoke(t, "new", "--seed", "3", "-o", path); got.code != exitOK {
		t.Fatalf("new -o: code %d, stderr %s", got.code, got.stderr)
	}

	again := invoke(t, "new", "--seed", "3", "-o", path)
	if again.code != exitError || !strings.Contains(again.stderr, "already exists") {
		t.Errorf("a second write was not refused: code %d, stderr %s", again.code, again.stderr)
	}

	if forced := invoke(t, "new", "--seed", "3", "-o", path, "--force"); forced.code != exitOK {
		t.Errorf("--force did not overwrite: code %d, stderr %s", forced.code, forced.stderr)
	}
}

// TestBatchDerivesEachSeedFromTheBase: every member replays on its own,
// without the batch it came from.
func TestBatchDerivesEachSeedFromTheBase(t *testing.T) {
	got := invoke(t, "batch", "--count", "3", "--seed", "100")
	if got.code != exitOK {
		t.Fatalf("batch: code %d, stderr %s", got.code, got.stderr)
	}

	lines := strings.Split(strings.TrimSuffix(got.stdout, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("batch wrote %d lines, want 3", len(lines))
	}

	for i, line := range lines {
		sub, err := worldgen.UnmarshalRecord([]byte(line))
		if err != nil {
			t.Fatalf("line %d does not parse: %v", i, err)
		}

		if want := uint64(100 + i); sub.RNG.Seed != want {
			t.Errorf("line %d has seed %d, want %d", i, sub.RNG.Seed, want)
		}

		if err := worldgen.Replay(sub, false); err != nil {
			t.Errorf("line %d does not replay: %v", i, err)
		}

		// A JSONL line has to be one line.
		if !json.Valid([]byte(line)) || strings.Contains(line, "\n") {
			t.Errorf("line %d is not a single JSON value", i)
		}
	}
}

func TestBatchToADirectory(t *testing.T) {
	dir := t.TempDir()

	if got := invoke(t, "batch", "--count", "2", "--seed", "5", "-o", dir); got.code != exitOK {
		t.Fatalf("batch -o dir: code %d, stderr %s", got.code, got.stderr)
	}

	for _, name := range []string{"subsector-01.json", "subsector-02.json"} {
		data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // a temp-dir path this test built
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}

		if _, err := worldgen.UnmarshalRecord(data); err != nil {
			t.Errorf("%s does not parse: %v", name, err)
		}
	}

	// A second run must refuse the whole batch rather than half-replace the
	// directory.
	again := invoke(t, "batch", "--count", "2", "--seed", "5", "-o", dir)
	if again.code != exitError || !strings.Contains(again.stderr, "already exists") {
		t.Errorf("a second batch into the same directory was not refused: %s", again.stderr)
	}

	if forced := invoke(t, "batch", "--count", "2", "--seed", "5", "-o", dir, "--force"); forced.code != exitOK {
		t.Errorf("--force did not overwrite the batch: %s", forced.stderr)
	}
}

func TestBatchNeedsACount(t *testing.T) {
	if got := invoke(t, "batch"); got.code != exitUsage {
		t.Errorf("batch with no --count: code %d, want %d", got.code, exitUsage)
	}
}

func TestRenderAndReplayReadARecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub.json")

	if got := invoke(t, "new", "--seed", "11", "--name", "Vega", "-o", path); got.code != exitOK {
		t.Fatalf("new: %s", got.stderr)
	}

	listing := invoke(t, "render", path)
	if listing.code != exitOK || !strings.Contains(listing.stdout, "# Vega Subsector") {
		t.Errorf("render: code %d, stdout starts %.40q", listing.code, listing.stdout)
	}

	history := invoke(t, "render", "--history", path)
	if history.code != exitOK || !strings.Contains(history.stdout, "generation record") {
		t.Errorf("render --history: code %d, stdout starts %.40q", history.code, history.stdout)
	}

	replay := invoke(t, "replay", path)
	if replay.code != exitOK || !strings.Contains(replay.stdout, "replay verified") {
		t.Errorf("replay: code %d, stdout %q, stderr %q", replay.code, replay.stdout, replay.stderr)
	}
}

func TestReplayFailsOnATamperedRecord(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub.json")

	if got := invoke(t, "new", "--seed", "11", "-o", path); got.code != exitOK {
		t.Fatalf("new: %s", got.stderr)
	}

	data, err := os.ReadFile(path) //nolint:gosec // a temp-dir path this test built
	if err != nil {
		t.Fatal(err)
	}

	tampered := filepath.Join(dir, "tampered.json")
	//nolint:gosec // a temp-dir path this test built
	if err := os.WriteFile(tampered, bytes.Replace(data, []byte(`"engine_version": "`+worldgen.EngineVersion+`"`),
		[]byte(`"engine_version": "0.0.0-elsewhere"`), 1), 0o600); err != nil {
		t.Fatal(err)
	}

	got := invoke(t, "replay", tampered)
	if got.code != exitError || !strings.Contains(got.stderr, "provenance") {
		t.Errorf("replay of a restamped record: code %d, stderr %s", got.code, got.stderr)
	}

	waived := invoke(t, "replay", "--ignore-provenance", tampered)
	if waived.code != exitOK {
		t.Errorf("replay --ignore-provenance: code %d, stderr %s", waived.code, waived.stderr)
	}
}

func TestRenderAndReplayRejectBadInput(t *testing.T) {
	dir := t.TempDir()

	junk := filepath.Join(dir, "junk.json")
	if err := os.WriteFile(junk, []byte("not a record"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{
		{"render"},                     // no filename
		{"replay"},                     // no filename
		{"render", "a.json", "b.json"}, // two filenames
		{"render", filepath.Join(dir, "absent.js")}, // missing file
		{"render", junk}, // unparseable
		{"replay", junk},
	} {
		if got := invoke(t, args...); got.code == exitOK {
			t.Errorf("%v succeeded", args)
		}
	}
}

func TestUsage(t *testing.T) {
	if got := invoke(t); got.code != exitUsage || !strings.Contains(got.stderr, "usage:") {
		t.Errorf("no arguments: code %d", got.code)
	}

	if got := invoke(t, "wander"); got.code != exitUsage || !strings.Contains(got.stderr, "unknown command") {
		t.Errorf("unknown command: code %d, stderr %s", got.code, got.stderr)
	}

	// -h is a request answered on stdout, not a usage error.
	for _, flag := range []string{"-h", "--help"} {
		if got := invoke(t, flag); got.code != exitOK || !strings.Contains(got.stdout, "usage:") {
			t.Errorf("%s: code %d", flag, got.code)
		}
	}
}

func TestVersionReportsTheStampsARecordCarries(t *testing.T) {
	version := invoke(t, "version")
	if version.code != exitOK {
		t.Fatalf("version: code %d", version.code)
	}

	for _, want := range []string{worldgen.SchemaVersion, worldgen.EngineVersion, worldgen.Ruleset} {
		if !strings.Contains(version.stdout, want) {
			t.Errorf("version does not report %q:\n%s", want, version.stdout)
		}
	}

	// There is no policy_version to report: the procedure has no choice
	// points (docs/PRD.md, "The architectural delta from ctchargen").
	if strings.Contains(version.stdout, "policy") {
		t.Errorf("version reports a policy version:\n%s", version.stdout)
	}
}

func TestBadFlagsAreUsageErrors(t *testing.T) {
	for _, args := range [][]string{
		{"new", "--nonesuch"},
		{"batch", "--count", "notanumber"},
		{"render", "--nonesuch", "x.json"},
		{"replay", "--nonesuch", "x.json"},
	} {
		if got := invoke(t, args...); got.code != exitUsage {
			t.Errorf("%v: code %d, want %d", args, got.code, exitUsage)
		}
	}

	for _, args := range [][]string{{"new", "-h"}, {"batch", "-h"}, {"render", "-h"}, {"replay", "-h"}} {
		if got := invoke(t, args...); got.code != exitOK {
			t.Errorf("%v: code %d, want %d", args, got.code, exitOK)
		}
	}
}

func TestRandomSeedDraws(t *testing.T) {
	a, err := randomSeed()
	if err != nil {
		t.Fatalf("randomSeed: %v", err)
	}

	b, err := randomSeed()
	if err != nil {
		t.Fatalf("randomSeed: %v", err)
	}

	if a == b {
		t.Error("two draws from OS entropy produced the same seed")
	}
}

// TestBatchCreatesADirectoryNamedWithATrailingSlash is the README's own
// example: `batch --count 16 -o sector/` on a path that does not exist yet.
func TestBatchCreatesADirectoryNamedWithATrailingSlash(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sector") + string(os.PathSeparator)

	if got := invoke(t, "batch", "--count", "2", "--seed", "5", "-o", dir); got.code != exitOK {
		t.Fatalf("batch -o sector/: code %d, stderr %s", got.code, got.stderr)
	}

	for _, name := range []string{"subsector-01.json", "subsector-02.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("%s was not written: %v", name, err)
		}
	}
}

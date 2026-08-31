package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/subsector"
)

const (
	aramis     = "Aramis"
	renderVerb = "render"
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

	out, _, err := exec(t, "new", "--seed", "1", "--name", aramis, "--occurrence-dm", "0")
	if err != nil {
		t.Fatal(err)
	}

	s := decode(t, out)
	if s.Seed != 1 || s.Name != aramis || s.OccurrenceDM != 0 {
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

// TestADrawnSeedSurvivesADoubleParser: a drawn seed is bounded to
// 2^53-1, so a reader that parses JSON numbers as IEEE-754 doubles cannot
// round it into a seed that reproduces a different subsector.
func TestADrawnSeedSurvivesADoubleParser(t *testing.T) {
	t.Parallel()

	for range 50 {
		out, _, err := exec(t, "new")
		if err != nil {
			t.Fatal(err)
		}

		seed := decode(t, out).Seed
		if seed > 1<<53-1 {
			t.Fatalf("drawn seed %d exceeds 2^53-1 and a double parser would round it", seed)
		}

		if float64(seed) != float64(seed)+0 || uint64(float64(seed)) != seed {
			t.Fatalf("drawn seed %d does not survive a round trip through a float64", seed)
		}
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

	for _, args := range [][]string{{}, {"nonesuch"}, {"generate"}} {
		_, stderr, err := exec(t, args...)
		if err == nil {
			t.Errorf("%v was accepted", args)
		}

		if !strings.Contains(stderr, "usage:") {
			t.Errorf("%v printed no usage", args)
		}
	}

	// batch and render are subcommands now, so they must not fall through
	// to the usage banner; they fail on their own terms instead.
	for _, args := range [][]string{{"batch"}, {renderVerb}} {
		_, stderr, err := exec(t, args...)
		if err == nil {
			t.Errorf("%v was accepted", args)
		}

		if strings.Contains(stderr, "usage:") {
			t.Errorf("%v printed the top-level usage rather than its own error", args)
		}
	}

	_, _, err := exec(t, "new", "--nonesuch")
	if err == nil {
		t.Error("an unknown flag was accepted")
	}
}

// TestHelpIsNotAFailure: -h is a request that the flag package has already
// answered on stderr. Reporting it as an error too would print a failure
// after the help text and exit non-zero on a command that did what was
// asked of it.
func TestHelpIsNotAFailure(t *testing.T) {
	t.Parallel()

	_, stderr, err := exec(t, "new", "-h")
	if err != nil {
		t.Errorf("new -h reported an error: %v", err)
	}

	if !strings.Contains(stderr, "occurrence-dm") {
		t.Errorf("new -h printed no flag help:\n%s", stderr)
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

// TestBatchMemberZeroIsTheBaseSeed: members are numbered from zero, so
// `batch --count 1 --seed N` produces exactly what `new --seed N` does.
func TestBatchMemberZeroIsTheBaseSeed(t *testing.T) {
	t.Parallel()

	alone, _, err := exec(t, "new", "--seed", "5", "--name", aramis)
	if err != nil {
		t.Fatal(err)
	}

	lines, _, err := exec(t, "batch", "--count", "1", "--seed", "5", "--name", aramis)
	if err != nil {
		t.Fatal(err)
	}

	member := decode(t, strings.TrimSpace(lines))
	if member.Seed != 5 {
		t.Errorf("member 0 carries seed %d, and the base was 5", member.Seed)
	}

	// The member is JSONL and `new` is indented, so compare the records
	// rather than the bytes: a subsector is its worlds and its lanes.
	solo := decode(t, alone)
	if len(solo.Worlds) != len(member.Worlds) || len(solo.Routes) != len(member.Routes) {
		t.Fatalf("member 0 has %d worlds and %d lanes; `new` at the same seed has %d and %d",
			len(member.Worlds), len(member.Routes), len(solo.Worlds), len(solo.Routes))
	}

	for index, world := range solo.Worlds {
		if world.Digits != member.Worlds[index].Digits || world.Hex != member.Worlds[index].Hex {
			t.Fatalf("member 0 differs from `new` at %s: %s vs %s",
				world.Hex, world.Digits, member.Worlds[index].Digits)
		}
	}
}

// TestBatchDerivesEachMemberFromTheBase: base + i, and every member is a
// complete record of its own.
func TestBatchDerivesEachMemberFromTheBase(t *testing.T) {
	t.Parallel()

	out, _, err := exec(t, "batch", "--count", "4", "--seed", "100", "--occurrence-dm", "-1")
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 4 {
		t.Fatalf("a batch of 4 emitted %d lines", len(lines))
	}

	for index, line := range lines {
		record := decode(t, line)
		if want := uint64(100 + index); record.Seed != want {
			t.Errorf("member %d carries seed %d, want %d", index, record.Seed, want)
		}

		if record.OccurrenceDM != -1 {
			t.Errorf("member %d carries DM %+d", index, record.OccurrenceDM)
		}
	}
}

func TestBatchWantsACount(t *testing.T) {
	t.Parallel()

	for _, count := range []string{"0", "-1"} {
		_, _, err := exec(t, "batch", "--count", count)
		if err == nil {
			t.Errorf("--count %s was accepted", count)
		}
	}

	_, _, err := exec(t, "batch")
	if err == nil {
		t.Error("batch without --count was accepted")
	}
}

// TestBatchNamesMembersForTheSubsector: the slug of --name and a
// zero-padded index, or "subsector" where the name slugs to nothing.
func TestBatchNamesMembersForTheSubsector(t *testing.T) {
	t.Parallel()

	for _, want := range []struct {
		name  string
		count string
		first string
		last  string
	}{
		{aramis, "3", "aramis-00.json", "aramis-02.json"},
		{"Aramis Sector!", "12", "aramis-sector-00.json", "aramis-sector-11.json"},
		{"", "3", "subsector-00.json", "subsector-02.json"},
		{"---", "3", "subsector-00.json", "subsector-02.json"},
		{aramis, "101", "aramis-000.json", "aramis-100.json"},
	} {
		t.Run(want.name+"/"+want.count, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			_, _, err := exec(t, "batch", "--count", want.count, "--seed", "1",
				"--name", want.name, "-o", dir+string(os.PathSeparator))
			if err != nil {
				t.Fatal(err)
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}

			if entries[0].Name() != want.first || entries[len(entries)-1].Name() != want.last {
				t.Errorf("the batch is %s ... %s, want %s ... %s",
					entries[0].Name(), entries[len(entries)-1].Name(), want.first, want.last)
			}
		})
	}
}

func TestBatchRefusesToOverwriteWithoutForce(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() + string(os.PathSeparator)

	_, _, err := exec(t, "batch", "--count", "2", "--seed", "1", "-o", dir)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = exec(t, "batch", "--count", "2", "--seed", "9", "-o", dir)
	if err == nil {
		t.Fatal("a batch overwrote an existing member without --force")
	}

	_, _, forceErr := exec(t, "batch", "--count", "2", "--seed", "9", "-o", dir, "--force")
	if forceErr != nil {
		t.Fatal(forceErr)
	}
}

// TestRenderReadsARecordAndWritesTheListing closes the loop: what `new`
// wrote, `render` reads back and turns into the referee's pages.
func TestRenderReadsARecordAndWritesTheListing(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "subsector.json")

	_, _, err := exec(t, "new", "--seed", "1977", "--name", aramis, "-o", path)
	if err != nil {
		t.Fatal(err)
	}

	out, _, err := exec(t, renderVerb, path)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"# Aramis", "## Worlds", "## Space lanes", "## The worlds in detail"} {
		if !strings.Contains(out, want) {
			t.Errorf("the listing has no %q", want)
		}
	}
}

func TestRenderWantsExactlyOneRecord(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "subsector.json")

	_, _, err := exec(t, "new", "--seed", "1", "-o", path)
	if err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{{renderVerb}, {renderVerb, path, path}} {
		_, _, argsErr := exec(t, args...)
		if argsErr == nil {
			t.Errorf("%v was accepted", args)
		}
	}

	_, _, err = exec(t, renderVerb, filepath.Join(t.TempDir(), "absent.json"))
	if err == nil {
		t.Error("render accepted a record that does not exist")
	}
}

// TestRenderRefusesABatch: a record is one document. `batch` writes JSONL,
// so handing its output to `render` is the easy mistake, and rendering the
// first member while discarding the rest would be a silent one.
func TestRenderRefusesABatch(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "batch.jsonl")

	_, _, err := exec(t, "batch", "--count", "3", "--seed", "1", "-o", path)
	if err != nil {
		t.Fatal(err)
	}

	_, _, renderErr := exec(t, renderVerb, path)
	if renderErr == nil {
		t.Error("render accepted a batch of three records and would have listed only the first")
	}
}

// TestRenderRefusesARecordFromANewerSchema is the read-time half of the
// two obligations: DisallowUnknownFields on the Go side, so a record the
// current engine could not have written fails loudly.
func TestRenderRefusesARecordFromANewerSchema(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "subsector.json")

	_, _, err := exec(t, "new", "--seed", "1", "-o", path)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := os.ReadFile(path) //nolint:gosec // a path this test created
	if err != nil {
		t.Fatal(err)
	}

	tampered := `{"surprise": 1,` + strings.TrimPrefix(string(encoded), "{")

	writeErr := os.WriteFile(path, []byte(tampered), 0o600) //nolint:gosec // a path this test created in its own TempDir
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	_, _, renderErr := exec(t, renderVerb, path)
	if renderErr == nil {
		t.Error("render accepted a record carrying a field the schema does not define")
	}
}

// TestSectorWritesOneRecordOnTheSectorGrid: the sector subcommand is the
// same shape as new -- same flags, same seed rule -- and writes one
// record covering sixteen subsectors (ERRATA E006).
func TestSectorWritesOneRecordOnTheSectorGrid(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer

	err := run([]string{"sector", "--seed", "1", "--name", "Aramis"}, &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}

	record, err := subsector.Decode(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatal(err)
	}

	if record.Grid != subsector.SectorGrid() {
		t.Errorf("the record is on a %dx%d grid, want %dx%d",
			record.Grid.Columns, record.Grid.Rows, subsector.SectorColumns, subsector.SectorRows)
	}

	if !slices.Contains(record.Errata, "E006") {
		t.Errorf("a sector record does not stamp E006: %v", record.Errata)
	}
}

// TestSectorTakesNoArguments: flags precede any filename, and sector
// takes none at all.
func TestSectorTakesNoArguments(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer

	err := run([]string{"sector", "somefile.json"}, &out, &errOut)
	if err == nil {
		t.Fatal("sector accepted a positional argument")
	}
}

// TestRenderWritesTheBooklet: --format pdf writes the printable booklet
// rather than the Markdown listing, and it goes to a file. A terminal is
// not where a binary goes, so the flag needs -o rather than defaulting to
// stdout as the listing does.
func TestRenderWritesTheBooklet(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	record := filepath.Join(dir, "subsector.json")

	_, _, err := exec(t, "new", "--seed", "1977", "--name", aramis, "-o", record)
	if err != nil {
		t.Fatal(err)
	}

	booklet := filepath.Join(dir, "subsector.pdf")

	_, _, err = exec(t, renderVerb, "--format", "pdf", "-o", booklet, record)
	if err != nil {
		t.Fatal(err)
	}

	written, err := os.ReadFile(booklet) //nolint:gosec // a path this test created
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.HasPrefix(written, []byte("%PDF-")) {
		t.Errorf("the booklet does not open as a PDF: %q", written[:min(len(written), 8)])
	}

	if !bytes.Contains(written, []byte(aramis)) {
		t.Error("the booklet does not carry the subsector's name")
	}
}

// TestRenderRefusesAFormatItDoesNotWrite: the flag takes two values, and
// an error that names them is what tells the operator which.
func TestRenderRefusesAFormatItDoesNotWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	record := filepath.Join(dir, "subsector.json")

	_, _, err := exec(t, "new", "--seed", "1", "-o", record)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = exec(t, renderVerb, "--format", "postscript", "-o", filepath.Join(dir, "out"), record)
	if err == nil {
		t.Fatal("--format postscript was accepted")
	}

	if !strings.Contains(err.Error(), "markdown or pdf") {
		t.Errorf("the error does not name the formats: %v", err)
	}

	// A booklet written to stdout would be a terminal full of binary.
	_, _, err = exec(t, renderVerb, "--format", "pdf", record)
	if err == nil {
		t.Fatal("--format pdf without -o was accepted")
	}
}

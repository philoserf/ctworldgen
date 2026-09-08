package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/starmap"
)

const (
	aramis       = "Aramis"
	renderVerb   = "render"
	areaFlagName = "--occurrence-area"
)

func exec(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var out, errs bytes.Buffer

	err := run(args, &out, &errs)

	return out.String(), errs.String(), err
}

func decode(t *testing.T, out string) *starmap.Record {
	t.Helper()

	s, err := starmap.Decode(strings.NewReader(out))
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

// TestBroadAreasReachTheRecord: --occurrence-area is repeatable, because
// p. 1 offers a DM "on broad areas within a subsector", plural.
//
// The leading dash is the thing worth asserting. A negative DM opens the
// value with the character flag parsing uses to introduce a flag, and Go's
// flag package takes the next argument for a non-boolean flag whatever it
// begins with -- which is behaviour this command depends on and does not
// control.
func TestBroadAreasReachTheRecord(t *testing.T) {
	t.Parallel()

	out, _, err := exec(t, "new", "--seed", "1", "--name", aramis,
		areaFlagName, "-1@0101-0805", areaFlagName, "+1@0106-0810")
	if err != nil {
		t.Fatal(err)
	}

	record := decode(t, out)
	if len(record.OccurrenceAreas) != 2 {
		t.Fatalf("the record carries %d broad areas; want 2", len(record.OccurrenceAreas))
	}

	// In the order they were given: nothing sorts them.
	for index, want := range []starmap.Area{
		{From: starmap.Hex{Col: 1, Row: 1}, To: starmap.Hex{Col: 8, Row: 5}, DM: -1},
		{From: starmap.Hex{Col: 1, Row: 6}, To: starmap.Hex{Col: 8, Row: 10}, DM: 1},
	} {
		if record.OccurrenceAreas[index] != want {
			t.Errorf("area %d is %+v; want %+v", index, record.OccurrenceAreas[index], want)
		}
	}

	if !slices.Contains(record.Errata, "E012") {
		t.Errorf("a record generated under broad areas did not stamp E012: %v", record.Errata)
	}
}

// TestARecordWithNoBroadAreasStampsNothingNew is the other half, and the
// one that keeps every golden still: a run without the flag reads no
// silence, so it stamps nothing and writes no field.
func TestARecordWithNoBroadAreasStampsNothingNew(t *testing.T) {
	t.Parallel()

	out, _, err := exec(t, "new", "--seed", "1", "--name", aramis)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(out, "occurrence_areas") {
		t.Error("a run with no --occurrence-area wrote the field anyway")
	}

	if slices.Contains(decode(t, out).Errata, "E012") {
		t.Error("a run with no --occurrence-area stamped E012")
	}
}

// TestRejectsBroadAreasTheReadingRefuses: the syntax is refused by
// starmap.ParseArea as the flag is read, and the set as a whole by
// gen.Inputs.Validate before any die is thrown (ERRATA E012).
func TestRejectsBroadAreasTheReadingRefuses(t *testing.T) {
	t.Parallel()

	for name, args := range map[string][]string{
		"no DM":                        {areaFlagName, "0101-0410"},
		"one corner":                   {areaFlagName, "-1@0101"},
		"not a hex":                    {areaFlagName, "-1@0000-0410"},
		"a DM the page does not offer": {areaFlagName, "2@0101-0410"},
		"a corner off the p. 3 grid":   {areaFlagName, "-1@0101-0910"},
		"two areas sharing a hex": {
			areaFlagName, "-1@0101-0410", areaFlagName, "+1@0401-0810",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, err := exec(t, append([]string{"new", "--seed", "1"}, args...)...)
			if err == nil {
				t.Errorf("%s was accepted", name)
			}
		})
	}
}

// TestTheAreaFlagPrintsWhatItHolds. flag calls String on a fresh zero
// value to decide whether a default is worth printing, so `-h` reaches the
// empty case and nothing reaches the other one: a flag whose value cannot
// be printed is a flag that reports the wrong thing in a usage message,
// and the compiler cannot say so.
func TestTheAreaFlagPrintsWhatItHolds(t *testing.T) {
	t.Parallel()

	var flagValue areaFlag

	if flagValue.String() != "" {
		t.Errorf("a flag holding no areas prints %q; want nothing", flagValue.String())
	}

	for _, text := range []string{"-1@0101-0805", "+1@0106-0810"} {
		err := flagValue.Set(text)
		if err != nil {
			t.Fatalf("%s: %v", text, err)
		}
	}

	// The rectangles, in the order they were given. The DM is not in it:
	// both documents write a DM through one formatter, and a flag with its
	// own would be a second place for "+1" to be written differently.
	if flagValue.String() != "0101-0805,0106-0810" {
		t.Errorf("the flag prints %q; want %q", flagValue.String(), "0101-0805,0106-0810")
	}
}

// TestSectorTakesNoBroadArea: an area is a rectangle of one grid's
// numbering, and a sector's sixteen members are each generated on their
// own p. 3 grid (ERRATA E006 part 1). The flag is not defined there, so
// asking for one fails by name rather than being ignored.
func TestSectorTakesNoBroadArea(t *testing.T) {
	t.Parallel()

	_, errs, err := exec(t, "sector", "--seed", "1", areaFlagName, "-1@0101-0805")
	if err == nil {
		t.Fatal("sector accepted --occurrence-area")
	}

	if !strings.Contains(errs, "occurrence-area") {
		t.Errorf("the error does not name the flag that was refused:\n%s", errs)
	}

	// And the flag `new` does take is still there, so this is the one
	// difference between the two rather than a broken flag set.
	_, _, sectorErr := exec(t, "sector", "--seed", "1", "--occurrence-dm", "-1")
	if sectorErr != nil {
		t.Errorf("sector refused --occurrence-dm as well: %v", sectorErr)
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

// TestForceReplacesTheRecordRatherThanTruncatingIt holds --force to the
// property that protects the referee's record. Opening it O_TRUNC and
// only then writing means a write that fails partway leaves a truncated
// file where the record was, the command reporting an error over content
// that is already gone (issue #19).
//
// What is asserted is the rename rather than the failure, because the
// failure cannot be induced without a seam for three call sites to carry.
// A truncating open writes into the file that is already there; a rename
// puts a different file at the name. So the record after --force must not
// be the same file as the record before it, which is the whole of the
// property and which no truncating implementation can satisfy.
func TestForceReplacesTheRecordRatherThanTruncatingIt(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "subsector.json")

	_, _, err := exec(t, "new", "--seed", "1", "-o", path)
	if err != nil {
		t.Fatal(err)
	}

	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = exec(t, "new", "--seed", "2", "-o", path, "--force")
	if err != nil {
		t.Fatal(err)
	}

	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if os.SameFile(before, after) {
		t.Error("--force wrote into the record that was already there; a write that failed partway would have truncated it")
	}

	// The permission of a record is the referee's own notebook page, and
	// the file it is renamed from has to be created at that permission
	// rather than at the temporary file's default.
	if after.Mode().Perm() != recordMode {
		t.Errorf("the replaced record is mode %o; want %o", after.Mode().Perm(), recordMode)
	}

	// Nothing left beside it. The temporary file is removed on every path
	// out, and a successful rename has already carried it off its name.
	left, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range left {
		if entry.Name() != "subsector.json" {
			t.Errorf("--force left %s beside the record", entry.Name())
		}
	}
}

func TestWriteReportsAnUnusablePath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "no-such-directory", "subsector.json")

	_, _, err := exec(t, "new", "--seed", "1", "-o", path)
	if err == nil {
		t.Error("writing into a directory that does not exist succeeded")
	}

	// --force takes the other route out of writeFile, and it has its own
	// two ways to fail: the temporary file beside the target cannot be
	// made, and it cannot be put in the target's place.
	_, _, err = exec(t, "new", "--seed", "1", "-o", path, "--force")
	if err == nil {
		t.Error("--force into a directory that does not exist succeeded")
	}

	// A directory inside dir rather than dir itself, so that the temporary
	// file --force makes beside its target lands in dir, where the sweep
	// below can see whether it was cleaned up.
	occupied := filepath.Join(dir, "a-directory")

	err = os.Mkdir(occupied, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = exec(t, "new", "--seed", "1", "-o", occupied, "--force")
	if err == nil {
		t.Error("--force over a directory succeeded")
	}

	left, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range left {
		if entry.Name() != "a-directory" {
			t.Errorf("a --force that could not finish left %s behind", entry.Name())
		}
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
		"engine     " + starmap.EngineVersion,
		"ruleset    " + starmap.Ruleset,
	}
	for _, want := range want {
		if !strings.Contains(out, want) {
			t.Errorf("version output has no %q:\n%s", want, out)
		}
	}
}

func TestUsage(t *testing.T) {
	t.Parallel()

	// `batch` is among them: it is a retired name, covered by `sector`,
	// and a retired subcommand that still ran would be worse than one that
	// never existed.
	for _, args := range [][]string{{}, {"nonesuch"}, {"generate"}, {"batch"}} {
		_, stderr, err := exec(t, args...)
		if err == nil {
			t.Errorf("%v was accepted", args)
		}

		if !strings.Contains(stderr, "usage:") {
			t.Errorf("%v printed no usage", args)
		}
	}

	// render is a subcommand, so it must not fall through to the usage
	// banner; it fails on its own terms instead.
	for _, args := range [][]string{{renderVerb}} {
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

	for _, want := range []string{"# Aramis", "## Worlds", "## Routes", "## The worlds in detail"} {
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

// TestRenderRefusesMoreThanOneRecord: a record is one document. Two of
// them in one file -- concatenated by hand, or by any tool that writes a
// stream of records -- must fail loudly, because listing the first and
// discarding the rest would be a silent wrong answer.
func TestRenderRefusesMoreThanOneRecord(t *testing.T) {
	t.Parallel()

	first, _, err := exec(t, "new", "--seed", "1")
	if err != nil {
		t.Fatal(err)
	}

	second, _, err := exec(t, "new", "--seed", "2")
	if err != nil {
		t.Fatal(err)
	}

	// Marshal ends every record with a newline, so this is two documents
	// one after the other, which is the shape the mistake takes.
	path := filepath.Join(t.TempDir(), "two-records.json")

	writeErr := os.WriteFile(path, []byte(first+second), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	_, _, renderErr := exec(t, renderVerb, path)
	if renderErr == nil {
		t.Error("render accepted two records and would have listed only the first")
	}
}

// TestRenderRefusesARecordCarryingAnUnknownField is the read-time half of the
// two obligations: DisallowUnknownFields on the Go side, so a record the
// current engine could not have written fails loudly.
func TestRenderRefusesARecordCarryingAnUnknownField(t *testing.T) {
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

	record, err := starmap.Decode(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatal(err)
	}

	if record.Grid != starmap.SectorGrid() {
		t.Errorf("the record is on a %dx%d grid, want %dx%d",
			record.Grid.Columns, record.Grid.Rows, starmap.SectorColumns, starmap.SectorRows)
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
// TestLanesChoosesWhatIsDrawn: legible is the default, --lanes all draws
// every lane the record carries, and anything else is refused (ERRATA
// E007). The record is the same file in all three cases.
func TestLanesChoosesWhatIsDrawn(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	record := filepath.Join(dir, "subsector.json")

	// Seed 1 at DM +1 is the dense case: 150 lanes, most of them redundant.
	_, _, err := exec(t, "new", "--seed", "1", "--occurrence-dm", "1", "-o", record)
	if err != nil {
		t.Fatal(err)
	}

	byDefault, _, err := exec(t, renderVerb, record)
	if err != nil {
		t.Fatal(err)
	}

	all, _, err := exec(t, renderVerb, "--lanes", "all", record)
	if err != nil {
		t.Fatal(err)
	}

	if rows := strings.Count(all, " | "); rows <= strings.Count(byDefault, " | ") {
		t.Error("--lanes all drew no more than the default, so the default suppressed nothing")
	}

	if !strings.Contains(byDefault, "not listed") || !strings.Contains(byDefault, "ERRATA E007") {
		t.Error("the default listing suppressed lanes without saying so")
	}

	if strings.Contains(all, "not listed") {
		t.Error("--lanes all reported suppressing something")
	}

	_, _, err = exec(t, renderVerb, "--lanes", "nonesuch", record)
	if err == nil {
		t.Fatal("--lanes nonesuch was accepted")
	}

	if !strings.Contains(err.Error(), "legible or all") {
		t.Errorf("the error does not name the lane modes: %v", err)
	}
}

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

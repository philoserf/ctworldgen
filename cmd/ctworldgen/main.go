// Command ctworldgen generates rules-accurate Classic Traveller
// subsectors from Book 3's Worlds chapter (pp. 1-12, (c) 1977 text).
package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/render"
	"github.com/philoserf/ctworldgen/subsector"
)

// recordMode is the permission a written record gets: the referee's own
// notebook page, not a world-readable one.
const (
	recordMode = 0o600
	dirMode    = 0o750

	// minIndexWidth is the narrowest a batch member index is written.
	minIndexWidth = 2
)

// maxSafeSeed is 2^53 - 1, the largest integer an IEEE-754 double holds
// exactly. A drawn seed is kept inside it because a record whose seed has
// been rounded by a reader that parses JSON numbers as doubles reproduces
// a different subsector -- the one corruption this record cannot afford,
// and a silent one. An explicit --seed is deliberately not bounded: it is
// the operator's own number, and Go reads it back exactly.
const maxSafeSeed = 1<<53 - 1

var (
	errNoSubcommand         = errors.New("no subcommand")
	errUnknownSubcommand    = errors.New("unknown subcommand")
	errFileExists           = errors.New("file exists")
	errTakesNoArguments     = errors.New("this subcommand takes no arguments (flags precede any filename)")
	errCountTooSmall        = errors.New("--count must be at least 1")
	errRenderWantsOneRecord = errors.New("render takes exactly one record to read (flags precede it)")
)

const usage = `ctworldgen generates Classic Traveller subsectors from Book 3 pp. 1-12.

usage:
  ctworldgen new   [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
  ctworldgen batch --count N [--seed N] [--name X] [--occurrence-dm N] [-o dir|file.jsonl] [--force]
  ctworldgen render [-o file] [--force] subsector.json
  ctworldgen version
`

func main() {
	err := run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ctworldgen:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, usage)

		return errNoSubcommand
	}

	switch args[0] {
	case "new":
		return newCmd(args[1:], stdout, stderr)
	case "batch":
		return batchCmd(args[1:], stdout, stderr)
	case "render":
		return renderCmd(args[1:], stdout, stderr)
	case "version":
		return versionCmd(stdout)
	default:
		_, _ = io.WriteString(stderr, usage)

		return fmt.Errorf("%w %q", errUnknownSubcommand, args[0])
	}
}

func newCmd(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("new", flag.ContinueOnError)
	flags.SetOutput(stderr)

	var (
		seed         = flags.Uint64("seed", 0, "seed for the dice stream; drawn from OS entropy and recorded when absent")
		name         = flags.String("name", "", "name of the subsector")
		occurrenceDM = flags.Int("occurrence-dm", 0, "world occurrence DM: -1, 0 or +1 (Book 3 p. 1)")
		out          = flags.String("o", "", "write to this file instead of stdout")
		force        = flags.Bool("force", false, "overwrite an existing output file")
	)

	err := flags.Parse(args)
	if err != nil {
		// -h is a request that flag has already answered on stderr, not a
		// failure: reporting it again would print an error after the help
		// and exit non-zero on a successful command.
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return fmt.Errorf("parsing flags: %w", err)
	}

	if flags.NArg() > 0 {
		return fmt.Errorf("%w: got %q", errTakesNoArguments, flags.Arg(0))
	}

	// A seed is always recorded, so a run is reproducible after the fact.
	// Drawing one from OS entropy is the single exception to the
	// seeded-stream rule, and it happens before the engine starts.
	if !isSet(flags, "seed") {
		drawn, drawErr := entropySeed()
		if drawErr != nil {
			return drawErr
		}

		*seed = drawn
	}

	engine, err := gen.New()
	if err != nil {
		return fmt.Errorf("building the engine: %w", err)
	}

	s, err := engine.Generate(gen.Inputs{Seed: *seed, Name: *name, OccurrenceDM: *occurrenceDM})
	if err != nil {
		return fmt.Errorf("generating the subsector: %w", err)
	}

	return write(s, *out, *force, stdout)
}

func isSet(fs *flag.FlagSet, name string) bool {
	set := false

	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})

	return set
}

func entropySeed() (uint64, error) {
	var seedBytes [8]byte

	_, err := rand.Read(seedBytes[:])
	if err != nil {
		return 0, fmt.Errorf("drawing a seed from OS entropy: %w", err)
	}

	return binary.BigEndian.Uint64(seedBytes[:]) & maxSafeSeed, nil
}

func write(record *subsector.Subsector, path string, force bool, stdout io.Writer) error {
	encoded, err := subsector.Marshal(record)
	if err != nil {
		return fmt.Errorf("rendering the record: %w", err)
	}

	if path == "" {
		_, writeErr := stdout.Write(encoded)
		if writeErr != nil {
			return fmt.Errorf("writing the record to stdout: %w", writeErr)
		}

		return nil
	}

	return writeFile(path, encoded, force)
}

// writeFile is the one place a file is written, so that "existing files
// are never overwritten without --force" holds for every subcommand
// rather than for whichever one remembered it.
func writeFile(path string, contents []byte, force bool) error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !force {
		flags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	}

	file, err := os.OpenFile(path, flags, recordMode) //nolint:gosec // path is the operator's own -o argument
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%w: %s; pass --force to overwrite it", errFileExists, path)
		}

		return fmt.Errorf("opening %s: %w", path, err)
	}

	_, writeErr := file.Write(contents)
	if writeErr != nil {
		_ = file.Close()

		return fmt.Errorf("writing %s: %w", path, writeErr)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("closing %s: %w", path, err)
	}

	return nil
}

func versionCmd(stdout io.Writer) error {
	build, revision, dirty := buildInfo()
	// The build is the binary's own provenance; the stamps are constants in
	// the code. An untagged or dirty build changes the first and cannot
	// change the second.
	var out strings.Builder

	fmt.Fprintf(&out, "ctworldgen %s\n", build)

	if revision != "" {
		fmt.Fprintf(&out, "revision   %s%s\n", revision, dirtySuffix(dirty))
	}

	fmt.Fprintf(&out, "engine     %s\n", subsector.EngineVersion)
	fmt.Fprintf(&out, "schema     %d\n", subsector.SchemaVersion)
	fmt.Fprintf(&out, "ruleset    %s\n", subsector.Ruleset)

	_, err := io.WriteString(stdout, out.String())
	if err != nil {
		return fmt.Errorf("writing the version report: %w", err)
	}

	return nil
}

func dirtySuffix(dirty bool) string {
	if dirty {
		return " (dirty)"
	}

	return ""
}

func buildInfo() (string, string, bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unknown)", "", false
	}

	version := info.Main.Version
	if version == "" {
		version = "(devel)"
	}

	var (
		revision string
		dirty    bool
	)

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			dirty = setting.Value == "true"
		}
	}

	return version, revision, dirty
}

// batchCmd produces N independent subsectors, each member's seed derived
// from the base seed and its index and recorded in its own record.
//
// Sixteen is the suggested count because a sector is sixteen subsectors,
// but no sector-level structure is implied or recorded: each member is a
// complete, independently reproducible record.
func batchCmd(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("batch", flag.ContinueOnError)
	flags.SetOutput(stderr)

	var (
		count        = flags.Int("count", 0, "number of subsectors to produce")
		seed         = flags.Uint64("seed", 0, "base seed; drawn from OS entropy and recorded when absent")
		name         = flags.String("name", "", "name of every subsector in the batch")
		occurrenceDM = flags.Int("occurrence-dm", 0, "world occurrence DM: -1, 0 or +1 (Book 3 p. 1)")
		out          = flags.String("o", "", "a directory for one file per subsector, or a path for a single JSONL file")
		force        = flags.Bool("force", false, "overwrite existing output files")
	)

	err := flags.Parse(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return fmt.Errorf("parsing flags: %w", err)
	}

	if flags.NArg() > 0 {
		return fmt.Errorf("%w: got %q", errTakesNoArguments, flags.Arg(0))
	}

	if *count < 1 {
		return fmt.Errorf("%w: got %d", errCountTooSmall, *count)
	}

	if !isSet(flags, "seed") {
		drawn, drawErr := entropySeed()
		if drawErr != nil {
			return drawErr
		}

		*seed = drawn
	}

	records, err := batch(*count, *seed, *name, *occurrenceDM)
	if err != nil {
		return err
	}

	if perFile(*out) {
		return writeMembers(records, *out, *name, *force)
	}

	return writeJSONL(records, *out, *force, stdout)
}

// batch derives each member's seed from the base and its index. Member i
// is seeded with base + i as an unsigned 64-bit addition, wrapping if it
// overflows, so member 0 carries the base seed itself and reproduces
// exactly what `new` with that seed produces.
func batch(count int, base uint64, name string, occurrenceDM int) ([]*subsector.Subsector, error) {
	engine, err := gen.New()
	if err != nil {
		return nil, fmt.Errorf("building the engine: %w", err)
	}

	// Deliberately not preallocated to count: the capacity would be sized
	// straight from an operator's --count, and makeslice panics rather than
	// erroring on an absurd one. Growing as members are generated bounds
	// the allocation by the work actually done.
	var records []*subsector.Subsector

	for index := range count {
		derived := base + uint64(index)

		record, genErr := engine.Generate(gen.Inputs{Seed: derived, Name: name, OccurrenceDM: occurrenceDM})
		if genErr != nil {
			return nil, fmt.Errorf("member %d: %w", index, genErr)
		}

		records = append(records, record)
	}

	return records, nil
}

// perFile reports whether -o names a directory rather than a file: an
// existing directory, or any path ending in a separator.
func perFile(out string) bool {
	if out == "" {
		return false
	}

	if strings.HasSuffix(out, string(os.PathSeparator)) {
		return true
	}

	info, err := os.Stat(out)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// memberName is the slug of --name with the member's index appended,
// zero-padded to the width the count needs so the files sort in order.
// An empty or unsluggable name gives "subsector".
func memberName(name string, index, count int) string {
	slug := slugify(name)

	// At least two digits, so a small batch still reads as a series and a
	// larger one sorts beside it.
	width := max(minIndexWidth, len(strconv.Itoa(count-1)))

	return fmt.Sprintf("%s-%0*d.json", slug, width, index)
}

// slugify lowercases a name and reduces every run of characters that is
// not a letter or a digit to a single hyphen.
func slugify(name string) string {
	var built strings.Builder

	hyphenated := false

	for _, char := range strings.ToLower(name) {
		switch {
		case (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9'):
			built.WriteRune(char)

			hyphenated = false
		case built.Len() > 0 && !hyphenated:
			built.WriteByte('-')

			hyphenated = true
		}
	}

	slug := strings.Trim(built.String(), "-")
	if slug == "" {
		return "subsector"
	}

	return slug
}

func writeMembers(records []*subsector.Subsector, dir, name string, force bool) error {
	err := os.MkdirAll(dir, dirMode)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	for index, record := range records {
		encoded, marshalErr := subsector.Marshal(record)
		if marshalErr != nil {
			return fmt.Errorf("member %d: %w", index, marshalErr)
		}

		path := filepath.Join(dir, memberName(name, index, len(records)))

		writeErr := writeFile(path, encoded, force)
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}

// writeJSONL emits one record per line, which is what a batch is when it
// is not a directory.
func writeJSONL(records []*subsector.Subsector, path string, force bool, stdout io.Writer) error {
	var built []byte

	for _, record := range records {
		encoded, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("marshaling a member: %w", err)
		}

		built = append(built, encoded...)
		built = append(built, '\n')
	}

	if path == "" {
		_, err := stdout.Write(built)
		if err != nil {
			return fmt.Errorf("writing the batch to stdout: %w", err)
		}

		return nil
	}

	return writeFile(path, built, force)
}

// renderCmd reads a record and writes the subsector listing.
func renderCmd(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("render", flag.ContinueOnError)
	flags.SetOutput(stderr)

	out := flags.String("o", "", "write to this file instead of stdout")
	force := flags.Bool("force", false, "overwrite an existing output file")

	err := flags.Parse(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return fmt.Errorf("parsing flags: %w", err)
	}

	if flags.NArg() != 1 {
		return errRenderWantsOneRecord
	}

	file, err := os.Open(flags.Arg(0))
	if err != nil {
		return fmt.Errorf("opening %s: %w", flags.Arg(0), err)
	}

	defer func() { _ = file.Close() }()

	record, err := subsector.Decode(file)
	if err != nil {
		return fmt.Errorf("reading %s: %w", flags.Arg(0), err)
	}

	renderer, err := render.New()
	if err != nil {
		return fmt.Errorf("building the renderer: %w", err)
	}

	var built strings.Builder

	err = renderer.Subsector(&built, record)
	if err != nil {
		return fmt.Errorf("rendering %s: %w", flags.Arg(0), err)
	}

	if *out == "" {
		_, writeErr := io.WriteString(stdout, built.String())
		if writeErr != nil {
			return fmt.Errorf("writing the listing: %w", writeErr)
		}

		return nil
	}

	return writeFile(*out, []byte(built.String()), *force)
}

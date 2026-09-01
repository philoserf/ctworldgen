// Command ctworldgen generates rules-accurate Classic Traveller
// subsectors from Book 3's Worlds chapter (pp. 1-12, (c) 1977 text).
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/render"
	"github.com/philoserf/ctworldgen/starmap"
)

// recordMode is the permission a written record gets: the referee's own
// notebook page, not a world-readable one.
const recordMode = 0o600

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
	errRenderWantsOneRecord = errors.New("render takes exactly one record to read (flags precede it)")
	errNotAFormat           = errors.New("--format is markdown or pdf and nothing else")
	errPDFWantsAFile        = errors.New("--format pdf writes a binary and needs -o")
)

// The two things render writes. The Markdown listing is the default
// because it is what the tool has always written and what a terminal can
// read.
const (
	formatMarkdown = "markdown"
	formatPDF      = "pdf"
)

const usage = `ctworldgen generates Classic Traveller subsectors from Book 3 pp. 1-12.

usage:
  ctworldgen new   [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
  ctworldgen sector [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
  ctworldgen render [--format markdown|pdf] [-o file] [--force] subsector.json
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
	case "sector":
		return sectorCmd(args[1:], stdout, stderr)
	case "render":
		return renderCmd(args[1:], stdout, stderr)
	case "version":
		return versionCmd(stdout)
	default:
		_, _ = io.WriteString(stderr, usage)

		return fmt.Errorf("%w %q", errUnknownSubcommand, args[0])
	}
}

// singleRecordCmd is `new` and `sector`. They take the same flags, draw a
// seed the same way, and write one record; the only difference is which
// pass of the engine fills it, so that is the only thing passed in.
func singleRecordCmd(
	subcommand, noun string, args []string, stdout, stderr io.Writer,
	fill func(*gen.Engine, gen.Inputs) (*starmap.Record, error),
) error {
	flags := flag.NewFlagSet(subcommand, flag.ContinueOnError)
	flags.SetOutput(stderr)

	var (
		seed         = flags.Uint64("seed", 0, "seed for the dice stream; drawn from OS entropy and recorded when absent")
		name         = flags.String("name", "", "name of the "+noun)
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

	record, err := fill(engine, gen.Inputs{Seed: *seed, Name: *name, OccurrenceDM: *occurrenceDM})
	if err != nil {
		return fmt.Errorf("generating the %s: %w", noun, err)
	}

	return write(record, *out, *force, stdout)
}

// newCmd writes one subsector: the whole of Book 3 pp. 1-12 on the p. 3
// grid.
func newCmd(args []string, stdout, stderr io.Writer) error {
	return singleRecordCmd("new", "subsector", args, stdout, stderr,
		func(e *gen.Engine, in gen.Inputs) (*starmap.Record, error) { return e.Generate(in) })
}

// sectorCmd writes one record covering sixteen subsectors on one grid,
// with the routes at their seams thrown for (ERRATA E006). Every member is
// the subsector `new --seed base+i` writes, so a sector is the sixteen
// subsectors a referee could have generated one at a time, plus the routes
// that generating them one at a time could not find.
func sectorCmd(args []string, stdout, stderr io.Writer) error {
	return singleRecordCmd("sector", "sector", args, stdout, stderr,
		func(e *gen.Engine, in gen.Inputs) (*starmap.Record, error) { return e.Sector(in) })
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

func write(record *starmap.Record, path string, force bool, stdout io.Writer) error {
	encoded, err := starmap.Marshal(record)
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

	fmt.Fprintf(&out, "engine     %s\n", starmap.EngineVersion)
	fmt.Fprintf(&out, "schema     %d\n", starmap.SchemaVersion)
	fmt.Fprintf(&out, "ruleset    %s\n", starmap.Ruleset)

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

// renderCmd reads a record and writes the subsector listing.
func renderCmd(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("render", flag.ContinueOnError)
	flags.SetOutput(stderr)

	out := flags.String("o", "", "write to this file instead of stdout")
	force := flags.Bool("force", false, "overwrite an existing output file")
	format := flags.String("format", formatMarkdown, "markdown listing or pdf booklet")

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

	err = holdFormat(*format, *out)
	if err != nil {
		return err
	}

	file, err := os.Open(flags.Arg(0))
	if err != nil {
		return fmt.Errorf("opening %s: %w", flags.Arg(0), err)
	}

	defer func() { _ = file.Close() }()

	record, err := starmap.Decode(file)
	if err != nil {
		return fmt.Errorf("reading %s: %w", flags.Arg(0), err)
	}

	renderer, err := render.New()
	if err != nil {
		return fmt.Errorf("building the renderer: %w", err)
	}

	if *format == formatPDF {
		return writeBooklet(renderer, record, flags.Arg(0), *out, *force)
	}

	return writeListing(renderer, record, flags.Arg(0), *out, *force, stdout)
}

// holdFormat refuses a format the tool does not write, and a booklet with
// nowhere to put it.
func holdFormat(format, out string) error {
	if format != formatMarkdown && format != formatPDF {
		return fmt.Errorf("%w: %s", errNotAFormat, format)
	}

	// A booklet is a binary, and a terminal is not where one goes.
	if format == formatPDF && out == "" {
		return errPDFWantsAFile
	}

	return nil
}

// writeBooklet writes the printable pages. It is built whole before
// anything is written, so a render that fails leaves no half a file.
func writeBooklet(renderer *render.Renderer, record *starmap.Record, from, out string, force bool) error {
	var booklet bytes.Buffer

	err := renderer.Booklet(&booklet, record)
	if err != nil {
		return fmt.Errorf("rendering %s: %w", from, err)
	}

	return writeFile(out, booklet.Bytes(), force)
}

// writeListing writes the Markdown, to stdout where no file was named.
func writeListing(
	renderer *render.Renderer, record *starmap.Record, from, out string, force bool, stdout io.Writer,
) error {
	var built strings.Builder

	err := renderer.Listing(&built, record)
	if err != nil {
		return fmt.Errorf("rendering %s: %w", from, err)
	}

	if out == "" {
		_, writeErr := io.WriteString(stdout, built.String())
		if writeErr != nil {
			return fmt.Errorf("writing the listing: %w", writeErr)
		}

		return nil
	}

	return writeFile(out, []byte(built.String()), force)
}

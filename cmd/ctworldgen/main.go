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
	"path/filepath"
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
	errNotALaneMode         = errors.New("--lanes is legible or all and nothing else")
	errPDFWantsAFile        = errors.New("--format pdf writes a binary and needs -o")
)

// The two things render writes. The Markdown listing is the default
// because it is what the tool has always written and what a terminal can
// read.
const (
	formatMarkdown = "markdown"
	formatPDF      = "pdf"
)

// The two ways to draw the commercial routes. Legible is the default: p. 2
// offers the map-drawer the choice of ignoring a lane whose worlds are
// already joined, and a dense subsector draws a hundred and sixty lanes
// over forty-six worlds without it (ERRATA E007). The record carries every
// lane either way, and `--lanes all` draws them.
const (
	lanesLegible = "legible"
	lanesAll     = "all"
)

const usage = `ctworldgen generates Classic Traveller subsectors from Book 3 pp. 1-12.

usage:
  ctworldgen new    [--seed N] [--name X] [--occurrence-dm N] [--occurrence-area DM@FROM-TO]... [-o file] [--force]
  ctworldgen sector [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
  ctworldgen render [--format markdown|pdf] [--lanes legible|all] [-o file] [--force] record.json
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

// areaFlag collects --occurrence-area, which the referee may give more
// than once: p. 1 offers a DM "on broad areas within a subsector", plural,
// and a flag is repeatable only by implementing [flag.Value].
//
// The parse itself is starmap's, beside ParseHex. This is the program's
// edge and nothing more.
type areaFlag struct{ areas []starmap.Area }

// String is what flag prints as the default. The zero value has no areas,
// which is the whole-subsector form and prints as nothing.
func (a *areaFlag) String() string {
	if a == nil || len(a.areas) == 0 {
		return ""
	}

	written := make([]string, 0, len(a.areas))
	for _, area := range a.areas {
		written = append(written, area.String())
	}

	return strings.Join(written, ",")
}

// Set reads one area and appends it. The set as a whole -- its DMs, its
// corners and whether any two of them overlap -- is held by
// gen.Inputs.Validate, which is the one place that rule lives.
func (a *areaFlag) Set(text string) error {
	area, err := starmap.ParseArea(text)
	if err != nil {
		return fmt.Errorf("reading --occurrence-area: %w", err)
	}

	a.areas = append(a.areas, area)

	return nil
}

// singleRecordCmd is `new` and `sector`. They draw a seed the same way and
// write one record; what differs is which pass of the engine fills it, and
// whether broad areas may be given at all.
//
// Only `new` takes them. A broad area is a rectangle of one grid's
// numbering, and a sector's sixteen members are each generated on their
// own p. 3 grid (ERRATA E006 part 1), so a sector-grid rectangle would
// have to be clipped into sixteen local ones. Leaving the flag undefined
// on `sector` means asking for one fails by name rather than silently.
func singleRecordCmd(
	subcommand, noun string, args []string, takesAreas bool, stdout, stderr io.Writer,
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

	var areas areaFlag

	if takesAreas {
		flags.Var(&areas, "occurrence-area",
			"a broad area with its own occurrence DM, as -1@0101-0410; repeatable, and areas may not overlap "+
				"(Book 3 p. 1, ERRATA E012)")
	}

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

	record, err := fill(engine, gen.Inputs{
		Seed: *seed, Name: *name, OccurrenceDM: *occurrenceDM, OccurrenceAreas: areas.areas,
	})
	if err != nil {
		return fmt.Errorf("generating the %s: %w", noun, err)
	}

	return write(record, *out, *force, stdout)
}

// newCmd writes one subsector: the whole of Book 3 pp. 1-12 on the p. 3
// grid.
func newCmd(args []string, stdout, stderr io.Writer) error {
	return singleRecordCmd("new", "subsector", args, true, stdout, stderr,
		func(e *gen.Engine, in gen.Inputs) (*starmap.Record, error) { return e.Generate(in) })
}

// sectorCmd writes one record covering sixteen subsectors on one grid,
// with the routes at their seams thrown for (ERRATA E006). Every member is
// the subsector `new --seed base+i` writes, so a sector is the sixteen
// subsectors a referee could have generated one at a time, plus the routes
// that generating them one at a time could not find.
func sectorCmd(args []string, stdout, stderr io.Writer) error {
	return singleRecordCmd("sector", "sector", args, false, stdout, stderr,
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
//
// Without --force, O_EXCL is that promise itself: the file is created or
// nothing happens. With --force it is replaceFile's, because a record is
// the one file the tool asks a referee to keep.
func writeFile(path string, contents []byte, force bool) error {
	if force {
		return replaceFile(path, contents)
	}

	//nolint:gosec // path is the operator's own -o argument
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, recordMode)
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

// replaceFile puts contents where path is, and leaves what is already
// there alone unless the whole of the new file was written.
//
// Opening the target O_TRUNC and writing into it means a write that fails
// partway -- a full disk, a signal, an I/O error -- leaves the referee
// with a truncated file where his record was: the command reports the
// error and the old content is already gone. That is the one file the
// tool asks him to keep. It is what render reads, and it may carry names
// and notes he wrote into it over several sessions.
//
// So the new record is written beside the old one and renamed over it,
// which is atomic on one filesystem. os.CreateTemp creates at 0600, which
// is recordMode, so the record is not widened on its way through.
//
// Two differences from a truncating open, and they are the cost of the
// pattern. A rename needs write permission on the directory rather than
// on the file, so --force into a directory the referee cannot write fails
// where a truncating open would succeed -- and that case could not keep
// his old file either. And where the target is a symbolic or hard link, a
// truncating open writes through it into the file it names, where a
// rename replaces the link with the new record and leaves what it pointed
// at alone.
func replaceFile(path string, contents []byte) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, ".ctworldgen-*")
	if err != nil {
		return fmt.Errorf("opening a temporary file in %s: %w", dir, err)
	}

	// On every path out, including the one where the rename already
	// carried the file off this name and there is nothing left to remove.
	defer func() { _ = os.Remove(tmp.Name()) }()

	_, writeErr := tmp.Write(contents)
	if writeErr != nil {
		_ = tmp.Close()

		return fmt.Errorf("writing %s: %w", tmp.Name(), writeErr)
	}

	err = tmp.Close()
	if err != nil {
		return fmt.Errorf("closing %s: %w", tmp.Name(), err)
	}

	err = os.Rename(tmp.Name(), path)
	if err != nil {
		return fmt.Errorf("putting %s in place: %w", path, err)
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
		suffix := ""
		if dirty {
			suffix = " (dirty)"
		}

		fmt.Fprintf(&out, "revision   %s%s\n", revision, suffix)
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
	lanes := flags.String("lanes", lanesLegible,
		"legible draws only the lanes that join something (ERRATA E007); all draws every lane")

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

	drawn, err := holdLanes(*lanes)
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

	renderer, err := render.New(drawn)
	if err != nil {
		return fmt.Errorf("building the renderer: %w", err)
	}

	if *format == formatPDF {
		return writeBooklet(renderer, record, flags.Arg(0), *out, *force)
	}

	return writeListing(renderer, record, flags.Arg(0), *out, *force, stdout)
}

// holdLanes reads which lanes the documents should draw, and refuses
// anything that is neither.
func holdLanes(lanes string) (render.Lanes, error) {
	switch lanes {
	case lanesLegible:
		return render.LegibleLanes, nil
	case lanesAll:
		return render.AllLanes, nil
	default:
		return render.LegibleLanes, fmt.Errorf("%w: %s", errNotALaneMode, lanes)
	}
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

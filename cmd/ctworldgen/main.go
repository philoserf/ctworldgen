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
	"runtime/debug"
	"strings"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/subsector"
)

// recordMode is the permission a written record gets: the referee's own
// notebook page, not a world-readable one.
const recordMode = 0o600

var (
	errNoSubcommand      = errors.New("no subcommand")
	errUnknownSubcommand = errors.New("unknown subcommand")
	errFileExists        = errors.New("file exists")
	errTakesNoArguments  = errors.New("new takes no arguments (flags precede any filename)")
)

const usage = `ctworldgen generates Classic Traveller subsectors from Book 3 pp. 1-12.

usage:
  ctworldgen new [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
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

	s, err := gen.Generate(gen.Inputs{Seed: *seed, Name: *name, OccurrenceDM: *occurrenceDM})
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

	return binary.BigEndian.Uint64(seedBytes[:]), nil
}

func write(s *subsector.Subsector, path string, force bool, stdout io.Writer) error {
	encoded, err := Marshal(s)
	if err != nil {
		return err
	}

	if path == "" {
		_, writeErr := stdout.Write(encoded)
		if writeErr != nil {
			return fmt.Errorf("writing the record to stdout: %w", writeErr)
		}

		return nil
	}

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

	_, writeErr := file.Write(encoded)
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

// Marshal renders a record as the JSON a golden holds: indented, with a
// trailing newline, so that a fixture is readable and diffs line by line.
func Marshal(s *subsector.Subsector) ([]byte, error) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling the record: %w", err)
	}

	return append(b, '\n'), nil
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

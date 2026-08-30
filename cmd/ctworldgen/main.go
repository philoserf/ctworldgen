// Command ctworldgen generates Classic Traveller subsectors (Book 3
// pp. 1-12, © 1977 text). See docs/PRD.md for the v1 contract.
//
// Subcommands: new, batch, render, replay, version.
package main

import (
	"bytes"
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
	"strings"

	"github.com/philoserf/ctworldgen/dice"
	"github.com/philoserf/ctworldgen/render"
	"github.com/philoserf/ctworldgen/worldgen"
)

// errFileExists refuses to overwrite output the caller did not ask to
// replace. Records are the product, so clobbering one silently would
// destroy work: --force is the only way past this.
var errFileExists = errors.New("already exists (use --force to overwrite)")

// Exit codes: 0 success, 1 operational error, 2 usage error (the flag
// package's own convention).
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

const usage = `usage:
  ctworldgen new [--seed N] [--name X] [--occurrence-dm N] [-o file] [--force]
                 (--occurrence-dm takes -1, 0, or +1: p. 1 offers those and no others)
  ctworldgen batch --count 16 [--seed N] [--name X] [--occurrence-dm N] [-o dir|file.jsonl] [--force]
  ctworldgen render [--history] subsector.json
  ctworldgen replay [--ignore-provenance] subsector.json
  ctworldgen version
`

func main() {
	os.Exit(run(os.Args[1:], randomSeed, os.Stdout, os.Stderr))
}

// randomSeed draws a seed from the OS entropy source: the one deliberate
// exception to the seeded-stream rule, which is engine-scoped. The chosen
// seed is recorded in the record's rng provenance, so replay stays exact.
func randomSeed() (uint64, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0, fmt.Errorf("drawing random seed: %w", err)
	}

	return binary.LittleEndian.Uint64(buf[:]), nil
}

func run(args []string, seedSource func() (uint64, error), stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)

		return exitUsage
	}

	switch args[0] {
	case "-h", "--help":
		// An answered request, not a usage error: usage is this run's
		// output, so it goes to stdout and exits clean.
		fmt.Fprint(stdout, usage)

		return exitOK
	case "new":
		return runNew(args[1:], seedSource, stdout, stderr)
	case "batch":
		return runBatch(args[1:], seedSource, stdout, stderr)
	case "render":
		return runRender(args[1:], stdout, stderr)
	case "replay":
		return runReplay(args[1:], stdout, stderr)
	case "version":
		return runVersion(stdout)
	default:
		fmt.Fprintf(stderr, "ctworldgen: unknown command %q\n", args[0])
		fmt.Fprint(stderr, usage)

		return exitUsage
	}
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Usage = func() {}

	return fs
}

// reportParse turns a flag parse failure into an exit code. -h is a
// request answered on stdout; anything else is a usage error.
func reportParse(fs *flag.FlagSet, err error, stdout, stderr io.Writer) int {
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(stdout, usage)

		return exitOK
	}

	fmt.Fprintf(stderr, "ctworldgen: %v\n", err)
	fmt.Fprint(stderr, usage)

	_ = fs

	return exitUsage
}

// config is the flag set every generating subcommand shares.
type config struct {
	seed         uint64
	name         string
	occurrenceDM int
	out          string
	force        bool
}

func (c *config) bind(fs *flag.FlagSet) {
	fs.Uint64Var(&c.seed, "seed", 0, "seed for the dice stream (default: drawn from OS entropy)")
	fs.StringVar(&c.name, "name", "", "name of the subsector")
	fs.IntVar(&c.occurrenceDM, "occurrence-dm", 0, "referee's subsector-wide world occurrence DM: -1, 0, or +1 (p. 1)")
	fs.StringVar(&c.out, "o", "", "write to this file instead of stdout")
	fs.BoolVar(&c.force, "force", false, "overwrite an existing output file")
}

func runNew(args []string, seedSource func() (uint64, error), stdout, stderr io.Writer) int {
	fs := newFlagSet("new")

	cfg := &config{}
	cfg.bind(fs)

	if err := fs.Parse(args); err != nil {
		return reportParse(fs, err, stdout, stderr)
	}

	if err := resolveSeed(fs, &cfg.seed, seedSource); err != nil {
		fmt.Fprintf(stderr, "ctworldgen: %v\n", err)

		return exitError
	}

	sub, err := worldgen.Generate(worldgen.Config{
		Seed: cfg.seed, Name: cfg.name, OccurrenceDM: cfg.occurrenceDM,
	})
	if err != nil {
		fmt.Fprintf(stderr, "ctworldgen: %v\n", err)

		return exitError
	}

	if err := emitRecord(sub, cfg.out, cfg.force, stdout); err != nil {
		fmt.Fprintf(stderr, "ctworldgen: %v\n", err)

		return exitError
	}

	return exitOK
}

func runBatch(args []string, seedSource func() (uint64, error), stdout, stderr io.Writer) int {
	fs := newFlagSet("batch")

	cfg := &config{}
	cfg.bind(fs)

	count := fs.Int("count", 0, "how many subsectors to generate (a sector is 16)")

	if err := fs.Parse(args); err != nil {
		return reportParse(fs, err, stdout, stderr)
	}

	if *count < 1 {
		fmt.Fprintf(stderr, "ctworldgen: batch needs --count of 1 or more\n")

		return exitUsage
	}

	if err := resolveSeed(fs, &cfg.seed, seedSource); err != nil {
		fmt.Fprintf(stderr, "ctworldgen: %v\n", err)

		return exitError
	}

	subs, err := generateBatch(cfg, *count)
	if err != nil {
		fmt.Fprintf(stderr, "ctworldgen: %v\n", err)

		return exitError
	}

	if err := emitBatch(subs, cfg.out, cfg.force, stdout); err != nil {
		fmt.Fprintf(stderr, "ctworldgen: %v\n", err)

		return exitError
	}

	return exitOK
}

// generateBatch derives each member's seed from the base seed plus its
// index, and records that derived seed in the member's own record — so
// every subsector of a batch replays on its own, without the batch.
func generateBatch(cfg *config, count int) ([]*worldgen.Subsector, error) {
	subs := make([]*worldgen.Subsector, 0, count)

	for i := range count {
		sub, err := worldgen.Generate(worldgen.Config{
			Seed:         cfg.seed + uint64(i),
			Name:         cfg.name,
			OccurrenceDM: cfg.occurrenceDM,
		})
		if err != nil {
			return nil, fmt.Errorf("subsector %d of %d: %w", i+1, count, err)
		}

		subs = append(subs, sub)
	}

	return subs, nil
}

func runReplay(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("replay")
	ignore := fs.Bool("ignore-provenance", false, "waive the version-stamp match (and nothing else)")

	if err := fs.Parse(args); err != nil {
		return reportParse(fs, err, stdout, stderr)
	}

	sub, code := readRecord(fs.Args(), stderr)
	if sub == nil {
		return code
	}

	if err := worldgen.Replay(sub, *ignore); err != nil {
		fmt.Fprintf(stderr, "ctworldgen: %v\n", err)

		return exitError
	}

	fmt.Fprintf(stdout, "replay verified: %d events reproduced from seed %d\n", len(sub.Events), sub.RNG.Seed)

	return exitOK
}

func runRender(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("render")
	history := fs.Bool("history", false, "render the generation transcript instead of the subsector listing")

	if err := fs.Parse(args); err != nil {
		return reportParse(fs, err, stdout, stderr)
	}

	sub, code := readRecord(fs.Args(), stderr)
	if sub == nil {
		return code
	}

	if *history {
		fmt.Fprint(stdout, render.History(sub))

		return exitOK
	}

	out, err := render.Listing(sub)
	if err != nil {
		fmt.Fprintf(stderr, "ctworldgen: %v\n", err)

		return exitError
	}

	fmt.Fprint(stdout, out)

	return exitOK
}

// readRecord reads the one filename the subcommand takes. It reports a nil
// subsector with the exit code on any failure.
func readRecord(args []string, stderr io.Writer) (*worldgen.Subsector, int) {
	if len(args) != 1 {
		fmt.Fprintf(stderr, "ctworldgen: expected exactly one subsector file\n")
		fmt.Fprint(stderr, usage)

		return nil, exitUsage
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "ctworldgen: %v\n", err)

		return nil, exitError
	}

	sub, err := worldgen.UnmarshalRecord(data)
	if err != nil {
		fmt.Fprintf(stderr, "ctworldgen: %v\n", err)

		return nil, exitError
	}

	return sub, exitOK
}

func runVersion(stdout io.Writer) int {
	build := "(unknown)"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		build = info.Main.Version
	}

	fmt.Fprintf(stdout, "ctworldgen %s\n", build)
	fmt.Fprintf(stdout, "schema_version %s\n", worldgen.SchemaVersion)
	fmt.Fprintf(stdout, "engine_version %s\n", worldgen.EngineVersion)
	fmt.Fprintf(stdout, "ruleset %s\n", worldgen.Ruleset)
	fmt.Fprintf(stdout, "rng %s\n", dice.Algorithm)

	return exitOK
}

// resolveSeed draws a seed only when --seed was not given, so that
// `--seed 0` is an explicit and reproducible choice rather than a request
// for a random one.
func resolveSeed(fs *flag.FlagSet, seed *uint64, seedSource func() (uint64, error)) error {
	given := false

	fs.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			given = true
		}
	})

	if given {
		return nil
	}

	drawn, err := seedSource()
	if err != nil {
		return err
	}

	*seed = drawn

	return nil
}

func emitRecord(sub *worldgen.Subsector, outPath string, force bool, stdout io.Writer) error {
	data, err := sub.MarshalRecord()
	if err != nil {
		return fmt.Errorf("rendering the record: %w", err)
	}

	if outPath == "" {
		_, err := stdout.Write(data)
		if err != nil {
			return fmt.Errorf("writing the record: %w", err)
		}

		return nil
	}

	return writeFile(outPath, data, force)
}

func emitBatch(subs []*worldgen.Subsector, outPath string, force bool, stdout io.Writer) error {
	if isDirTarget(outPath) {
		return writeBatchDir(subs, outPath, force)
	}

	data, err := batchJSONL(subs)
	if err != nil {
		return err
	}

	if outPath == "" {
		if _, err := stdout.Write(data); err != nil {
			return fmt.Errorf("writing the batch: %w", err)
		}

		return nil
	}

	return writeFile(outPath, data, force)
}

// isDirTarget reports whether -o names a directory to fill rather than a
// file to write. A trailing separator says so even when the directory does
// not exist yet — writeBatchDir creates it — so the documented
// `-o sector/` works on a fresh checkout rather than failing obscurely on
// a missing path.
func isDirTarget(outPath string) bool {
	if outPath == "" || strings.HasSuffix(outPath, ".jsonl") {
		return false
	}

	if strings.HasSuffix(outPath, string(os.PathSeparator)) || strings.HasSuffix(outPath, "/") {
		return true
	}

	// The error is checked in its own statement rather than short-circuited
	// into the return: NilAway reads the two-value Stat as an unguarded
	// dereference otherwise, and an explicit guard is the honest shape.
	info, err := os.Stat(outPath)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// batchJSONL is one record per line: the canonical record with its
// indentation removed, which is what makes a line a record.
func batchJSONL(subs []*worldgen.Subsector) ([]byte, error) {
	var b strings.Builder

	for _, sub := range subs {
		data, err := sub.MarshalRecord()
		if err != nil {
			return nil, fmt.Errorf("rendering a batch record: %w", err)
		}

		compact, err := compactJSON(data)
		if err != nil {
			return nil, err
		}

		b.Write(compact)
		b.WriteString("\n")
	}

	return []byte(b.String()), nil
}

// compactJSON strips the canonical record's indentation so that one
// record occupies one line.
func compactJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		return nil, fmt.Errorf("compacting a batch record: %w", err)
	}

	return buf.Bytes(), nil
}

func writeBatchDir(subs []*worldgen.Subsector, dir string, force bool) error {
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec // a directory of records a referee reads
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	paths := make([]string, 0, len(subs))
	for i := range subs {
		paths = append(paths, filepath.Join(dir, fmt.Sprintf("subsector-%02d.json", i+1)))
	}

	if !force {
		if err := checkNoneExist(paths); err != nil {
			return err
		}
	}

	for i, sub := range subs {
		data, err := sub.MarshalRecord()
		if err != nil {
			return fmt.Errorf("rendering %s: %w", paths[i], err)
		}

		if err := writeFile(paths[i], data, true); err != nil {
			return err
		}
	}

	return nil
}

// checkNoneExist refuses the whole batch if any member's file is already
// there, so a partial write cannot leave a directory half replaced.
func checkNoneExist(paths []string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return fmt.Errorf("%s %w", p, errFileExists)
		}
	}

	return nil
}

func writeFile(path string, data []byte, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s %w", path, errFileExists)
		}
	}

	if err := os.WriteFile(path, data, 0o644); err != nil { //nolint:gosec // a record a referee reads and shares
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

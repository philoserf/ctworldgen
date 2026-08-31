// Command ratchet compares uncovered statements per package against a
// checked-in baseline.
//
// It counts uncovered statements rather than a percentage, because a
// percentage can hold still while a guarded branch adds one covered
// statement and one uncovered one. It fails in both directions: a package
// that gains uncovered statements has lost coverage, and one that loses
// them has gained coverage the baseline should record.
//
//	ratchet -profile coverage.out -baseline coverage.baseline
//	ratchet -profile coverage.out -baseline coverage.baseline -update
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	fileMode = 0o600

	// a baseline line is "<package> <count>".
	baselineFields = 2
)

var (
	errUnreadableLine = errors.New("cannot read the line")
	errCoverageMoved  = errors.New("coverage moved")
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ratchet:", err)
		os.Exit(1)
	}
}

func run() error {
	profile := flag.String("profile", "coverage.out", "coverage profile to read")
	baseline := flag.String("baseline", "coverage.baseline", "checked-in baseline to compare against")
	update := flag.Bool("update", false, "rewrite the baseline from the profile")

	flag.Parse()

	current, err := uncovered(*profile)
	if err != nil {
		return err
	}

	if *update {
		return writeBaseline(*baseline, current)
	}

	want, err := readBaseline(*baseline)
	if err != nil {
		return fmt.Errorf("%w (run `task ratchet:update` to create it)", err)
	}

	return compare(want, current)
}

// uncovered counts the statements a profile reports as never executed,
// per package.
func uncovered(profile string) (map[string]int, error) {
	file, err := os.Open(profile) //nolint:gosec // the profile path is a flag of this tool
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", profile, err)
	}

	defer func() { _ = file.Close() }()

	counts := map[string]int{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}

		block, parseErr := parseProfileLine(line)
		if parseErr != nil {
			return nil, fmt.Errorf("%s: %w", profile, parseErr)
		}

		if _, seen := counts[block.pkg]; !seen {
			counts[block.pkg] = 0
		}

		if block.hits == 0 {
			counts[block.pkg] += block.statements
		}
	}

	err = scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", profile, err)
	}

	return counts, nil
}

// block is one line of a coverage profile: the package the block belongs
// to, how many statements it holds, and how many times it ran.
type block struct {
	pkg        string
	statements int
	hits       int
}

// profileFields is the field count of a profile line:
// "name.go:line.col,line.col numberOfStatements count".
const profileFields = 3

func parseProfileLine(line string) (block, error) {
	colon := strings.LastIndex(line, ":")

	fields := strings.Fields(line)
	if colon < 0 || len(fields) != profileFields {
		return block{}, fmt.Errorf("%w: %q", errUnreadableLine, line)
	}

	statements, err := strconv.Atoi(fields[1])
	if err != nil {
		return block{}, fmt.Errorf("statement count %q: %w", fields[1], err)
	}

	hits, err := strconv.Atoi(fields[2])
	if err != nil {
		return block{}, fmt.Errorf("hit count %q: %w", fields[2], err)
	}

	return block{pkg: path.Dir(line[:colon]), statements: statements, hits: hits}, nil
}

func writeBaseline(path string, counts map[string]int) error {
	var out strings.Builder

	out.WriteString("# Uncovered statements per package. Written by `task ratchet:update`.\n")
	out.WriteString("# The gate fails in both directions: gaining uncovered statements is\n")
	out.WriteString("# lost coverage, and losing them is coverage the baseline should record.\n")

	for _, pkg := range sorted(counts) {
		fmt.Fprintf(&out, "%s %d\n", pkg, counts[pkg])
	}

	err := os.WriteFile(path, []byte(out.String()), fileMode)
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "wrote", path)

	return nil
}

func readBaseline(path string) (map[string]int, error) {
	encoded, err := os.ReadFile(path) //nolint:gosec // the baseline path is a flag of this tool
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	counts := map[string]int{}

	for line := range strings.SplitSeq(string(encoded), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != baselineFields {
			return nil, fmt.Errorf("%s: %w: %q", path, errUnreadableLine, line)
		}

		n, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}

		counts[fields[0]] = n
	}

	return counts, nil
}

func compare(want, got map[string]int) error {
	var problems []string

	for _, pkg := range sorted(want) {
		current, measured := got[pkg]
		if !measured {
			problems = append(problems, pkg+" is in the baseline and the profile does not cover it")

			continue
		}

		switch {
		case current > want[pkg]:
			problems = append(problems,
				fmt.Sprintf("%s has %d uncovered statements, up from %d", pkg, current, want[pkg]))
		case current < want[pkg]:
			problems = append(problems,
				fmt.Sprintf("%s has %d uncovered statements, down from %d: run `task ratchet:update`",
					pkg, current, want[pkg]))
		}
	}

	for _, pkg := range sorted(got) {
		if _, known := want[pkg]; !known {
			problems = append(problems, pkg+" is not in the baseline: run `task ratchet:update`")
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("%w:\n  %s", errCoverageMoved, strings.Join(problems, "\n  "))
	}

	_, _ = fmt.Fprintln(os.Stdout, "coverage holds at the baseline")

	return nil
}

func sorted(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

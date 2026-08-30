package worldgen_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/worldgen"
)

// The documents in docs/ are held to the code here, so that a reading the
// engine applies and the document does not describe — or the reverse —
// fails the gate rather than waiting to mislead a reader.

var errataHeading = regexp.MustCompile(`(?m)^## (E\d{3}) — (.+)$`)

// TestErrataIDsMatchTheDocument: every reading the engine can stamp has an
// entry, and every entry is a reading the engine can stamp.
func TestErrataIDsMatchTheDocument(t *testing.T) {
	doc := readDoc(t, filepath.Join("..", "docs", "ERRATA.md"))

	found := errataHeading.FindAllStringSubmatch(doc, -1)

	documented := make([]string, 0, len(found))
	for _, m := range found {
		documented = append(documented, m[1])
	}

	stampable := worldgen.ErrataIDs()

	slices.Sort(documented)
	slices.Sort(stampable)

	if !slices.Equal(documented, stampable) {
		t.Errorf("docs/ERRATA.md documents %v, the engine stamps %v", documented, stampable)
	}
}

// TestErrataEntriesCiteAPage: a reading without a page cite cannot be
// checked against the book, which is the only thing that makes it a
// reading rather than an invention.
func TestErrataEntriesCiteAPage(t *testing.T) {
	doc := readDoc(t, filepath.Join("..", "docs", "ERRATA.md"))
	cite := regexp.MustCompile(`\bpp?\. ?\d`)

	for _, m := range errataHeading.FindAllStringSubmatch(doc, -1) {
		if !cite.MatchString(m[2]) {
			t.Errorf("%s's heading cites no page: %q", m[1], m[2])
		}
	}
}

// TestErrataDocumentSaysWhenEachIsStamped: the array in a record is only
// useful if a reader can find out, from the entry, why it is there.
func TestErrataDocumentSaysWhenEachIsStamped(t *testing.T) {
	doc := readDoc(t, filepath.Join("..", "docs", "ERRATA.md"))
	// Split drops the headings and their capture groups, so the first
	// chunk is the preamble and each one after it is an entry's body.
	sections := errataHeading.Split(doc, -1)
	headings := errataHeading.FindAllStringSubmatch(doc, -1)

	if len(sections) != len(headings)+1 {
		t.Fatalf("split docs/ERRATA.md into %d chunks for %d entries", len(sections), len(headings))
	}

	for i, heading := range headings {
		if !strings.Contains(sections[i+1], "Stamped on") {
			t.Errorf("%s does not say when it is stamped", heading[1])
		}
	}
}

// TestCoverageNamesRealTests: COVERAGE.md's whole value is that its test
// column can be followed, so a renamed or deleted test must fail here
// rather than leave a dead reference.
func TestCoverageNamesRealTests(t *testing.T) {
	doc := readDoc(t, filepath.Join("..", "docs", "COVERAGE.md"))
	declared := testFunctionsInRepo(t)

	named := regexp.MustCompile("`(Test[A-Za-z0-9_]+)`").FindAllStringSubmatch(doc, -1)
	if len(named) == 0 {
		t.Fatal("COVERAGE.md names no tests at all")
	}

	for _, m := range named {
		if !slices.Contains(declared, m[1]) {
			t.Errorf("COVERAGE.md names %s, which no test declares", m[1])
		}
	}
}

// TestCoverageNamesRealFiles: the implementation column too.
func TestCoverageNamesRealFiles(t *testing.T) {
	doc := readDoc(t, filepath.Join("..", "docs", "COVERAGE.md"))

	named := regexp.MustCompile("`([a-z][a-z0-9_/]*\\.(?:go|json))`").FindAllStringSubmatch(doc, -1)
	if len(named) == 0 {
		t.Fatal("COVERAGE.md names no source files at all")
	}

	for _, m := range named {
		if _, err := os.Stat(filepath.Join("..", m[1])); err != nil {
			t.Errorf("COVERAGE.md names %s, which is not in the repository", m[1])
		}
	}
}

func readDoc(t *testing.T, path string) string {
	t.Helper()

	raw, err := os.ReadFile(path) //nolint:gosec // a docs/ path this test names
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	return string(raw)
}

// testFunctionsInRepo scans every _test.go file for the tests it declares.
func testFunctionsInRepo(t *testing.T) []string {
	t.Helper()

	decl := regexp.MustCompile(`(?m)^func (Test[A-Za-z0-9_]+)\(`)

	var found []string

	err := filepath.WalkDir("..", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		raw, err := os.ReadFile(path) //nolint:gosec // walking this repository's own tree
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		for _, m := range decl.FindAllStringSubmatch(string(raw), -1) {
			found = append(found, m[1])
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walking the repository: %v", err)
	}

	return found
}

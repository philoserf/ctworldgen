// Package audit holds the checks that read the repository itself rather
// than a package's behaviour: that every record the engine writes matches
// the published schema, and that the readings cited in the code and the
// documents are exactly the readings ERRATA.md records. It is test
// support, not architecture.
package audit

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var errNoModule = errors.New("no go.mod above the working directory")

// Root returns the repository root, found by walking up from the working
// directory to the go.mod.
func Root() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("finding the working directory: %w", err)
	}

	for {
		_, err := os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%w: %s", errNoModule, dir)
		}

		dir = parent
	}
}

// scannable reports whether a file can cite a reading. ERRATA.md itself
// is excluded: it is where the readings are recorded, not where they are
// applied.
func scannable(rel, path string) bool {
	if rel == filepath.Join("docs", "ERRATA.md") {
		return false
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".md", ".json", ".yml", ".yaml":
		return true
	default:
		return false
	}
}

// erratumID matches the identifiers ERRATA.md gives its readings.
var erratumID = regexp.MustCompile(`\bE\d{3}\b`)

// headingID matches the identifier at the head of an ERRATA.md entry.
var headingID = regexp.MustCompile(`(?m)^## (E\d{3}) `)

// Headings returns the readings ERRATA.md records, in document order.
func Headings(root string) ([]string, error) {
	path := filepath.Join(root, "docs", "ERRATA.md")

	encoded, err := os.ReadFile(path) //nolint:gosec // a fixed path inside the repository
	if err != nil {
		return nil, fmt.Errorf("reading docs/ERRATA.md: %w", err)
	}

	var ids []string

	for _, m := range headingID.FindAllStringSubmatch(string(encoded), -1) {
		ids = append(ids, m[1])
	}

	return ids, nil
}

// Citations returns every reading cited anywhere in the repository except
// ERRATA.md itself, mapped to the files that cite it.
func Citations(root string) (map[string][]string, error) {
	cited := map[string][]string{}

	err := filepath.WalkDir(root, collect(root, cited))
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", root, err)
	}

	for id := range cited {
		slices.Sort(cited[id])
	}

	return cited, nil
}

// collect returns the walk function that records each file's citations.
func collect(root string, cited map[string][]string) fs.WalkDirFunc {
	skipDirs := map[string]bool{".git": true, "testdata": true}

	return func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if skipDirs[entry.Name()] {
				return fs.SkipDir
			}

			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("relative path for %s: %w", path, err)
		}

		if !scannable(rel, path) {
			return nil
		}

		b, err := os.ReadFile(path) //nolint:gosec // a path this walk produced, inside the repository
		if err != nil {
			return fmt.Errorf("reading %s: %w", rel, err)
		}

		for _, id := range erratumID.FindAllString(string(b), -1) {
			if !slices.Contains(cited[id], rel) {
				cited[id] = append(cited[id], rel)
			}
		}

		return nil
	}
}

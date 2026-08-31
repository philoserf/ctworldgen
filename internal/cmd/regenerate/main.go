// Command regenerate rewrites the golden fixtures from the roster in
// internal/fixture. Fixtures move only by regeneration, never by hand:
// run it, then read the diff.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
)

const (
	dirMode  = 0o750
	fileMode = 0o600
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "regenerate:", err)
		os.Exit(1)
	}
}

func run() error {
	dir := filepath.Join("gen", "testdata")

	err := os.MkdirAll(dir, dirMode)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	for _, golden := range fixture.Goldens() {
		record, err := gen.Generate(gen.Inputs{Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM})
		if err != nil {
			return fmt.Errorf("%s: %w", golden.File, err)
		}

		encoded, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			return fmt.Errorf("%s: %w", golden.File, err)
		}

		path := filepath.Join(dir, golden.File+".json")

		err = os.WriteFile(path, append(encoded, '\n'), fileMode)
		if err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		_, _ = fmt.Fprintln(os.Stdout, "wrote", path, "--", len(record.Worlds), "worlds")
	}

	return nil
}

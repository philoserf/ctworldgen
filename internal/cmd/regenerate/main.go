// Command regenerate rewrites the golden fixtures, and the complete
// example record shipped beside the schema, from the roster in
// internal/fixture. Both move only by regeneration, never by hand: run it,
// then read the diff.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/subsector"
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

	engine, err := gen.New()
	if err != nil {
		return fmt.Errorf("building the engine: %w", err)
	}

	for _, golden := range fixture.Goldens() {
		record, err := engine.Generate(gen.Inputs{Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM})
		if err != nil {
			return fmt.Errorf("%s: %w", golden.File, err)
		}

		encoded, err := subsector.Marshal(record)
		if err != nil {
			return fmt.Errorf("%s: %w", golden.File, err)
		}

		path := filepath.Join(dir, golden.File+".json")

		err = os.WriteFile(path, encoded, fileMode)
		if err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		_, _ = fmt.Fprintln(os.Stdout, "wrote", path, "--", len(record.Worlds), "worlds")
	}

	return writeExample()
}

// writeExample rewrites the complete example record shipped beside the
// schema. It is documentation, but it is documentation the engine writes.
func writeExample() error {
	example := fixture.CompleteExample()

	engine, err := gen.New()
	if err != nil {
		return fmt.Errorf("building the engine: %w", err)
	}

	record, err := engine.Generate(gen.Inputs{Seed: example.Seed, Name: example.Name, OccurrenceDM: example.OccurrenceDM})
	if err != nil {
		return fmt.Errorf("%s: %w", fixture.CompleteExamplePath(), err)
	}

	encoded, err := subsector.Marshal(record)
	if err != nil {
		return fmt.Errorf("%s: %w", fixture.CompleteExamplePath(), err)
	}

	err = os.WriteFile(fixture.CompleteExamplePath(), encoded, fileMode)
	if err != nil {
		return fmt.Errorf("writing %s: %w", fixture.CompleteExamplePath(), err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "wrote", fixture.CompleteExamplePath(), "--", len(record.Worlds), "worlds")

	return nil
}

// Command regenerate rewrites the golden fixtures, and the complete
// example record shipped beside the schema, from the roster in
// internal/fixture. Both move only by regeneration, never by hand: run it,
// then read the diff.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/render"
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
		record, genErr := engine.Generate(gen.Inputs{
			Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM,
		})
		if genErr != nil {
			return fmt.Errorf("%s: %w", golden.File, genErr)
		}

		encoded, marshalErr := subsector.Marshal(record)
		if marshalErr != nil {
			return fmt.Errorf("%s: %w", golden.File, marshalErr)
		}

		path := filepath.Join(dir, golden.File+".json")

		err = os.WriteFile(path, encoded, fileMode)
		if err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		_, _ = fmt.Fprintln(os.Stdout, "wrote", path, "--", len(record.Worlds), "worlds")
	}

	err = writeExample()
	if err != nil {
		return err
	}

	return writeListings()
}

// writeListings rewrites the Markdown golden for every fixture. The two
// golden trees are driven from the one roster, so they cannot come to
// describe different subsectors under the same name.
func writeListings() error {
	dir := filepath.Join("render", "testdata")

	err := os.MkdirAll(dir, dirMode)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	engine, err := gen.New()
	if err != nil {
		return fmt.Errorf("building the engine: %w", err)
	}

	renderer, rendererErr := render.New()
	if rendererErr != nil {
		return fmt.Errorf("building the renderer: %w", rendererErr)
	}

	for _, golden := range fixture.Goldens() {
		record, genErr := engine.Generate(gen.Inputs{
			Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM,
		})
		if genErr != nil {
			return fmt.Errorf("%s: %w", golden.File, genErr)
		}

		var built strings.Builder

		renderErr := renderer.Subsector(&built, record)
		if renderErr != nil {
			return fmt.Errorf("%s: %w", golden.File, renderErr)
		}

		path := filepath.Join(dir, golden.File+".md")

		writeErr := os.WriteFile(path, []byte(built.String()), fileMode)
		if writeErr != nil {
			return fmt.Errorf("writing %s: %w", path, writeErr)
		}

		_, _ = fmt.Fprintln(os.Stdout, "wrote", path)
	}

	return nil
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

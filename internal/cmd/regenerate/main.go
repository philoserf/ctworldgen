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

// run rewrites both golden trees. Each fixture is generated once and
// written twice -- the JSON record and the Markdown listing rendered from
// that same record -- so the two trees cannot come to describe different
// subsectors under the same name.
func run() error {
	recordDir := filepath.Join("gen", "testdata")
	listingDir := filepath.Join("render", "testdata")

	for _, dir := range []string{recordDir, listingDir} {
		err := os.MkdirAll(dir, dirMode)
		if err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}

	engine, err := gen.New()
	if err != nil {
		return fmt.Errorf("building the engine: %w", err)
	}

	renderer, err := render.New()
	if err != nil {
		return fmt.Errorf("building the renderer: %w", err)
	}

	for _, golden := range fixture.Goldens() {
		err = writeGolden(engine, renderer, golden, recordDir, listingDir)
		if err != nil {
			return err
		}
	}

	return writeExample()
}

// writeGolden generates one fixture and writes both of its goldens.
func writeGolden(
	engine *gen.Engine, renderer *render.Renderer, golden fixture.Golden, recordDir, listingDir string,
) error {
	record, err := engine.Generate(gen.Inputs{
		Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", golden.File, err)
	}

	encoded, err := subsector.Marshal(record)
	if err != nil {
		return fmt.Errorf("%s: %w", golden.File, err)
	}

	recordPath := filepath.Join(recordDir, golden.File+".json")

	err = os.WriteFile(recordPath, encoded, fileMode)
	if err != nil {
		return fmt.Errorf("writing %s: %w", recordPath, err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "wrote", recordPath, "--", len(record.Worlds), "worlds")

	var built strings.Builder

	err = renderer.Subsector(&built, record)
	if err != nil {
		return fmt.Errorf("%s: %w", golden.File, err)
	}

	listingPath := filepath.Join(listingDir, golden.File+".md")

	err = os.WriteFile(listingPath, []byte(built.String()), fileMode)
	if err != nil {
		return fmt.Errorf("writing %s: %w", listingPath, err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "wrote", listingPath)

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

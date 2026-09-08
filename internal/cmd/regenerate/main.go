// Command regenerate rewrites the golden fixtures, and the complete
// example record shipped beside the schema, from the roster in
// internal/fixture. Both move only by regeneration, never by hand: run it,
// then read the diff.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/render"
	"github.com/philoserf/ctworldgen/starmap"
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

	renderer, err := render.New(render.LegibleLanes)
	if err != nil {
		return fmt.Errorf("building the renderer: %w", err)
	}

	for _, golden := range fixture.Goldens() {
		err = writeGolden(engine, renderer, golden, recordDir, listingDir)
		if err != nil {
			return err
		}
	}

	// The sector is generated once and handed to both writers. Deriving
	// it separately in each made the seam golden and the listing slice
	// agree only because two derivations happened to match -- and
	// internal/fixture exists so that two goldens cannot come to describe
	// different subsectors under one name. Here alone that held by
	// coincidence rather than by construction.
	sector := fixture.SectorGolden()

	sectorRecord, err := engine.Sector(gen.Inputs{
		Seed: sector.Seed, Name: sector.Name, OccurrenceDM: sector.OccurrenceDM,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", sector.File, err)
	}

	err = writeSeams(sectorRecord, sector.File)
	if err != nil {
		return err
	}

	err = writeSectorSlice(renderer, sectorRecord)
	if err != nil {
		return err
	}

	return writeExample(engine)
}

// writeSectorSlice pins one member's section of a sector's listing: the
// decomposition, the ring of neighbours its map draws, and the crossing
// lanes it shares with its neighbours (ERRATA E008).
func writeSectorSlice(renderer *render.Renderer, record *starmap.Record) error {
	var whole strings.Builder

	err := renderer.Listing(&whole, record)
	if err != nil {
		return fmt.Errorf("rendering the sector: %w", err)
	}

	path := fixture.SectorSlicePath()

	err = os.WriteFile(path, []byte(fixture.SectorSlice(whole.String())), fileMode)
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "wrote", path, "-- subsector", fixture.SectorSliceMember)

	return nil
}

// writeSeams pins the one stream a sector adds: the route pass over pairs
// that straddle two members (ERRATA E006). The members themselves are
// pinned by being identical to the subsectors `new` writes, which is a
// test rather than a fixture.
func writeSeams(record *starmap.Record, file string) error {
	crossing := gen.CrossingRoutes(record)

	encoded, err := json.MarshalIndent(crossing, "", "  ")
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	path := fixture.SeamsPath()

	err = os.WriteFile(path, append(encoded, '\n'), fileMode)
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "wrote", path, "--", len(crossing), "routes at the seams")

	return nil
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

	encoded, err := starmap.Marshal(record)
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

	err = renderer.Listing(&built, record)
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
func writeExample(engine *gen.Engine) error {
	example := fixture.CompleteExample()

	record, err := engine.Generate(gen.Inputs{Seed: example.Seed, Name: example.Name, OccurrenceDM: example.OccurrenceDM})
	if err != nil {
		return fmt.Errorf("%s: %w", fixture.CompleteExamplePath(), err)
	}

	encoded, err := starmap.Marshal(record)
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

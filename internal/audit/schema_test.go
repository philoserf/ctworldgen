package audit_test

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/starmap"
)

func schema(t *testing.T, root string) *jsonschema.Schema {
	t.Helper()

	path := filepath.Join(root, "docs", "record.schema.json")

	file, err := os.Open(path) //nolint:gosec // a fixed path inside the repository
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = file.Close() }()

	doc, unmarshalErr := jsonschema.UnmarshalJSON(file)
	if unmarshalErr != nil {
		t.Fatal(unmarshalErr)
	}

	compiler := jsonschema.NewCompiler()

	addErr := compiler.AddResource("record.schema.json", doc)
	if addErr != nil {
		t.Fatal(addErr)
	}

	compiled, compileErr := compiler.Compile("record.schema.json")
	if compileErr != nil {
		t.Fatal(compileErr)
	}

	return compiled
}

func validate(t *testing.T, compiled *jsonschema.Schema, path string) {
	t.Helper()

	b, err := os.ReadFile(path) //nolint:gosec // a fixed path inside the repository
	if err != nil {
		t.Fatal(err)
	}

	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}

	err = compiled.Validate(inst)
	if err != nil {
		t.Errorf("%s does not validate:\n%v", path, err)
	}
}

// TestGoldensValidate checks every record the engine writes against the
// published schema.
func TestGoldensValidate(t *testing.T) {
	t.Parallel()

	repoRoot := root(t)

	s := schema(t, repoRoot)
	for _, g := range fixture.Goldens() {
		t.Run(g.File, func(t *testing.T) {
			t.Parallel()
			validate(t, s, filepath.Join(repoRoot, "gen", "testdata", g.File+".json"))
		})
	}
}

// TestExamplesValidate checks the two example records shipped beside the
// schema. The minimal one is an empty subsector, which is a result: a run
// whose eighty throws place no world produces a valid record with no
// worlds, and nothing rerolls.
func TestExamplesValidate(t *testing.T) {
	t.Parallel()

	repoRoot := root(t)

	s := schema(t, repoRoot)
	for _, name := range []string{"minimal.json", "complete.json"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			validate(t, s, filepath.Join(repoRoot, "docs", "examples", name))
		})
	}
}

// TestTheCompleteExampleIsAGeneratedRecord pins the example shipped beside
// the schema the way TestGoldens pins a golden. It is documentation, but it
// is documentation `ctworldgen new` writes, and an example that had drifted
// from what the engine produces would document a record shape that does not
// exist. It moves only by `task regenerate`.
func TestTheCompleteExampleIsAGeneratedRecord(t *testing.T) {
	t.Parallel()

	example := fixture.CompleteExample()

	want, err := os.ReadFile(filepath.Join(root(t), fixture.CompleteExamplePath()))
	if err != nil {
		t.Fatal(err)
	}

	engine, err := gen.New()
	if err != nil {
		t.Fatal(err)
	}

	in := gen.Inputs{
		Seed: example.Seed, Name: example.Name, OccurrenceDM: example.OccurrenceDM,
		OccurrenceAreas: example.OccurrenceAreas,
	}

	generated, err := engine.Generate(in)
	if err != nil {
		t.Fatal(err)
	}

	got, err := starmap.Marshal(generated)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(want) {
		t.Errorf("%s does not match what the engine writes for seed %d, %q and DM %+d.\n"+
			"If this change was intended, run `task regenerate` and read the diff.",
			fixture.CompleteExamplePath(), example.Seed, example.Name, example.OccurrenceDM)
	}
}

// TestSchemaRejectsUnknownFields is half of the two obligations: the
// schema says additionalProperties false at every level, and the Go side
// says DisallowUnknownFields. A schema alone rejects nothing at read time,
// so both are required.
func TestSchemaRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	repoRoot := root(t)
	compiled := schema(t, repoRoot)

	for _, bad := range slices.Concat(badRecords(), badAreaRecords()) {
		t.Run(bad.name, func(t *testing.T) {
			t.Parallel()

			record := loadRecord(t, filepath.Join(repoRoot, "docs", "examples", "complete.json"))

			bad.mutate(record, firstObject(t, record, "worlds"), firstObject(t, record, "routes"))

			err := compiled.Validate(any(record))
			if err == nil {
				t.Errorf("the schema accepted %s", bad.name)
			}
		})
	}
}

// badRecord is one mutation of the complete example that the schema must
// refuse, named for what it does.
type badRecord struct {
	name   string
	mutate func(record, world, route map[string]any)
}

// badRecords is the table. It sits outside the test so that the list can
// grow without the test itself growing with it.
// offTheNumbering is a four-digit identifier the grid never prints: no
// hex is in row 00.
const offTheNumbering = "0900"

func badRecords() []badRecord {
	return []badRecord{
		{"an unknown field at the top level", func(record, _, _ map[string]any) {
			record["surprise"] = 1
		}},
		{"an unknown field on a world", func(_, world, _ map[string]any) {
			world["surprise"] = 1
		}},
		// Not "off the p. 3 grid": the pattern now spans the sector grid
		// too, so 0910 is a well-formed identifier and it is
		// starmap.Decode, not the schema, that refuses it on a p. 3
		// record. What the pattern still refuses is a number the grid
		// never prints at all.
		{"a hex the grid numbering does not print", func(_, world, _ map[string]any) {
			world["hex"] = offTheNumbering
		}},
		{"a hex that is not four digits", func(_, world, _ map[string]any) {
			world["hex"] = "1-5"
		}},
		{"a starport the book does not print", func(_, world, _ map[string]any) {
			world["starport"] = "Q"
		}},
		{"an occurrence DM the book does not offer", func(record, _, _ map[string]any) {
			record["occurrence_dm"] = 2
		}},
		// The complete example carries no broad areas, so these add the
		// field rather than mutating one. What the schema can state about
		// an area is its shape, its two corners and its DM; the corners
		// being low and high, and two areas not overlapping, are ERRATA
		// E012's and starmap.Decode's.
		{"a ruleset that is not the held pages", func(record, _, _ map[string]any) {
			record["ruleset"] = "mongoose-2022"
		}},
		{"an unrecognised erratum identifier", func(record, _, _ map[string]any) {
			record["errata"] = []any{"E2"}
		}},
		{"a missing required field", func(record, _, _ map[string]any) {
			delete(record, "seed")
		}},
		{"an unknown field on a route", func(_, _, route map[string]any) {
			route["surprise"] = 1
		}},
		{"a route reaching a hex the grid numbering does not print", func(_, _, route map[string]any) {
			route["from"] = offTheNumbering
		}},
		// The two dimensions are enumerated separately, so without the
		// oneOf beside them the schema accepts a pair nothing prints --
		// and starmap.Decode refuses it, which is the schema and the
		// reader disagreeing about the same record.
		{"a grid pair nothing prints", func(record, _, _ map[string]any) {
			record["grid"] = map[string]any{"columns": 8, "rows": 40}
		}},
		{"the other grid pair nothing prints", func(record, _, _ map[string]any) {
			record["grid"] = map[string]any{"columns": 32, "rows": 10}
		}},
		// int, not the Parsecs the table constant carries: the validator
		// reads a Go value the way it reads a decoded document, and a named
		// integer type is not a number to it -- it would reject a distance
		// the schema allows, and this case would pass without the maximum
		// it means to test.
		{"a route longer than the jump routes table states", func(_, _, route map[string]any) {
			route["distance"] = int(starmap.MaxJump) + 1
		}},
	}
}

// badAreaRecords is the same table for the broad areas of p. 1 (ERRATA
// E012). They sit apart because the complete example carries no areas, so
// every one of them adds the field rather than mutating one.
//
// What this schema can state about an area is its shape, its two corners
// and its DM. The corners being the low and the high hex, and no two areas
// overlapping, are things JSON Schema cannot say at all; those are held by
// starmap.Decode, which is the other half of the two obligations.
func badAreaRecords() []badRecord {
	const (
		lowCorner  = "from"
		highCorner = "to"
		modifier   = "dm"
	)

	area := func(fields map[string]any) func(record, _, _ map[string]any) {
		return func(record, _, _ map[string]any) {
			record["occurrence_areas"] = []any{fields}
		}
	}

	return []badRecord{
		{
			"a broad area with a DM the book does not offer",
			area(map[string]any{lowCorner: "0101", highCorner: "0410", modifier: 2}),
		},
		{
			"a broad area reaching a hex the grid numbering does not print",
			area(map[string]any{lowCorner: offTheNumbering, highCorner: "0410", modifier: -1}),
		},
		{
			"a broad area with an unknown field",
			area(map[string]any{lowCorner: "0101", highCorner: "0410", modifier: -1, "surprise": 1}),
		},
		{
			"a broad area missing its DM",
			area(map[string]any{lowCorner: "0101", highCorner: "0410"}),
		},
		// Absent is how a record with no areas is written, and an empty
		// list would be a second way to say it -- which would let a record
		// write the key where every golden does not.
		{"an empty list of broad areas", func(record, _, _ map[string]any) {
			record["occurrence_areas"] = []any{}
		}},
	}
}

// TestTheUnmutatedExampleStillValidates guards the mutations above: if the
// example's shape drifts so that a mutation no longer applies, the checks
// would pass by accident.
func TestTheUnmutatedExampleStillValidates(t *testing.T) {
	t.Parallel()

	r := root(t)

	record := loadRecord(t, filepath.Join(r, "docs", "examples", "complete.json"))

	err := schema(t, r).Validate(any(record))
	if err != nil {
		t.Fatalf("the complete example does not validate before mutation: %v", err)
	}

	for _, required := range []string{"seed", "ruleset", "occurrence_dm", "errata", "worlds", "routes"} {
		if _, ok := record[required]; !ok {
			t.Errorf("the complete example has no %q field for the mutations to target", required)
		}
	}
}

// firstObject returns the first entry of one of the record's arrays, which
// is the object a mutation targets. A mutation aimed at an array that had
// emptied would test nothing, so an empty one fails here.
func firstObject(t *testing.T, record map[string]any, field string) map[string]any {
	t.Helper()

	entries, isSlice := record[field].([]any)
	if !isSlice || len(entries) == 0 {
		t.Fatalf("the complete example has no %s for a mutation to target", field)
	}

	object, isObject := entries[0].(map[string]any)
	if !isObject {
		t.Fatalf("the complete example's first %s entry is not an object", field)
	}

	return object
}

func loadRecord(t *testing.T, path string) map[string]any {
	t.Helper()

	b, err := os.ReadFile(path) //nolint:gosec // a fixed path inside the repository
	if err != nil {
		t.Fatal(err)
	}

	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}

	record, ok := inst.(map[string]any)
	if !ok {
		t.Fatalf("%s is not a JSON object", path)
	}

	return record
}

// TestASectorRecordValidates: the schema was widened for sector records
// -- a 32x40 grid, and hexes running to 3240 -- and no golden is one, so
// nothing else ever holds the widened schema against a record the engine
// actually writes. A pattern or an enum that had been widened wrongly
// would sit in the document unexercised.
func TestASectorRecordValidates(t *testing.T) {
	t.Parallel()

	engine, err := gen.New()
	if err != nil {
		t.Fatal(err)
	}

	golden := fixture.SectorGolden()

	record, err := engine.Sector(gen.Inputs{
		Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM, OccurrenceAreas: nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	if record.Grid != starmap.SectorGrid() {
		t.Fatalf("the record under test is on a %dx%d grid, so it is not the shape this test means to hold",
			record.Grid.Columns, record.Grid.Rows)
	}

	encoded, err := starmap.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}

	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}

	err = schema(t, root(t)).Validate(inst)
	if err != nil {
		t.Errorf("a sector record does not validate:\n%v", err)
	}
}

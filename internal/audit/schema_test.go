package audit_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/subsector"
	"github.com/philoserf/ctworldgen/tables"
)

func schema(t *testing.T, root string) *jsonschema.Schema {
	t.Helper()

	path := filepath.Join(root, "docs", "subsector.schema.json")

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

	addErr := compiler.AddResource("subsector.schema.json", doc)
	if addErr != nil {
		t.Fatal(addErr)
	}

	compiled, compileErr := compiler.Compile("subsector.schema.json")
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

	in := gen.Inputs{Seed: example.Seed, Name: example.Name, OccurrenceDM: example.OccurrenceDM}

	generated, err := engine.Generate(in)
	if err != nil {
		t.Fatal(err)
	}

	got, err := subsector.Marshal(generated)
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

	for _, bad := range []struct {
		name   string
		mutate func(record, world, route map[string]any)
	}{
		{"an unknown field at the top level", func(record, _, _ map[string]any) {
			record["surprise"] = 1
		}},
		{"an unknown field on a world", func(_, world, _ map[string]any) {
			world["surprise"] = 1
		}},
		{"a hex off the p. 3 grid", func(_, world, _ map[string]any) {
			world["hex"] = "0900"
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
		{"a lane reaching a hex off the p. 3 grid", func(_, _, route map[string]any) {
			route["from"] = "0900"
		}},
		// int, not the Parsecs the table constant carries: the validator
		// reads a Go value the way it reads a decoded document, and a named
		// integer type is not a number to it -- it would reject a distance
		// the schema allows, and this case would pass without the maximum
		// it means to test.
		{"a lane longer than the jump routes table states", func(_, _, route map[string]any) {
			route["distance"] = int(tables.MaxJump) + 1
		}},
	} {
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

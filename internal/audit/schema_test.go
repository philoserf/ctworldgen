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
		mutate func(record map[string]any, world map[string]any)
	}{
		{"an unknown field at the top level", func(record, _ map[string]any) {
			record["surprise"] = 1
		}},
		{"an unknown field on a world", func(_, world map[string]any) {
			world["surprise"] = 1
		}},
		{"a hex off the p. 3 grid", func(_, world map[string]any) {
			world["hex"] = "0900"
		}},
		{"a hex that is not four digits", func(_, world map[string]any) {
			world["hex"] = "1-5"
		}},
		{"a starport the book does not print", func(_, world map[string]any) {
			world["starport"] = "Q"
		}},
		{"an occurrence DM the book does not offer", func(record, _ map[string]any) {
			record["occurrence_dm"] = 2
		}},
		{"a ruleset that is not the held pages", func(record, _ map[string]any) {
			record["ruleset"] = "mongoose-2022"
		}},
		{"an unrecognised erratum identifier", func(record, _ map[string]any) {
			record["errata"] = []any{"E2"}
		}},
		{"a missing required field", func(record, _ map[string]any) {
			delete(record, "seed")
		}},
	} {
		t.Run(bad.name, func(t *testing.T) {
			t.Parallel()

			record := loadRecord(t, filepath.Join(repoRoot, "docs", "examples", "complete.json"))

			worlds, isSlice := record["worlds"].([]any)
			if !isSlice || len(worlds) == 0 {
				t.Fatal("the complete example has no worlds to mutate")
			}

			world, isSlice := worlds[0].(map[string]any)
			if !isSlice {
				t.Fatal("the complete example's first world is not an object")
			}

			bad.mutate(record, world)

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

	for _, required := range []string{"seed", "ruleset", "occurrence_dm", "errata", "worlds"} {
		if _, ok := record[required]; !ok {
			t.Errorf("the complete example has no %q field for the mutations to target", required)
		}
	}
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

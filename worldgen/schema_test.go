package worldgen_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/worldgen"
)

// The schema is documentation of what the engine writes; this pins it to
// the structs so neither drifts. A field added to a struct without a
// schema property (or vice versa) fails here.
func TestSchemaMatchesStructs(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "docs", "subsector.schema.json"))
	if err != nil {
		t.Fatal(err)
	}

	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}

	root := properties(t, schema)
	defs := node(t, schema, "$defs")

	checkKeys(t, "root", root, reflect.TypeFor[worldgen.Subsector]())
	checkKeys(t, "rng", properties(t, child(t, root, "rng")), reflect.TypeFor[worldgen.RNG]())
	checkKeys(t, "inputs", properties(t, child(t, root, "inputs")), reflect.TypeFor[worldgen.Inputs]())
	checkKeys(t, "world", properties(t, child(t, defs, "world")), reflect.TypeFor[worldgen.World]())
	checkKeys(t, "route", properties(t, child(t, defs, "route")), reflect.TypeFor[worldgen.Route]())
	checkKeys(t, "event", properties(t, child(t, defs, "event")), reflect.TypeFor[worldgen.Event]())
	checkKeys(t, "dm", properties(t, child(t, defs, "dm")), reflect.TypeFor[worldgen.EventDM]())
}

// TestSchemaExamplesParse: the examples beside the schema are
// documentation of the record's shape, so they must at least be records
// this engine's own strict decoder accepts — which rejects unknown fields,
// and so catches a property renamed in one place and not the other.
//
// Unlike the sibling's, these examples are hand-written rather than engine
// output, and so are not replay-verified: a real record of this engine
// runs past a third of a megabyte, which is not a document anyone reads.
// The golden fixtures under worldgen/testdata are the engine output that
// carries that guarantee, and TestGoldenRecordsReplay is where it is
// checked.
func TestSchemaExamplesParse(t *testing.T) {
	for _, name := range []string{"subsector.minimal.json", "subsector.complete.json"} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "docs", name)) //nolint:gosec // a docs/ path this test names
			if err != nil {
				t.Fatal(err)
			}

			sub, err := worldgen.UnmarshalRecord(raw)
			if err != nil {
				t.Fatalf("UnmarshalRecord: %v", err)
			}

			if sub.SchemaVersion != worldgen.SchemaVersion {
				t.Errorf("example schema_version is %q, this build writes %q", sub.SchemaVersion, worldgen.SchemaVersion)
			}

			if sub.EngineVersion != worldgen.EngineVersion {
				t.Errorf("example engine_version is %q, this build writes %q", sub.EngineVersion, worldgen.EngineVersion)
			}

			if sub.Ruleset != worldgen.Ruleset {
				t.Errorf("example ruleset is %q, this build writes %q", sub.Ruleset, worldgen.Ruleset)
			}

			for _, id := range sub.Errata {
				if !slices.Contains(worldgen.ErrataIDs(), id) {
					t.Errorf("example stamps %q, which is not a reading this engine applies", id)
				}
			}
		})
	}
}

func node(t *testing.T, parent map[string]any, name string) map[string]any {
	t.Helper()

	child, ok := parent[name].(map[string]any)
	if !ok {
		t.Fatalf("schema node %q is missing or not an object", name)
	}

	return child
}

func properties(t *testing.T, n map[string]any) map[string]any {
	t.Helper()

	return node(t, n, "properties")
}

func child(t *testing.T, props map[string]any, name string) map[string]any {
	t.Helper()

	return node(t, props, name)
}

// checkKeys asserts the schema's property names and the struct's JSON tags
// are the same set.
func checkKeys(t *testing.T, label string, props map[string]any, typ reflect.Type) {
	t.Helper()

	schemaKeys := make([]string, 0, len(props))
	for k := range props {
		schemaKeys = append(schemaKeys, k)
	}

	slices.Sort(schemaKeys)

	structKeys := jsonTags(typ)
	slices.Sort(structKeys)

	if !slices.Equal(schemaKeys, structKeys) {
		t.Errorf("%s: schema has %v, %s has %v", label, schemaKeys, typ.Name(), structKeys)
	}
}

func jsonTags(typ reflect.Type) []string {
	out := []string{}

	for field := range typ.Fields() {
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// Cut, not SplitN indexed at 0: it returns a plain string, so the
		// name cannot come back through a slice NilAway has to reason about.
		name, _, _ := strings.Cut(tag, ",")
		out = append(out, name)
	}

	return out
}

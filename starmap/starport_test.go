package starmap_test

import (
	"encoding/json"
	"testing"

	"github.com/philoserf/ctworldgen/starmap"
)

func TestStarportParsesAndMarshals(t *testing.T) {
	t.Parallel()

	for _, printed := range []string{"A", "B", "C", "D", "E", "X"} {
		port, err := starmap.ParseStarport(printed)
		if err != nil {
			t.Fatal(err)
		}

		if port.String() != printed {
			t.Errorf("%q parsed to %q", printed, port)
		}

		encoded, err := json.Marshal(port)
		if err != nil {
			t.Fatal(err)
		}

		if string(encoded) != `"`+printed+`"` {
			t.Errorf("%q marshaled to %s", printed, encoded)
		}

		var back starmap.Starport

		err = json.Unmarshal(encoded, &back)
		if err != nil {
			t.Fatal(err)
		}

		if back != port {
			t.Errorf("%q round-tripped to %q", printed, back)
		}
	}
}

func TestStarportRejectsWhatTheBookDoesNotPrint(t *testing.T) {
	t.Parallel()

	for _, bad := range []string{"", "a", "F", "AB", "0", " "} {
		_, err := starmap.ParseStarport(bad)
		if err == nil {
			t.Errorf("ParseStarport(%q) succeeded", bad)
		}
	}

	_, err := json.Marshal(starmap.Starport(0))
	if err == nil {
		t.Error("marshaling the zero Starport succeeded")
	}

	var port starmap.Starport

	err = json.Unmarshal([]byte(`"a"`), &port)
	if err == nil {
		t.Error(`unmarshaling "a" as a starport succeeded`)
	}

	err = json.Unmarshal([]byte(`3`), &port)
	if err == nil {
		t.Error("unmarshaling a number as a starport succeeded")
	}

	if got := starmap.Starport('Q').String(); got != "Starport(81)" {
		t.Errorf("Starport('Q').String() = %q", got)
	}
}

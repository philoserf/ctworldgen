package tables

import (
	"errors"
	"strings"
	"testing"

	"github.com/philoserf/ctworldgen/dice"
)

// The validators run at load against data this repository writes, so a
// failure is a build defect rather than a runtime condition. These tests
// feed them the defects on purpose, because a validator that never
// rejected anything would be indistinguishable from no validator.

func TestCheckLabelRowsRejectsBrokenTables(t *testing.T) {
	for name, rows := range map[string][]labelRow{
		"empty":          {},
		"not from zero":  {{Value: "1", Label: "one"}},
		"a gap":          {{Value: "0", Label: "zero"}, {Value: "2", Label: "two"}},
		"a blank label":  {{Value: "0", Label: "zero"}, {Value: "1", Label: "   "}},
		"a wrong digit":  {{Value: "0", Label: "zero"}, {Value: "b", Label: "lowercase"}},
		"a missing zero": {{Value: "0", Label: "zero"}, {Value: "1", Label: "one"}, {Value: "3", Label: "three"}},
	} {
		if err := checkLabelRows("size", rows); !errors.Is(err, ErrInvalidData) {
			t.Errorf("%s: err = %v, want ErrInvalidData", name, err)
		}
	}

	good := []labelRow{{Value: "0", Label: "zero"}, {Value: "1", Label: "one"}}
	if err := checkLabelRows("size", good); err != nil {
		t.Errorf("a well-formed table was rejected: %v", err)
	}
}

func TestAddRoutePairRejectsBrokenRows(t *testing.T) {
	for name, row := range map[string]routePair{
		"no hyphen":      {Pair: "AA", Targets: []int{1, 2, 3, 4}},
		"an X":           {Pair: "A-X", Targets: []int{1, 2, 3, 4}},
		"unknown type":   {Pair: "A-Q", Targets: []int{1, 2, 3, 4}},
		"too few cells":  {Pair: "A-A", Targets: []int{1, 2, 3}},
		"too many cells": {Pair: "A-A", Targets: []int{1, 2, 3, 4, 5}},
		"a seven":        {Pair: "A-A", Targets: []int{7, 2, 3, 4}},
		"a negative":     {Pair: "A-A", Targets: []int{-1, 2, 3, 4}},
	} {
		c := &Charts{routeTargets: map[string][]int{}}
		if err := c.addRoutePair(row); !errors.Is(err, ErrInvalidData) {
			t.Errorf("%s: err = %v, want ErrInvalidData", name, err)
		}
	}

	c := &Charts{routeTargets: map[string][]int{}}

	row := routePair{Pair: "A-C", Targets: []int{1, 4, 6, 0}}
	if err := c.addRoutePair(row); err != nil {
		t.Fatalf("a well-formed row was rejected: %v", err)
	}

	// The table prints each unordered pair once, so C-A is the same row.
	if err := c.addRoutePair(routePair{Pair: "C-A", Targets: []int{1, 4, 6, 0}}); !errors.Is(err, ErrInvalidData) {
		t.Errorf("the reversed duplicate of a pair was accepted: %v", err)
	}
}

func TestCheckRoutesCompleteRejectsAPartialTable(t *testing.T) {
	c := &Charts{routeTargets: map[string][]int{"A-A": {1, 2, 4, 5}}}
	if err := c.checkRoutesComplete(); !errors.Is(err, ErrInvalidData) {
		t.Errorf("a table missing fourteen rows was accepted: %v", err)
	}
}

func TestCheckTechCompleteRejectsAPartialMatrix(t *testing.T) {
	c := &Charts{tech: map[string]techRow{"0": {Value: "0"}}}

	err := c.checkTechComplete()
	if err == nil || !errors.Is(err, ErrInvalidData) {
		t.Fatalf("a matrix missing fifteen rows was accepted: %v", err)
	}

	if !strings.Contains(err.Error(), "no row for value") {
		t.Errorf("the message does not name the missing row: %v", err)
	}
}

func TestLoadStarportThrowsRejectsBrokenTables(t *testing.T) {
	for name, throws := range map[string][]starportThrow{
		"an unknown type":   {{Die: 2, Type: "Q"}},
		"a one-die total":   {{Die: 1, Type: "A"}},
		"a thirteen":        {{Die: 13, Type: "A"}},
		"a duplicate throw": {{Die: 2, Type: "A"}, {Die: 2, Type: "B"}},
		"an incomplete run": {{Die: 2, Type: "A"}},
	} {
		c := &Charts{
			starportByThrow: map[int]string{},
			starports:       map[string]*Starport{"A": {Type: "A"}, "B": {Type: "B"}},
		}

		if err := c.loadStarportThrows(throws); !errors.Is(err, ErrInvalidData) {
			t.Errorf("%s: err = %v, want ErrInvalidData", name, err)
		}
	}
}

// TestLoadBaseTargetsRejectsUnparseableNotation: the p. 5 chart's throws
// are held in the book's own notation and parsed at load, so a typo in the
// data file is a build failure rather than a base that silently never
// appears.
func TestLoadBaseTargetsRejectsUnparseableNotation(t *testing.T) {
	for name, bad := range map[string]Starport{
		"naval": {Type: StarportA, NavalBase: "eight or better", ScoutBase: "10+"},
		"scout": {Type: StarportA, NavalBase: "8+", ScoutBase: "0+"},
	} {
		c := &Charts{
			starports:   map[string]*Starport{},
			navalTarget: map[string]dice.Target{},
			scoutTarget: map[string]dice.Target{},
		}

		for _, tp := range starportTypes {
			c.starports[tp] = &Starport{Type: tp}
		}

		row := bad
		c.starports[StarportA] = &row

		if err := c.loadBaseTargets(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("%s: err = %v, want ErrInvalidData", name, err)
		}
	}
}

// TestCheckStarportFacilitiesRejectsABrokenChart: render decides between
// "no fuel" and "<grade> fuel" by comparing the fuel column to "none"
// exactly, so the value has to come from a closed set — a chart that said
// "None" would otherwise load and then print "None fuel".
func TestCheckStarportFacilitiesRejectsABrokenChart(t *testing.T) {
	for name, bad := range map[string]Starport{
		"unknown fuel grade": {Fuel: "None", Quality: "q", Shipyard: "none"},
		"empty fuel grade":   {Fuel: "", Quality: "q", Shipyard: "none"},
		"no quality":         {Fuel: "none", Quality: "  ", Shipyard: "none"},
		"no shipyard":        {Fuel: "none", Quality: "q", Shipyard: ""},
	} {
		c := &Charts{starports: map[string]*Starport{}}

		for _, tp := range starportTypes {
			c.starports[tp] = &Starport{Type: tp, Fuel: "none", Quality: "q", Shipyard: "none"}
		}

		row := bad
		row.Type = StarportA
		c.starports[StarportA] = &row

		if err := c.checkStarportFacilities(); !errors.Is(err, ErrInvalidData) {
			t.Errorf("%s: err = %v, want ErrInvalidData", name, err)
		}
	}

	// The guard for a type the presence check should already have caught.
	empty := &Charts{starports: map[string]*Starport{}}
	if err := empty.checkStarportFacilities(); !errors.Is(err, ErrInvalidData) {
		t.Errorf("a chart with no rows: err = %v, want ErrInvalidData", err)
	}
}

// TestTheEmbeddedChartsLoad is the assertion the rest of the program
// assumes: this repository's own data files are valid.
func TestTheEmbeddedChartsLoad(t *testing.T) {
	if _, err := Load(); err != nil {
		t.Fatalf("the embedded Book 3 charts do not load: %v", err)
	}
}

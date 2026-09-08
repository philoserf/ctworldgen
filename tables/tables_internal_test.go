package tables

import (
	"strconv"
	"strings"
	"testing"
)

// The second transcription of the two technological levels tables, and
// the reason it is an internal test rather than sitting beside the other
// second transcriptions in tables_test.go.
//
// What has to be checked here is what the pages *print*: which level each
// column's entry appears at, and no other. The exported reading applies
// ERRATA E009 and returns the last entry at or below a level, so a test
// written against it would compare a derived value and could not tell a
// mis-keyed cell from a mis-read rule. The ladders are the transcription,
// and they are unexported, so the check has to be in here.
//
// Each column is retyped in a function of its own, so that the unit of
// this file is one column of one page -- which is also the unit a reader
// checking it against the book works in.

// T5's band names, as p. 232 brackets them.
const (
	vlow  = "Vlow"
	low   = "Low"
	mid   = "Mid"
	high  = "High"
	vhigh = "Vhigh"
	xhigh = "Xhigh"
)

// TestTheTechnologicalLevelsTableIsThePrintedPages is Book 3 pp. 10-11,
// retyped: every cell the pages print, at the level they print it, and no
// cell they do not.
func TestTheTechnologicalLevelsTableIsThePrintedPages(t *testing.T) {
	t.Parallel()

	levels := load(t).TechLevels

	// p. 10, Weapons.
	assertLadder(t, "personal", levels.personal, printedPersonal())
	assertLadder(t, "armor", levels.armor, printedArmor())
	assertLadder(t, "special", levels.special, printedSpecial())
	assertLadder(t, "computers", levels.computers, printedComputers())
	assertLadder(t, "communication", levels.communication, printedCommunication())

	// p. 11, Transportation.
	assertLadder(t, "water", levels.water, printedWater())
	assertLadder(t, "land", levels.land, printedLand())
	assertLadder(t, "air", levels.air, printedAir())
	assertLadder(t, "space", levels.space, printedSpace())
	assertLadder(t, "fuels", levels.fuels, printedFuels())

	// Row 16, printed across the water, land and air columns rather than
	// in one of them (E010 part 2).
	const matterTransport = 16

	if levels.spanLevel != matterTransport {
		t.Errorf("matter transport is carried at level %v; p. 11 prints it at %d",
			levels.spanLevel, matterTransport)
	}
}

// TestTheBorrowedTechnologyChartIsThePrintedPages is T5 Core Book 2
// pp. 230-232, retyped. It is here for the same reason and read the same
// way; what it is doing in this repository at all is ERRATA E011.
//
// The fractional levels are the point of half of it: 1.6 and 3.6 are what
// give a world at index 2 its cities and one at index 4 its railroads.
func TestTheBorrowedTechnologyChartIsThePrintedPages(t *testing.T) {
	t.Parallel()

	levels := load(t).TechLevels

	assertLadder(t, "era", levels.era, printedEra())
	assertLadder(t, "energy", levels.energy, printedEnergy())
	assertLadder(t, "society", levels.society, printedSociety())
	assertLadder(t, "environ", levels.environ, printedEnviron())
	assertLadder(t, "transport", levels.transport, printedTransport())
	assertLadder(t, "computing", levels.borrowedComputers, printedComputing())
}

// TestTheBandsAreTheBracketsOnPageTwoThirtyTwo is the last thing T5's
// chart carries and the one an extraction cannot see at all: the bands
// are drawn as brackets down the margin rather than printed in a column.
func TestTheBandsAreTheBracketsOnPageTwoThirtyTwo(t *testing.T) {
	t.Parallel()

	levels := load(t).TechLevels

	for level, want := range map[int]string{
		0: vlow, 1: vlow, 2: vlow, 3: vlow,
		4: low, 5: low, 6: low,
		7: mid, 8: mid, 9: mid,
		10: high, 11: high, 12: high,
		13: vhigh, 14: vhigh, 15: vhigh,
		16: xhigh, 17: xhigh, 18: xhigh,
	} {
		if got := levels.Level(level).Borrowed.Band; got != want {
			t.Errorf("level %d is in band %q; p. 232 brackets it as %q", level, got, want)
		}
	}
}

// TestAMalformedTechnologyTableIsRefusedAtLoad holds the promise Load
// makes for every chart here: a table that is not the page fails when it
// is read, not at the render that needed it. These are the checks a
// mis-edited data file trips, and each one is a way the transcription
// could be wrong without any cell being misspelled.
func TestAMalformedTechnologyTableIsRefusedAtLoad(t *testing.T) {
	t.Parallel()

	for _, refused := range []struct {
		what string
		json string
	}{
		{"not JSON at all", `{`},
		{"too few rows", `{"rows":[{"level":0}],"spanning":{"level":16}}`},
		{"no matter transport row", rowsThroughEighteen() + `,"spanning":{}}`},
		{
			"rows out of order",
			strings.Replace(rowsThroughEighteen(), `{"level":0}`, `{"level":5}`, 1) +
				`,"spanning":{"level":16}}`,
		},
	} {
		t.Run(refused.what, func(t *testing.T) {
			t.Parallel()

			var levels TechLevels

			if levels.loadHeld([]byte(refused.json)) == nil {
				t.Errorf("a table with %s was accepted", refused.what)
			}
		})
	}

	for _, refused := range []struct {
		what string
		json string
	}{
		{"not JSON at all", `{`},
		{"no bands", `{"bands":{},"rows":[]}`},
		{"a band that is not a first and a last", `{"bands":{"Low":[4]},"rows":[]}`},
		{"levels that do not ascend", `{"bands":{"All":[0,18]},"rows":[{"level":2,"band":"All"},{"level":1,"band":"All"}]}`},
		{"a row in a band the chart does not bracket", `{"bands":{"Low":[4,6]},"rows":[{"level":4,"band":"Vlow"}]}`},
		{"a level in no band at all", `{"bands":{"Low":[4,6]},"rows":[{"level":4,"band":"Low"}]}`},
	} {
		t.Run(refused.what, func(t *testing.T) {
			t.Parallel()

			var levels TechLevels

			if levels.loadBorrowed([]byte(refused.json)) == nil {
				t.Errorf("a chart with %s was accepted", refused.what)
			}
		})
	}
}

// rowsThroughEighteen is the right number of rows in the right order and
// nothing else, so that a test of some other rejection is not answered by
// the row count.
func rowsThroughEighteen() string {
	rows := make([]string, 0, maxTechLevel-minTechLevel+1)

	for level := minTechLevel; level <= maxTechLevel; level++ {
		rows = append(rows, `{"level":`+strconv.Itoa(level)+`}`)
	}

	return `{"rows":[` + strings.Join(rows, ",") + `]`
}

// Book 3 p. 10, Weapons.

func printedPersonal() ladder {
	return ladder{
		{0, "Club, cudgel, Spear"},
		{1, "Dagger, pike, Sword"},
		{2, "Halberd, Broadsword"},
		{3, "Foil, cutlass, Blade, bayonet"},
		{4, "Revolver, Shotgun"},
		{5, "Carbine, Rifle, Pistol, SMG"},
		{6, "Auto Rifle"},
		{7, "Body Pistol"},
		{8, "Laser Carbine"},
		{9, "Laser Rifle"},
	}
}

func printedArmor() ladder {
	return ladder{
		{1, "Jack"},
		{4, "Cloth"},
		{7, "Mesh"},
		{9, "Ablat"},
		{10, "Reflec"},
		{13, "Battle Dress"},
	}
}

func printedSpecial() ladder {
	return ladder{
		{1, "Catapult"},
		{2, "Cannon"},
		{4, "Artillery"},
		{5, "Sandcasters, Mortars"},
		{6, "Missiles, Rocket Launchers"},
		{7, "Pulse Laser"},
		{8, "Auto-Cannon"},
		{9, "Beam Laser"},
	}
}

func printedComputers() ladder {
	return ladder{
		{1, "Abacus"},
		{4, "Adding Machine"},
		{5, "Model/1"},
		{6, "Model/1 bis"},
		{7, "Model/2"},
		{8, "Model/2 bis"},
		{9, "Model/3"},
		{10, "Model/4"},
		{11, "Model/5"},
		{12, "Model/6"},
		{13, "Model/7"},
		{17, "Artificial Intelligence"},
	}
}

func printedCommunication() ladder {
	return ladder{
		{0, "Runners"},
		{1, "Heliograph"},
		{4, "Telephones"},
		{5, "Radio"},
		{6, "Television"},
	}
}

// Book 3 p. 11, Transportation.

func printedWater() ladder {
	return ladder{
		{0, "Canoes"},
		{1, "Galley"},
		{3, "Sailing Ships"},
		{4, "Steamships"},
		{6, "Submersibles"},
		{7, "Hovercraft"},
	}
}

func printedLand() ladder {
	return ladder{
		{0, "Carts"},
		{1, "Wagons"},
		{4, "Trains"},
		{5, "Ground cars"},
		{6, "ATV, AFV"},
		{7, "Hovercraft"},
	}
}

func printedAir() ladder {
	return ladder{
		{3, "Hot air balloon"},
		{4, "Dirigibles"},
		{5, "Fixed wing aircraft"},
		{6, "Rotary wing aircraft"},
		{8, "Air/Raft"},
		{12, "Grav belts"},
	}
}

func printedSpace() ladder {
	return ladder{
		{7, "Non-starships"},
		{9, "Starships"},
		{10, "Drives H or less"},
		{11, "Drives K or less"},
		{12, "Drives N or less"},
		{13, "Drives O or less"},
		{14, "Drives U or less"},
		{15, "All drives"},
	}
}

func printedFuels() ladder {
	return ladder{
		{0, "Muscle"},
		{2, "Wind"},
		{3, "Water wheel"},
		{4, "Coal"},
		{5, "Oil"},
		{6, "Fission"},
		{7, "Solar"},
		{8, "Fusion"},
		{17, "Anti-Matter"},
	}
}

// T5 Core Book 2 p. 230, Tech Level Chart 1.

func printedEra() ladder {
	return ladder{
		{0, "Primitive Stone Age"},
		{1, "Bronze Age 3500 BC"},
		{1.3, "Iron Age 1300 BC"},
		{1.6, "Middle Ages 600 AD"},
		{2, "Age Of Sail 1500 AD"},
		{3, "Industrial Revolution 1700 AD"},
		{3.3, "1800 AD"},
		{3.6, "1850 AD"},
		{4, "Mechanization 1900 AD"},
		{5, "1930 AD"},
		{6, "Nuclear Age 1950 AD"},
		{7, "1975 AD"},
		{8, "2000 AD"},
		{9, "2050 AD"},
		{10, "2100 AD"},
		{11, "Imperial Average Circa Year Zero"},
		{13, "Imperial Maximum Circa 550"},
		{15, "Imperial Maximum Circa 1107"},
		{16, "Darrian Maximum"},
	}
}

func printedEnergy() ladder {
	return ladder{
		{0, "Personal Effort. Fire"},
		{1, "Water Power"},
		{2, "Wind. Sail."},
		{3, "Coal. Steam."},
		{4, "Electricity."},
		{5, "Oil. Petrochemicals."},
		{6, "Nuclear Fission."},
		{7, "Geothermal. Solar."},
		{8, "Renewables."},
		{9, "Early Fusion."},
		{10, "Practical Fusion."},
		{11, "[FusionPlus]."},
		{14, "Exotics. Collectors."},
		{16, "Experimental AM."},
		{18, "Exotics. Collectors."},
	}
}

func printedSociety() ladder {
	return ladder{
		{0, "Tribe. Clan."},
		{1, "Ethnic Groups."},
		{1.6, "Kingdoms."},
		{2, "Nations."},
		{3, "Democracies."},
		{5, "Dictators."},
		{6, "Superpowers."},
		{10, "Non-Geographic Communities."},
		{13, "Robots."},
		{14, "Temporary Personality Transfer"},
		{15, "Mindwipe."},
		{16, "Artificial Persons. The Under Society."},
		{17, "Permanent Personality Transfer"},
	}
}

func printedEnviron() ladder {
	return ladder{
		{0, "Natural. Crude Shelters."},
		{1, "Settlement. Villages."},
		{1.3, "Towns. Roads. Canals."},
		{1.6, "Cities."},
		{4, "Skyscrapers"},
		{6, "Suburbs."},
		{9, "Arcologies"},
	}
}

// T5 Core Book 2 p. 231, Tech Level Chart 2.

func printedTransport() ladder {
	return ladder{
		{0, "Walking."},
		{1, "Beasts of Burden."},
		{1.3, "Wheel."},
		{1.6, "Galleys."},
		{2, "Sailing Ships."},
		{3.3, "Steamships."},
		{3.6, "Railroads."},
		{5, "Groundcars."},
		{7, "Rockets to Orbit."},
		{9, "NAFAL."},
		{10, "Gravity Manipulation Lifters to Orbit."},
	}
}

func printedComputing() ladder {
	return ladder{
		{0, "Counting."},
		{1, "Abacus Quipu."},
		{2, "Algebra."},
		{3, "Calculus."},
		{4, "Analog Computers."},
		{5, "Electric Calculators."},
		{6, "Model /1."},
		{7, "Model /2."},
		{9, "Model /3."},
		{10, "Model /4."},
		{11, "Semi-Organic Brain. Model /5."},
		{12, "Positronic Brain. Model /6."},
		{13, "Wafer Technology. Model /7."},
		{14, "Self Aware Model /8."},
		{15, "Model /9."},
		{16, "True AI Artificial Intelligence."},
	}
}

func assertLadder(t *testing.T, column string, got, want ladder) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("the %s column carries %d entries and the page prints %d:\ngot  %v\nwant %v",
			column, len(got), len(want), got, want)

		return
	}

	for index, printed := range want {
		if got[index] != printed {
			t.Errorf("the %s column carries %v where the page prints %v",
				column, got[index], printed)
		}
	}
}

func load(t *testing.T) *Tables {
	t.Helper()

	loaded, err := Load()
	if err != nil {
		t.Fatalf("loading the tables: %v", err)
	}

	return loaded
}

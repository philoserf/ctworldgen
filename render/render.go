// Package render writes a record's listing: the Markdown a referee can run
// from, for a sub-sector or for a sector alike.
//
// It stands in for what p. 4 asks him to keep -- "at least one (and
// preferably several) pages in a central notebook maintained by the
// referee" -- so the listing carries the roster, the routes, and a
// page of detail per world with the labels its Book 3 tables give.
//
// The JSON record is the source of truth; this is a render of it. Nothing
// here throws a die or decides a rule.
package render

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/philoserf/ctworldgen/starmap"
	"github.com/philoserf/ctworldgen/tables"
)

// Renderer holds the descriptive tables of pp. 5-7, loaded once.
type Renderer struct {
	charts *tables.Tables
}

// New loads and validates the charts the listing reads its labels from.
func New() (*Renderer, error) {
	charts, err := tables.Load()
	if err != nil {
		return nil, fmt.Errorf("loading the Book 3 charts: %w", err)
	}

	return &Renderer{charts: charts}, nil
}

// Listing writes the Markdown listing.
func (r *Renderer) Listing(out io.Writer, record *starmap.Record) error {
	var built strings.Builder

	names := namesByHex(record)

	r.heading(&built, record)
	r.grid(&built, record)
	r.roster(&built, record)
	r.routes(&built, names, record)
	r.details(&built, names, record)

	_, err := io.WriteString(out, built.String())
	if err != nil {
		return fmt.Errorf("writing the listing: %w", err)
	}

	return nil
}

func (r *Renderer) heading(built *strings.Builder, record *starmap.Record) {
	name := record.Name
	if name == "" {
		name = untitled(record)
	}

	fmt.Fprintf(built, "# %s\n\n", name)
	fmt.Fprintf(built, "%d worlds, %d routes. Generated from seed %d at occurrence DM %s.\n\n",
		len(record.Worlds), len(record.Routes), record.Seed, occurrenceDM(record.OccurrenceDM))
}

// untitled is the heading a record the referee has not named gets. It is
// what the record is, and a sector is not a subsector: those two grids
// are the only shapes a record takes (ERRATA E006), and heading a 32x40
// one "Subsector" is the listing misreporting its own subject.
func untitled(record *starmap.Record) string {
	if record.Grid == starmap.SectorGrid() {
		return "Sector"
	}

	return "Subsector"
}

// The map is drawn as p. 3 prints the grid: a hex is one slot wide, the
// odd-numbered columns sit on their row's first line, and the
// even-numbered ones are pushed half a slot right and half a row down.
// Getting that parity backwards draws a map that disagrees with
// Hex.Distance for half the grid and looks entirely reasonable.
const (
	mapSlot     = 10 // characters one hex occupies on a map line
	mapHalfSlot = mapSlot / 2
)

// mapNote says what the map shows and, as much to the point, what it does
// not: p. 2 asks for "a line connecting the two worlds on the map" and
// this one draws none.
//
// The first sentence names the grid actually drawn. The rest is true of
// either, because the parity it describes is the p. 3 parity in both
// cases -- which is exactly what makes the sector translation safe
// (ERRATA E006 part 2).
func mapNote(record *starmap.Record) string {
	if record.Grid == starmap.SectorGrid() {
		return sectorGridNote + mapNoteTail
	}

	return pageThreeGridNote + mapNoteTail
}

const (
	pageThreeGridNote = "The p. 3 sub-sector hex grid."
	sectorGridNote    = "Sixteen p. 3 sub-sector hex grids on one sheet, 0101 through 3240 (ERRATA E006)."
)

const mapNoteTail = " The odd-numbered columns sit high and the " +
	"even-numbered ones half a hex below them, which is how the page prints it. " +
	"A world carries the letter of its starport -- p. 1 marks the hex with the " +
	"letter the starports table gives -- and a hex with no world is left blank, " +
	"which is what p. 1 says to leave it. P. 2 also asks for a line drawn between the " +
	"worlds a route joins; a monospace grid has nowhere to put one, so this map " +
	"draws none and the route table below carries them instead. " +
	"`render --format pdf` draws them.\n\n"

// grid draws the subsector map. It marks what p. 1 says to mark and
// nothing else, so it is a render of the record like every other section.
func (r *Renderer) grid(built *strings.Builder, record *starmap.Record) {
	built.WriteString("## The map\n\n")
	built.WriteString(mapNote(record))
	built.WriteString("```text\n")

	marked := make(map[starmap.Hex]starmap.Starport, len(record.Worlds))
	for _, world := range record.Worlds {
		marked[world.Hex] = world.Starport
	}

	for row := 1; row <= record.Grid.Rows; row++ {
		// Odd columns first, then the even ones half a slot right: one
		// printed row of the grid is two lines of the map.
		built.WriteString(gridLine(marked, record.Grid.Columns, row, 1))
		built.WriteString(gridLine(marked, record.Grid.Columns, row, 0))
	}

	built.WriteString("```\n\n")
}

// gridLine draws the columns of one parity across a single row of the
// record's grid. parity is the remainder that selects them: 1 for the
// high columns, 0 for the low.
func gridLine(marked map[starmap.Hex]starmap.Starport, columns, row, parity int) string {
	var line strings.Builder

	if parity == 0 {
		line.WriteString(strings.Repeat(" ", mapHalfSlot))
	}

	for col := 1; col <= columns; col++ {
		if col%2 != parity {
			continue
		}

		hex := starmap.Hex{Col: col, Row: row}

		cell := hex.String() + " "
		if starport, ok := marked[hex]; ok {
			cell += starport.String()
		}

		// max: a Starport the schema would have rejected prints as
		// "Starport(0)", which is wider than a slot. The rest of the
		// listing degrades on such a record rather than stopping, so the
		// map does too -- a negative count here would panic.
		line.WriteString(cell + strings.Repeat(" ", max(mapSlot-len(cell), 0)))
	}

	return strings.TrimRight(line.String(), " ") + "\n"
}

// roster is the world roster: hexes, names, and strings of digits.
func (r *Renderer) roster(built *strings.Builder, record *starmap.Record) {
	built.WriteString("## Worlds\n\n")

	if len(record.Worlds) == 0 {
		built.WriteString("No world was placed. An empty subsector is a result.\n\n")

		return
	}

	built.WriteString("| Hex | Name | Digits | Bases |\n| --- | --- | --- | --- |\n")

	for _, world := range record.Worlds {
		fmt.Fprintf(built, "| %s | %s | %s | %s |\n", world.Hex, cell(world.Name), world.Digits, bases(world))
	}

	built.WriteString("\n")
}

// occurrenceDM writes the referee's DM the way p. 1 offers it: +1, 0 or
// -1, rather than the "+0" a signed format would give.
func occurrenceDM(value int) string {
	if value == 0 {
		return "0"
	}

	return fmt.Sprintf("%+d", value)
}

// oneLine flattens a value to a single line. A world's name is the one
// field a referee writes in himself, so it is the one that can carry a
// line break, and a name broken over two lines is not a name any document
// here can set.
func oneLine(value string) string {
	return strings.NewReplacer("\n", " ", "\r", " ").Replace(value)
}

// cell writes a value into a Markdown table cell, where a pipe would
// otherwise break the row into columns the table does not have.
//
// This is Markdown's own syntax and nothing else's: the escape must not
// reach a document with no pipes to escape. It did once -- named() called
// this, so the PDF booklet drew a backslash the referee never typed,
// while the roster and the map beside it drew the same name plain.
func cell(value string) string {
	return strings.ReplaceAll(oneLine(value), "|", `\|`)
}

// namesByHex indexes the names the referee wrote in, so the route table
// and the detail headings can carry them too. P. 12 step 3 asks him to
// name each world and prints no table for it, so the name is his and the
// hex is the tool's; the listing prints the hex first everywhere, because
// that is what the map is labelled with.
func namesByHex(record *starmap.Record) map[starmap.Hex]string {
	names := make(map[starmap.Hex]string, len(record.Worlds))

	for _, world := range record.Worlds {
		if world.Name != "" {
			names[world.Hex] = oneLine(world.Name)
		}
	}

	return names
}

// named writes a world where its hex alone was the referee's only handle
// on it. An unnamed world is still the bare hex, so a record he has not
// annotated reads exactly as it did.
func named(names map[starmap.Hex]string, hex starmap.Hex) string {
	name, ok := names[hex]
	if !ok {
		return hex.String()
	}

	return hex.String() + " " + name
}

// bases names the bases a world has, which p. 5 prints throws for at
// starports A through D and nowhere else.
func bases(world starmap.World) string {
	var present []string

	if world.NavalBase {
		present = append(present, "naval")
	}

	if world.ScoutBase {
		present = append(present, "scout")
	}

	if len(present) == 0 {
		return "--"
	}

	return strings.Join(present, ", ")
}

func (r *Renderer) routes(built *strings.Builder, names map[starmap.Hex]string, record *starmap.Record) {
	built.WriteString("## Routes\n\n")

	if len(record.Routes) == 0 {
		built.WriteString("No route was drawn.\n\n")

		return
	}

	built.WriteString("| From | To | Parsecs |\n| --- | --- | --- |\n")

	for _, route := range record.Routes {
		fmt.Fprintf(built, "| %s | %s | %d |\n",
			cell(named(names, route.From)), cell(named(names, route.To)), route.Distance)
	}

	built.WriteString("\n")
}

func (r *Renderer) details(built *strings.Builder, names map[starmap.Hex]string, record *starmap.Record) {
	if len(record.Worlds) == 0 {
		return
	}

	built.WriteString("## The worlds in detail\n\n")
	built.WriteString(techIndexNote)

	for _, world := range record.Worlds {
		r.world(built, names, world)
	}
}

// techIndexNote says once what the technological index line does not
// repeat on every world: the index is generated (p. 9), but the tables
// saying what one means during play (pp. 10-11) are not transcribed, so
// the listing carries the digit alone. That is a gap and not a boundary
// -- issue 1 asks for the gloss, and p. 11 is the line players ask about
// -- and the tables are printed with holes besides, which p. 11 asks the
// referee or the players to fill in as play discovers them.
const techIndexNote = "The technological index carries its digit and no description. " +
	"The technological levels tables of pp. 10-11 say what an index means " +
	"during play rather than how it is generated, so this tool does not read " +
	"them; p. 11 asks the referee or the players to fill in their holes as " +
	"play discovers them.\n\n"

func (r *Renderer) world(built *strings.Builder, names map[starmap.Hex]string, world starmap.World) {
	fmt.Fprintf(built, "### %s &mdash; %s\n\n", named(names, world.Hex), world.Digits)

	starport := "no chart row"

	row, err := r.charts.StarportChart.Row(world.Starport)
	if err == nil {
		starport = row.Description
	}

	fmt.Fprintf(built, "- **Starport %s.** %s\n", world.Starport, starport)

	for _, line := range []struct {
		name   string
		value  int
		labels tables.Labels
	}{
		{"Size", world.Size, r.charts.Size},
		{"Atmosphere", world.Atmosphere, r.charts.Atmosphere},
		{"Hydrographics", world.Hydrographics, r.charts.Hydrographics},
		{"Population", world.Population, r.charts.Population},
		{"Government", world.Government, r.charts.Government},
		{"Law level", world.LawLevel, r.charts.LawLevels},
	} {
		fmt.Fprintf(built, "- **%s %s.**%s\n", line.name, digit(line.value), described(line.labels, line.value))
	}

	// No description: techIndexNote said once, at the head of the section,
	// why pp. 10-11 supply none.
	fmt.Fprintf(built, "- **Technological index %s.**\n", digit(world.TechIndex))

	fmt.Fprintf(built, "- **Bases.** %s\n", bases(world))

	for _, clamp := range world.Clamps {
		fmt.Fprintf(built, "- **Clamped.** %s threw %d and is recorded as %d.\n",
			clamp.Characteristic, clamp.Raw, clamp.Value)
	}

	built.WriteString("\n")
}

// described returns a value's label, or the note that the book prints
// none. R14 lets a generated value exceed a table's printed range -- an
// atmosphere of 13, a government of 14 -- and the book prints no label
// for one. That is a gap in the page, not an error to correct, and the
// listing says so: a bare digit reads as a defect in the tool exactly
// where the tool is being scrupulous.
func described(labels tables.Labels, value int) string {
	if label, ok := labels.Label(value); ok {
		return " " + label
	}

	// Getting past Label means above the last printed row: R14 floors
	// every characteristic at 0 and every table prints a row 0, so a value
	// the engine wrote can only run off the top. A record hand-edited
	// below 0 is outside the contract the schema states, and Decode does
	// not check it; it would read this note in the wrong direction.
	return aboveThePrintedRows
}

// aboveThePrintedRows is the remedy p. 8 prints for a value its tables do
// not cover -- "either the players or referee will generate a rationale
// which explains the situation, or an alternative description should be
// made" -- which is the whole of ERRATA E004's second part.
const aboveThePrintedRows = " Above the last row its table prints; p. 8 leaves the description " +
	"to the referee, to explain or to replace (ERRATA E004)."

// digit writes a value in the notation of Book 1 p. 8 extended by Book 3
// p. 2, which is how the string of digits writes it.
func digit(value int) string {
	written, err := starmap.NewDigit(value)
	if err != nil {
		return strconv.Itoa(value)
	}

	return written.String()
}

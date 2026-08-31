// Package render writes the subsector listing: the Markdown a referee can
// run from.
//
// It stands in for what p. 4 asks him to keep -- "at least one (and
// preferably several) pages in a central notebook maintained by the
// referee" -- so the listing carries the roster, the space lanes, and a
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

	"github.com/philoserf/ctworldgen/subsector"
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

// Subsector writes the listing.
func (r *Renderer) Subsector(out io.Writer, record *subsector.Subsector) error {
	var built strings.Builder

	names := namesByHex(record)

	r.heading(&built, record)
	r.roster(&built, record)
	r.lanes(&built, names, record)
	r.details(&built, names, record)

	_, err := io.WriteString(out, built.String())
	if err != nil {
		return fmt.Errorf("writing the listing: %w", err)
	}

	return nil
}

func (r *Renderer) heading(built *strings.Builder, record *subsector.Subsector) {
	name := record.Name
	if name == "" {
		name = "Subsector"
	}

	fmt.Fprintf(built, "# %s\n\n", name)
	fmt.Fprintf(built, "%d worlds, %d space lanes. Generated from seed %d at occurrence DM %s.\n\n",
		len(record.Worlds), len(record.Routes), record.Seed, occurrenceDM(record.OccurrenceDM))
}

// roster is the world roster: hexes, names, and strings of digits.
func (r *Renderer) roster(built *strings.Builder, record *subsector.Subsector) {
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

// cell writes a value into a Markdown table cell. A world's name is the
// one field a referee writes in himself, so it is the one that can carry
// a pipe or a line break, either of which would otherwise break the row
// into columns the table does not have.
func cell(value string) string {
	replaced := strings.NewReplacer("|", `\|`, "\n", " ", "\r", " ")

	return replaced.Replace(value)
}

// namesByHex indexes the names the referee wrote in, so the lane table
// and the detail headings can carry them too. P. 12 step 3 asks him to
// name each world and prints no table for it, so the name is his and the
// hex is the tool's; the listing prints the hex first everywhere, because
// that is what the map is labelled with.
func namesByHex(record *subsector.Subsector) map[subsector.Hex]string {
	names := make(map[subsector.Hex]string, len(record.Worlds))

	for _, world := range record.Worlds {
		if world.Name != "" {
			names[world.Hex] = world.Name
		}
	}

	return names
}

// named writes a world where its hex alone was the referee's only handle
// on it. An unnamed world is still the bare hex, so a record he has not
// annotated reads exactly as it did.
func named(names map[subsector.Hex]string, hex subsector.Hex) string {
	name, ok := names[hex]
	if !ok {
		return hex.String()
	}

	return hex.String() + " " + cell(name)
}

// bases names the bases a world has, which p. 5 prints throws for at
// starports A through D and nowhere else.
func bases(world subsector.World) string {
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

func (r *Renderer) lanes(built *strings.Builder, names map[subsector.Hex]string, record *subsector.Subsector) {
	built.WriteString("## Space lanes\n\n")

	if len(record.Routes) == 0 {
		built.WriteString("No space lane was drawn.\n\n")

		return
	}

	built.WriteString("| From | To | Parsecs |\n| --- | --- | --- |\n")

	for _, route := range record.Routes {
		fmt.Fprintf(built, "| %s | %s | %d |\n", named(names, route.From), named(names, route.To), route.Distance)
	}

	built.WriteString("\n")
}

func (r *Renderer) details(built *strings.Builder, names map[subsector.Hex]string, record *subsector.Subsector) {
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
// saying what one means during play (pp. 10-11) are out of the scope
// PRD.md declares, so the listing carries the digit alone. Those tables
// are printed with holes besides, which p. 11 asks the referee or the
// players to fill in as play discovers items or devices of interest.
const techIndexNote = "The technological index carries its digit and no description. " +
	"The technological levels tables of pp. 10-11 say what an index means " +
	"during play rather than how it is generated, so this tool does not read " +
	"them; p. 11 asks the referee or the players to fill in their holes as " +
	"play discovers them.\n\n"

func (r *Renderer) world(built *strings.Builder, names map[subsector.Hex]string, world subsector.World) {
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
	written, err := subsector.NewDigit(value)
	if err != nil {
		return strconv.Itoa(value)
	}

	return written.String()
}

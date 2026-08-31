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

	r.heading(&built, record)
	r.roster(&built, record)
	r.lanes(&built, record)
	r.details(&built, record)

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
		fmt.Fprintf(built, "| %s | %s | %s | %s |\n", world.Hex, world.Name, world.Digits, bases(world))
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

func (r *Renderer) lanes(built *strings.Builder, record *subsector.Subsector) {
	built.WriteString("## Space lanes\n\n")

	if len(record.Routes) == 0 {
		built.WriteString("No space lane was drawn.\n\n")

		return
	}

	built.WriteString("| From | To | Parsecs |\n| --- | --- | --- |\n")

	for _, route := range record.Routes {
		fmt.Fprintf(built, "| %s | %s | %d |\n", route.From, route.To, route.Distance)
	}

	built.WriteString("\n")
}

func (r *Renderer) details(built *strings.Builder, record *subsector.Subsector) {
	if len(record.Worlds) == 0 {
		return
	}

	built.WriteString("## The worlds in detail\n\n")

	for _, world := range record.Worlds {
		r.world(built, world)
	}
}

func (r *Renderer) world(built *strings.Builder, world subsector.World) {
	fmt.Fprintf(built, "### %s &mdash; %s\n\n", world.Hex, world.Digits)

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

	// The technological level tables of pp. 10-11 say what an index means
	// during play, which is referee reference and not part of generation,
	// so the listing carries the digit alone.
	fmt.Fprintf(built, "- **Technological index %s.**\n", digit(world.TechIndex))

	fmt.Fprintf(built, "- **Bases.** %s\n", bases(world))

	for _, clamp := range world.Clamps {
		fmt.Fprintf(built, "- **Clamped.** %s threw %d and is recorded as %d.\n",
			clamp.Characteristic, clamp.Raw, clamp.Value)
	}

	built.WriteString("\n")
}

// described returns a value's label, or nothing at all. R14 lets a
// generated value exceed a table's printed range -- an atmosphere of 13,
// a government of 14 -- and the book prints no label for one. That is a
// gap in the page, not an error to correct, so the listing prints the
// digit and no description.
func described(labels tables.Labels, value int) string {
	if label, ok := labels.Label(value); ok {
		return " " + label
	}

	return ""
}

// digit writes a value in the notation of Book 1 p. 8 extended by Book 3
// p. 2, which is how the string of digits writes it.
func digit(value int) string {
	written, err := subsector.NewDigit(value)
	if err != nil {
		return strconv.Itoa(value)
	}

	return written.String()
}

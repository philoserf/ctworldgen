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

// Renderer holds the descriptive tables of pp. 5-7, loaded once, and which
// lanes the documents draw.
type Renderer struct {
	charts *tables.Tables
	lanes  Lanes
}

// New loads and validates the charts the listing reads its labels from, and
// fixes which lanes the documents draw.
//
// The lanes are an argument rather than a default because neither answer is
// obviously the zero one, and a rendering that quietly stopped drawing half
// a map would be the worst way to find that out.
func New(lanes Lanes) (*Renderer, error) {
	charts, err := tables.Load()
	if err != nil {
		return nil, fmt.Errorf("loading the Book 3 charts: %w", err)
	}

	return &Renderer{charts: charts, lanes: lanes}, nil
}

// listing is the Markdown typesetter's document, and it is deliberately
// the shape of booklet: the state a document is written from, held once,
// rather than threaded through every section's parameters. The two
// typesetters share a middle -- bullets, member/members, legible, summary,
// named, bases -- and structural parallelism is what keeps that middle
// discoverable. Before this, five of the Markdown sections took six
// parameters on a *Renderer receiver they never touched, so the next thing
// that has to agree between the two backends had an obvious home on one
// side and none on the other.
type listing struct {
	built  strings.Builder
	charts *tables.Tables
	record *starmap.Record

	// drawn is the lanes this listing shows, which is every one the record
	// carries unless the renderer was built for legible lanes (ERRATA
	// E007). The record itself is unchanged.
	drawn []starmap.Route

	names map[starmap.Hex]string
}

// Listing writes the Markdown listing.
//
// A subsector's listing is one map, one roster, one lane table and a block
// per world. A sector's is an index of the whole grid and then the sixteen
// sub-sector listings its members would have had, because that is what a
// referee runs from: he reads the sector to place a campaign and one
// sub-sector to run a session (ERRATA E008).
func (r *Renderer) Listing(out io.Writer, record *starmap.Record) error {
	doc := &listing{
		built:  strings.Builder{},
		charts: r.charts,
		record: record,
		drawn:  r.drawn(record.Routes),
		names:  namesByHex(record),
	}

	doc.heading()

	if record.Grid.IsSector() {
		doc.sectorSections()
	} else {
		doc.grid(wholeGrid(record.Grid), everywhere, "## The map", mapNote(record))
		doc.roster("## Worlds", record.Worlds)
		doc.routes("## Routes", record.Routes, doc.drawn, nil)
		doc.details("## The worlds in detail", "###", record.Worlds)
	}

	_, err := io.WriteString(out, doc.built.String())
	if err != nil {
		return fmt.Errorf("writing the listing: %w", err)
	}

	return nil
}

// sectorSections writes the index of the whole sector and then a section
// per member (ERRATA E008).
func (l *listing) sectorSections() {
	gathered := members(l.record, l.drawn)

	l.sectorIndex(gathered)

	for index := range gathered {
		l.memberSection(&gathered[index])
	}
}

// memberSection writes one of a sector's sixteen sub-sectors as the
// listing it would have had on its own: its own p. 3 grid, its own roster,
// its own lanes and its own worlds in detail.
func (l *listing) memberSection(part *member) {
	fmt.Fprintf(&l.built, "## Subsector %d &mdash; %s to %s\n\n", part.Index, part.First, part.Last)
	fmt.Fprintf(&l.built, "%s\n\n", part.provenance(l.record.Seed))
	fmt.Fprintf(&l.built, "%s\n\n", part.summary())

	l.grid(memberWindow(part.Index), part.shows, "### The map", part.mapNote())
	l.roster("### Worlds", part.Worlds)
	l.routes("### Routes", part.Carried, part.Lanes, leavingFor(part.Index))
	l.details("### The worlds in detail", "####", part.Worlds)
}

// drawn returns the lanes this renderer puts on the page. The record is not
// changed and carries every one of them (ERRATA E007).
func (r *Renderer) drawn(routes []starmap.Route) []starmap.Route {
	if r.lanes == AllLanes {
		return routes
	}

	return legible(routes)
}

func (l *listing) heading() {
	name := l.record.Name
	if name == "" {
		name = untitled(l.record)
	}

	fmt.Fprintf(&l.built, "# %s\n\n", name)
	fmt.Fprintf(&l.built, "%s\n\n", summary(l.record, l.drawn))

	// The referee's note about the map as a whole, under the summary and
	// above everything the tool generated (issue 1 #6).
	if l.record.Notes != "" {
		fmt.Fprintf(&l.built, "%s\n\n", oneLine(l.record.Notes))
	}
}

// summary is the sentence both documents open with, so that the two report
// the same record in the same words. It says how many lanes were drawn
// whenever that is fewer than the record carries: a document that quietly
// showed less than it had would be worse than an unreadable one.
func summary(record *starmap.Record, drawn []starmap.Route) string {
	if len(drawn) == len(record.Routes) {
		return fmt.Sprintf("%d worlds, %d routes. Generated from seed %d at occurrence DM %s.",
			len(record.Worlds), len(record.Routes), record.Seed, occurrenceDM(record.OccurrenceDM))
	}

	return fmt.Sprintf("%d worlds, %d routes, %d drawn. Generated from seed %d at occurrence DM %s.",
		len(record.Worlds), len(record.Routes), len(drawn), record.Seed, occurrenceDM(record.OccurrenceDM))
}

// untitled is the heading a record the referee has not named gets. It is
// what the record is, and a sector is not a subsector: those two grids
// are the only shapes a record takes (ERRATA E006), and heading a 32x40
// one "Subsector" is the listing misreporting its own subject.
func untitled(record *starmap.Record) string {
	if record.Grid.IsSector() {
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

// noWorlds is what a roster with nothing in it says. Every roster either
// document writes is a sub-sector's -- a record's own, or one of a
// sector's sixteen members -- so the sentence is true wherever it appears.
const noWorlds = "No world was placed. An empty subsector is a result."

// mapNote says what the map shows and, as much to the point, what it does
// not: p. 2 asks for "a line connecting the two worlds on the map" and
// this one draws none.
//
// The first sentence names the grid actually drawn. The rest is true of
// either, because the parity it describes is the p. 3 parity in both
// cases -- which is exactly what makes the sector translation safe
// (ERRATA E006 part 2).
func mapNote(record *starmap.Record) string {
	if record.Grid.IsSector() {
		return sectorIndexNote
	}

	return pageThreeGridNote + mapNoteTail
}

const pageThreeGridNote = "The p. 3 sub-sector hex grid."

// sectorIndexNote heads the map of the whole sector, which is an index and
// says so. It carries no hex numbers: sixteen grids on one sheet leave a
// four-digit number no room, and the member maps below carry them one grid
// at a time (ERRATA E008 part 4).
const sectorIndexNote = "Sixteen p. 3 sub-sector hex grids on one sheet, 0101 through 3240 " +
	"(ERRATA E006). This is an index and not a grid to read a hex off: a hex carries the letter " +
	"of its starport, or a full stop where p. 1 says to leave it blank, and the bars and blank " +
	"lines mark the seams between the sixteen. Each sub-sector is drawn again below, on its own " +
	"numbered grid (ERRATA E008). P. 2 also asks for a line drawn between the worlds a route " +
	"joins; a monospace grid has nowhere to put one, so this map draws none. " +
	"`render --format pdf` draws them.\n\n"

// mapNoteTail is the half of a map's note that is true of every map any
// document here draws, because the parity it describes is the p. 3 parity
// in all of them -- which is exactly what makes the sector translation
// safe (ERRATA E006 part 2).
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
func (l *listing) grid(draw window, shows func(starmap.Hex) bool, heading, note string) {
	fmt.Fprintf(&l.built, "%s\n\n%s", heading, note)
	l.built.WriteString("```text\n")

	marked := marks(l.record)

	for row := draw.FromRow; row <= draw.ToRow; row++ {
		// A member's window rings it with a hex of bleed, and a member on
		// the edge of the sector has none on that side: those rows are not
		// on the l.record's grid and are not drawn.
		if row < 1 || row > l.record.Grid.Rows {
			continue
		}

		// Odd columns first, then the even ones half a slot right: one
		// printed row of the grid is two lines of the map.
		l.built.WriteString(gridLine(marked, l.record.Grid, draw, shows, row, 1))
		l.built.WriteString(gridLine(marked, l.record.Grid, draw, shows, row, 0))
	}

	l.built.WriteString("```\n\n")
}

// marks is the starport letter p. 1 says to write in each hex that has a
// world, which is the whole of what any of these maps draws.
func marks(record *starmap.Record) map[starmap.Hex]starmap.Starport {
	marked := make(map[starmap.Hex]starmap.Starport, len(record.Worlds))
	for _, world := range record.Worlds {
		marked[world.Hex] = world.Starport
	}

	return marked
}

// gridLine draws the columns of one parity across a single row of the
// record's grid. parity is the remainder that selects them: 1 for the
// high columns, 0 for the low.
func gridLine(
	marked map[starmap.Hex]starmap.Starport, grid starmap.Grid,
	draw window, shows func(starmap.Hex) bool, row, parity int,
) string {
	var line strings.Builder

	// Column c is drawn half a slot right of column c-1, so the line whose
	// first column is not the window's own gets the half-slot indent.
	//
	// Which line that is depends on the window: a subsector's begins at
	// column 1 and its low line is indented, and every member's begins at
	// an even column -- a member's first column is 8k+1, and the ring puts
	// the window one to the left of it -- so its high line is. Indenting
	// the low line always shifts every member map by a whole slot, into a
	// grid that looks entirely tidy and disagrees with Hex.Distance.
	if parity != draw.FromCol%2 {
		line.WriteString(strings.Repeat(" ", mapHalfSlot))
	}

	for col := draw.FromCol; col <= draw.ToCol; col++ {
		// Parity is the hex's own column and never its offset in the
		// window, for the same reason.
		if col%2 != parity {
			continue
		}

		hex := starmap.Hex{Col: col, Row: row}

		// A bleed column off the sector's own grid keeps its slot, so that
		// the columns of an edge member line up with an interior one's.
		cell := ""
		if grid.Contains(hex) && shows(hex) {
			cell = hex.String() + " "
			if starport, ok := marked[hex]; ok {
				cell += starport.String()
			}
		}

		line.WriteString(cell)

		// max: a Starport the schema would have rejected prints as
		// "Starport(0)", which is wider than a slot. The rest of the
		// listing degrades on such a record rather than stopping, so the
		// map does too -- a negative count here would panic.
		line.WriteString(strings.Repeat(" ", max(mapSlot-len(cell), 0)))
	}

	return strings.TrimRight(line.String(), " ") + "\n"
}

// roster is the world roster: hexes, names, and strings of digits.
func (l *listing) roster(heading string, worlds []starmap.World) {
	fmt.Fprintf(&l.built, "%s\n\n", heading)

	if len(worlds) == 0 {
		fmt.Fprintf(&l.built, "%s\n\n", noWorlds)

		return
	}

	l.built.WriteString("| Hex | Name | Digits | Bases |\n| --- | --- | --- | --- |\n")

	for _, world := range worlds {
		fmt.Fprintf(&l.built, "| %s | %s | %s | %s |\n", world.Hex, cell(world.Name), world.Digits, bases(world))
	}

	l.built.WriteString("\n")
}

// lanesNote says what the table is not showing and how to see it. P. 2
// offers the map-drawer this and the tool takes it by default, so the
// document has to say so where a referee reading the table would look
// (ERRATA E007).
func lanesNote(all, suppressed int) string {
	return fmt.Sprintf(
		"%d of these %d lanes are not listed: each joins two worlds already joined by shorter lanes, "+
			"which p. 2 says may be ignored in the drawing (ERRATA E007). The record carries every one of "+
			"them, and `render --lanes=all` lists them.",
		suppressed, all)
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

func (l *listing) routes(heading string, carried, inked []starmap.Route, into func(starmap.Route) string) {
	fmt.Fprintf(&l.built, "%s\n\n", heading)

	if len(carried) == 0 {
		l.built.WriteString("No route was drawn.\n\n")

		return
	}

	if suppressed := len(carried) - len(inked); suppressed > 0 {
		fmt.Fprintf(&l.built, "%s\n\n", lanesNote(len(carried), suppressed))
	}

	if into == nil {
		l.built.WriteString("| From | To | Parsecs |\n| --- | --- | --- |\n")
	} else {
		l.built.WriteString("| From | To | Parsecs | Into |\n| --- | --- | --- | --- |\n")
	}

	for _, route := range inked {
		fmt.Fprintf(&l.built, "| %s | %s | %d",
			cell(named(l.names, route.From)), cell(named(l.names, route.To)), route.Distance)

		if into != nil {
			fmt.Fprintf(&l.built, " | %s", into(route))
		}

		l.built.WriteString(" |\n")
	}

	l.built.WriteString("\n")
}

// leavingFor names the sub-sector a lane's far end sits in, which is the
// Into column of a member's lane table. A lane both of whose ends are at
// home leaves for nowhere and says so (ERRATA E008 part 3).
func leavingFor(home int) func(starmap.Route) string {
	return func(route starmap.Route) string {
		for _, end := range []starmap.Hex{route.From, route.To} {
			if other := starmap.MemberOf(end); other != home {
				return fmt.Sprintf("subsector %d", other)
			}
		}

		return "--"
	}
}

func (l *listing) details(heading, level string, worlds []starmap.World) {
	if len(worlds) == 0 {
		return
	}

	fmt.Fprintf(&l.built, "%s\n\n%s", heading, technologyNote)

	for _, world := range worlds {
		l.world(level, world)
	}
}

// technologyNote says once, at the head of the section, what the
// technological index line on every world is made of. It is the citation
// ERRATA E011 requires: a reader meets the borrowed book here, before he
// meets anything it says.
const technologyNote = "What a technological index means is described in two halves. " +
	"The first is T5's -- Core Book 2 pp. 230-232, cited for description alone and never " +
	"for a throw (ERRATA E011): the band, the era it anchors the level to, and the level's " +
	"energy, society and settlements. The second is pp. 10-11, read downward -- an entry " +
	"printed at a level is the best of its kind until the next one (E009), and a hole is the " +
	"page inviting the referee to fill it rather than an absence (E010). Where both books " +
	"cover the same ground the held page speaks.\n\n"

// technological glosses an index: what a world at that level is like, and
// what it can build.
//
// The index is thrown from the p. 9 matrix (R12) and neither table
// generates anything, so this is description and only description.
//
// It is in two halves because it comes from two books, and it says so:
// the first names T5, which technologyNote has already introduced. The
// halves are also divided by subject rather than arbitrarily. **Where the
// two books cover the same ground, the held page speaks.** T5's transport
// and computing columns are the two this does not print, because Book 3's
// own air, space and computers columns say the same thing in the held
// page's words -- grav belts and drives N or less against "gravity
// manipulation lifters to orbit", Model/6 against "Model /6".
func technological(charts *tables.Tables, index int) string {
	level := charts.TechLevels.Level(index)

	borrowed := clauses(
		level.Borrowed.Energy,
		level.Borrowed.Society,
		level.Borrowed.Environ,
	)

	held := clauses(
		level.Held.Personal,
		level.Held.Armor,
		level.Held.Computers,
		level.Held.Air,
		level.Held.Space,
	)

	if level.Held.MatterTransport {
		held = append(held, "matter transport")
	}

	var built strings.Builder

	fmt.Fprintf(&built, "%s Tech (T5), %s",
		level.Borrowed.Band, strings.TrimSuffix(level.Borrowed.Era, "."))

	if len(borrowed) > 0 {
		fmt.Fprintf(&built, ": %s", strings.Join(borrowed, "; "))
	}

	built.WriteString(".")

	if len(held) > 0 {
		fmt.Fprintf(&built, " Pp. 10-11: %s.", strings.Join(held, "; "))
	}

	return built.String()
}

// clauses drops the columns a page prints nothing in at or below the
// level -- a hole is not an absence, and saying nothing is what the page
// does (ERRATA E010) -- and drops the full stop each T5 cell ends in, so
// that a run of them reads as one sentence rather than as five.
//
// It also collapses a run of neighbouring columns carrying the same
// entry. Water and land are both Hovercraft from level 7, and "Hovercraft;
// Hovercraft" reads as a fault in the tool rather than as the page.
//
// The clauses are joined with semicolons rather than commas because half
// of these cells are lists already -- "Carbine, Rifle, Pistol, SMG" is one
// cell of p. 10 -- and a comma between them hides where one column ends
// and the next begins.
func clauses(entries ...string) []string {
	kept := make([]string, 0, len(entries))

	for _, entry := range entries {
		trimmed := strings.TrimSuffix(entry, ".")
		if trimmed == "" || (len(kept) > 0 && kept[len(kept)-1] == trimmed) {
			continue
		}

		kept = append(kept, trimmed)
	}

	return kept
}

func (l *listing) world(level string, world starmap.World) {
	fmt.Fprintf(&l.built, "%s %s &mdash; %s\n\n", level, named(l.names, world.Hex), world.Digits)

	for _, line := range bullets(l.charts, world) {
		l.built.WriteString(line.markdown())
	}

	l.built.WriteString("\n")
}

// bullet is one line of a world's detail: the label both documents set in
// bold, and the description the charts give for it. The description is
// stored trimmed, because it is a value here rather than a fragment of
// either document's punctuation.
type bullet struct {
	label       string
	description string
}

// markdown writes one bullet as the listing sets it: the label in bold,
// then a single space and the description.
//
// Every bullet has one. The technological index line was the exception
// until pp. 10-11 were read, and the branch that carried it is gone with
// it; bullets is what holds the promise, and a test sweeps the goldens
// for a bullet that ends at its label.
func (b bullet) markdown() string {
	return "- **" + b.label + "** " + b.description + "\n"
}

// bullets is a world's detail lines, in the order the p. 4 Planetary
// Characteristics box lists them.
//
// One list, two typesetters. The Markdown listing sets these as bold
// labels through markdown() above, and the booklet lays the same slice out
// as a block of wrapped lines. They were written out twice and had to
// agree by convention -- two suites checking two copies -- and a change to
// one was a change the other's tests could not see. Now they agree by
// construction.
func bullets(charts *tables.Tables, world starmap.World) []bullet {
	starport := "no chart row"

	row, err := charts.StarportChart.Row(world.Starport)
	if err == nil {
		starport = row.Description
	}

	// One line for the starport, six for the characteristics of pp. 5-8,
	// one for the technological index, one for the bases, and one for
	// every clamp that bound.
	const fixedLines = 9

	lines := make([]bullet, 0, fixedLines+len(world.Clamps))

	lines = append(lines, bullet{
		label:       fmt.Sprintf("Starport %s.", world.Starport),
		description: starport,
	})

	for _, line := range []struct {
		name   string
		value  int
		labels tables.Labels
	}{
		{"Size", world.Size, charts.Size},
		{"Atmosphere", world.Atmosphere, charts.Atmosphere},
		{"Hydrographics", world.Hydrographics, charts.Hydrographics},
		{"Population", world.Population, charts.Population},
		{"Government", world.Government, charts.Government},
		{"Law level", world.LawLevel, charts.LawLevels},
	} {
		lines = append(lines, bullet{
			label:       fmt.Sprintf("%s %s.", line.name, digit(line.value)),
			description: strings.TrimSpace(described(line.labels, line.value)),
		})
	}

	lines = append(lines,
		bullet{
			label:       fmt.Sprintf("Technological index %s.", digit(world.TechIndex)),
			description: technological(charts, world.TechIndex),
		},
		bullet{label: "Bases.", description: bases(world)},
	)

	for _, clamp := range world.Clamps {
		lines = append(lines, bullet{
			label: "Clamped.",
			description: fmt.Sprintf("%s threw %d and is recorded as %d.",
				clamp.Characteristic, clamp.Raw, clamp.Value),
		})
	}

	// Last, and only where he wrote one: everything above is the book's,
	// and this is the referee's (issue 1 #6). Flattened as his name for a
	// world already is -- a line break would end the Markdown list, and the
	// two documents must carry the same line.
	if world.Notes != "" {
		lines = append(lines, bullet{label: "Notes.", description: oneLine(world.Notes)})
	}

	return lines
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

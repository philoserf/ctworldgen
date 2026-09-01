package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/philoserf/ctworldgen/starmap"
)

// member is one of a sector's sixteen sub-sectors, gathered out of the
// record the sector wrote.
//
// Nothing here is decided. Which member a world or a lane belongs to is
// read straight off the hex, because starmap.Place put member i's hexes in
// band i and starmap.MemberOf reads the band back (ERRATA E006 part 1).
// The record is not changed and carries no member field: a sector that had
// to remember its own decomposition could disagree with its own grid.
type member struct {
	Index int
	First starmap.Hex
	Last  starmap.Hex

	Worlds []starmap.World

	// Carried is every lane of the record with an end in this member, and
	// Lanes is the subset the documents draw (ERRATA E007). A lane that
	// crosses a seam is in both its members' lists, because a referee
	// reading one sub-sector needs to see the road out of it (ERRATA E008
	// part 3) -- so the sixteen tables together list more rows than the
	// record has lanes, and the section that carries them says so.
	Carried []starmap.Route
	Lanes   []starmap.Route

	// Crossing counts the lanes of Carried with an end in another member.
	Crossing int

	// OffMap counts the drawn lanes this member's map cannot draw. A lane
	// runs up to four parsecs (p. 2) and the map reaches one hex past the
	// member's border, so a long lane out of the sub-sector has no far end
	// on the sheet. It is in the lane table with the sub-sector it leaves
	// for named, and the map's note says how many (ERRATA E008 part 2).
	OffMap int
}

// members gathers the sixteen out of the record, in index order.
func members(record *starmap.Record, drawn []starmap.Route) []member {
	gathered := make([]member, starmap.Members)

	for index := range gathered {
		first, last := starmap.MemberBounds(index)

		gathered[index] = member{
			Index: index, First: first, Last: last,
			Worlds: nil, Carried: nil, Lanes: nil, Crossing: 0, OffMap: 0,
		}
	}

	for _, world := range record.Worlds {
		at := starmap.MemberOf(world.Hex)

		gathered[at].Worlds = append(gathered[at].Worlds, world)
	}

	for _, route := range record.Routes {
		home, other, crosses := ends(route)

		gathered[home].Carried = append(gathered[home].Carried, route)

		if crosses {
			gathered[home].Crossing++

			gathered[other].Carried = append(gathered[other].Carried, route)
			gathered[other].Crossing++
		}
	}

	for _, route := range drawn {
		home, other, crosses := ends(route)

		gathered[home].hold(route)

		if crosses {
			gathered[other].hold(route)
		}
	}

	return gathered
}

// hold adds a drawn lane to a member, counting it as one its map cannot
// draw when either end falls outside the hexes that map shows.
func (m *member) hold(route starmap.Route) {
	m.Lanes = append(m.Lanes, route)

	if !m.shows(route.From) || !m.shows(route.To) {
		m.OffMap++
	}
}

// shows reports whether a hex is on this member's map (ERRATA E008 part 2).
func (m *member) shows(hex starmap.Hex) bool { return onMemberMap(m.Index, hex) }

// ends returns the members a lane's two ends sit in, and whether they
// differ. A lane that crosses is carried by both (ERRATA E008 part 3).
func ends(route starmap.Route) (int, int, bool) {
	from := starmap.MemberOf(route.From)
	to := starmap.MemberOf(route.To)

	return from, to, from != to
}

// memberSeed is the seed that writes a member on its own: the sector's
// base plus the member's index (ERRATA E006 part 1). The index is one of
// sixteen, so the conversion cannot overflow, and the bound is asserted
// rather than asserted in a comment.
func memberSeed(base uint64, index int) uint64 {
	if index < 0 || index >= starmap.Members {
		return base
	}

	return base + uint64(index)
}

// provenance is the line that makes a member section usable on its own:
// the seed that writes this sub-sector standalone, which is the identity
// E006 part 1 keeps and the README already promises. A referee who wants
// this one sub-sector and nothing else can generate it from here.
func (m *member) provenance(seed uint64) string {
	return fmt.Sprintf(
		"Member %d of this sector. `ctworldgen new --seed %d` writes this sub-sector on its own "+
			"p. 3 grid, where its first hex is 0101; here it is laid on the sector's, and that hex "+
			"is %s (ERRATA E006).",
		m.Index, memberSeed(seed, m.Index), m.First)
}

// summary counts this member the way the document's own heading counts the
// sector, and separates the lanes that stay at home from the ones that
// leave, because the second kind is what a sector adds.
func (m *member) summary() string {
	return fmt.Sprintf("%d worlds, %d lanes within it, %d crossing into its neighbours.",
		len(m.Worlds), len(m.Carried)-m.Crossing, m.Crossing)
}

// mapNote heads a member's own map. It says what the ring of neighbouring
// hexes is for and, where there are any, how many drawn lanes reach past
// it -- because a map that quietly stopped drawing some of its lanes would
// be worse than one that drew none (ERRATA E008 part 2).
func (m *member) mapNote() string {
	note := "This sub-sector's eighty hexes, numbered where the sector lays them, with one hex of " +
		"each neighbour drawn around the edge, so that a lane crossing a seam to an adjacent hex " +
		"has a far end that can be looked up (ERRATA E008 part 2)."

	if m.OffMap > 0 {
		note += fmt.Sprintf(" A lane runs up to four parsecs (p. 2) and this map reaches one, so %d "+
			"of its lanes end past the edge and are drawn here as no line; the table below carries "+
			"them, and names the sub-sector each one leaves for.", m.OffMap)
	}

	return note + mapNoteTail
}

// sectorIndex writes the map of the whole sector and the table of what is
// in each of its sixteen (ERRATA E008 part 4).
func (r *Renderer) sectorIndex(built *strings.Builder, record *starmap.Record, gathered []member) {
	built.WriteString("## The sector\n\n")
	built.WriteString(mapNote(record))

	indexMap(built, record)

	built.WriteString(sectorContentsNote)
	built.WriteString("| Subsector | Hexes | Worlds | Lanes within | Crossing | Seed |\n")
	built.WriteString("| --- | --- | --- | --- | --- | --- |\n")

	for _, part := range gathered {
		fmt.Fprintf(built, "| %d | %s to %s | %d | %d | %d | %d |\n",
			part.Index, part.First, part.Last, len(part.Worlds),
			len(part.Carried)-part.Crossing, part.Crossing, memberSeed(record.Seed, part.Index))
	}

	built.WriteString("\n")
}

const sectorContentsNote = "A lane that crosses a seam is listed under both the sub-sectors it " +
	"joins, so the sixteen tables below carry more rows between them than the record has lanes " +
	"(ERRATA E008 part 3).\n\n"

// The index map's slot. Two characters: the starport letter p. 1 says to
// mark a hex with, and the space that carries the half-slot offset of the
// p. 3 parity. A four-digit number would make the sheet a hundred and
// sixty columns wide, which is what this map exists not to be.
const (
	indexSlot     = 2
	indexHalfSlot = indexSlot / 2
	indexEmpty    = "."
)

// indexMap draws the whole sector as an index: starport letters, the seams
// between the sixteen, and no hex numbers (ERRATA E008 part 4).
func indexMap(built *strings.Builder, record *starmap.Record) {
	built.WriteString("```text\n")

	marked := marks(record)

	for row := 1; row <= record.Grid.Rows; row++ {
		built.WriteString(indexLine(marked, record.Grid, row, 1))
		built.WriteString(indexLine(marked, record.Grid, row, 0))

		// The seams that run across the sheet, between the bands of
		// members. A blank line rather than a rule, because a rule would
		// have to interleave with the half-row offset and a blank line
		// reads at a glance.
		if row%starmap.Rows == 0 && row < record.Grid.Rows {
			built.WriteString("\n")
		}
	}

	built.WriteString("```\n\n")
}

// indexLine draws the columns of one parity across a single row.
func indexLine(marked map[starmap.Hex]starmap.Starport, grid starmap.Grid, row, parity int) string {
	var line strings.Builder

	if parity == 0 {
		line.WriteString(strings.Repeat(" ", indexHalfSlot))
	}

	for col := 1; col <= grid.Columns; col++ {
		if col%2 != parity {
			continue
		}

		hex := starmap.Hex{Col: col, Row: row}

		cell := indexEmpty
		if starport, ok := marked[hex]; ok {
			cell = starport.String()
		}

		line.WriteString(cell)
		line.WriteByte(' ')

		// The seam goes where this line's next hex is in another member,
		// which is after column 7 on the high line and after column 8 on
		// the low one -- a line carries only every second column, so a bar
		// placed by column number alone draws a staircase.
		next := starmap.Hex{Col: col + indexSlot, Row: row}
		if grid.Contains(next) && starmap.MemberOf(hex) != starmap.MemberOf(next) {
			line.WriteString("| ")
		}
	}

	return strings.TrimRight(line.String(), " ") + "\n"
}

// sectorPages writes the booklet a sector gets: the index of the whole
// grid, the table of what is in each of the sixteen, and then the sixteen
// sub-sector booklets themselves (ERRATA E008).
func (b *booklet) sectorPages() {
	gathered := members(b.record, b.drawn)

	b.indexPage()
	b.contentsPage(gathered)

	for index := range gathered {
		b.memberPages(&gathered[index])
	}
}

// indexPage is the sector's first page: the title, the summary the listing
// opens with, and the index map across the whole width.
func (b *booklet) indexPage() {
	b.newPage()

	name := b.record.Name
	if name == "" {
		name = untitled(b.record)
	}

	b.pdf.SetTitle(b.encode(name), false)
	b.pdf.SetFont("Helvetica", "B", titleSize)
	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)
	b.text(pageMargin, b.y+titleSize, name)

	b.y += titleLead

	b.pdf.SetFont("Helvetica", "", noteSize)
	b.pdf.SetTextColor(inkNote, inkNote, inkNote)
	b.text(pageMargin, b.y+noteSize, summary(b.record, b.drawn))

	b.y += sectionGap

	if b.record.Notes != "" {
		b.body(oneLine(b.record.Notes))
	}

	b.rule(pageMargin, contentRight)

	b.y += blockGap

	b.drawIndex(box{X: pageMargin, Y: b.y, Width: contentWidth, Height: contentBottom - b.y})
}

// contentsPage is what a booklet of a hundred and seventy pages owes its
// reader: which sub-sector is which, how much is in it, and the seed that
// writes it on its own.
func (b *booklet) contentsPage(gathered []member) {
	b.heading("The sixteen sub-sectors")
	b.body(sectorContentsNote)

	b.pdf.SetFont("Helvetica", "B", rosterSize)
	b.contentsCells("Subsector", "Hexes", "Worlds", "Lanes within", "Crossing", "Seed")

	b.y += blockGap
	b.rule(pageMargin, contentRight)

	b.y += blockGap / two

	for _, part := range gathered {
		b.pdf.SetFont("Helvetica", "", rosterSize)
		b.contentsCells(
			strconv.Itoa(part.Index),
			part.First.String()+" to "+part.Last.String(),
			strconv.Itoa(len(part.Worlds)),
			strconv.Itoa(len(part.Carried)-part.Crossing),
			strconv.Itoa(part.Crossing),
			strconv.FormatUint(memberSeed(b.record.Seed, part.Index), 10),
		)
	}
}

// contentsColumn is the width of one column of the contents table. Six
// equal columns rather than fitted ones: every cell is a short number or a
// pair of hexes, and nothing here is the referee's own text.
const contentsColumn = contentWidth / 6

func (b *booklet) contentsCells(cells ...string) {
	baseline := b.y + rosterSize

	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)

	for at, text := range cells {
		b.text(pageMargin+contentsColumn*float64(at), baseline, b.clip(text, contentsColumn-blockGap))
	}

	b.y += rosterLead
}

// memberPages writes one of the sixteen as the booklet it would have had
// on its own: its map beside its roster, then its lanes, then its worlds.
func (b *booklet) memberPages(part *member) {
	b.newPage()

	b.pdf.SetFont("Helvetica", "B", headingSize)
	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)
	b.text(pageMargin, b.y+headingSize,
		fmt.Sprintf("Subsector %d — %s to %s", part.Index, part.First, part.Last))

	b.y += headingLead

	b.body(part.provenance(b.record.Seed))
	b.body(part.summary())
	b.body(part.mapNote())
	b.rule(pageMargin, contentRight)

	b.y += blockGap

	b.drawMap(box{X: rightColumnX, Y: b.y, Width: columnWidth, Height: contentBottom - b.y},
		memberWindow(part.Index), part.shows)
	b.rosterSection(pageMargin, columnWidth, b.y, part.Worlds)

	b.routesSection(fmt.Sprintf("Subsector %d — routes", part.Index),
		part.Carried, part.Lanes, leavingFor(part.Index))
	b.detailsSection(fmt.Sprintf("Subsector %d — the worlds in detail", part.Index), part.Worlds)
}

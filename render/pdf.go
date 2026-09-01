package render

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-pdf/fpdf"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"

	"github.com/philoserf/ctworldgen/starmap"
	"github.com/philoserf/ctworldgen/tables"
)

// The page. US Letter portrait, in PostScript points, which is what
// fpdf.New is asked for below.
const (
	pageWidth  = 612.0
	pageHeight = 792.0
	pageMargin = 40.0

	contentWidth  = pageWidth - 2*pageMargin
	contentRight  = pageWidth - pageMargin
	contentBottom = pageHeight - pageMargin

	// The booklet page puts the roster beside the map, which is how the
	// listing a referee already buys is laid out.
	columnGutter = 16.0
	columnWidth  = (contentWidth - columnGutter) / 2
	rightColumnX = pageMargin + columnWidth + columnGutter
)

// Type sizes and the leading each is set on.
const (
	titleSize   = 16.0
	titleLead   = 22.0
	noteSize    = 8.0
	headingSize = 11.0
	headingLead = 18.0
	bodySize    = 8.5
	bodyLead    = 10.5
	rosterSize  = 7.5
	rosterLead  = 9.5

	sectionGap = 14.0
	blockGap   = 5.0
)

// The type drawn inside a hex is a fraction of the hex rather than a fixed
// size. The booklet now draws three windows -- a sub-sector's eighty
// hexes, a member's eighty and the ring around them, and the sector's
// twelve hundred and eighty -- and a size that suits the first overruns
// its own hex on the last, which is what a sector booklet did.
//
// The fractions are the sizes the sub-sector map has always used divided
// by the side a p. 3 grid fits in the booklet's column, so that map is
// drawn exactly as it was.
const (
	pageThreeSide = columnWidth / (columnStep*starmap.Columns + halfColumn)

	hexNumberOfSide = 4.5 / pageThreeSide
	starportOfSide  = 8.5 / pageThreeSide
	worldNameOfSide = 5.0 / pageThreeSide

	// A member's number is set across its whole band, so it is a multiple
	// of a hex rather than a fraction of one.
	memberNumberOfSide = 3.0
)

// hexNumber, starport and worldName are those fractions of a fitted hex.
func (f mapFit) hexNumber() float64 { return f.Side * hexNumberOfSide }
func (f mapFit) starport() float64  { return f.Side * starportOfSide }
func (f mapFit) worldName() float64 { return f.Side * worldNameOfSide }

// The inks. Grey rather than black for everything the worlds are read
// over: the grid and the routes are the paper a referee writes on, and a
// dense subsector draws a hundred and fifty routes across it.
const (
	inkBlack = 0
	inkNote  = 95
	inkRoute = 125
	inkGrid  = 185
	inkBand  = 215
	white    = 255
)

// The roster's fixed columns. Name takes whatever is left, because it is
// the one column the referee fills in himself and the one whose width the
// page can afford to vary.
const (
	hexColumnWidth    = 28.0
	digitsColumnWidth = 54.0
	basesColumnWidth  = 56.0
	rosterFixedWidth  = hexColumnWidth + digitsColumnWidth + basesColumnWidth
)

// Line weights.
const (
	gridWeight  = 0.4
	routeWeight = 0.6
	ruleWeight  = 0.8
	seamWeight  = 1.6
)

// two is the divisor that halves a measure: a centred string starts half
// its width left of centre, and a hexagon's short radius is half its
// side. Named because the linter reads a bare 2 as a magic number, and
// because "/ two" says what it is doing.
const two = 2.0

// firstEvenColumn is the leftmost of the columns p. 3 draws half a hex
// low, which is where the bottom of a drawn grid is.
const firstEvenColumn = 2

// epoch is the creation and modification date every booklet is stamped
// with. A PDF carries both, and time.Now in either makes two renders of
// one record differ in bytes -- which would mean a referee could not
// check that the file on his disk is the file this record produces.
// TestTheBookletIsDeterministic is what holds this, and it holds it by
// reading the stamped dates rather than by comparing two renders: both
// renders finish inside the same second, and fpdf writes to the second,
// so an unpinned clock passes that comparison.
func epoch() time.Time { return time.Unix(0, 0).UTC() }

// Booklet writes the record as the pages a referee prints: the map beside
// its roster, then the routes, then a block per world.
//
// It renders the same record the Markdown listing renders, from the same
// charts, and it decides nothing the listing does not. The one thing it
// adds is the line p. 2 asks for -- "a line connecting the two worlds on
// the map" -- which a monospace grid has nowhere to put.
func (r *Renderer) Booklet(out io.Writer, record *starmap.Record) error {
	book := &booklet{
		pdf:    newPDF(),
		charts: r.charts,
		record: record,
		drawn:  r.drawn(record.Routes),
		names:  namesByHex(record),
		latin:  windows1252(),
		y:      pageMargin,
	}

	if record.Grid == starmap.SectorGrid() {
		book.sectorPages()
	} else {
		book.firstPage()
		book.routesSection("Routes", record.Routes, book.drawn, nil)
		book.detailsSection("The worlds in detail", record.Worlds)
	}

	err := book.pdf.Output(out)
	if err != nil {
		return fmt.Errorf("writing the booklet: %w", err)
	}

	return nil
}

// newPDF returns a document pinned to reproduce byte for byte.
//
// Compression is off for the same reason: an uncompressed content stream
// is plain text, so the checks that every world and route reaches the page
// can read the drawn strings straight out of the bytes rather than
// bringing in a PDF parser to answer one question.
func newPDF() *fpdf.Fpdf {
	pdf := fpdf.New("P", "pt", "Letter", "")

	pdf.SetCompression(false)
	pdf.SetCatalogSort(true)
	pdf.SetCreationDate(epoch())
	pdf.SetModificationDate(epoch())
	pdf.SetProducer("ctworldgen", false)
	pdf.SetCreator("ctworldgen", false)

	// Every break is taken deliberately below. Automatic ones would push a
	// roster row off the bottom of the column it is laid beside the map
	// in, and land it on a page with no map.
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(pageMargin, pageMargin, pageMargin)

	return pdf
}

// windows1252 is the encoding fpdf's core fonts are indexed by: a PDF
// drawn with Helvetica carries single bytes, not UTF-8, so an em-dash
// handed straight to fpdf is drawn as the three characters its UTF-8
// spells.
//
// One is made per booklet and held on it. The encoding is stateless, and
// every string drawn goes through it -- a sector draws its 1280 hex
// numbers alone -- so building one for each was an allocation per stamp.
func windows1252() *encoding.Encoder { return charmap.Windows1252.NewEncoder() }

// encode writes a string in the alphabet the page can draw.
//
// Almost everything here is ASCII and passes through untouched. Two
// things are not: the em-dash and ellipsis this file sets deliberately,
// which Windows-1252 has characters for, and a referee's own name for a
// world, which is the one field with no alphabet the record can promise.
// A character the encoding cannot carry is written as a question mark
// rather than dropped, so a name stays the length and shape he gave it.
func (b *booklet) encode(text string) string {
	written, err := b.latin.String(text)
	if err == nil {
		return written
	}

	var built strings.Builder

	for _, char := range text {
		one, oneErr := b.latin.String(string(char))
		if oneErr != nil {
			built.WriteByte('?')

			continue
		}

		built.WriteString(one)
	}

	return built.String()
}

// booklet is one document being laid out. y is the pen: every section
// draws from it and leaves it below what it drew.
type booklet struct {
	pdf    *fpdf.Fpdf
	charts *tables.Tables
	record *starmap.Record

	// drawn is the lanes this booklet inks, which is every one the record
	// carries unless the renderer was built for legible lanes (ERRATA
	// E007). The record itself is unchanged.
	drawn []starmap.Route

	names map[starmap.Hex]string
	latin *encoding.Encoder
	y     float64
}

// text draws a string, and width measures one, in the page's own
// alphabet. Every string drawn goes through here: fpdf takes the bytes it
// is given, so a single Text call that skipped this would be the one that
// mojibakes.
func (b *booklet) text(x, y float64, s string) { b.pdf.Text(x, y, b.encode(s)) }

func (b *booklet) width(s string) float64 { return b.pdf.GetStringWidth(b.encode(s)) }

// drawable rewrites a string in the alphabet the page can draw: every rune
// Windows-1252 has no character for becomes a question mark, and the result
// is still UTF-8. A question mark rather than nothing, so a line keeps the
// length and shape the referee gave it -- the same promise encode makes.
//
// This is not only tidiness. fpdf indexes its character widths into a
// 256-entry table by rune, so measuring a rune above 255 panics outright.
// That was unreachable while every wrapped paragraph was a Book 3 table
// label; notes let the referee's own text reach one, and an arrow in a
// note crashed the render.
func (b *booklet) drawable(text string) string {
	_, err := b.latin.String(text)
	if err == nil {
		return text
	}

	var built strings.Builder

	for _, char := range text {
		_, charErr := b.latin.String(string(char))
		if charErr != nil {
			built.WriteByte('?')

			continue
		}

		built.WriteRune(char)
	}

	return built.String()
}

// split wraps a paragraph to a width. It measures the drawable form rather
// than the encoded one, because fpdf indexes its widths by rune -- and
// rather than the string as written, because a rune the page cannot draw
// is a rune fpdf cannot measure.
func (b *booklet) split(s string, width float64) []string {
	return b.pdf.SplitText(b.drawable(s), width)
}

// firstPage is the booklet page: the title, the summary the listing opens
// with, and then the map with as much of the roster as fits beside it.
func (b *booklet) firstPage() {
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

	// The referee's note about the map as a whole, where the listing puts
	// it: under the summary, above everything the tool generated (issue 1
	// #6). body wraps it, and the map below is fitted to whatever height is
	// left rather than to a fixed box, so a long note costs hex size and
	// never runs off the page.
	if b.record.Notes != "" {
		b.body(oneLine(b.record.Notes))
	}

	b.rule(pageMargin, contentRight)

	b.y += blockGap

	b.drawMap(box{X: rightColumnX, Y: b.y, Width: columnWidth, Height: contentBottom - b.y},
		wholeGrid(b.record.Grid), everywhere)
	b.rosterSection(pageMargin, columnWidth, b.y, b.record.Worlds)
}

// drawMap draws the p. 3 grid inside a box: the hexes and their numbers,
// then the routes over them, then the worlds over those.
//
// The order is the order a referee draws it in and the order that stays
// legible: a route crossing a hex must not black out the starport letter
// at either end of it.
func (b *booklet) drawMap(within box, draw window, shows func(starmap.Hex) bool) {
	fit := fitMap(draw, within)

	b.drawHexes(fit, draw, shows, true)
	b.drawRoutes(fit, shows)
	b.drawWorlds(fit, shows)
}

// drawIndex draws the whole sector as an index rather than as a grid to
// read a hex off: the same hexes with no numbers in them, the seams
// between the sixteen members drawn heavy, and each member's number in its
// band (ERRATA E008 part 4).
//
// Thirty-two columns fitted to a page give a hex seventeen points across.
// A four-digit number drawn in one runs out into the hex beside it, which
// is what a sector booklet used to do, and the numbers are on the member
// maps overleaf where there is room for them.
func (b *booklet) drawIndex(within box) {
	draw := wholeGrid(b.record.Grid)
	fit := fitMap(draw, within)

	b.drawHexes(fit, draw, everywhere, false)
	b.drawRoutes(fit, everywhere)
	b.drawWorlds(fit, everywhere)
	b.drawSeams(fit)
	b.drawMemberNumbers(fit)
}

// drawHexes draws the grid a map is laid on, and its numbers where the map
// is one a hex can be read off.
//
// A window may reach past the record's own grid -- a member on the edge of
// the sector has no neighbour on that side -- and those hexes are not
// drawn. The window is still fitted whole, so all sixteen members draw at
// one size.
func (b *booklet) drawHexes(fit mapFit, draw window, shows func(starmap.Hex) bool, numbered bool) {
	b.pdf.SetLineWidth(gridWeight)
	b.pdf.SetDrawColor(inkGrid, inkGrid, inkGrid)
	b.pdf.SetTextColor(inkGrid, inkGrid, inkGrid)
	b.pdf.SetFont("Helvetica", "", fit.hexNumber())

	for col := draw.FromCol; col <= draw.ToCol; col++ {
		for row := draw.FromRow; row <= draw.ToRow; row++ {
			hex := starmap.Hex{Col: col, Row: row}
			if !b.record.Grid.Contains(hex) || !shows(hex) {
				continue
			}

			center := fit.hexCenter(hex)

			b.pdf.Polygon(polygon(hexOutline(center, fit.Side)), "D")

			if !numbered {
				continue
			}

			// The number goes inside the upper-left, which is where p. 3
			// prints it.
			b.text(
				center.X-fit.Side/two+fit.hexNumber()/two,
				center.Y-root3*fit.Side/two+fit.hexNumber()+1,
				hex.String(),
			)
		}
	}
}

// drawSeams draws the three lines down and the three across that divide a
// sector's index into its sixteen members (ERRATA E008 part 4).
//
// The true border between two bands is not straight -- the even-numbered
// columns sit half a hex low, so it steps -- and the index draws it
// straight, because it is dividing an index and not tracing a hex edge.
func (b *booklet) drawSeams(fit mapFit) {
	b.pdf.SetLineWidth(seamWeight)
	b.pdf.SetDrawColor(inkBlack, inkBlack, inkBlack)

	top := fit.hexCenter(starmap.Hex{Col: 1, Row: 1}).Y - root3*fit.Side/two
	// Measured on an even column, which is the one drawn half a hex
	// low and so reaches furthest down the sheet.
	bottom := fit.hexCenter(starmap.Hex{Col: firstEvenColumn, Row: b.record.Grid.Rows}).Y + root3*fit.Side/two
	left := fit.hexCenter(starmap.Hex{Col: 1, Row: 1}).X - fit.Side
	right := fit.hexCenter(starmap.Hex{Col: b.record.Grid.Columns, Row: 1}).X + fit.Side

	for band := 1; band < starmap.SectorAcross; band++ {
		// Halfway between the last column of one band and the first of the
		// next.
		last := fit.hexCenter(starmap.Hex{Col: band * starmap.Columns, Row: 1}).X
		next := fit.hexCenter(starmap.Hex{Col: band*starmap.Columns + 1, Row: 1}).X
		b.pdf.Line((last+next)/two, top, (last+next)/two, bottom)

		// And halfway between the last row of one band and the first of
		// the next, measured on an odd column so the half-hex step of the
		// even ones does not enter it.
		lastRow := fit.hexCenter(starmap.Hex{Col: 1, Row: band * starmap.Rows}).Y
		nextRow := fit.hexCenter(starmap.Hex{Col: 1, Row: band*starmap.Rows + 1}).Y
		b.pdf.Line(left, (lastRow+nextRow)/two, right, (lastRow+nextRow)/two)
	}
}

// drawMemberNumbers writes each member's number across its band of the
// index, so the page says which sub-sector is which (ERRATA E008 part 1).
//
// Set large and very pale, under nothing and over nothing: the index is
// read to choose a sub-sector, and the number is what the section overleaf
// is headed with. Without it the sixteen bands are indistinguishable, and
// an index laid out in columns instead of rows would look perfectly well.
func (b *booklet) drawMemberNumbers(fit mapFit) {
	b.pdf.SetFont("Helvetica", "B", fit.Side*memberNumberOfSide)
	b.pdf.SetTextColor(inkBand, inkBand, inkBand)

	for index := range starmap.Members {
		first, last := starmap.MemberBounds(index)

		at := fit.hexCenter(first)
		to := fit.hexCenter(last)

		number := strconv.Itoa(index)
		b.text((at.X+to.X)/two-b.width(number)/two, (at.Y+to.Y)/two, number)
	}
}

// drawRoutes draws p. 2's line between the two worlds a route joins. The
// text map draws none and says so; this is the one thing the drawn map
// exists for.
func (b *booklet) drawRoutes(fit mapFit, shows func(starmap.Hex) bool) {
	b.pdf.SetLineWidth(routeWeight)
	b.pdf.SetDrawColor(inkRoute, inkRoute, inkRoute)

	for _, route := range b.drawn {
		// Both ends must be on this map. A lane can be four parsecs long
		// (p. 2) and a member's map reaches one hex past its own border,
		// so a long lane out of the sub-sector has no far end to draw to.
		// It is not hidden: it is in the member's own lane table with the
		// sub-sector it leaves for named, and the map's note counts the
		// ones it could not draw (ERRATA E008 part 2).
		if !shows(route.From) || !shows(route.To) {
			continue
		}

		from := fit.hexCenter(route.From)
		to := fit.hexCenter(route.To)

		b.pdf.Line(from.X, from.Y, to.X, to.Y)
	}
}

// drawWorlds marks each world's hex with the letter of its starport,
// which is what p. 1 says to mark it with, and with the name underneath
// where the referee has written one in.
func (b *booklet) drawWorlds(fit mapFit, shows func(starmap.Hex) bool) {
	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)

	for _, world := range b.record.Worlds {
		if !shows(world.Hex) {
			continue
		}

		center := fit.hexCenter(world.Hex)

		b.pdf.SetFont("Helvetica", "B", fit.starport())

		letter := world.Starport.String()
		b.text(center.X-b.width(letter)/two, center.Y+fit.starport()/two/two, letter)

		name, named := b.names[world.Hex]
		if !named {
			continue
		}

		// Trimmed to the column the hex sits in. A referee's own name has
		// no length the record can promise, and one drawn whole runs
		// across the hexes beside it and off the edge of the map.
		b.pdf.SetFont("Helvetica", "", fit.worldName())

		name = b.clip(name, columnStep*fit.Side)
		b.text(center.X-b.width(name)/two, center.Y+fit.Side/two, name)
	}
}

// polygon converts a hex outline to the points fpdf draws.
func polygon(outline [hexSides]Point) []fpdf.PointType {
	points := make([]fpdf.PointType, 0, hexSides)
	for _, point := range outline {
		points = append(points, fpdf.PointType{X: point.X, Y: point.Y})
	}

	return points
}

// rosterSection writes the world roster, beginning beside the map and
// continuing full width overleaf for as many worlds as it takes. A
// subsector of eighty worlds is a legal record and a sector carries
// hundreds, so the roster's length is not something the page can assume.
func (b *booklet) rosterSection(left, width, top float64, worlds []starmap.World) {
	b.y = top

	if len(worlds) == 0 {
		b.pdf.SetFont("Helvetica", "", bodySize)
		b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)
		b.text(left, b.y+bodySize, noWorlds)

		b.y += bodyLead

		return
	}

	b.rosterHead(left, width)

	for _, world := range worlds {
		if b.y+rosterLead > contentBottom {
			b.newPage()

			// Past the map, the roster has the whole width to itself.
			left, width = pageMargin, contentWidth

			b.rosterHead(left, width)
		}

		b.rosterRow(left, width, world)
	}
}

// rosterHead writes the column headings and the rule beneath them.
func (b *booklet) rosterHead(left, width float64) {
	b.pdf.SetFont("Helvetica", "B", rosterSize)
	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)

	b.rosterCells(left, width, "Hex", "Name", "Digits", "Bases")

	b.y += blockGap

	b.rule(left, left+width)

	b.y += blockGap / two
}

// rosterRow writes one world.
func (b *booklet) rosterRow(left, width float64, world starmap.World) {
	b.pdf.SetFont("Helvetica", "", rosterSize)
	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)

	b.rosterCells(left, width, world.Hex.String(), b.names[world.Hex], world.Digits, bases(world))
}

// rosterCells lays one line of the roster across its four columns and
// leaves the pen on the next.
func (b *booklet) rosterCells(left, width float64, hex, name, digits, baseList string) {
	nameWidth := width - rosterFixedWidth
	baseline := b.y + rosterSize

	b.text(left, baseline, hex)
	b.text(left+hexColumnWidth, baseline, b.clip(name, nameWidth))
	b.text(left+hexColumnWidth+nameWidth, baseline, digits)
	b.text(left+hexColumnWidth+nameWidth+digitsColumnWidth, baseline, b.clip(baseList, basesColumnWidth))

	b.y += rosterLead
}

// clip trims a value to the column it is written in. A referee's own name
// for a world is the one field with no length the record can promise, and
// a long one would otherwise run through the column beside it.
func (b *booklet) clip(value string, width float64) string {
	if value == "" || b.width(value) <= width {
		return value
	}

	const ellipsis = "…"

	// A rune at a time, not a byte. Cutting inside a multi-byte character
	// leaves invalid UTF-8, which encode reads as the replacement rune and
	// draws as a question mark -- so trimming a name would silently
	// destroy the character it stopped on, which is the one thing encode
	// promises it will not do.
	for value != "" {
		_, last := utf8.DecodeLastRuneInString(value)

		value = value[:len(value)-last]

		if b.width(value+ellipsis) <= width {
			return value + ellipsis
		}
	}

	return ""
}

// routesSection writes the route table, which carries every route whether or
// not the map above it drew a line a reader can follow.
func (b *booklet) routesSection(
	heading string, carried, drawn []starmap.Route, into func(starmap.Route) string,
) {
	b.heading(heading)

	if len(carried) == 0 {
		b.body("No route was drawn.")

		return
	}

	if suppressed := len(carried) - len(drawn); suppressed > 0 {
		b.body(lanesNote(len(carried), suppressed))
	}

	b.routeHead(into != nil)

	for _, route := range drawn {
		// The head goes on every page the table reaches, as the roster's
		// does. A subsector at DM +1 carries a hundred and fifty routes and
		// a sector carries hundreds, so a table without this prints pages
		// of unlabelled columns.
		if b.y+rosterLead > contentBottom {
			b.newPage()
			b.routeHead(into != nil)
		}

		leaves := ""
		if into != nil {
			leaves = into(route)
		}

		b.pdf.SetFont("Helvetica", "", rosterSize)
		b.routeCells(named(b.names, route.From), named(b.names, route.To), fmt.Sprint(route.Distance), leaves)
	}
}

// routeHead writes the route table's column headings and the rule beneath
// them, as rosterHead does for the roster.
func (b *booklet) routeHead(leaving bool) {
	into := ""
	if leaving {
		into = "Into"
	}

	b.pdf.SetFont("Helvetica", "B", rosterSize)
	b.routeCells("From", "To", "Parsecs", into)

	b.y += blockGap
	b.rule(pageMargin, contentRight)

	b.y += blockGap / two
}

// routeColumnWidth is what one end of a route is written in: a hex and the
// name the referee gave it.
const (
	routeColumnWidth   = 200.0
	parsecsColumnWidth = 44.0
)

func (b *booklet) routeCells(from, to, parsecs, into string) {
	baseline := b.y + rosterSize

	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)
	b.text(pageMargin, baseline, b.clip(from, routeColumnWidth))
	b.text(pageMargin+routeColumnWidth, baseline, b.clip(to, routeColumnWidth))
	b.text(pageMargin+routeColumnWidth*two, baseline, parsecs)

	if into != "" {
		b.text(pageMargin+routeColumnWidth*two+parsecsColumnWidth, baseline, into)
	}

	b.y += rosterLead
}

// detailsSection writes a block per world, carrying exactly the lines the
// Markdown listing carries.
func (b *booklet) detailsSection(heading string, worlds []starmap.World) {
	if len(worlds) == 0 {
		return
	}

	b.heading(heading)
	b.body(techIndexNote)

	for _, world := range worlds {
		b.worldBlock(world)
	}
}

func (b *booklet) worldBlock(world starmap.World) {
	lines := bullets(b.charts, world)

	// The heading and its first line stay together: a world whose name is
	// the last thing on a page and whose starport is the first thing on
	// the next is harder to read than one moved whole.
	b.ensure(bodyLead*two + blockGap)

	b.pdf.SetFont("Helvetica", "B", bodySize)
	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)
	b.text(pageMargin, b.y+bodySize, named(b.names, world.Hex)+" — "+world.Digits)

	b.y += bodyLead

	for _, line := range lines {
		b.bulletLine(line)
	}

	b.y += blockGap
}

// bulletIndent is where a world's lines are set, and how far the label
// holds the description off the margin.
const (
	bulletIndent = 12.0
	labelGap     = 3.0
)

func (b *booklet) bulletLine(line bullet) {
	b.pdf.SetFont("Helvetica", "B", bodySize)

	labelWidth := b.width(line.label) + labelGap
	textX := pageMargin + bulletIndent + labelWidth

	b.pdf.SetFont("Helvetica", "", bodySize)

	wrapped := b.split(line.description, contentRight-textX)

	if line.description == "" {
		wrapped = nil
	}

	b.ensure(bodyLead * float64(max(len(wrapped), 1)))

	b.pdf.SetFont("Helvetica", "B", bodySize)
	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)
	b.text(pageMargin+bulletIndent, b.y+bodySize, line.label)

	if len(wrapped) == 0 {
		b.y += bodyLead

		return
	}

	b.pdf.SetFont("Helvetica", "", bodySize)

	for _, text := range wrapped {
		b.text(textX, b.y+bodySize, text)

		b.y += bodyLead
	}
}

// heading opens a section on a page of its own, because both sections
// that use it run to a page or more on any record with worlds in it.
func (b *booklet) heading(text string) {
	b.newPage()

	b.pdf.SetFont("Helvetica", "B", headingSize)
	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)
	b.text(pageMargin, b.y+headingSize, text)

	b.y += headingLead

	b.rule(pageMargin, contentRight)

	b.y += blockGap
}

// body writes a paragraph across the content width.
func (b *booklet) body(text string) {
	b.pdf.SetFont("Helvetica", "", bodySize)
	b.pdf.SetTextColor(inkNote, inkNote, inkNote)

	for _, line := range b.split(strings.TrimSpace(text), contentWidth) {
		b.ensure(bodyLead)
		b.text(pageMargin, b.y+bodySize, line)

		b.y += bodyLead
	}

	b.y += blockGap
}

// rule draws the hairline a heading or a table head sits above.
func (b *booklet) rule(from, to float64) {
	b.pdf.SetLineWidth(ruleWeight)
	b.pdf.SetDrawColor(inkBlack, inkBlack, inkBlack)
	b.pdf.Line(from, b.y, to, b.y)
}

// ensure starts a new page when what is about to be drawn would not fit
// on this one.
func (b *booklet) ensure(height float64) {
	if b.y+height > contentBottom {
		b.newPage()
	}
}

// newPage starts a page with the pen at its top margin.
func (b *booklet) newPage() {
	b.pdf.AddPage()
	b.pdf.SetFillColor(white, white, white)

	b.y = pageMargin
}

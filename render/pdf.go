package render

import (
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-pdf/fpdf"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"

	"github.com/philoserf/ctworldgen/subsector"
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
	titleSize     = 16.0
	titleLead     = 22.0
	noteSize      = 8.0
	headingSize   = 11.0
	headingLead   = 18.0
	bodySize      = 8.5
	bodyLead      = 10.5
	rosterSize    = 7.5
	rosterLead    = 9.5
	hexNumberSize = 4.5
	starportSize  = 8.5
	worldNameSize = 5.0

	sectionGap = 14.0
	blockGap   = 5.0
)

// The inks. Grey rather than black for everything the worlds are read
// over: the grid and the lanes are the paper a referee writes on, and a
// dense subsector draws a hundred and fifty lanes across it.
const (
	inkBlack = 0
	inkNote  = 95
	inkLane  = 125
	inkGrid  = 185
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
	gridWeight = 0.4
	laneWeight = 0.6
	ruleWeight = 0.8
)

// two is the divisor that halves a measure: a centred string starts half
// its width left of centre, and a hexagon's short radius is half its
// side. Named because the linter reads a bare 2 as a magic number, and
// because "/ two" says what it is doing.
const two = 2.0

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
// its roster, then the space lanes, then a block per world.
//
// It renders the same record the Markdown listing renders, from the same
// charts, and it decides nothing the listing does not. The one thing it
// adds is the line p. 2 asks for -- "a line connecting the two worlds on
// the map" -- which a monospace grid has nowhere to put.
func (r *Renderer) Booklet(out io.Writer, record *subsector.Subsector) error {
	book := &booklet{
		pdf:    newPDF(),
		charts: r.charts,
		record: record,
		names:  namesByHex(record),
		latin:  windows1252(),
		y:      pageMargin,
	}

	book.firstPage()
	book.lanesSection()
	book.detailsSection()

	err := book.pdf.Output(out)
	if err != nil {
		return fmt.Errorf("writing the booklet: %w", err)
	}

	return nil
}

// newPDF returns a document pinned to reproduce byte for byte.
//
// Compression is off for the same reason: an uncompressed content stream
// is plain text, so the checks that every world and lane reaches the page
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
	record *subsector.Subsector
	names  map[subsector.Hex]string
	latin  *encoding.Encoder
	y      float64
}

// text draws a string, and width measures one, in the page's own
// alphabet. Every string drawn goes through here: fpdf takes the bytes it
// is given, so a single Text call that skipped this would be the one that
// mojibakes.
func (b *booklet) text(x, y float64, s string) { b.pdf.Text(x, y, b.encode(s)) }

func (b *booklet) width(s string) float64 { return b.pdf.GetStringWidth(b.encode(s)) }

// split wraps a paragraph to a width. It measures the string as it was
// written rather than as it is encoded, because fpdf indexes its widths
// by rune and Windows-1252 agrees with Unicode over the range a
// description reaches.
func (b *booklet) split(s string, width float64) []string {
	return b.pdf.SplitText(s, width)
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
	b.text(pageMargin, b.y+noteSize, summary(b.record))

	b.y += sectionGap

	b.rule(pageMargin, contentRight)

	b.y += blockGap

	// A sector is thirty-two columns across (ERRATA E006). Fitted into
	// half the page it draws hexes too small to carry a starport letter,
	// so it takes the whole width and the roster begins overleaf.
	beside := b.record.Grid == subsector.PageThreeGrid()

	rosterX, rosterWidth, rosterTop := pageMargin, columnWidth, b.y

	if beside {
		b.drawMap(box{X: rightColumnX, Y: b.y, Width: columnWidth, Height: contentBottom - b.y})
	} else {
		b.drawMap(box{X: pageMargin, Y: b.y, Width: contentWidth, Height: contentBottom - b.y})
		b.newPage()

		rosterX, rosterWidth, rosterTop = pageMargin, contentWidth, b.y
	}

	b.rosterSection(rosterX, rosterWidth, rosterTop)
}

// summary is the sentence the Markdown listing opens with, so that the
// two documents report the same record in the same words.
func summary(record *subsector.Subsector) string {
	return fmt.Sprintf("%d worlds, %d space lanes. Generated from seed %d at occurrence DM %s.",
		len(record.Worlds), len(record.Routes), record.Seed, occurrenceDM(record.OccurrenceDM))
}

// drawMap draws the p. 3 grid inside a box: the hexes and their numbers,
// then the lanes over them, then the worlds over those.
//
// The order is the order a referee draws it in and the order that stays
// legible: a lane crossing a hex must not black out the starport letter
// at either end of it.
func (b *booklet) drawMap(within box) {
	fit := fitMap(b.record.Grid, within)

	b.pdf.SetLineWidth(gridWeight)
	b.pdf.SetDrawColor(inkGrid, inkGrid, inkGrid)
	b.pdf.SetTextColor(inkGrid, inkGrid, inkGrid)
	b.pdf.SetFont("Helvetica", "", hexNumberSize)

	for col := 1; col <= b.record.Grid.Columns; col++ {
		for row := 1; row <= b.record.Grid.Rows; row++ {
			hex := subsector.Hex{Col: col, Row: row}
			center := fit.hexCenter(hex)

			b.pdf.Polygon(polygon(hexOutline(center, fit.Side)), "D")

			// The number goes inside the upper-left, which is where p. 3
			// prints it.
			b.text(
				center.X-fit.Side/two+hexNumberSize/two,
				center.Y-root3*fit.Side/two+hexNumberSize+1,
				hex.String(),
			)
		}
	}

	b.drawLanes(fit)
	b.drawWorlds(fit)
}

// drawLanes draws p. 2's line between the two worlds a lane joins. The
// text map draws none and says so; this is the one thing the drawn map
// exists for.
func (b *booklet) drawLanes(fit mapFit) {
	b.pdf.SetLineWidth(laneWeight)
	b.pdf.SetDrawColor(inkLane, inkLane, inkLane)

	for _, route := range b.record.Routes {
		from := fit.hexCenter(route.From)
		to := fit.hexCenter(route.To)

		b.pdf.Line(from.X, from.Y, to.X, to.Y)
	}
}

// drawWorlds marks each world's hex with the letter of its starport,
// which is what p. 1 says to mark it with, and with the name underneath
// where the referee has written one in.
func (b *booklet) drawWorlds(fit mapFit) {
	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)

	for _, world := range b.record.Worlds {
		center := fit.hexCenter(world.Hex)

		b.pdf.SetFont("Helvetica", "B", starportSize)

		letter := world.Starport.String()
		b.text(center.X-b.width(letter)/two, center.Y+starportSize/two/two, letter)

		name, named := b.names[world.Hex]
		if !named {
			continue
		}

		// Trimmed to the column the hex sits in. A referee's own name has
		// no length the record can promise, and one drawn whole runs
		// across the hexes beside it and off the edge of the map.
		b.pdf.SetFont("Helvetica", "", worldNameSize)

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
func (b *booklet) rosterSection(left, width, top float64) {
	b.y = top

	if len(b.record.Worlds) == 0 {
		b.pdf.SetFont("Helvetica", "", bodySize)
		b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)
		b.text(left, b.y+bodySize, "No world was placed. An empty subsector is a result.")

		b.y += bodyLead

		return
	}

	b.rosterHead(left, width)

	for _, world := range b.record.Worlds {
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
func (b *booklet) rosterRow(left, width float64, world subsector.World) {
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

// lanesSection writes the lane table, which carries every lane whether or
// not the map above it drew a line a reader can follow.
func (b *booklet) lanesSection() {
	b.heading("Space lanes")

	if len(b.record.Routes) == 0 {
		b.body("No space lane was drawn.")

		return
	}

	b.laneHead()

	for _, route := range b.record.Routes {
		// The head goes on every page the table reaches, as the roster's
		// does. A subsector at DM +1 carries a hundred and fifty lanes and
		// a sector carries hundreds, so a table without this prints pages
		// of unlabelled columns.
		if b.y+rosterLead > contentBottom {
			b.newPage()
			b.laneHead()
		}

		b.pdf.SetFont("Helvetica", "", rosterSize)
		b.laneCells(named(b.names, route.From), named(b.names, route.To), fmt.Sprint(route.Distance))
	}
}

// laneHead writes the lane table's column headings and the rule beneath
// them, as rosterHead does for the roster.
func (b *booklet) laneHead() {
	b.pdf.SetFont("Helvetica", "B", rosterSize)
	b.laneCells("From", "To", "Parsecs")

	b.y += blockGap
	b.rule(pageMargin, contentRight)

	b.y += blockGap / two
}

// laneColumnWidth is what one end of a lane is written in: a hex and the
// name the referee gave it.
const laneColumnWidth = 200.0

func (b *booklet) laneCells(from, to, parsecs string) {
	baseline := b.y + rosterSize

	b.pdf.SetTextColor(inkBlack, inkBlack, inkBlack)
	b.text(pageMargin, baseline, b.clip(from, laneColumnWidth))
	b.text(pageMargin+laneColumnWidth, baseline, b.clip(to, laneColumnWidth))
	b.text(pageMargin+laneColumnWidth*two, baseline, parsecs)

	b.y += rosterLead
}

// detailsSection writes a block per world, carrying exactly the lines the
// Markdown listing carries.
func (b *booklet) detailsSection() {
	if len(b.record.Worlds) == 0 {
		return
	}

	b.heading("The worlds in detail")
	b.body(techIndexNote)

	for _, world := range b.record.Worlds {
		b.worldBlock(world)
	}
}

// bullet is one line of a world's block: the label the listing sets in
// bold, and the description the charts give for it.
type bullet struct {
	label       string
	description string
}

func (b *booklet) worldBlock(world subsector.World) {
	lines := b.bullets(world)

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

// bullets is the world's lines, in the order the p. 4 Planetary
// Characteristics box lists them, which is the order the Markdown listing
// writes them in.
func (b *booklet) bullets(world subsector.World) []bullet {
	starport := "no chart row"

	row, err := b.charts.StarportChart.Row(world.Starport)
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
		{"Size", world.Size, b.charts.Size},
		{"Atmosphere", world.Atmosphere, b.charts.Atmosphere},
		{"Hydrographics", world.Hydrographics, b.charts.Hydrographics},
		{"Population", world.Population, b.charts.Population},
		{"Government", world.Government, b.charts.Government},
		{"Law level", world.LawLevel, b.charts.LawLevels},
	} {
		lines = append(lines, bullet{
			label:       fmt.Sprintf("%s %s.", line.name, digit(line.value)),
			description: strings.TrimSpace(described(line.labels, line.value)),
		})
	}

	// No description: techIndexNote said once, at the head of the section,
	// why pp. 10-11 supply none.
	lines = append(lines,
		bullet{label: fmt.Sprintf("Technological index %s.", digit(world.TechIndex)), description: ""},
		bullet{label: "Bases.", description: bases(world)},
	)

	for _, clamp := range world.Clamps {
		lines = append(lines, bullet{
			label: "Clamped.",
			description: fmt.Sprintf("%s threw %d and is recorded as %d.",
				clamp.Characteristic, clamp.Raw, clamp.Value),
		})
	}

	return lines
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

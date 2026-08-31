package render_test

import (
	"bytes"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/philoserf/ctworldgen/gen"
	"github.com/philoserf/ctworldgen/internal/fixture"
	"github.com/philoserf/ctworldgen/render"
	"github.com/philoserf/ctworldgen/subsector"
)

// The booklet is checked by reading the drawing back out of the PDF and
// measuring it, rather than by comparing bytes against a golden.
//
// Two reasons, and the second is the one that matters. A golden of a
// binary defeats the point of `task regenerate`, which exists so that a
// human reads the diff. And a golden is regenerated from the code under
// test, so a map drawn with its parity inverted would simply produce a
// new golden -- which is exactly the trap CLAUDE.md records the drawn
// grid falling into.
//
// So the geometry below is derived from the drawing itself: where the
// hexes were actually put on the page, never from the constants the
// renderer put them there with. A check fed those constants agrees with a
// map drawn upside down.

// stamp is one string drawn on a page, at the place the content stream
// put it. Coordinates are PDF user space -- x to the right, **y upward**,
// which is the opposite of the way the page is laid out. Keeping the
// stream's own direction here means no assertion below needs the page
// height to convert, and the one direction that matters is anchored
// explicitly by TestTheDrawnMapIsTheGridPrintedOnPageThree.
type stamp struct {
	X, Y float64
	Text string
}

// segment is one straight line drawn on a page, in the same coordinates.
type segment struct {
	FromX, FromY float64
	ToX, ToY     float64
}

func drawn(t *testing.T, record *subsector.Subsector) []byte {
	t.Helper()

	renderer, err := render.New()
	if err != nil {
		t.Fatalf("building the renderer: %v", err)
	}

	var out bytes.Buffer

	err = renderer.Booklet(&out, record)
	if err != nil {
		t.Fatalf("rendering the booklet: %v", err)
	}

	return out.Bytes()
}

var (
	objectPattern   = regexp.MustCompile(`(?s)(\d+) 0 obj\n(.*?)\nendobj`)
	kidsPattern     = regexp.MustCompile(`/Kids \[([^\]]*)\]`)
	kidPattern      = regexp.MustCompile(`(\d+) 0 R`)
	contentsPattern = regexp.MustCompile(`/Contents (\d+) 0 R`)
	streamPattern   = regexp.MustCompile(`(?s)stream\n(.*)\nendstream`)

	textPattern = regexp.MustCompile(`BT (-?[\d.]+) (-?[\d.]+) Td \(((?:[^()\\]|\\.)*)\) Tj ET`)
	linePattern = regexp.MustCompile(`(-?[\d.]+) (-?[\d.]+) m (-?[\d.]+) (-?[\d.]+) l S`)

	// The dates a PDF carries, and the format fpdf writes them in.
	datePattern = regexp.MustCompile(`/(CreationDate|ModDate) \(D:(\d{14})\)`)
)

// pages returns each page's content stream, in the order the document's
// page tree lists them.
//
// The streams are plain text because Booklet turns compression off. That
// is what lets these checks read the drawing without a PDF parser, and it
// is the reason SetCompression(false) is not merely a convenience.
func pages(t *testing.T, doc []byte) []string {
	t.Helper()

	bodies := map[string]string{}

	for _, match := range objectPattern.FindAllStringSubmatch(string(doc), -1) {
		bodies[match[1]] = match[2]
	}

	var tree string

	for _, body := range bodies {
		if strings.Contains(body, "/Type /Pages") {
			tree = body
		}
	}

	kids := kidsPattern.FindStringSubmatch(tree)
	if kids == nil {
		t.Fatal("the document has no page tree")
	}

	streams := make([]string, 0, len(kidPattern.FindAllStringSubmatch(kids[1], -1)))

	for _, kid := range kidPattern.FindAllStringSubmatch(kids[1], -1) {
		contents := contentsPattern.FindStringSubmatch(bodies[kid[1]])
		if contents == nil {
			t.Fatalf("page object %s has no contents", kid[1])
		}

		stream := streamPattern.FindStringSubmatch(bodies[contents[1]])
		if stream == nil {
			t.Fatalf("content object %s is not a stream", contents[1])
		}

		streams = append(streams, stream[1])
	}

	if len(streams) == 0 {
		t.Fatal("the document has no pages")
	}

	return streams
}

func stamps(t *testing.T, stream string) []stamp {
	t.Helper()

	matches := textPattern.FindAllStringSubmatch(stream, -1)
	found := make([]stamp, 0, len(matches))

	for _, match := range matches {
		found = append(found, stamp{X: number(t, match[1]), Y: number(t, match[2]), Text: match[3]})
	}

	return found
}

func segments(t *testing.T, stream string) []segment {
	t.Helper()

	matches := linePattern.FindAllStringSubmatch(stream, -1)
	found := make([]segment, 0, len(matches))

	for _, match := range matches {
		found = append(found, segment{
			FromX: number(t, match[1]), FromY: number(t, match[2]),
			ToX: number(t, match[3]), ToY: number(t, match[4]),
		})
	}

	return found
}

func number(t *testing.T, text string) float64 {
	t.Helper()

	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		t.Fatalf("reading %q from the content stream: %v", text, err)
	}

	return value
}

// hexLabel is the four-digit grid number every hex of the map carries, in
// the upper-left, as p. 3 prints it.
var hexLabel = regexp.MustCompile(`^\d{4}$`)

// mapRegion is where on page one the map is drawn, which is the right
// column for a subsector and the whole width for a sector. It is a fact
// about the page rather than a geometry constant: if it were wrong, the
// count assertion in mapLabels would fail rather than quietly reading the
// roster's hex column as if it were the map.
func mapRegion(record *subsector.Subsector) float64 {
	if record.Grid == subsector.PageThreeGrid() {
		return 300
	}

	return 0
}

// mapLabels reads back where each hex of the grid was actually drawn.
//
// Every hex of the grid must be found -- not every hex with a world in
// it. That count is what makes the region filter safe: picking up the
// roster's hex column instead would find only the worlds, and picking up
// both would find too many.
func mapLabels(t *testing.T, record *subsector.Subsector, page string) map[subsector.Hex]stamp {
	t.Helper()

	left := mapRegion(record)
	places := map[subsector.Hex]stamp{}

	for _, found := range stamps(t, page) {
		if found.X < left || !hexLabel.MatchString(found.Text) {
			continue
		}

		hex, err := subsector.ParseHex(found.Text)
		if err != nil {
			t.Fatalf("the map drew %q, which is not a hex: %v", found.Text, err)
		}

		if _, twice := places[hex]; twice {
			t.Fatalf("the map drew hex %s twice", hex)
		}

		places[hex] = found
	}

	want := record.Grid.Columns * record.Grid.Rows
	if len(places) != want {
		t.Fatalf("the map drew %d hexes of a %dx%d grid; want %d",
			len(places), record.Grid.Columns, record.Grid.Rows, want)
	}

	return places
}

// near reports whether two measures agree to within a hairline, which is
// as exactly as a content stream's two decimal places can state them.
func near(a, b float64) bool { return math.Abs(a-b) < 0.05 }

// TestTheDrawnMapIsTheGridPrintedOnPageThree is the parity check the drawn
// map needs, and it is the trap CLAUDE.md names twice: draw the
// odd-numbered columns low instead of the even ones and every hex still
// lands in a tidy grid, while half the map disagrees with Hex.Distance.
//
// Three things are asserted, and the first two are what pin the
// direction. Adjacency alone is symmetric under a flip, so it cannot tell
// the printed grid from its mirror.
func TestTheDrawnMapIsTheGridPrintedOnPageThree(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			record := generated(t, golden)
			places := mapLabels(t, record, pages(t, drawn(t, record))[0])

			label := func(col, row int) stamp { return places[hexOf(t, col, row)] }

			// 1. A later row is drawn further down the page. Content stream
			// y runs upward, so down the page is a smaller y.
			if label(1, 2).Y >= label(1, 1).Y {
				t.Fatalf("row 2 of column 1 is not below row 1: %v then %v", label(1, 1), label(1, 2))
			}

			// 2. The even-numbered columns sit half a hex low, which is how
			// p. 3 prints the grid. This is the assertion that a mirrored
			// map fails and nothing else does.
			step := label(1, 1).Y - label(1, 2).Y
			drop := label(1, 1).Y - label(2, 1).Y

			if !near(drop, step/2) {
				t.Fatalf("column 2 sits %.2f below column 1; a row step is %.2f, so it should be %.2f",
					drop, step, step/2)
			}

			// 3. Every hex the drawing puts against another is a hex
			// Hex.Distance calls one parsec away, and every hex it does not
			// is not.
			assertDrawnNeighboursAreOneParsec(t, places)
		})
	}
}

// assertDrawnNeighboursAreOneParsec holds the drawing against the
// page-anchored distance, over the whole grid.
//
// The drawn neighbours are found by measure -- the six places a hex's own
// spacing puts against it -- rather than by asking the renderer where it
// thinks they are.
func assertDrawnNeighboursAreOneParsec(t *testing.T, places map[subsector.Hex]stamp) {
	t.Helper()

	first := places[hexOf(t, 1, 1)]
	rowStep := first.Y - places[hexOf(t, 1, 2)].Y
	colStep := places[hexOf(t, 2, 1)].X - first.X

	// A neighbour's centre is one row step away, or one column step across
	// and half a row step up or down. Anything further is not touching.
	const reach = 1.2

	for hex, place := range places {
		for other, against := range places {
			if hex == other {
				continue
			}

			apart := math.Hypot(place.X-against.X, place.Y-against.Y)
			touching := apart < reach*math.Hypot(colStep, rowStep/2)
			oneParsec := hex.Distance(other) == 1

			if touching != oneParsec {
				t.Fatalf("%s and %s: the drawing puts them %.2f apart (touching=%v) and Hex.Distance says %d parsecs",
					hex, other, apart, touching, hex.Distance(other))
			}
		}
	}
}

// nearest returns the hex whose drawn label is closest to a place, which
// is how a starport letter and a lane end are attributed back to a hex
// without asking the renderer where it drew them.
func nearest(places map[subsector.Hex]stamp, atX, atY float64) subsector.Hex {
	var (
		found subsector.Hex
		best  = math.Inf(1)
	)

	for hex, place := range places {
		apart := math.Hypot(place.X-atX, place.Y-atY)
		if apart < best {
			best, found = apart, hex
		}
	}

	return found
}

// TestEveryWorldIsDrawnInItsOwnHex: p. 1 says to mark a world's hex with
// the letter of its starport, and the letter has to land in that hex.
func TestEveryWorldIsDrawnInItsOwnHex(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			record := generated(t, golden)
			page := pages(t, drawn(t, record))[0]
			places := mapLabels(t, record, page)
			left := mapRegion(record)

			marked := map[subsector.Hex]string{}

			for _, found := range stamps(t, page) {
				if found.X < left || len(found.Text) != 1 {
					continue
				}

				hex := nearest(places, found.X, found.Y)
				if was, twice := marked[hex]; twice {
					t.Fatalf("hex %s was marked twice, %q and %q", hex, was, found.Text)
				}

				marked[hex] = found.Text
			}

			if len(marked) != len(record.Worlds) {
				t.Fatalf("the map marked %d hexes; the record has %d worlds", len(marked), len(record.Worlds))
			}

			for _, world := range record.Worlds {
				if marked[world.Hex] != world.Starport.String() {
					t.Fatalf("hex %s carries a %s starport and the map drew %q",
						world.Hex, world.Starport, marked[world.Hex])
				}
			}
		})
	}
}

// TestEveryLaneIsDrawn: p. 2 asks for "a line connecting the two worlds
// on the map", which the Markdown listing says outright that it draws
// none of. This is the one thing the drawn map exists for.
func TestEveryLaneIsDrawn(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			record := generated(t, golden)
			doc := drawn(t, record)
			page := pages(t, doc)[0]
			places := mapLabels(t, record, page)
			left := mapRegion(record)

			joined := map[string]int{}

			for _, line := range segments(t, page) {
				// The hairlines under a heading and a table head are drawn
				// the same way and start left of the map.
				if line.FromX < left || line.ToX < left {
					continue
				}

				from := nearest(places, line.FromX, line.FromY)
				until := nearest(places, line.ToX, line.ToY)

				if until.Less(from) {
					from, until = until, from
				}

				joined[from.String()+"-"+until.String()]++
			}

			for _, route := range record.Routes {
				key := route.From.String() + "-" + route.To.String()
				if joined[key] == 0 {
					t.Fatalf("the lane %s to %s was not drawn", route.From, route.To)
				}

				joined[key]--
				if joined[key] == 0 {
					delete(joined, key)
				}
			}

			if len(joined) != 0 {
				t.Fatalf("the map drew lines no lane accounts for: %v", joined)
			}
		})
	}
}

// TestEveryWorldAndLaneReachesTheBooklet is the check a drawing cannot
// make: the map marks a world's hex and the roster names it, and dropping
// half the roster would leave the map untouched.
func TestEveryWorldAndLaneReachesTheBooklet(t *testing.T) {
	t.Parallel()

	for _, golden := range fixture.Goldens() {
		t.Run(golden.File, func(t *testing.T) {
			t.Parallel()

			record := generated(t, golden)
			doc := drawn(t, record)

			sheets := pages(t, doc)
			written := make([]stamp, 0, len(sheets))

			for _, page := range sheets {
				written = append(written, stamps(t, page)...)
			}

			// The string of digits is unique to a world and is written once
			// in the roster and once at the head of the world's own block.
			const timesEachWorldIsWritten = 2

			for _, world := range record.Worlds {
				count := 0

				for _, found := range written {
					if strings.Contains(found.Text, world.Digits) {
						count++
					}
				}

				if count != timesEachWorldIsWritten {
					t.Fatalf("the world at %s is written %d times; want %d (the roster and its own block)",
						world.Hex, count, timesEachWorldIsWritten)
				}
			}

			assertLaneTableIsWhole(t, record, written)
		})
	}
}

// assertLaneTableIsWhole counts the lane table's rows by their Parsecs
// column, which is the one column of the booklet nothing else writes in.
func assertLaneTableIsWhole(t *testing.T, record *subsector.Subsector, written []stamp) {
	t.Helper()

	// The column is placed by the renderer, and this reads it back: the x
	// that carries the most bare distances is it.
	byColumn := map[float64]int{}

	for _, found := range written {
		if len(found.Text) == 1 && found.Text[0] >= '1' && found.Text[0] <= '4' {
			byColumn[found.X]++
		}
	}

	best := 0
	for _, count := range byColumn {
		best = max(best, count)
	}

	if best != len(record.Routes) {
		t.Fatalf("the lane table has %d rows; the record has %d lanes", best, len(record.Routes))
	}
}

// TestTheBookletIsDeterministic: a PDF carries a creation date and a
// modification date, and time.Now in either would mean a referee could
// not check that the file on his disk is the file this record produces.
//
// Comparing two renders is not enough on its own, and this is the check
// that was dead when it was written: both renders finish inside the same
// second, and fpdf writes its dates to the second, so an unpinned
// time.Now passed. The dates are therefore read back and held against the
// clock -- a date the clock could not have produced is the only proof
// that no clock was read.
func TestTheBookletIsDeterministic(t *testing.T) {
	t.Parallel()

	record := generated(t, fixture.Goldens()[0])

	first := drawn(t, record)
	second := drawn(t, record)

	if !bytes.Equal(first, second) {
		t.Fatal("two renders of one record produced different bytes")
	}

	dates := datePattern.FindAllStringSubmatch(string(first), -1)
	if len(dates) == 0 {
		t.Fatal("the document carries no dates, so this check has nothing to hold")
	}

	for _, date := range dates {
		stamped, err := time.Parse(pdfDate, date[2])
		if err != nil {
			t.Fatalf("reading %s %q: %v", date[1], date[2], err)
		}

		// Any reading of the clock lands in the year the test is run.
		if !stamped.Before(beforeTheTool()) {
			t.Fatalf("%s is %s, which is a reading of the clock rather than a pinned date",
				date[1], stamped.Format(time.RFC3339))
		}
	}
}

const pdfDate = "20060102150405"

// beforeTheTool is a date no run of this tool can have happened on, which
// is what makes "the stamp is older than this" mean "no clock was read".
func beforeTheTool() time.Time { return time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC) }

// TestTheBookletRendersASector: a sector is thirty-two columns by forty
// on one grid (ERRATA E006), and the same drawing has to hold it. A hex
// size that suits a subsector overruns the page here.
func TestTheBookletRendersASector(t *testing.T) {
	t.Parallel()

	golden := fixture.SectorGolden()

	engine, err := gen.New()
	if err != nil {
		t.Fatalf("building the engine: %v", err)
	}

	record, err := engine.Sector(gen.Inputs{Seed: golden.Seed, Name: golden.Name, OccurrenceDM: golden.OccurrenceDM})
	if err != nil {
		t.Fatalf("generating the sector: %v", err)
	}

	doc := drawn(t, record)
	places := mapLabels(t, record, pages(t, doc)[0])

	assertDrawnNeighboursAreOneParsec(t, places)
	assertNothingLeavesThePage(t, doc)
}

// TestAFullSubsectorPaginates: eighty worlds is a legal record -- every
// hex of the grid placed one -- and the roster is laid beside the map in
// half a column. It does not fit, and what does not fit has to continue
// overleaf rather than run off the bottom.
func TestAFullSubsectorPaginates(t *testing.T) {
	t.Parallel()

	record := everyHexAWorld(t)

	assertNothingLeavesThePage(t, drawn(t, record))
}

// everyHexAWorld is the subsector where all eighty throws placed a world.
// It is built by hand rather than generated because no seed is worth
// hunting for one: the record is legal, the page has to hold it, and what
// is on each world does not matter to the question.
func everyHexAWorld(t *testing.T) *subsector.Subsector {
	t.Helper()

	record := subsector.New(1, "Aramis", 0)

	for col := 1; col <= subsector.Columns; col++ {
		for row := 1; row <= subsector.Rows; row++ {
			hex := hexOf(t, col, row)

			world := subsector.World{
				Hex: hex, Name: "Regina Aleph", Starport: subsector.StarportB,
				NavalBase: true, ScoutBase: true,
				Size: 7, Atmosphere: 5, Hydrographics: 5,
				Population: 6, Government: 6, LawLevel: 6, TechIndex: 9,
				Digits: "", Clamps: nil,
			}

			digits, err := world.DigitString()
			if err != nil {
				t.Fatalf("the world at %s: %v", hex, err)
			}

			world.Digits = digits
			record.Worlds = append(record.Worlds, world)
		}
	}

	return record
}

// assertNothingLeavesThePage holds every drawn thing inside the margins.
// A roster that ran past the bottom of the page would still be in the
// file and would print as nothing at all.
func assertNothingLeavesThePage(t *testing.T, doc []byte) {
	t.Helper()

	// The page is US Letter portrait and the margin is what the booklet
	// sets. Both are measured here as the page's own edges rather than
	// taken from the renderer, so a page laid out to the wrong size is a
	// failure rather than an agreement.
	const (
		width  = 612
		height = 792
		margin = 40
		slack  = 1
	)

	for number, page := range pages(t, doc) {
		for _, found := range stamps(t, page) {
			if found.X < margin-slack || found.X > width-margin+slack {
				t.Fatalf("page %d draws %q at x %.2f, outside the margins", number+1, found.Text, found.X)
			}

			if found.Y < margin-slack || found.Y > height-margin+slack {
				t.Fatalf("page %d draws %q at y %.2f, outside the margins", number+1, found.Text, found.Y)
			}
		}
	}
}

// TestABookletHoldsARefereesOwnName: a world's name is the one field the
// referee writes in himself -- p. 12 step 3 says to name each world and
// prints no table for it -- so it is the one field with neither a length
// nor an alphabet the record can promise.
//
// Two things follow, and both are invisible until a referee writes a name
// the tool did not expect. A name longer than its column is trimmed
// rather than run through the column beside it. And a character the page
// cannot draw is written as a question mark rather than as the several
// bytes its UTF-8 spells, which is what an em-dash did before encode
// existed.
func TestABookletHoldsARefereesOwnName(t *testing.T) {
	t.Parallel()

	const (
		accented    = "Régina"
		unwriteable = "Regina Łąka"
		overlong    = "Regina Aleph Beth Gimel Daleth He Waw Zayin Heth"
		piped       = "North | South"
	)

	// Long enough to be trimmed, and every character of it two bytes
	// wide, so the trim can only land inside one.
	accentedAndOverlong := strings.Repeat("é", 60)

	record := subsector.New(1, "Aramis", 0)

	for index, name := range []string{accented, unwriteable, overlong, piped, accentedAndOverlong} {
		record.Worlds = append(record.Worlds, world(t, hexOf(t, 1, index+1), name))
	}

	written := everyStamp(t, drawn(t, record))

	// Windows-1252 carries é at 0xE9, which is the single byte Helvetica
	// is indexed by. Three UTF-8 bytes here would be the mojibake.
	if !anyStamp(written, "R\xe9gina") {
		t.Error("the accented name did not reach the page in the alphabet it can be drawn in")
	}

	// L-with-stroke and a-with-ogonek are outside Windows-1252, and the
	// k and a beside them are not: a name is not dropped for carrying a
	// character the page cannot draw.
	if !anyStamp(written, "Regina ??ka") {
		t.Error("a character the page cannot draw was not written as a question mark")
	}

	if anyStamp(written, overlong) {
		t.Error("a name longer than its column was drawn whole, through the column beside it")
	}

	// Trimmed, not dropped: the ellipsis Windows-1252 carries at 0x85.
	if !anyStampWith(written, "Regina Aleph") || !anyStampWith(written, "\x85") {
		t.Error("the overlong name was not trimmed to its column")
	}

	// A pipe is Markdown's syntax and no document here has any. Escaping
	// it stamped a backslash the referee never typed, and stamped it in
	// the lane table and the detail heading only -- so the same booklet
	// spelled one world two ways.
	if anyStampWith(written, `\|`) {
		t.Error("a name was drawn with Markdown's pipe escape in it")
	}

	if !anyStampWith(written, piped) {
		t.Error("a name carrying a pipe did not reach the page as it was written")
	}

	assertTrimmingKeepsWholeCharacters(t, written)
}

// assertTrimmingKeepsWholeCharacters: a trim that cuts inside a
// multi-byte character leaves invalid UTF-8, which encode reads as the
// replacement rune and draws as a question mark. That would silently
// destroy the character the trim stopped on, which is the one thing
// encode promises it will not do.
func assertTrimmingKeepsWholeCharacters(t *testing.T, written []stamp) {
	t.Helper()

	// Windows-1252 draws e-acute as the single byte 0xE9, so a run of
	// them that survived the trim is a run of 0xE9 and nothing else.
	trimmed := 0

	for _, found := range written {
		if !strings.HasPrefix(found.Text, "\xe9") {
			continue
		}

		trimmed++

		if strings.Contains(found.Text, "?") {
			t.Errorf("a trim cut inside a character: %q", found.Text)
		}

		if !strings.HasSuffix(found.Text, "\x85") {
			t.Errorf("the trimmed name does not end in an ellipsis: %q", found.Text)
		}
	}

	// The roster cell and the map label, both trimmed to their own width.
	const placesTheNameIsDrawn = 2

	if trimmed != placesTheNameIsDrawn {
		t.Errorf("the accented name was drawn %d times; want %d", trimmed, placesTheNameIsDrawn)
	}
}

// TestTheLaneTableIsLabelledOnEveryPageItReaches: a subsector at DM +1
// carries a hundred and fifty lanes and a sector carries hundreds, so the
// table runs past one page. The roster beside it repeats its head; this
// is the check that the lane table does too, rather than printing pages
// of unlabelled columns.
func TestTheLaneTableIsLabelledOnEveryPageItReaches(t *testing.T) {
	t.Parallel()

	record := generated(t, fixture.Golden{File: "dm-plus-one", Seed: 1, Name: "Aramis", OccurrenceDM: 1})

	sheets := pages(t, drawn(t, record))
	labelled := 0

	for number, page := range sheets {
		written := stamps(t, page)

		if !carriesALaneRow(written) {
			continue
		}

		if !anyStamp(written, "Parsecs") {
			t.Errorf("page %d carries lane rows and no column heading", number+1)
		}

		labelled++
	}

	// The table has to have reached a second page, or this proves nothing.
	const atLeast = 2

	if labelled < atLeast {
		t.Fatalf("the lane table reached %d pages; the check needs at least %d to mean anything", labelled, atLeast)
	}
}

// carriesALaneRow reports whether a page has a row of the lane table on
// it, which is a bare distance written in the table's third column.
func carriesALaneRow(written []stamp) bool {
	const parsecsColumn = 440.0

	for _, found := range written {
		if near(found.X, parsecsColumn) && found.Text != "Parsecs" {
			return true
		}
	}

	return false
}

// TestAnEmptyBookletSaysSo: a run whose eighty throws place no world
// produces a valid record, and the booklet says so rather than printing
// an empty page.
func TestAnEmptyBookletSaysSo(t *testing.T) {
	t.Parallel()

	record := subsector.New(1, "Aramis", 0)

	written := everyStamp(t, drawn(t, record))

	for _, want := range []string{
		"No world was placed. An empty subsector is a result.",
		"No space lane was drawn.",
	} {
		if !anyStamp(written, want) {
			t.Errorf("the booklet does not say %q", want)
		}
	}

	for _, found := range written {
		if strings.Contains(found.Text, "The worlds in detail") {
			t.Error("the booklet opened a detail section for a subsector with no worlds")
		}
	}
}

// world is a named world at a hex, for a record built by hand.
func world(t *testing.T, hex subsector.Hex, name string) subsector.World {
	t.Helper()

	built := subsector.World{
		Hex: hex, Name: name, Starport: subsector.StarportB,
		NavalBase: false, ScoutBase: false,
		Size: 7, Atmosphere: 5, Hydrographics: 5,
		Population: 6, Government: 6, LawLevel: 6, TechIndex: 9,
		Digits: "", Clamps: nil,
	}

	digits, err := built.DigitString()
	if err != nil {
		t.Fatalf("the world at %s: %v", hex, err)
	}

	built.Digits = digits

	return built
}

func anyStamp(written []stamp, text string) bool {
	for _, found := range written {
		if found.Text == text {
			return true
		}
	}

	return false
}

func anyStampWith(written []stamp, text string) bool {
	for _, found := range written {
		if strings.Contains(found.Text, text) {
			return true
		}
	}

	return false
}

// everyStamp is every string the whole booklet drew, in page order.
func everyStamp(t *testing.T, doc []byte) []stamp {
	t.Helper()

	sheets := pages(t, doc)
	written := make([]stamp, 0, len(sheets))

	for _, page := range sheets {
		written = append(written, stamps(t, page)...)
	}

	return written
}

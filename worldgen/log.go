package worldgen

import (
	"fmt"
	"strconv"

	"github.com/philoserf/ctworldgen/dice"
	"github.com/philoserf/ctworldgen/tables"
)

// stampErratum records that a docs/ERRATA.md reading governed this run.
func (g *generator) stampErratum(id string) { g.errata[id] = true }

// stamped is the readings that governed, in the document's own order, so
// two records that applied the same readings list them the same way.
func (g *generator) stamped() []string {
	out := []string{}

	for _, id := range errataIDs {
		if g.errata[id] {
			out = append(out, id)
		}
	}

	return out
}

// Plural writes a count with its noun.
//
// Exported because render says the same counts in its listing, and the two
// must not drift: the event log's text is compared verbatim by Replay, so a
// wording fixed on one side alone would leave the rendered listing and the
// record it renders disagreeing, with nothing to catch it. "1 worlds" would
// otherwise be frozen into every record written before anyone fixed it.
func Plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}

	return strconv.Itoa(n) + " " + word + "s"
}

func (g *generator) next() int {
	g.seq++

	return g.seq
}

// step opens a procedure step in the log.
func (g *generator) step(step, text string) {
	g.sub.Events = append(g.sub.Events, Event{Seq: g.next(), Kind: KindStep, Step: step, Text: text})
}

// outcome records a consequence, referencing the throw that caused it
// (ref 0 for a consequence the rules impose without a throw).
func (g *generator) outcome(step, hex, text string, ref int) {
	g.sub.Events = append(g.sub.Events, Event{
		Seq: g.next(), Kind: KindOutcome, Step: step, Hex: hex, Text: text, Ref: ref,
	})
}

// describe records a characteristic's value with the label its Book 3
// table gives it. A value past the table's last printed row has no label —
// law level is the one that can be, since E004 leaves it uncapped — and
// the outcome then states the bare value.
func (g *generator) describe(hex, label, table string, value, ref int) {
	text := fmt.Sprintf("%s %d", label, value)

	if desc, ok := g.charts.Label(table, value); ok {
		text = fmt.Sprintf("%s %d: %s", label, value, desc)
	}

	g.outcome(StepWorldCreation, hex, text, ref)
}

// throwSpec is one throw to make and log: which step and world it belongs
// to, how many dice, the DMs, and the target if there is one. A throw with
// no target is a table lookup — the starport type, every characteristic —
// and its total is the result rather than a pass or a fail.
type throwSpec struct {
	step   string
	label  string
	hex    string
	count  int
	dms    []EventDM
	target *dice.Target
}

// roll makes one throw, logs it, and reports the modified total, whether
// the target was met (always false for a throw with no target), and the
// sequence number the log gave it so a consequence can reference it.
//
// This is the only place dice are consumed. Every roll is drawn here in
// procedure order, which is what makes that order load-bearing for replay.
func (g *generator) roll(spec throwSpec) (int, bool, int) {
	rolled := make([]int, 0, spec.count)

	total := 0

	for range spec.count {
		d := g.stream.One()
		rolled = append(rolled, d)
		total += d
	}

	for _, dm := range spec.dms {
		total += dm.Value
	}

	event := Event{
		Seq: g.next(), Kind: KindThrow, Step: spec.step, Label: spec.label, Hex: spec.hex,
		Dice: rolled, DMs: spec.dms, Total: total,
	}

	met := false

	if spec.target != nil {
		met = spec.target.Met(total)
		event.Target = spec.target.String()
		event.Success = &met
	}

	g.sub.Events = append(g.sub.Events, event)

	return total, met, event.Seq
}

// clamped holds a generated value inside the printed range of its own
// Book 3 table (docs/ERRATA.md E004), reading the ceiling off the table
// rather than from a second transcription of it. A clamp that actually
// binds is logged and stamps the reading; one that does not is silent,
// which is why size and population — whose formulas cannot leave their
// tables — pass through without ever stamping it.
func (g *generator) clamped(hex, table string, raw, ref int) (int, error) {
	maximum, ok := g.charts.MaxValue(table)
	if !ok {
		return 0, fmt.Errorf("hex %s: no %s table to clamp against: %w", hex, table, tables.ErrNoSuchValue)
	}

	return g.clampRange(hex, table, raw, 0, maximum, ref), nil
}

// floored applies E004's lower bound alone: law level's table ends at 9
// but the level is not capped there.
func (g *generator) floored(hex, label string, raw, ref int) int {
	if raw < 0 {
		return g.noteClamp(hex, label, raw, 0, "floor", ref)
	}

	return raw
}

// clampRange is the clamp itself.
func (g *generator) clampRange(hex, label string, raw, low, high, ref int) int {
	switch {
	case raw < low:
		return g.noteClamp(hex, label, raw, low, "floor", ref)
	case raw > high:
		return g.noteClamp(hex, label, raw, high, "ceiling", ref)
	}

	return raw
}

// noteClamp is the one place E004 is stamped.
func (g *generator) noteClamp(hex, label string, raw, value int, bound string, ref int) int {
	g.stampErratum(ErrataClamp)
	g.outcome(StepWorldCreation, hex,
		fmt.Sprintf("%s throw of %d clamped to %d, the %s of the printed table (docs/ERRATA.md E004)",
			label, raw, value, bound), ref)

	return value
}

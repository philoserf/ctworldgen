// Package render turns a subsector record into Markdown: the subsector
// listing a referee keeps, and the transcript of the generation that
// produced it. The JSON record is the source of truth; everything here is
// a view of it (docs/PRD.md).
package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/philoserf/ctworldgen/tables"
	"github.com/philoserf/ctworldgen/worldgen"
)

// Listing renders the subsector: the world roster, the space lanes, and a
// page of detail for each world — "at least one (and preferably several)
// pages in a central notebook maintained by the referee" (p. 4), which is
// what this stands in for.
func Listing(sub *worldgen.Subsector) (string, error) {
	charts, err := tables.Load()
	if err != nil {
		return "", fmt.Errorf("loading the Book 3 charts: %w", err)
	}

	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", title(sub))
	fmt.Fprintf(&b, "%s\n\n", summary(sub))

	writeWorldTable(&b, sub)
	writeLaneTable(&b, sub)

	if len(sub.Worlds) > 0 {
		b.WriteString("## World details\n\n")

		for _, w := range sub.Worlds {
			writeWorldDetail(&b, charts, w)
		}
	}

	return b.String(), nil
}

func title(sub *worldgen.Subsector) string {
	if sub.Name == "" {
		return "Subsector"
	}

	return sub.Name + " Subsector"
}

func summary(sub *worldgen.Subsector) string {
	lines := []string{
		fmt.Sprintf("%s in %d hexes; %s.",
			plural(len(sub.Worlds), "world"), worldgen.Hexes, plural(len(sub.Routes), "space lane")),
		fmt.Sprintf("Generated from seed %d, occurrence DM %+d.", sub.RNG.Seed, sub.Inputs.OccurrenceDM),
		"Book 3 pp. 1–12, © 1977 text. Characteristics are written as the string of",
		"digits of p. 4: starport, size, atmosphere, hydrographics, population,",
		"government, law level, technological index (docs/ERRATA.md E005).",
	}

	return strings.Join(lines, "\n")
}

func writeWorldTable(b *strings.Builder, sub *worldgen.Subsector) {
	b.WriteString("## Worlds\n\n")

	if len(sub.Worlds) == 0 {
		// Eighty throws that placed nothing is a result, not a failure
		// (docs/PRD.md, Decisions).
		b.WriteString("No world is present in any hex of the subsector.\n\n")

		return
	}

	b.WriteString("| Hex | Name | Profile | Bases |\n| --- | --- | --- | --- |\n")

	for _, w := range sub.Worlds {
		fmt.Fprintf(b, "| %s | %s | %s | %s |\n", w.Hex, w.Name, w.Profile, bases(w))
	}

	b.WriteString("\n")
}

func writeLaneTable(b *strings.Builder, sub *worldgen.Subsector) {
	b.WriteString("## Space lanes\n\n")

	if len(sub.Routes) == 0 {
		b.WriteString("No space lane connects any pair of worlds.\n\n")

		return
	}

	b.WriteString("| From | To | Jump |\n| --- | --- | --- |\n")

	for _, r := range sub.Routes {
		fmt.Fprintf(b, "| %s | %s | %d |\n", r.From, r.To, r.Distance)
	}

	b.WriteString("\n")
}

func bases(w worldgen.World) string {
	var present []string

	if w.NavalBase {
		present = append(present, "naval")
	}

	if w.ScoutBase {
		present = append(present, "scout")
	}

	if len(present) == 0 {
		return "—"
	}

	return strings.Join(present, ", ")
}

func writeWorldDetail(b *strings.Builder, charts *tables.Charts, w worldgen.World) {
	heading := w.Hex + " " + w.Profile
	if w.Name != "" {
		heading = w.Name + " (" + heading + ")"
	}

	b.WriteString("### " + heading + "\n\n")

	for _, line := range append([]string{starportLine(charts, w)}, characteristicLines(charts, w)...) {
		fmt.Fprintf(b, "- %s\n", line)
	}

	b.WriteString("\n")
}

// starportLine describes the starport from the p. 5 chart, and the bases
// the E001 throws placed at it.
func starportLine(charts *tables.Charts, w worldgen.World) string {
	sp, err := charts.Starport(w.Starport)
	if err != nil {
		return "Starport " + w.Starport
	}

	parts := []string{sp.Quality}

	if sp.Fuel == "none" {
		parts = append(parts, "no fuel")
	} else {
		parts = append(parts, sp.Fuel+" fuel")
	}

	if sp.Overhaul {
		parts = append(parts, "annual overhaul")
	}

	if sp.Shipyard != "none" {
		parts = append(parts, "shipyard: "+sp.Shipyard)
	}

	parts = append(parts, "bases: "+bases(w))

	return fmt.Sprintf("**Starport %s** — %s", w.Starport, strings.Join(parts, "; "))
}

// characteristicLines states each of the seven remaining characteristics
// with the label its Book 3 table gives it. Law level can hold a value
// past the last row the p. 7 table prints (docs/ERRATA.md E004 leaves it
// uncapped), and then the line says so rather than inventing a label.
func characteristicLines(charts *tables.Charts, w worldgen.World) []string {
	rows := []struct {
		name  string
		table string
		value int
	}{
		{"Planetary size", tables.Size, w.Size},
		{"Atmosphere", tables.Atmosphere, w.Atmosphere},
		{"Hydrographics", tables.Hydrographics, w.Hydrographics},
		{"Population", tables.Population, w.Population},
		{"Government", tables.Government, w.Government},
		{"Law level", tables.LawLevel, w.LawLevel},
	}

	out := make([]string, 0, len(rows)+1)

	for _, row := range rows {
		line := fmt.Sprintf("%s %d", row.name, row.value)

		if label, ok := charts.Label(row.table, row.value); ok {
			line += ": " + label
		} else {
			line += ": above the last row the printed table gives (docs/ERRATA.md E004)"
		}

		out = append(out, line)
	}

	return append(out, fmt.Sprintf("Technological index %d", w.TechnologicalIndex))
}

// History renders the generation transcript: every step, throw, and
// outcome in the order the engine reached them (FR17).
func History(sub *worldgen.Subsector) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s — generation record\n\n", title(sub))
	fmt.Fprintf(&b, "Seed %d, engine %s, ruleset %s.\n",
		sub.RNG.Seed, sub.EngineVersion, sub.Ruleset)

	if len(sub.Errata) > 0 {
		fmt.Fprintf(&b, "Readings applied (docs/ERRATA.md): %s.\n", strings.Join(sub.Errata, ", "))
	}

	b.WriteString("\n")

	step := ""

	for _, ev := range sub.Events {
		if ev.Kind == worldgen.KindStep {
			step = ev.Step

			fmt.Fprintf(&b, "## %s\n\n%s\n\n", ev.Step, ev.Text)

			continue
		}

		if step == "" {
			// An event before any step opened: log it rather than drop it.
			b.WriteString("## (no step)\n\n")

			step = "-"
		}

		fmt.Fprintf(&b, "%d. %s\n", ev.Seq, describeEvent(ev))
	}

	return b.String()
}

func describeEvent(ev worldgen.Event) string {
	where := ""
	if ev.Hex != "" {
		where = ev.Hex + ": "
	}

	if ev.Kind == worldgen.KindThrow {
		return where + describeThrow(ev)
	}

	out := where + ev.Text
	if ev.Ref != 0 {
		out += fmt.Sprintf(" (from throw %d)", ev.Ref)
	}

	return out
}

func describeThrow(ev worldgen.Event) string {
	dice := make([]string, 0, len(ev.Dice))
	for _, d := range ev.Dice {
		dice = append(dice, strconv.Itoa(d))
	}

	var out strings.Builder
	fmt.Fprintf(&out, "%s: threw %s", ev.Label, strings.Join(dice, "+"))

	for _, dm := range ev.DMs {
		fmt.Fprintf(&out, ", %+d %s", dm.Value, dm.Source)
	}

	fmt.Fprintf(&out, " = %d", ev.Total)

	if ev.Target != "" {
		verdict := "failed"
		if ev.Success != nil && *ev.Success {
			verdict = "made"
		}

		fmt.Fprintf(&out, " vs %s — %s", ev.Target, verdict)
	}

	return out.String()
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}

	return strconv.Itoa(n) + " " + word + "s"
}

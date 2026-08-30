# ERRATA

Interpretations of and deviations from the held text (Book 3 pp. 1–12,
© 1977 text, FFE reprints; Book 1 pp. 2–3 and p. 8 where Book 3 is written
on top of them), each with its page cite. Nothing here is applied silently:
every record stamps, in its `errata` array, the identifiers of the entries
that governed its generation. Page cites are printed page numbers of Book 3
unless marked B1 (PDF page N+5 in Book 3, N+6 in Book 1).

## E001 — Where the base throws sit in the procedure (pp. 1, 5, 12)

The star mapping and world creation checklist (p. 12) has three steps and
no base throw. But p. 1 says starports "will be accompanied by naval or
scout bases," and the starport chart prints the throws as rules: naval base
8+ at starport A and B; scout base 10+ at A, 9+ at B, 8+ at C, 7+ at D
(p. 5). Bases are generated; the checklist is a summary that omits them.

The reading: the throws are made **immediately after the starport type**,
in the star-mapping pass, because a base is a property of the starport and
the starport chart is where the throw is printed — naval first, then scout,
the order the chart lists them. Starports E and X throw for neither. The
throws are two dice each: the targets run to 10+, which one die cannot
reach, and B1 p. 2 makes two dice the unqualified throw.

Stream position is load-bearing for replay, so this is a version-bumping
reading, not a cosmetic one. Stamped on records where any base throw was
made.

## E002 — The hex scan order (p. 1)

"Systematically check each hex, throwing one die" (p. 1) fixes that every
hex is checked and not the order in which. The order decides which hex
consumes which die, so it is load-bearing for replay.

The reading: ascending hex number — the p. 3 grid's own numbering, column
by column, row within column. 0101, 0102, … 0110, 0201, … 0810. That is
the order the grid numbers the hexes and the order a referee reading the
sheet would work in. Stamped on every record.

## E003 — Which pairs of worlds are examined for lanes, and in what order (p. 2)

P. 2 says to consult the jump routes table for "each world … and its
neighbors," that "each specific pair of worlds should be examined for jump
routes only once," and that the procedure "is followed for most worlds
within four hexes of each other." It then observes that because a jump-2
can be made over two jump-1 links, "it is possible to ignore some potential
connections because they are already present," which "may well help in the
creation of legible subsector maps."

Three things are unfixed, and all three are load-bearing:

1. **Which pairs.** The reading: every pair of worlds at a hex distance of
   1, 2, 3, or 4. The legibility passage is advice to a referee drawing a
   map by hand, not a mechanic — it states no rule for which connections to
   skip, and any tool that skipped some would be inventing one. "Most
   worlds" is the same kind of hedge.
2. **Starport X.** The jump routes table prints rows for A–A through E–E
   and none for X, and p. 5 gives X "no provision … for any starship
   landings." The reading: a pair with an X starport on either side has no
   row, is not examined, and consumes no die. Commercial lanes need a
   starport at both ends.
3. **The order.** The reading: ascending hex number of the first world,
   then of the second — the same order as E002, over the worlds already
   placed.

Stamped on records with two or more worlds — the point at which there is
a pair for the reading to govern at all.

## E004 — Generated values are clamped to their tables' printed ranges (pp. 4–9)

The world creation arithmetic can produce values that no table row
describes. Atmosphere is 2D − 7 + size, so a size-10 world can throw to 15
against a table that ends at C (12). Hydrographics is the same expression
against a table that ends at A (10) — and 150% of a surface covered by
ocean is not a description of anything. Government is 2D − 7 + population
and reaches 15 against a table ending at D (13). Every one of them can also
go negative, and the technological index can, at 1D − 4 (starport X) − 2
(government D), reach −5 against tables that begin at 0.

The reading: **a generated value is clamped to the printed range of its own
table** — floor at 0, ceiling at the last row the book prints. So
atmosphere at C, hydrographics at A, government at D, technological index
at 18. Planetary size and population cannot exceed their tables under the
printed formulas (2D − 2 tops out at 10 = A) and so are floored only.

Two warrants, both on the page:

- P. 8 anticipates exactly this: "At times, the referee (or the players)
  will find combinations of features which may seem contradictory or
  unreasonable. Common sense should rule in such cases."
- The technological index matrix (p. 9) requires it. The matrix has a row
  for every value in a table's printed range and none outside it, so a
  government of 15 or an atmosphere of 15 has no DM to contribute. Clamping
  is what makes the matrix total.

**Law level is the exception and is not capped above.** Its table ends at
9, but the note beneath it — "Each law level includes all prohibitions and
conditions of levels numbered lower than it. Thus, shotguns are prohibited
at all law levels from 7 higher" (p. 7) — is written for levels above the
last printed row, and law level feeds no matrix. A law level of 15 is a
world where everything at 9 is prohibited and the enforcement throw is
harder; that is meaningful, so it stands. It is floored at 0 like the rest.

The technological index's ceiling of 18 is the book's own: "Technological
index may vary from zero to 18" (p. 9), and the technological level tables
(pp. 10–11) print rows 0 through 18 and stop.

Stamped on records where any clamp actually bound a value.

## E005 — The string of digits: order, alphabet, and the absent hyphen (p. 4; B1 p. 8)

"For convenience, planetary characteristics should be expressed as a string
of digits, in much the same manner as the Universal Personality Profile is
used for the easy identification of persons" (p. 4). That fixes neither the
order of the characteristics in the string, nor how a value above 9 is
written, nor whether anything separates them.

The reading, in three parts:

1. **Order**: as the p. 4 Planetary Characteristics box lists them —
   starport type, planetary size, planetary atmosphere, hydrographics,
   population, government, law level, technological index. Eight characters.
2. **Alphabet**: one character per characteristic, digits 0–9 then letters,
   per the UPP's own hexadecimal convention (B1 p. 8: 10–15 as A–F) extended
   by p. 2's letter set for values above 15 — "letters (A through Z,
   omitting O and I as they may be confused with numbers)". So 10 is A, 15
   is F, 16 is G, 17 is H, and 18 — the technological index's ceiling — is
   J, I having been skipped. Starport type is already a letter and is
   written as it is.
3. **No separator.** The UPP is six characters with nothing between them,
   and this is to be "in much the same manner." The familiar hyphen before
   the technological index — `A867A69-8` — is not in the held text, and
   this tool does not add it. A world's string of digits is eight
   characters: `A867A698`.

Part 3 is the one a reader will notice, and it is the same call the sibling
makes when the held page says survival failure is death: the remembered
edition does not govern. The individual characteristics are all stored
numerically in the record and printed in labelled columns by `render`, so
nothing is lost by the format — only gained back if a later convention is
wanted.

Stamped on every record with at least one world.

## Noted discrepancies (not stamped)

**A population-0 world still receives a technological index (pp. 8, 9).**
The technological index matrix gives population 0 no DM — not a penalty,
not a zeroing — so an uninhabited world's index is thrown and recorded like
any other's. Later printings zero the index of an unpopulated world; the
held text prints no such rule, and the held text governs. The record says
what the matrix and the die said.

**"Greater than 9" and "A+" are the same hydrographics DM (pp. 4, 12).**
The prose says the −4 DM applies "if planetary atmosphere is 0, 1, or
greater than 9"; the checklist writes the same condition as "Atmosphere 0, 1
or A+". Those agree — A is 10 — and are implemented as one condition. No
reading is required; it is recorded here because the two wordings look
different on the page.

**The em-dash and minus glyphs in the held PDFs.** Not a rule at all, but
the trap that would most easily put a wrong table in this repository. The
PDFs' embedded font maps the em-dash to `4` and the minus sign to `3`, so a
text extraction renders the jump routes table's empty cells as the digit 4
(`B-E 4 4 4 4` for `4 — — —`) and the size formula `2D − 2` as `2D32`. Both
are wrong and both look like data. Every table in `tables/data/` was
transcribed from a visual read of the page. Any future table work does the
same.

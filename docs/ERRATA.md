# ERRATA

Readings of the held text where it is ambiguous or silent, each with the
printed page it is read from. Nothing here is applied silently: every
subsector record stamps, in its `errata` array, the identifiers of the
entries that governed its generation, and each entry below states the
condition under which it is stamped.

The held text is Book 3 _Worlds and Adventures_ pp. 1–12 and Book 1
_Characters and Combat_ pp. 2–3 and p. 8 (FFE reprints of the © 1977
text). Page cites are Book 3 printed pages unless marked B1. Printed page
N is PDF page N+5 in Book 3 and N+6 in Book 1.

## E001 — Where the base throws sit in the procedure (pp. 1, 5, 12)

The star mapping and world creation checklist (p. 12) has three steps and
no base throw. But p. 1 says starports "will be accompanied by naval or
scout bases," and the starport chart prints the throws as rules: naval
base present on a throw of 8+ at starport A and B; scout base present on
a throw of 10+ at A, 9+ at B, 8+ at C, 7+ at D (p. 5). Bases are
generated; the checklist is a summary that omits them.

What the pages do not fix is where in the procedure the throws are made,
and that decides which throw consumes which die.

**The reading:** the base throws are made **immediately after the
starport type**, within the star-mapping pass, because a base is a
property of the starport and the starport chart is where the throw is
printed. Naval first, then scout — the order the chart lists them.
Starports E and X print no throw and are thrown for at neither. Each
throw is two dice: the scout target at starport A is 10+, which one die
cannot reach, and B1 p. 2 makes two dice the unqualified throw.

_Stamped on records where any base throw was made._

## E002 — The order of the procedure's passes, and of the hexes and worlds within them (pp. 1, 12)

"Systematically check each hex, throwing one die and marking the hex with
a circle if the result is a 4, 5, or 6" (p. 1). That fixes that every hex
is checked, and not the order. Nor does any page fix the order in which
worlds are visited once placed — the starport pass and world creation each
iterate over them and neither says how. Every one of these decides which
throw consumes which die.

**The reading, in three parts:**

1. **The passes.** The p. 12 checklist fixes the boundaries and this tool
   follows them exactly: 1.A throws for every hex in the subsector, then
   1.B determines a starport type for every world found, then 1.C
   determines the routes, and only then does 2 generate the specific
   worlds. Each is a complete pass over the whole subsector before the
   next begins; nothing is interleaved. The base throws sit inside the
   starport pass (E001), and the route pass has an order of its own (E003).

2. **Hexes and worlds within a pass.** Ascending hex number — the p. 3
   grid's own numbering, column by column and row within column: 0101,
   0102, … 0110, 0201, … 0810. That is the order the grid prints the
   numbers in and the order a referee working down the sheet would follow.
   Because the occurrence scan runs in that order, the worlds it places
   are already in it, and every later pass keeps it.

3. **Characteristics within a world.** The order the p. 12 checklist lists
   them at 2.B through 2.H: size, atmosphere, hydrographics, population,
   government, law level, technological index. Each world is finished
   before the next is begun.

_Stamped on every record._

## E003 — Which pairs of worlds are examined for routes, when a die is thrown, and in what order (p. 2)

P. 2 says to note the starport type "for it and for its neighbors," to
consult the jump routes table "throwing one die," that "each specific
pair of worlds should be examined for jump routes only once," and that
the procedure "is followed for most worlds within four hexes of each
other." The p. 12 checklist writes the same step as "Determine space
lanes; check all possible jump routes."

Four things are unfixed, and all four decide the dice stream.

1. **Which pairs.** The reading: every pair of worlds at a hex distance
   of 1, 2, 3, or 4. P. 2 goes on to observe that because a jump-2 can be
   made over two jump-1 links, "it is possible to ignore some potential
   connections because they are already present," which "may well help in
   the creation of legible subsector maps." That is advice to a referee
   drawing a map by hand: it states no rule for which connections to
   skip, and a tool that skipped some would be inventing one. "Most
   worlds" is the same kind of hedge, and p. 12's "check all possible
   jump routes" is the checklist's own reading of it.

2. **Starport X.** The jump routes table prints rows for A–A through E–E
   and none for X, and p. 5 gives starport X "no provision … for any
   starship landings." The reading: a pair with an X starport at either
   end has no row, is not examined, and consumes no die. A commercial
   route needs a starport at both ends.

3. **A dash cell.** Twenty-nine of the table's sixty cells print an
   em-dash rather than a number — A–C at jump-4, A–D at jump-3 and
   jump-4, and so on down to E–E, which prints a number only at jump-1.
   The reading: a dash consumes no die. P. 2 describes the throw as "At the intersection
   of the distance column and the world pair row, a number is stated. If
   the one die throw is equal to, or greater than the number, a space
   lane exists." Where a dash is printed no number is stated, so there is
   no target to throw against and no route is possible at that distance.
   The alternative reading — that "Consult the jump routes table,
   throwing one die" makes the throw unconditional, the die thrown and
   discarded — is answered by that sentence: the die is thrown against a
   stated number, and a dash states none.

4. **The order.** The reading: ascending hex number of the first world,
   then of the second — the order of E002, over the worlds already
   placed.

_Stamped on records with two or more worlds, the point at which there is
a pair for the reading to govern._

## E004 — Generated values are floored at zero, and capped only where the book prints a cap (pp. 4–9)

The world creation arithmetic produces values that no descriptive table
describes. Planetary atmosphere is 2D − 7 + planetary size, so a size-10
world reaches 15 against a table that ends at C (12). Hydrographics is the
same expression against a table ending at A (10). Planetary government is
2D − 7 + population and reaches 15 against a table ending at D (13). Law
level is 2D − 7 + government and reaches 20 against a table ending at 9.
Every one of them can also fall below zero, and the technological index at
1D − 4 (starport X) − 2 (government D) reaches −5.

**The reading, in three parts.**

### 1. Every value is floored at 0

This one is forced rather than chosen. P. 4 requires the characteristics
to be "expressed as a string of digits," and the notation those digits are
written in — B1 p. 8's hexadecimal, extended by p. 2's letters — has no
character for a negative number. Every printed table begins at 0 besides.

Atmosphere, hydrographics, government, law level and the technological
index can all go negative; planetary size and population cannot, 2D − 2
having a floor of 0 already.

**A floored value is the value.** It is what the record carries and what
every later step consumes: government feeds law level, atmosphere feeds
the hydrographics DM, and all six feed the technological index matrix. The
raw negative is recorded beside it as a clamp and used for nothing.

### 2. Nothing is capped because a table stops describing it

A generated value that no table names is a gap in the descriptive table,
not a bound on the throw. P. 8 says as much, and says what to do instead:

> At times, the referee (or the players) will find combinations of
> features which may seem contradictory or unreasonable. Common sense
> should rule in such cases; either the players or referee will generate a
> rationale which explains the situation, or an alternative description
> should be made.

The remedy the book prints for a value its tables do not cover is the
referee's own description of it — not a clamp. An atmosphere of D is one
more exotic; a hydrographic percentage of B is a world whose referee
explains it. Two further things on the page point the same way:

- The law levels note (p. 7) — "Each law level includes all prohibitions
  and conditions of levels numbered lower than it. Thus, shotguns are
  prohibited at all law levels from 7 higher" — is written for levels
  above the last printed row.
- The technological index matrix (p. 9) prints an atmosphere DM of +1 at
  Value rows D (13) and E (14), values the p. 5 atmosphere table does not
  describe. The book's own machinery expects atmospheres above C, and
  capping there would make two printed DMs unreachable.

So atmosphere, hydrographics and government run 0 to 15, and law level 0
to 20.

### 3. One cap: the technological index at 18

"Technological index may vary from zero to 18" (p. 9), and the
technological levels tables print rows 0 through 18 and stop (pp. 10–11).
This is the one characteristic for which the book prints a range of the
value itself rather than a table that happens to end; the others never get
one.

The cap binds. The matrix's DMs reach +14 — starport A +6, size 0 +2,
atmosphere 0 +1, hydrographics 0 nothing, population A +4, government 5 +1
— which with a die of 6 gives 20. That is an asteroid belt with a first
class starport and ten billion inhabitants, so the cap is rare; it is
still reachable, and R6's size 0 makes the atmosphere and hydrographics of
that combination automatic rather than lucky.

### A value with no matrix row contributes no DM

The matrix's Value column runs 0 through 9, A through E, and X, so 15 is
the only generated value with no row at all, reachable by atmosphere,
hydrographics and government. An absent row contributes nothing, which is
what the matrix's printed dashes already mean.

_Stamped on records where a floor or the cap actually bound a value. The
clamp is also recorded on the world it bound, with the raw value and the
value kept._

## E005 — The string of digits: order, alphabet, and the absent hyphen (p. 4; B1 p. 8)

"For convenience, planetary characteristics should be expressed as a
string of digits, in much the same manner as the Universal Personality
Profile is used for the easy identification of persons" (p. 4). That
fixes neither the order of the characteristics in the string, nor how a
value above 9 is written, nor whether anything separates them.

**The reading, in three parts:**

1. **Order** — as the p. 4 Planetary Characteristics box lists them:
   starport type, planetary size, planetary atmosphere, hydrographics,
   population, government, law level, technological index. Eight
   characters.

2. **Alphabet** — one character per characteristic. B1 p. 8 gives the
   UPP's own convention: "hexadecimal (base 16) notation. In hexadecimal
   notation, the digits 0 through 9 are represented by common arabic
   numbers; the digits 10 through 15 are represented by the letters A
   through F." Above 15 the UPP has nothing to say, and Book 3 p. 2 does:
   "single digits (the numbers 0 through 9) and letters (A through Z,
   omitting O and I as they may be confused with numbers)." So 10 is A,
   15 is F, 16 is G, 17 is H, 18 is J — I having been skipped — 19 is K
   and 20 is L. The alphabet must reach 20, which is law level's maximum:
   law level is 2D − 7 + government against an uncapped government of 15
   (E004). Starport type is already a letter and is written as it is.

3. **No separator.** B1 p. 8 gives the UPP as "a string of 6 digits" with
   nothing between them, and p. 4 asks for "much the same manner." The
   familiar hyphen before the technological index — `A867A69-8` — is not
   in the held text, and is not added here. A world's string of digits is
   eight characters: `A867A698`.

Part 3 is the one a reader will notice first, and it is the reading the
held printing forces: the remembered edition does not govern. Every
characteristic is stored numerically in the record besides, so the format
loses nothing.

_Stamped on records with at least one world._

## E006 — A sector: sixteen subsectors on one grid, and the routes at the seams (pp. 1–2)

The book charts a subsector and stops. P. 1 maps "in convenient segments,
called subsectors", takes "a convenient size" from the p. 3 grid sheet,
and charts growth outward as "ultimately, travellers will venture into
unknown areas and additional subsectors will have to be charted." It
prints no sector grid and no procedure that crosses one subsector into
the next.

But the route rule has no border in it. P. 2 says "For each world, note the
starport type for it and for its neighbors", and "This procedure is
followed for most worlds within four hexes of each other." The jump routes
table is read on the starport pair and the distance, and neither knows
where a subsector ends. The border is an artifact of generating one
subsector at a time, not a term in the rule.

So a sector is the book's own two halves put together: its subsectors,
charted one at a time as p. 1 charts them, and its route rule applied to
neighbours as p. 2 states it. **The reading, in four parts.**

### 1. Sixteen subsectors, each generated whole

Four across and four down. Member _i_ is generated exactly as a lone
subsector is — the same eighty occurrence throws, the same starports, the
same bases, the same interior route pass, the same world creation — on
seed `base + i`. So every member of a sector reproduces on its own: member
5 of `sector --seed N` is `new --seed N+5`, and nothing about being in a
sector changes a world.

Member _i_ sits at column band `i mod 4` and row band `i div 4`, which
numbers the members left to right and then down, as a reader numbers
anything on a page.

Nothing is re-thrown at sector level. No world is placed, no starport
graded, no characteristic generated by the sector pass. It only assembles
and then examines seams.

### 2. The hexes translate, and the p. 3 parity survives it

A local hex (_c_, _r_) in band (_mx_, _my_) becomes (`mx*8 + c`,
`my*10 + r`) on the sector grid, 0101 through 3240.

A subsector is **eight** columns wide, and eight is even, so a column's
odd-or-even parity is the same before and after translation. That is what
makes the translation safe: the p. 3 layout puts the even-numbered columns
half a hex low, so a distance measured in sector coordinates equals the
distance measured locally, for every interior pair. An odd band width
would have flipped the parity of every second band and quietly changed
interior distances — which is the trap `Hex.cube` already carries, arriving
by a second road.

### 3. One more route pass, at the seams only

P. 2 says "Each specific pair of worlds should be examined for jump routes
only once." An interior pair was examined inside its own member, so the
sector pass examines only pairs whose two worlds are in **different**
members, and only those four hexes apart or fewer.

Everything else about the throw is E003 unchanged: the same jump routes
table, one die against a stated number, no row for an X starport and no
throw at a dash cell, and the order is E003's — ascending grid number of
the first world and then of the second, now read in sector coordinates.

### 4. The seam pass has its own stream

The sixteen members consume the streams of seeds `base` through
`base + 15`. The seam pass consumes a seventeenth, seeded `base + 16`. It
is the next seed after the members and it is stated here because nothing
on the page fixes it, and because a sector must reproduce from its one
recorded seed like everything else this tool writes.

_Stamped on every sector record. A sector also carries the readings its
members stamped, because the members are the record's worlds._

## Noted discrepancies (not stamped)

These need no reading. They are recorded because each one looks, on the
page, like a mistake or a rule and is neither.

**A population-0 world still receives a technological index (pp. 8, 9).**
The technological index matrix gives population 0 no DM — not a penalty,
not a zeroing; the cell is a dash like any other absent modifier — so an
uninhabited world's index is thrown and recorded like any other's. Later
printings zero the index of an unpopulated world. The held text prints no
such rule, and the held text governs.

**"The single inhabited world" and a population of 0 (pp. 2, 8).** P. 2
says the process "applies only to the single inhabited world in a star
system," and p. 8's population table prints a row 0, "No inhabitants."
The book generates uninhabited worlds by its own procedure and describes
them by its own table. No reconciliation is attempted: the throw is made
and the result recorded.

**The world occurrence throw is a target, not a set (p. 1).** The page
says to mark the hex "if the result is a 4, 5, or 6," which reads as
membership in a set of three faces. It is a 4+ target. The same paragraph
offers the referee a DM of +1 or −1 "making them more frequent or less
frequent," and under set membership a +1 DM would be a no-op — naturals 3,
4 and 5 succeeding instead of 4, 5 and 6, three faces either way — while a
natural 6 modified to 7 would stop placing a world. Only a target reading
makes the DM do what the page says it does. No reading is required; it is
recorded here because the set is what the page prints.

**"Greater than 9" and "A+" are the same hydrographics DM (pp. 4, 12).**
The prose applies the −4 DM "if planetary atmosphere is 0, 1, or greater
than 9"; the p. 12 checklist writes the same condition as "Atmosphere 0,
1 or A+". Those agree — A is 10 — and are one condition.

**A world with no atmosphere and a hydrographic percentage (pp. 4, 8).**
The automatic zero belongs to planetary size, not to atmosphere. P. 4
gives "a planetary size of O or 1" an "automatic result of 0", and gives
an atmosphere of "0, 1, or greater than 9" a DM of -4; the p. 12 checklist
writes the same split. So an airless world can carry surface liquid.

It is uncommon rather than freakish -- about one world in sixty -- and it
cannot be very wet. An atmosphere of 0 needs 2D - 7 + size at or below
zero, so the size is 5 or less; hydrographics is then 2D - 11 + size,
which reaches size + 1. Six is the ceiling, at a size-5 world that throws
2 for its atmosphere and 12 for its water.

It looks like a contradiction, and p. 8 answers it: "Common sense should
rule in such cases; either the players or referee will generate a
rationale which explains the situation, or an alternative description
should be made."

Worth noting where the book's own help stops. P. 4 glosses what the liquid
might be -- "for normal worlds, this will be water; on other worlds (with
corrosive or exotic atmospheres), it may instead be other liquids of
fluids such as ammonia" -- and that gloss is offered for the _high_ end of
the same DM, atmospheres A through C. The low end gets the DM and no
gloss. What covers a vacuum world is the referee's rationale, not a
printed one, and the tool supplies none: the throw is made and the result
recorded.

**The planetary size table is wider than the throw that fills it
(pp. 4, 5).** P. 4 gives size as 2D − 2, "the resulting value ranges from
0 to 10," while the p. 5 table prints rows 0 through C (12000 miles
diameter). Rows B and C are unreachable by the printed formula. They are
transcribed anyway, because the table is the table; nothing generates
into them.

**The em-dash and minus glyphs in the held PDFs.** Not a rule, but the
trap most likely to put a wrong table in this repository. The PDFs'
embedded font maps the em-dash to the glyph `4` and the minus sign to
`3`, so a text extraction renders the jump routes table's empty cells as
the digit 4 — `B-E 4 4 4 4` for `4 — — —` — and the size formula `2D − 2`
as `2D32`. Both readings are wrong and both look like data. Every table
is transcribed from a visual read of the page, and then a second time
inside the table package's tests.

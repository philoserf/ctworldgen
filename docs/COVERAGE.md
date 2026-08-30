# COVERAGE

Maps every step and rule of Book 3 pp. 1–12 — the Worlds chapter — to its
page cite, implementation, and test. Every such rule is mapped and
implemented. Cites are Book 3 printed pages unless marked B1.

Two stretches of that range are deliberately absent.

**Pp. 10–11, the technological levels tables.** They say what a
technological index _means_: which weapons, armour, computers, vehicles,
and drives a world can build at each level. That is referee reference
during play, not a step of generation, and generating a subsector is the
whole of this tool (docs/PRD.md, Non-goals). Only the index's printed range
of 0–18 is read from them, and it is read as the clamp ceiling
(docs/ERRATA.md E004).

**The referee's own latitude on p. 1 and p. 8.** The book invites the
referee to write his own starports table, to impose worlds deliberately
rather than roll them, to detail the territories of a balkanized world
separately, and to place alternate world forms — ringworlds and the rest —
which it states are "not included in the world creation sequence." None of
these is a mechanic with a throw and a table; each is an instruction to the
referee. The one piece of that latitude the book _does_ make mechanical, the
±1 world occurrence DM, is implemented as an input.

## Implemented

| Rule                                                                                                                              | Page    | Implementation                                                   | Test                                                                                                               |
| --------------------------------------------------------------------------------------------------------------------------------- | ------- | ---------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Subsector grid: eight columns of ten rows, numbered 0101–0810                                                                     | 3       | `worldgen/hex.go` `AllHexes`, `ParseHex`                         | `TestAllHexesIsTheWholeGridInScanOrder`, `TestParseHexRejectsWhatIsNotOnTheGrid`                                   |
| One hex is one parsec; distance on the printed offset layout                                                                      | 1, 3    | `worldgen/hex.go` `Distance`                                     | `TestDistanceMatchesThePrintedGrid`                                                                                |
| World occurrence: one die per hex, a world on 4, 5 or 6                                                                           | 1       | `worldgen/engine.go` `occurrence`                                | golden fixtures, `TestWorldsAreInScanOrderAndOnTheGrid`                                                            |
| Hex scan order (reading E002)                                                                                                     | 1       | `worldgen/hex.go` `AllHexes`, `worldgen/engine.go` `occurrence`  | `TestWorldsAreInScanOrderAndOnTheGrid`                                                                             |
| The referee's subsector-wide occurrence DM of +1 or −1, and no other                                                              | 1       | `worldgen/engine.go` `Generate`, `occurrence`                    | `TestOccurrenceDMIsBoundedByThePage`, `TestNewRefusesAnOccurrenceDMThePageDoesNotOffer`                            |
| Starports table: two dice, 2–4 A, 5–6 B, 7–8 C, 9 D, 10–11 E, 12 X                                                                | 1       | `tables/data/starports.json`, `worldgen/engine.go` `starports`   | `TestStarportsTableMatchesThePage`                                                                                 |
| Starport chart: quality, fuel, overhaul and shipyard by type                                                                      | 5       | `tables/data/starports.json`, `render/render.go` `starportLine`  | `TestStarportChartIsComplete`, golden renders                                                                      |
| Naval and scout base throws, and their place in the stream (reading E001)                                                         | 1, 5    | `worldgen/engine.go` `bases`                                     | `TestBaseThrowsMatchTheStarportChart`, `TestBasesOnlyWhereTheChartPrintsAThrow`                                    |
| Jump routes table: one die per pair, four distance columns, em-dash cells                                                         | 2       | `tables/data/routes.json`, `worldgen/engine.go` `lane`           | `TestJumpRoutesTableMatchesThePage`, `TestRouteTableIsSymmetricAndBounded`                                         |
| Which pairs are examined, X starports excluded, each pair once, in order (reading E003)                                           | 2       | `worldgen/engine.go` `routes`, `lane`                            | `TestRoutesFollowThePrintedTable`                                                                                  |
| Planetary size: 2D − 2                                                                                                            | 4, 12   | `worldgen/worlds.go` `size`                                      | golden fixtures, `TestCharacteristicsStayInTheirPrintedRanges`                                                     |
| Planetary atmosphere: 2D − 7 + size; size 0 gives atmosphere 0 automatically                                                      | 4, 12   | `worldgen/worlds.go` `atmosphere`                                | `TestAutomaticValuesConsumeNoDie`                                                                                  |
| Hydrographics: 2D − 7 + size, −4 for atmosphere 0, 1 or above 9; size 0 or 1 gives 0 automatically                                | 4, 12   | `worldgen/worlds.go` `hydrographics`, `vacuumOrExoticAtmosphere` | `TestAutomaticValuesConsumeNoDie`                                                                                  |
| Population: 2D − 2, an exponent of 10                                                                                             | 8, 12   | `worldgen/worlds.go` `population`                                | golden fixtures                                                                                                    |
| Planetary government: 2D − 7 + population                                                                                         | 8, 12   | `worldgen/worlds.go` `government`                                | golden fixtures                                                                                                    |
| Law level: 2D − 7 + government type                                                                                               | 8, 12   | `worldgen/worlds.go` `lawLevel`                                  | golden fixtures                                                                                                    |
| Technological index: one die plus the matrix's six columns of DMs                                                                 | 9       | `worldgen/worlds.go` `technologicalIndex`, `techDMs`             | `TestTechnologicalIndexMatrixMatchesThePage`                                                                       |
| Clamping every value to its table's printed range; law level floored only (reading E004)                                          | 4–9     | `worldgen/log.go` `clamped`, `floored`, `clampRange`             | `TestCharacteristicsStayInTheirPrintedRanges`, `TestClampsAreStampedAndLogged`, `TestMaxValuesAreThePrintedRanges` |
| Every clamped value has a row in the technological index matrix                                                                   | 9       | `tables/load.go` `checkTechComplete`                             | `TestEveryClampedValueHasAMatrixRow`                                                                               |
| The descriptive tables of size, atmosphere, hydrographics, population, government and law level                                   | 5–7     | `tables/data/characteristics.json`                               | `TestLabelsCoverEveryValueAndStopThere`                                                                            |
| The string of digits: order, alphabet, no separator (reading E005)                                                                | 4; B1 8 | `worldgen/worlds.go` `profile`, `tables/tables.go` `Digit`       | `TestProfileRestatesTheCharacteristics`, `TestDigitNotation`                                                       |
| Naming each world: the referee's, the book printing no table                                                                      | 12      | `worldgen/record.go` `World.Name`, `render/render.go`            | `TestListingNamesAWorldThatHasOne`, `TestReplayToleratesRefereeNames`                                              |
| Die roll conventions: one and two dice, DMs, N+/N−/exact targets                                                                  | B1 2–3  | `dice/dice.go`                                                   | `dice` package tests                                                                                               |
| A subsector where no hex holds a world is a valid record                                                                          | 1       | `worldgen/engine.go` `occurrence`, `render/render.go`            | `TestListingOfAnEmptySubsector`                                                                                    |
| Record: the subsector listing renders from the canonical JSON, `--history` for the transcript                                     | —       | `render/render.go`, `cmd/ctworldgen/main.go` `runRender`         | `TestGoldenRenders`, `TestRenderAndReplayReadARecord`                                                              |
| Batch: member seeds = base + index, JSONL or per-file                                                                             | —       | `cmd/ctworldgen/main.go` `runBatch`                              | `TestBatchDerivesEachSeedFromTheBase`, `TestBatchToADirectory`                                                     |
| Replay: recompute every throw, non-zero exit at the first mismatch; `--ignore-provenance` waives the four stamps and nothing else | —       | `worldgen/replay.go`, `cmd/ctworldgen/main.go` `runReplay`       | `TestReplayRoundTrip`, `TestReplayChecksProvenance`, `TestIgnoreProvenanceDoesNotWaiveDivergence`                  |
| Documents held to the code: ERRATA ids stamped and documented both ways, COVERAGE test names                                      | —       | `docs/ERRATA.md`, `docs/COVERAGE.md`                             | `TestErrataIDsMatchTheDocument`, `TestErrataEntriesCiteAPage`, `TestCoverageNamesRealTests`                        |

All four PRD milestones are implemented; nothing is pending.

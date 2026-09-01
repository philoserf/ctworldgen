# COVERAGE

Every rule of Book 3 pp. 1-12 mapped to its printed page, the code that
implements it, and the test that holds it. Page cites are Book 3 printed
pages unless marked B1. Printed page N is PDF page N+5 in Book 3, N+6 in
Book 1.

Status is one of **done**, **data only** (the chart is transcribed,
validated and tested, but no step consumes it yet), or **not built**, with
the reason and where it sits in the backlog. A page inside pp. 1-12 that
this tool does not implement has a row saying so; the gap in the document
is otherwise indistinguishable from an oversight.

This document is living. A rule is not implemented until it has a row here
with a test named in it.

## Star mapping (pp. 1-3)

| Rule                                                   | Page | Implementation                                                        | Test                                                                                                                       | Status |
| ------------------------------------------------------ | ---- | --------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ------ |
| R1 The subsector grid, 0101-0810, one hex one parsec   | 1, 3 | `starmap.Hex`, `starmap.Columns`/`Rows`                               | `TestNewHexRejectsOffGrid`, `TestParseHex`, `TestHexRoundTripsThroughJSON`                                                 | done   |
| R1 Hex distance, even columns half a hex lower         | 3    | `starmap.Hex.cube`, `Hex.Distance`                                    | `TestDistanceAgainstPrintedGrid` -- measured by hand off the printed grid                                                  | done   |
| R1 The grid drawn: the map in the listing              | 1, 3 | `render.Renderer.grid`, `render.gridLine`                             | `TestTheMapIsTheGridPrintedOnPageThree` -- held against the page-anchored distance, `TestTheMapMarksWhatPageOneSaysToMark` | done   |
| R1 The grid drawn: the map in the booklet              | 1, 3 | `render.fitMap`, `render.mapFit.hexCenter`, `render.Renderer.Booklet` | `TestTheDrawnMapIsTheGridPrintedOnPageThree` -- measured off the drawing, `TestEveryWorldIsDrawnInItsOwnHex`               | done   |
| R1a The line drawn between the worlds a route joins    | 2    | `render.booklet.drawRoutes`                                           | `TestEveryRouteIsDrawn`                                                                                                    | done   |
| R2 World occurrence, one die at 4+, DM -1/0/+1         | 1    | `gen.occurrenceTarget`, `gen.scan`                                    | `TestOccurrenceDMChangesTheStream`, `TestRejectsDMsTheBookDoesNotOffer`, `TestInvariantsOverManySeeds`                     | done   |
| R3 Starport type, two dice against the starports table | 1    | `tables.Starports.Type`, `gen.Generate`                               | `TestStarportsTable`, `TestStarportDistributionFollowsThePage`                                                             | done   |
| R4 Naval and scout base throws                         | 5    | `tables.StarportChart.NavalBase`/`ScoutBase`, `gen.Engine.Generate`   | `TestStarportChartBaseThrows`, `TestBaseThrowsFollowTheChart`, `assertBasesFollowTheChart`                                 | done   |
| R5 Routes, one die against the jump routes table       | 2    | `tables.JumpRoutes.Target`, `tables.MaxJump`, `gen.Engine.routes`     | `TestJumpRoutesTable`, `TestJumpRoutesHasNoRowForX`, `assertRoutesFollowTheTable`, `assertRoutesTheTableMakesCertain`      | done   |

## World creation (pp. 4-9)

| Rule                                                     | Page    | Implementation                                                | Test                                                                                                                                           | Status |
| -------------------------------------------------------- | ------- | ------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| R6 Planetary size, 2D-2                                  | 4, 12   | `gen.Engine.detail`                                           | `assertWithinFormula`                                                                                                                          | done   |
| R7 Planetary atmosphere, 2D-7+size                       | 4, 12   | `gen.Engine.detail`                                           | `assertWithinFormula`, `assertAutomaticZeros`                                                                                                  | done   |
| R8 Hydrographic percentage, 2D-7+size, DM -4             | 4, 12   | `gen.Engine.detail`                                           | `assertWithinFormula`, `assertAutomaticZeros`                                                                                                  | done   |
| R9 Population, 2D-2                                      | 8, 12   | `gen.Engine.detail`                                           | `assertWithinFormula`                                                                                                                          | done   |
| R10 Planetary government, 2D-7+population                | 8, 12   | `gen.Engine.detail`                                           | `assertWithinFormula`                                                                                                                          | done   |
| R11 Law level, 2D-7+government                           | 8, 12   | `gen.Engine.detail`                                           | `assertWithinFormula`                                                                                                                          | done   |
| R12 Technological index, one die plus the matrix DMs     | 9       | `gen.Engine.detail`, `tables.TechIndexMatrix`                 | `TestTechnologicalIndexMatrix`, `assertTechIndexIsTheMatrix`                                                                                   | done   |
| R13 The two throws not made                              | 4       | `gen.Engine.detail`                                           | `assertAutomaticZeros`                                                                                                                         | done   |
| R14 Floored at 0; the technological index capped at 18   | 4-9     | `gen.clamp`, `starmap.Clamp`                                  | `assertClampsAreHonest`, `TestTheClampsThatBindAreTheOnesR14Names`                                                                             | done   |
| R15 The string of digits, eight characters, no separator | 4, B1 8 | `starmap.World.DigitString`                                   | `TestDigitAlphabet`, `assertDigitsSpellTheWorld`                                                                                               | done   |
| R16 The descriptive tables                               | 5-7     | `tables.Labels`, `render.bullets` -- one list, both documents | `TestEveryPrintedValueHasALabel`, `TestLabelsComeFromTheTables` (the listing), `TestTheBookletsBulletsCarryTheirLabelsAndTables` (the booklet) | done   |

## Technological levels (pp. 10-11)

These two pages carry no rule and so get no R-number, but they are inside
pp. 1-12 and so get a row. They were declared out of scope by the
retired PRD; issue 1 #4 reopened them, and they are now simply not built.

| Pages                                      | Page  | Implementation                       | Test                                               | Status                  |
| ------------------------------------------ | ----- | ------------------------------------ | -------------------------------------------------- | ----------------------- |
| The technological levels tables, rows 0-18 | 10-11 | none; `render.techIndexNote` says so | `TestTheListingSaysWhyTheTechnologicalIndexIsBare` | not built -- issue 1 #4 |

Two tables, rows 0 through 18: Weapons (Personal, Armor, Special,
Computers, Communication) on p. 10 and Transportation (Water, Land, Air,
Space, Fuels) on p. 11. They say what a technological index _means_ during
play -- what a world can build -- and they generate nothing: R12 produces
the index from the p. 9 matrix, and no step of pp. 1-12 reads back from
these two pages. So they are not transcribed, and the listing prints the
index digit with no description and one sentence saying why.

The pages themselves are printed incomplete, and say so: "The
technological level tables have several spaces or holes, and such gaps
should be filled in by the referee or the players when they discover items
or devices of interest" (p. 11). That is the invitation p. 8 makes for the
descriptive tables (E004), made again, and answered here the same way: by
the referee at the table, not by the tool.

## The tool

| Rule                                                            | Page   | Implementation                                                                      | Test                                                                                                                                                                                                                                                                          | Status |
| --------------------------------------------------------------- | ------ | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| R17 Dice engine, one and two dice, cumulative DMs, N+ targets   | B1 2-3 | `dice.Stream.Die`/`D2`, `dice.Sum`, `dice.Target`                                   | `TestDieIsOneDie`, `TestD2ConsumesTwoDiceInOrder`, `TestTargetIsNPlus`, `TestSumIsCumulative`                                                                                                                                                                                 | done   |
| R18 The record and its provenance                               | --     | `starmap.Record`, `starmap.World`, `starmap.Route`, `starmap.New`, `starmap.Decode` | `TestNewRecordCarriesItsProvenance`, `TestGoldensValidate`, `TestDecodeRejectsUnknownFields`, `TestDecodeRejectsMoreThanOneDocument`, `TestDecodeRejectsAHexOffTheRecordsGrid`, `TestDecodeRejectsARouteEndOffTheRecordsGrid`                                                 | done   |
| R18a The read path holds a record to the schema it claims       | --     | `starmap.Record.carriesThisToolsProvenance`, `carriesTheFieldsTheSchemaRequires`    | `TestDecodeRejectsAnotherToolsProvenance`, `TestDecodeRejectsARecordMissingARequiredField`, `TestDecodeRejectsAWorldWithNoStarport`, `TestAnEmptyArrayIsNotAMissingOne`, `TestARecordWithNoGridIsASubsector`                                                                  | done   |
| R20 The listing                                                 | 4      | `render.Renderer`                                                                   | `TestListings`, `TestEveryWorldAndRouteReachesTheListing`, `TestLabelsComeFromTheTables`, `TestARefereesNameReachesEveryPlaceTheHexAppears`, `TestAnEmptySubsectorRenders`                                                                                                    | done   |
| R21 Sector: sixteen subsectors on one grid, routes at the seams | 1-2    | `gen.Engine.Sector`, `starmap.Place`, `starmap.MemberOf`, `starmap.MemberBounds`, `cmd/ctworldgen.sectorCmd`        | `TestASectorsMembersAreTheSubsectorsNewWrites`, `TestASectorsMembersKeepTheirOwnRoutes`, `TestASectorIsInTheOrderPageTwoReads`, `TestRoutesCrossTheSeams`, `TestTranslationKeepsThePageThreeParity`, `TestEveryPairIsExaminedOnce`, `TestTheSeamsGolden`, `TestTheMemberBandsAreLeftToRightThenDown`, `TestMemberBoundsAreTheCornersOfTheBand`, `TestTheListingSaysWhichGridItDrew`, `TestASectorRecordValidates` | done   |
| R23 Which lanes the documents draw | 2 | `render.legible`, `render.Lanes`, `cmd/ctworldgen.holdLanes` | `TestTheDefaultListingDrawsLegibleLanes`, `TestLegibleLanesDoNotDependOnRouteOrder`, `TestLegibleLanesKeepEveryWorldReachable`, `TestEveryJumpOneLaneIsDrawn`, `TestTheSummaryReportsWhatWasNotDrawn` | done |
| R24 A sector's documents: the sixteen sub-sectors, the ring, the index map | 1-4 | `render.member`, `render.members`, `render.memberSection`, `render.sectorIndex`, `render.indexMap`, `render.onMemberMap`, `render.memberWindow`, `render.window`, `booklet.memberPages`, `booklet.drawIndex`, `booklet.drawSeams`, `booklet.drawMemberNumbers` | `TestASectorsMemberMapsAreThePrintedGrid`, `TestASectorsMemberSectionsPartitionItsWorlds`, `TestACrossingLaneIsListedUnderBothItsSubsectors`, `TestTheBookletsMemberMapsAreTheGridPrintedOnPageThree`, `TestTheSectorIndexCarriesNoHexNumbers`, `TestTheSectorIndexNumbersItsSixteenBands`, `TestTheSectorSliceGolden` | done |
| R22 The referee's own notes, on a world and on the record | -- | `starmap.Record.Notes`, `starmap.World.Notes`, `render.bullets`, `render.Renderer.heading` | `TestNotesRoundTripAndDoNotLoosenTheRecord`, `TestARecordWithNoNotesIsUnchanged`, `TestTheRefereesNotesReachBothDocuments` | done |

## Determinism and provenance

| Guarantee                                          | Implementation                               | Test                                                               |
| -------------------------------------------------- | -------------------------------------------- | ------------------------------------------------------------------ |
| One seed fills both PCG words                      | `dice.NewStream`                             | `TestSameSeedSameStream`, `TestSameSeedSameSubsector`              |
| One die is one `IntN(6)` draw; 2D is two, in order | `dice.Stream.Die`, `dice.Stream.D2`          | `TestD2ConsumesTwoDiceInOrder`                                     |
| Consumption order is pinned                        | `gen.Generate`                               | `TestGoldens`                                                      |
| A seed is always recorded                          | `cmd/ctworldgen.newCmd`                      | `TestASeedIsAlwaysRecorded`, `TestSeedZeroIsAChoice`               |
| A record reproduces from its own seed and inputs   | --                                           | `TestRegenerationRoundTrip`                                        |
| The shipped example is a record the engine wrote   | `starmap.Marshal`, `internal/cmd/regenerate` | `TestTheCompleteExampleIsAGeneratedRecord`                         |
| Unknown fields are rejected on both sides          | `starmap.Decode`, `docs/record.schema.json`  | `TestDecodeRejectsUnknownFields`, `TestSchemaRejectsUnknownFields` |
| Coverage does not drift                            | `internal/cmd/ratchet`                       | `TestCompareFailsInBothDirections`                                 |

## Readings

Each entry of `ERRATA.md` states its own stamping condition. The record's
`errata` array grows with the milestones the way the schema does: a
reading is stamped only once the engine implements the step it governs.

| Reading | Governs                                                           | Cited in                                               | Stamped                                      |
| ------- | ----------------------------------------------------------------- | ------------------------------------------------------ | -------------------------------------------- |
| E001    | Where the base throws sit in the procedure                        | `gen/gen.go`, `tables/tables.go`                       | where any base throw was made, now           |
| E002    | The order of the passes, hexes and characteristics                | `gen/gen.go`, `starmap/hex.go`                         | every record, now                            |
| E003    | Which pairs are examined, when a die is thrown, in what order     | `gen/gen.go`, `tables/tables.go`                       | records with two or more worlds, now         |
| E004    | Floored at 0, capped only at the technological index              | `gen/gen.go`, `tables/tables.go`, `render/render.go`   | where a floor or the cap actually bound, now |
| E005    | The string of digits: order, alphabet, no separator               | `starmap/record.go`, `starmap/digit.go`                | records with at least one world, now         |
| E006    | Sixteen subsectors on one grid, and the route pass at their seams | `gen/sector.go`, `starmap/hex.go`, `starmap/record.go` | every sector record                          |
| E007    | Which lanes the documents draw                                    | `render/lanes.go`, `render/render.go`, `cmd/ctworldgen/main.go` | never -- it governs the documents            |
| E008    | How a sector's documents present its sixteen sub-sectors          | `render/sector.go`, `render/layout.go`, `render/render.go`, `render/pdf.go` | never -- it governs the documents            |

Both directions are checked by `internal/audit`: every `E00N` cited in the
code or the documents resolves to a heading in `ERRATA.md`, and every
heading is cited at least once.

## The font trap

Every table in `tables/data/` was transcribed from a **visual** read of the
page and then transcribed a second time inside `tables/tables_test.go`.
The two must agree. The held PDFs' embedded font maps the em-dash to the
glyph `4` and the minus sign to `3`, so a text extraction renders the jump
routes table's empty cells as the digit 4 and the size formula `2D - 2` as
`2D32`. Both are wrong and both look like data. No table in this
repository may be produced by extraction.

The one exception is the descriptive **labels** of pp. 5-7, which are
editorial abbreviations of the book's prose rather than transcriptions of
it: retyping them in a test would compare an abbreviation against itself.
They are checked by re-reading the page, and the suite asserts only that
every value in a printed range has a non-empty label.

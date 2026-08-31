# COVERAGE

Every rule of Book 3 pp. 1-12 mapped to its printed page, the code that
implements it, and the test that holds it. Page cites are Book 3 printed
pages unless marked B1. Printed page N is PDF page N+5 in Book 3, N+6 in
Book 1.

Status is one of **done**, **data only** (the chart is transcribed,
validated and tested, but no step consumes it yet), or **not yet**, with
the milestone that will do it.

This document is living. A rule is not implemented until it has a row here
with a test named in it.

## Star mapping (pp. 1-3)

| Rule                                                   | Page | Implementation                               | Test                                                                                                   | Status                  |
| ------------------------------------------------------ | ---- | -------------------------------------------- | ------------------------------------------------------------------------------------------------------ | ----------------------- |
| R1 The subsector grid, 0101-0810, one hex one parsec   | 1, 3 | `subsector.Hex`, `subsector.Columns`/`Rows`  | `TestNewHexRejectsOffGrid`, `TestParseHex`, `TestHexRoundTripsThroughJSON`                             | done                    |
| R1 Hex distance, even columns half a hex lower         | 3    | `subsector.Hex.cube`, `Hex.Distance`         | `TestDistanceAgainstPrintedGrid` -- measured by hand off the printed grid                              | done                    |
| R2 World occurrence, one die at 4+, DM -1/0/+1         | 1    | `gen.occurrenceTarget`, `gen.scan`           | `TestOccurrenceDMChangesTheStream`, `TestRejectsDMsTheBookDoesNotOffer`, `TestInvariantsOverManySeeds` | done                    |
| R3 Starport type, two dice against the starports table | 1    | `tables.Starports.Type`, `gen.Generate`      | `TestStarportsTable`, `TestStarportDistributionFollowsThePage`                                         | done                    |
| R4 Naval and scout base throws | 5 | `tables.StarportChart.NavalBase`/`ScoutBase`, `gen.Engine.Generate` | `TestStarportChartBaseThrows`, `TestBaseThrowsFollowTheChart`, `assertBasesFollowTheChart` | done |
| R5 Space lanes, one die against the jump routes table | 2 | `tables.JumpRoutes.Target`, `tables.MaxJump`, `gen.Engine.lanes` | `TestJumpRoutesTable`, `TestJumpRoutesHasNoRowForX`, `assertLanesFollowTheTable`, `assertLanesTheTableMakesCertain` | done |

## World creation (pp. 4-9)

| Rule                                                     | Page    | Implementation                                 | Test                                                               | Status                  |
| -------------------------------------------------------- | ------- | ---------------------------------------------- | ------------------------------------------------------------------ | ----------------------- |
| R6 Planetary size, 2D-2 | 4, 12 | `gen.Engine.detail` | `assertWithinFormula` | done |
| R7 Planetary atmosphere, 2D-7+size | 4, 12 | `gen.Engine.detail` | `assertWithinFormula`, `assertAutomaticZeros` | done |
| R8 Hydrographic percentage, 2D-7+size, DM -4 | 4, 12 | `gen.Engine.detail` | `assertWithinFormula`, `assertAutomaticZeros` | done |
| R9 Population, 2D-2 | 8, 12 | `gen.Engine.detail` | `assertWithinFormula` | done |
| R10 Planetary government, 2D-7+population | 8, 12 | `gen.Engine.detail` | `assertWithinFormula` | done |
| R11 Law level, 2D-7+government | 8, 12 | `gen.Engine.detail` | `assertWithinFormula` | done |
| R12 Technological index, one die plus the matrix DMs | 9 | `gen.Engine.detail`, `tables.TechIndexMatrix` | `TestTechnologicalIndexMatrix`, `assertTechIndexIsTheMatrix` | done |
| R13 The two throws not made | 4 | `gen.Engine.detail` | `assertAutomaticZeros` | done |
| R14 Floored at 0; the technological index capped at 18 | 4-9 | `gen.clamp`, `subsector.Clamp` | `assertClampsAreHonest`, `TestTheClampsThatBindAreTheOnesR14Names` | done |
| R15 The string of digits, eight characters, no separator | 4, B1 8 | `subsector.World.DigitString` | `TestDigitAlphabet`, `assertDigitsSpellTheWorld` | done |
| R16 The descriptive tables                               | 5-7     | `tables.Labels`, `tables.StarportChart.Row`    | `TestEveryPrintedValueHasALabel`, `TestStarportChartDescriptions`  | data only (milestone 4) |

## The tool

| Rule                                                          | Page   | Implementation                                             | Test                                                                                          | Status                    |
| ------------------------------------------------------------- | ------ | ---------------------------------------------------------- | --------------------------------------------------------------------------------------------- | ------------------------- |
| R17 Dice engine, one and two dice, cumulative DMs, N+ targets | B1 2-3 | `dice.Stream.Die`/`D2`, `dice.Sum`, `dice.Target`          | `TestDieIsOneDie`, `TestD2ConsumesTwoDiceInOrder`, `TestTargetIsNPlus`, `TestSumIsCumulative` | done                      |
| R18 The subsector record and its provenance | -- | `subsector.Subsector`, `subsector.World`, `subsector.Route`, `subsector.New`, `subsector.Decode` | `TestNewRecordCarriesItsProvenance`, `TestGoldensValidate`, `TestDecodeRejectsUnknownFields` | grows with each milestone |
| R19 Batch                                                     | --     | --                                                         | --                                                                                            | not yet (milestone 4)     |
| R20 The subsector listing                                     | 4      | --                                                         | --                                                                                            | not yet (milestone 4)     |

## Determinism and provenance

| Guarantee                                          | Implementation                                   | Test                                                  |
| -------------------------------------------------- | ------------------------------------------------ | ----------------------------------------------------- |
| One seed fills both PCG words                      | `dice.NewStream`                                 | `TestSameSeedSameStream`, `TestSameSeedSameSubsector` |
| One die is one `IntN(6)` draw; 2D is two, in order | `dice.Stream.Die`, `dice.Stream.D2`              | `TestD2ConsumesTwoDiceInOrder`                        |
| Consumption order is pinned                        | `gen.Generate`                                   | `TestGoldens`                                         |
| A seed is always recorded                          | `cmd/ctworldgen.newCmd`                          | `TestASeedIsAlwaysRecorded`, `TestSeedZeroIsAChoice`  |
| A record reproduces from its own seed and inputs   | --                                               | `TestRegenerationRoundTrip`                           |
| The shipped example is a record the engine wrote   | `subsector.Marshal`, `internal/cmd/regenerate`   | `TestTheCompleteExampleIsAGeneratedRecord`            |
| Unknown fields are rejected on both sides          | `subsector.Decode`, `docs/subsector.schema.json` | `TestDecodeRejectsUnknownFields`, `TestSchemaRejectsUnknownFields` |
| Coverage does not drift                            | `internal/cmd/ratchet`                           | `TestCompareFailsInBothDirections`                    |

## Readings

Each entry of `ERRATA.md` states its own stamping condition. The record's
`errata` array grows with the milestones the way the schema does: a
reading is stamped only once the engine implements the step it governs.

| Reading | Governs                                                       | Cited in                                | Stamped           |
| ------- | ------------------------------------------------------------- | --------------------------------------- | ----------------- |
| E001 | Where the base throws sit in the procedure | `gen/gen.go`, `tables/tables.go` | where any base throw was made, now |
| E002    | The order of the passes, hexes and characteristics            | `gen/gen.go`, `subsector/hex.go`        | every record, now |
| E003 | Which pairs are examined, when a die is thrown, in what order | `gen/gen.go`, `tables/tables.go` | records with two or more worlds, now |
| E004 | Floored at 0, capped only at the technological index | `gen/gen.go`, `tables/tables.go` | where a floor or the cap actually bound, now |
| E005 | The string of digits: order, alphabet, no separator | `subsector/record.go`, `subsector/digit.go` | records with at least one world, now |

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

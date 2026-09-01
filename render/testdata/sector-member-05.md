## Subsector 5 &mdash; 0911 to 1620

Member 5 of this sector. `ctworldgen new --seed 6` writes this sub-sector on its own p. 3 grid, where its first hex is 0101; here it is laid on the sector's, and that hex is 0911 (ERRATA E006).

42 worlds, 85 lanes within it, 63 crossing into its neighbours.

### The map

This sub-sector's eighty hexes, numbered where the sector lays them, with one hex of each neighbour drawn around the edge, so that a lane crossing a seam to an adjacent hex has a far end that can be looked up (ERRATA E008 part 2). A lane runs up to four parsecs (p. 2) and this map reaches one, so 1 of its lanes end past the edge and are drawn here as no line; the table below carries them, and names the sub-sector each one leaves for. The odd-numbered columns sit high and the even-numbered ones half a hex below them, which is how the page prints it. A world carries the letter of its starport -- p. 1 marks the hex with the letter the starports table gives -- and a hex with no world is left blank, which is what p. 1 says to leave it. P. 2 also asks for a line drawn between the worlds a route joins; a monospace grid has nowhere to put one, so this map draws none and the route table below carries them instead. `render --format pdf` draws them.

```text
     0910 X    1110 C    1310      1510
0810 C    1010 E    1210 C    1410 B    1610 A
     0911 X    1111 E    1311 A    1511 E    1711 D
0811      1011      1211 A    1411 A    1611
     0912      1112      1312 D    1512      1712
0812 C    1012 E    1212 C    1412      1612 B
     0913      1113 B    1313      1513 E    1713 A
0813 C    1013 E    1213 A    1413      1613 D
     0914 B    1114      1314 X    1514      1714
0814      1014      1214 B    1414      1614
     0915 E    1115 E    1315      1515 E    1715
0815      1015 C    1215      1415 E    1615 C
     0916 A    1116 C    1316      1516 A    1716
0816 E    1016      1216 D    1416 A    1616 C
     0917      1117 E    1317      1517      1717 B
0817 C    1017      1217 B    1417      1617
     0918      1118      1318      1518 A    1718
0818 B    1018 A    1218 C    1418      1618
     0919      1119      1319      1519      1719
0819 C    1019 C    1219 A    1419      1619
     0920      1120 B    1320 E    1520 C    1720 C
0820      1020      1220 B    1420 A    1620
     0921 C    1121      1321 C    1521 A    1721 C
          1021      1221      1421      1621
```

### Worlds

| Hex | Name | Digits | Bases |
| --- | --- | --- | --- |
| 0911 |  | X10048A6 | -- |
| 0914 |  | B3543008 | -- |
| 0915 |  | E4245967 | -- |
| 0916 |  | A310544B | naval |
| 1012 |  | E6168882 | -- |
| 1013 |  | E1401048 | -- |
| 1015 |  | C230642A | scout |
| 1018 |  | A300598B | naval, scout |
| 1019 |  | C6943219 | -- |
| 1111 |  | E7775744 | -- |
| 1113 |  | B7773547 | -- |
| 1115 |  | E3536417 | -- |
| 1116 |  | C3756326 | -- |
| 1117 |  | E6893645 | -- |
| 1120 |  | B341534A | -- |
| 1211 |  | A689212C | -- |
| 1212 |  | C52356B6 | scout |
| 1213 |  | A211005E | -- |
| 1214 |  | B4113509 | scout |
| 1216 |  | D6583566 | scout |
| 1217 |  | B53148B9 | -- |
| 1218 |  | C000546C | scout |
| 1219 |  | A4558A6A | -- |
| 1220 |  | B3108689 | naval |
| 1311 |  | A220440E | naval, scout |
| 1312 |  | D69878A2 | scout |
| 1314 |  | X0004584 | -- |
| 1320 |  | E4203265 | -- |
| 1411 |  | A322465D | naval |
| 1415 |  | E8887443 | -- |
| 1416 |  | A8777458 | -- |
| 1420 |  | A3256139 | naval, scout |
| 1511 |  | E6786696 | -- |
| 1513 |  | E6341467 | -- |
| 1515 |  | E63A455B | -- |
| 1516 |  | A678333B | scout |
| 1518 |  | A100047B | scout |
| 1520 |  | C9591135 | -- |
| 1612 |  | B45298A8 | -- |
| 1613 |  | D56A7956 | -- |
| 1615 |  | C7785695 | scout |
| 1616 |  | C6661007 | -- |

### Routes

78 of these 148 lanes are not listed: each joins two worlds already joined by shorter lanes, which p. 2 says may be ignored in the drawing (ERRATA E007). The record carries every one of them, and `render --lanes=all` lists them.

| From | To | Parsecs | Into |
| --- | --- | --- | --- |
| 0813 | 0914 | 1 | subsector 4 |
| 0816 | 0916 | 1 | subsector 4 |
| 0817 | 1018 | 2 | subsector 4 |
| 0818 | 1018 | 2 | subsector 4 |
| 0818 | 1019 | 2 | subsector 4 |
| 0914 | 0916 | 2 | -- |
| 0914 | 1015 | 2 | -- |
| 0914 | 1113 | 2 | -- |
| 0915 | 0916 | 1 | -- |
| 0916 | 1015 | 1 | -- |
| 0921 | 1120 | 2 | subsector 9 |
| 1012 | 1113 | 1 | -- |
| 1018 | 1019 | 1 | -- |
| 1019 | 1120 | 1 | -- |
| 1110 | 1111 | 1 | subsector 1 |
| 1111 | 1210 | 1 | subsector 1 |
| 1111 | 1211 | 1 | -- |
| 1113 | 1213 | 1 | -- |
| 1115 | 1116 | 1 | -- |
| 1116 | 1214 | 2 | -- |
| 1116 | 1217 | 2 | -- |
| 1117 | 1217 | 1 | -- |
| 1120 | 1219 | 1 | -- |
| 1120 | 1220 | 1 | -- |
| 1210 | 1211 | 1 | subsector 1 |
| 1210 | 1311 | 1 | subsector 1 |
| 1211 | 1212 | 1 | -- |
| 1211 | 1311 | 1 | -- |
| 1211 | 1312 | 1 | -- |
| 1212 | 1213 | 1 | -- |
| 1213 | 1214 | 1 | -- |
| 1216 | 1217 | 1 | -- |
| 1217 | 1218 | 1 | -- |
| 1217 | 1416 | 2 | -- |
| 1218 | 1219 | 1 | -- |
| 1219 | 1220 | 1 | -- |
| 1219 | 1320 | 1 | -- |
| 1220 | 1320 | 1 | -- |
| 1220 | 1321 | 1 | subsector 9 |
| 1311 | 1312 | 1 | -- |
| 1311 | 1410 | 1 | subsector 1 |
| 1311 | 1411 | 1 | -- |
| 1312 | 1411 | 1 | -- |
| 1320 | 1420 | 1 | -- |
| 1321 | 1420 | 1 | subsector 9 |
| 1410 | 1411 | 1 | subsector 1 |
| 1411 | 1511 | 1 | -- |
| 1411 | 1612 | 2 | -- |
| 1415 | 1416 | 1 | -- |
| 1415 | 1516 | 1 | -- |
| 1416 | 1516 | 1 | -- |
| 1416 | 1518 | 2 | -- |
| 1420 | 1422 | 2 | subsector 9 |
| 1420 | 1520 | 1 | -- |
| 1420 | 1521 | 1 | subsector 9 |
| 1511 | 1610 | 1 | subsector 1 |
| 1513 | 1612 | 1 | -- |
| 1515 | 1516 | 1 | -- |
| 1515 | 1615 | 1 | -- |
| 1516 | 1518 | 2 | -- |
| 1516 | 1615 | 1 | -- |
| 1516 | 1616 | 1 | -- |
| 1518 | 1520 | 2 | -- |
| 1518 | 1717 | 2 | subsector 6 |
| 1520 | 1521 | 1 | subsector 9 |
| 1610 | 1612 | 2 | subsector 1 |
| 1612 | 1613 | 1 | -- |
| 1612 | 1713 | 1 | subsector 6 |
| 1613 | 1713 | 1 | subsector 6 |
| 1616 | 1717 | 1 | subsector 6 |

### The worlds in detail

The technological index carries its digit and no description. The technological levels tables of pp. 10-11 say what an index means during play rather than how it is generated, so this tool does not read them; p. 11 asks the referee or the players to fill in their holes as play discovers them.

#### 0911 &mdash; X10048A6

- **Starport X.** No starport. No provision is made for any starship landings.
- **Size 1.** 1000 miles diameter.
- **Atmosphere 0.** No atmosphere.
- **Hydrographics 0.** No free standing water.
- **Population 4.** 10,000
- **Government 8.** Civil Service Bureaucracy. Ruling functions are performed by government agencies employing individuals selected for their expertise.
- **Law level A.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index 6.**
- **Bases.** --

#### 0914 &mdash; B3543008

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 4.** 40%
- **Population 3.** 1,000
- **Government 0.** No government structure. In many cases, family bonds will predominate.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index 8.**
- **Bases.** --
- **Clamped.** government threw -2 and is recorded as 0.
- **Clamped.** law_level threw -2 and is recorded as 0.

#### 0915 &mdash; E4245967

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 4.** 40%
- **Population 5.** 100,000
- **Government 9.** Impersonal Bureaucracy. Ruling functions are performed by agencies which have become insulated from the governed citizens.
- **Law level 6.** Most firearms (all except shotguns) are prohibited. The carrying of any type of weapon openly is discouraged.
- **Technological index 7.**
- **Bases.** --

#### 0916 &mdash; A310544B

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 0.** No free standing water.
- **Population 5.** 100,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index B.**
- **Bases.** naval
- **Clamped.** hydrographics threw -4 and is recorded as 0.

#### 1012 &mdash; E6168882

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 6.** 60%
- **Population 8.** 100,000,000
- **Government 8.** Civil Service Bureaucracy. Ruling functions are performed by government agencies employing individuals selected for their expertise.
- **Law level 8.** Long bladed weapons (all blade weapons except daggers) are strictly controlled. Open possession in public is prohibited. Ownership is, however, not restricted.
- **Technological index 2.**
- **Bases.** --

#### 1013 &mdash; E1401048

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 1.** 1000 miles diameter.
- **Atmosphere 4.** Thin, tainted.
- **Hydrographics 0.** No free standing water.
- **Population 1.** 10
- **Government 0.** No government structure. In many cases, family bonds will predominate.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index 8.**
- **Bases.** --

#### 1015 &mdash; C230642A

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 2.** 2000 miles diameter.
- **Atmosphere 3.** Very thin.
- **Hydrographics 0.** No free standing water.
- **Population 6.** 1,000,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 2.** Portable energy weapons, such as laser rifles or carbines are prohibited. Ship's gunnery is not affected.
- **Technological index A.**
- **Bases.** scout
- **Clamped.** hydrographics threw -3 and is recorded as 0.

#### 1018 &mdash; A300598B

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 0.** No atmosphere.
- **Hydrographics 0.** No free standing water.
- **Population 5.** 100,000
- **Government 9.** Impersonal Bureaucracy. Ruling functions are performed by agencies which have become insulated from the governed citizens.
- **Law level 8.** Long bladed weapons (all blade weapons except daggers) are strictly controlled. Open possession in public is prohibited. Ownership is, however, not restricted.
- **Technological index B.**
- **Bases.** naval, scout
- **Clamped.** hydrographics threw -1 and is recorded as 0.

#### 1019 &mdash; C6943219

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 9.** Dense, tainted.
- **Hydrographics 4.** 40%
- **Population 3.** 1,000
- **Government 2.** Participating Democracy. Ruling function decisions are reached by the advice and consent of the citizenry directly.
- **Law level 1.** Certain weapons are prohibited, including specifically 1) body pistols which are undetectable by standard detectors, 2) explosive weapons such as bombs or grenades, and 3) poison gas.
- **Technological index 9.**
- **Bases.** --

#### 1111 &mdash; E7775744

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 7.** 7000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 7.** 70%
- **Population 5.** 100,000
- **Government 7.** Balkanization. No central ruling authority exists; rival governments compete for control. Law level refers to government nearest the starport.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index 4.**
- **Bases.** --

#### 1113 &mdash; B7773547

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 7.** 7000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 7.** 70%
- **Population 3.** 1,000
- **Government 5.** Feudal Technocracy. Ruling functions are performed by specific individuals for persons who agree to be ruled by them. Relationships are based on the performance of technical activities which are mutually beneficial.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index 7.**
- **Bases.** --

#### 1115 &mdash; E3536417

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 3.** 30%
- **Population 6.** 1,000,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 1.** Certain weapons are prohibited, including specifically 1) body pistols which are undetectable by standard detectors, 2) explosive weapons such as bombs or grenades, and 3) poison gas.
- **Technological index 7.**
- **Bases.** --

#### 1116 &mdash; C3756326

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 5.** 50%
- **Population 6.** 1,000,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 2.** Portable energy weapons, such as laser rifles or carbines are prohibited. Ship's gunnery is not affected.
- **Technological index 6.**
- **Bases.** --

#### 1117 &mdash; E6893645

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 8.** Dense.
- **Hydrographics 9.** 90%
- **Population 3.** 1,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index 5.**
- **Bases.** --

#### 1120 &mdash; B341534A

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 4.** Thin, tainted.
- **Hydrographics 1.** 10%
- **Population 5.** 100,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index A.**
- **Bases.** --

#### 1211 &mdash; A689212C

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 8.** Dense.
- **Hydrographics 9.** 90%
- **Population 2.** 100
- **Government 1.** Company/Corporation. Ruling functions are assumed by a company managerial elite, and most citizenry are company employees or dependents.
- **Law level 2.** Portable energy weapons, such as laser rifles or carbines are prohibited. Ship's gunnery is not affected.
- **Technological index C.**
- **Bases.** --

#### 1212 &mdash; C52356B6

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 5.** 5000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 3.** 30%
- **Population 5.** 100,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level B.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index 6.**
- **Bases.** scout

#### 1213 &mdash; A211005E

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 2.** 2000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 1.** 10%
- **Population 0.** 0. No inhabitants.
- **Government 0.** No government structure. In many cases, family bonds will predominate.
- **Law level 5.** Personal concealable firearms (such as pistols and revolvers) are prohibited.
- **Technological index E.**
- **Bases.** --
- **Clamped.** government threw -1 and is recorded as 0.

#### 1214 &mdash; B4113509

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 1.** 10%
- **Population 3.** 1,000
- **Government 5.** Feudal Technocracy. Ruling functions are performed by specific individuals for persons who agree to be ruled by them. Relationships are based on the performance of technical activities which are mutually beneficial.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index 9.**
- **Bases.** scout

#### 1216 &mdash; D6583566

- **Starport D.** Poor quality installation. Only unrefined fuel available. No repair or shipyard facilities present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 8.** 80%
- **Population 3.** 1,000
- **Government 5.** Feudal Technocracy. Ruling functions are performed by specific individuals for persons who agree to be ruled by them. Relationships are based on the performance of technical activities which are mutually beneficial.
- **Law level 6.** Most firearms (all except shotguns) are prohibited. The carrying of any type of weapon openly is discouraged.
- **Technological index 6.**
- **Bases.** scout

#### 1217 &mdash; B53148B9

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 5.** 5000 miles diameter.
- **Atmosphere 3.** Very thin.
- **Hydrographics 1.** 10%
- **Population 4.** 10,000
- **Government 8.** Civil Service Bureaucracy. Ruling functions are performed by government agencies employing individuals selected for their expertise.
- **Law level B.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index 9.**
- **Bases.** --

#### 1218 &mdash; C000546C

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 0.** Asteroid/Planetoid Complex.
- **Atmosphere 0.** No atmosphere.
- **Hydrographics 0.** No free standing water.
- **Population 5.** 100,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 6.** Most firearms (all except shotguns) are prohibited. The carrying of any type of weapon openly is discouraged.
- **Technological index C.**
- **Bases.** scout

#### 1219 &mdash; A4558A6A

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 5.** 50%
- **Population 8.** 100,000,000
- **Government A.** Charismatic Dictator. Ruling functions are performed by agencies directed by a single leader who enjoys the overwhelming confidence of the citizens.
- **Law level 6.** Most firearms (all except shotguns) are prohibited. The carrying of any type of weapon openly is discouraged.
- **Technological index A.**
- **Bases.** --

#### 1220 &mdash; B3108689

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 0.** No free standing water.
- **Population 8.** 100,000,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 8.** Long bladed weapons (all blade weapons except daggers) are strictly controlled. Open possession in public is prohibited. Ownership is, however, not restricted.
- **Technological index 9.**
- **Bases.** naval
- **Clamped.** hydrographics threw -2 and is recorded as 0.

#### 1311 &mdash; A220440E

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 2.** 2000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 0.** No free standing water.
- **Population 4.** 10,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index E.**
- **Bases.** naval, scout
- **Clamped.** hydrographics threw -1 and is recorded as 0.
- **Clamped.** law_level threw -1 and is recorded as 0.

#### 1312 &mdash; D69878A2

- **Starport D.** Poor quality installation. Only unrefined fuel available. No repair or shipyard facilities present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 9.** Dense, tainted.
- **Hydrographics 8.** 80%
- **Population 7.** 10,000,000
- **Government 8.** Civil Service Bureaucracy. Ruling functions are performed by government agencies employing individuals selected for their expertise.
- **Law level A.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index 2.**
- **Bases.** scout

#### 1314 &mdash; X0004584

- **Starport X.** No starport. No provision is made for any starship landings.
- **Size 0.** Asteroid/Planetoid Complex.
- **Atmosphere 0.** No atmosphere.
- **Hydrographics 0.** No free standing water.
- **Population 4.** 10,000
- **Government 5.** Feudal Technocracy. Ruling functions are performed by specific individuals for persons who agree to be ruled by them. Relationships are based on the performance of technical activities which are mutually beneficial.
- **Law level 8.** Long bladed weapons (all blade weapons except daggers) are strictly controlled. Open possession in public is prohibited. Ownership is, however, not restricted.
- **Technological index 4.**
- **Bases.** --

#### 1320 &mdash; E4203265

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 0.** No free standing water.
- **Population 3.** 1,000
- **Government 2.** Participating Democracy. Ruling function decisions are reached by the advice and consent of the citizenry directly.
- **Law level 6.** Most firearms (all except shotguns) are prohibited. The carrying of any type of weapon openly is discouraged.
- **Technological index 5.**
- **Bases.** --
- **Clamped.** hydrographics threw -1 and is recorded as 0.

#### 1411 &mdash; A322465D

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 2.** 20%
- **Population 4.** 10,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 5.** Personal concealable firearms (such as pistols and revolvers) are prohibited.
- **Technological index D.**
- **Bases.** naval

#### 1415 &mdash; E8887443

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 8.** 8000 miles diameter.
- **Atmosphere 8.** Dense.
- **Hydrographics 8.** 80%
- **Population 7.** 10,000,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index 3.**
- **Bases.** --

#### 1416 &mdash; A8777458

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 8.** 8000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 7.** 70%
- **Population 7.** 10,000,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 5.** Personal concealable firearms (such as pistols and revolvers) are prohibited.
- **Technological index 8.**
- **Bases.** --

#### 1420 &mdash; A3256139

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 5.** 50%
- **Population 6.** 1,000,000
- **Government 1.** Company/Corporation. Ruling functions are assumed by a company managerial elite, and most citizenry are company employees or dependents.
- **Law level 3.** Weapons of a strict military nature (such as machine guns or automatic rifles, though not submachine guns) are prohibited.
- **Technological index 9.**
- **Bases.** naval, scout

#### 1511 &mdash; E6786696

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 8.** 80%
- **Population 6.** 1,000,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 9.** Possession of any weapon outside of one's home is prohibited.
- **Technological index 6.**
- **Bases.** --

#### 1513 &mdash; E6341467

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 3.** Very thin.
- **Hydrographics 4.** 40%
- **Population 1.** 10
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 6.** Most firearms (all except shotguns) are prohibited. The carrying of any type of weapon openly is discouraged.
- **Technological index 7.**
- **Bases.** --

#### 1515 &mdash; E63A455B

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 3.** Very thin.
- **Hydrographics A.** All water. No land masses.
- **Population 4.** 10,000
- **Government 5.** Feudal Technocracy. Ruling functions are performed by specific individuals for persons who agree to be ruled by them. Relationships are based on the performance of technical activities which are mutually beneficial.
- **Law level 5.** Personal concealable firearms (such as pistols and revolvers) are prohibited.
- **Technological index B.**
- **Bases.** --

#### 1516 &mdash; A678333B

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 8.** 80%
- **Population 3.** 1,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 3.** Weapons of a strict military nature (such as machine guns or automatic rifles, though not submachine guns) are prohibited.
- **Technological index B.**
- **Bases.** scout

#### 1518 &mdash; A100047B

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 1.** 1000 miles diameter.
- **Atmosphere 0.** No atmosphere.
- **Hydrographics 0.** No free standing water.
- **Population 0.** 0. No inhabitants.
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 7.** Shotguns are prohibited.
- **Technological index B.**
- **Bases.** scout

#### 1520 &mdash; C9591135

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 9.** 9000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 9.** 90%
- **Population 1.** 10
- **Government 1.** Company/Corporation. Ruling functions are assumed by a company managerial elite, and most citizenry are company employees or dependents.
- **Law level 3.** Weapons of a strict military nature (such as machine guns or automatic rifles, though not submachine guns) are prohibited.
- **Technological index 5.**
- **Bases.** --

#### 1612 &mdash; B45298A8

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 2.** 20%
- **Population 9.** 1,000,000,000
- **Government 8.** Civil Service Bureaucracy. Ruling functions are performed by government agencies employing individuals selected for their expertise.
- **Law level A.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index 8.**
- **Bases.** --

#### 1613 &mdash; D56A7956

- **Starport D.** Poor quality installation. Only unrefined fuel available. No repair or shipyard facilities present.
- **Size 5.** 5000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics A.** All water. No land masses.
- **Population 7.** 10,000,000
- **Government 9.** Impersonal Bureaucracy. Ruling functions are performed by agencies which have become insulated from the governed citizens.
- **Law level 5.** Personal concealable firearms (such as pistols and revolvers) are prohibited.
- **Technological index 6.**
- **Bases.** --

#### 1615 &mdash; C7785695

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 7.** 7000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 8.** 80%
- **Population 5.** 100,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 9.** Possession of any weapon outside of one's home is prohibited.
- **Technological index 5.**
- **Bases.** scout

#### 1616 &mdash; C6661007

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics 6.** 60%
- **Population 1.** 10
- **Government 0.** No government structure. In many cases, family bonds will predominate.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index 7.**
- **Bases.** --
- **Clamped.** government threw -1 and is recorded as 0.


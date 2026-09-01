# Aramis

24 worlds, 24 routes. Generated from seed 1 at occurrence DM -1.

## The map

The p. 3 sub-sector hex grid. The odd-numbered columns sit high and the even-numbered ones half a hex below them, which is how the page prints it. A world carries the letter of its starport -- p. 1 marks the hex with the letter the starports table gives -- and a hex with no world is left blank, which is what p. 1 says to leave it. P. 2 also asks for a line drawn between the worlds a route joins; a monospace grid has nowhere to put one, so this map draws none and the route table below carries them instead. `render --format pdf` draws them.

```text
0101 B    0301      0501      0701
     0201      0401 E    0601      0801 B
0102      0302      0502      0702 E
     0202 C    0402      0602 B    0802
0103 C    0303 D    0503      0703
     0203      0403      0603      0803
0104 B    0304      0504      0704
     0204      0404      0604      0804
0105      0305      0505 B    0705
     0205      0405      0605 B    0805
0106 C    0306 E    0506 A    0706 C
     0206 E    0406 E    0606      0806 D
0107 A    0307 B    0507      0707
     0207      0407      0607 C    0807 B
0108      0308      0508      0708
     0208      0408      0608      0808
0109      0309      0509 C    0709
     0209      0409      0609      0809
0110      0310      0510      0710
     0210      0410      0610      0810 A
```

## Worlds

| Hex | Name | Digits | Bases |
| --- | --- | --- | --- |
| 0101 |  | B9E578AA | naval, scout |
| 0103 |  | C5511239 | -- |
| 0104 |  | B2600107 | -- |
| 0106 |  | C8B6200B | -- |
| 0107 |  | A11069BB | -- |
| 0202 |  | C3352579 | -- |
| 0206 |  | E4555005 | -- |
| 0303 |  | D68A7A97 | scout |
| 0306 |  | E2123338 | -- |
| 0307 |  | BAA46559 | -- |
| 0401 |  | E3438436 | -- |
| 0406 |  | E9774352 | -- |
| 0505 |  | B67579BA | -- |
| 0506 |  | A97C120B | -- |
| 0509 |  | C6A4312A | -- |
| 0602 |  | B100551E | -- |
| 0605 |  | B7B2204D | naval, scout |
| 0607 |  | C4237ABA | -- |
| 0702 |  | E3105376 | -- |
| 0706 |  | C6613314 | -- |
| 0801 |  | B86B8CC5 | scout |
| 0806 |  | DAC46646 | -- |
| 0807 |  | B6747A9A | -- |
| 0810 |  | A777447A | -- |

## Routes

| From | To | Parsecs |
| --- | --- | --- |
| 0101 | 0202 | 2 |
| 0103 | 0104 | 1 |
| 0104 | 0505 | 4 |
| 0106 | 0107 | 1 |
| 0106 | 0307 | 2 |
| 0107 | 0307 | 2 |
| 0107 | 0505 | 4 |
| 0107 | 0506 | 4 |
| 0306 | 0307 | 1 |
| 0307 | 0505 | 3 |
| 0307 | 0605 | 3 |
| 0307 | 0607 | 3 |
| 0406 | 0506 | 1 |
| 0505 | 0506 | 1 |
| 0505 | 0602 | 3 |
| 0505 | 0605 | 1 |
| 0506 | 0605 | 1 |
| 0506 | 0607 | 2 |
| 0506 | 0807 | 3 |
| 0509 | 0807 | 3 |
| 0602 | 0605 | 3 |
| 0605 | 0706 | 1 |
| 0806 | 0807 | 1 |
| 0807 | 0810 | 3 |

## The worlds in detail

The technological index carries its digit and no description. The technological levels tables of pp. 10-11 say what an index means during play rather than how it is generated, so this tool does not read them; p. 11 asks the referee or the players to fill in their holes as play discovers them.

### 0101 &mdash; B9E578AA

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 9.** 9000 miles diameter.
- **Atmosphere E.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Hydrographics 5.** 50%
- **Population 7.** 10,000,000
- **Government 8.** Civil Service Bureaucracy. Ruling functions are performed by government agencies employing individuals selected for their expertise.
- **Law level A.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index A.**
- **Bases.** naval, scout

### 0103 &mdash; C5511239

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 5.** 5000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 1.** 10%
- **Population 1.** 10
- **Government 2.** Participating Democracy. Ruling function decisions are reached by the advice and consent of the citizenry directly.
- **Law level 3.** Weapons of a strict military nature (such as machine guns or automatic rifles, though not submachine guns) are prohibited.
- **Technological index 9.**
- **Bases.** --

### 0104 &mdash; B2600107

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 2.** 2000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics 0.** No free standing water.
- **Population 0.** 0. No inhabitants.
- **Government 1.** Company/Corporation. Ruling functions are assumed by a company managerial elite, and most citizenry are company employees or dependents.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index 7.**
- **Bases.** --
- **Clamped.** hydrographics threw -1 and is recorded as 0.
- **Clamped.** law_level threw -4 and is recorded as 0.

### 0106 &mdash; C8B6200B

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 8.** 8000 miles diameter.
- **Atmosphere B.** Corrosive.
- **Hydrographics 6.** 60%
- **Population 2.** 100
- **Government 0.** No government structure. In many cases, family bonds will predominate.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index B.**
- **Bases.** --
- **Clamped.** government threw -3 and is recorded as 0.

### 0107 &mdash; A11069BB

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 1.** 1000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 0.** No free standing water.
- **Population 6.** 1,000,000
- **Government 9.** Impersonal Bureaucracy. Ruling functions are performed by agencies which have become insulated from the governed citizens.
- **Law level B.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index B.**
- **Bases.** --

### 0202 &mdash; C3352579

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 3.** Very thin.
- **Hydrographics 5.** 50%
- **Population 2.** 100
- **Government 5.** Feudal Technocracy. Ruling functions are performed by specific individuals for persons who agree to be ruled by them. Relationships are based on the performance of technical activities which are mutually beneficial.
- **Law level 7.** Shotguns are prohibited.
- **Technological index 9.**
- **Bases.** --

### 0206 &mdash; E4555005

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 5.** 50%
- **Population 5.** 100,000
- **Government 0.** No government structure. In many cases, family bonds will predominate.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index 5.**
- **Bases.** --
- **Clamped.** law_level threw -1 and is recorded as 0.

### 0303 &mdash; D68A7A97

- **Starport D.** Poor quality installation. Only unrefined fuel available. No repair or shipyard facilities present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 8.** Dense.
- **Hydrographics A.** All water. No land masses.
- **Population 7.** 10,000,000
- **Government A.** Charismatic Dictator. Ruling functions are performed by agencies directed by a single leader who enjoys the overwhelming confidence of the citizens.
- **Law level 9.** Possession of any weapon outside of one's home is prohibited.
- **Technological index 7.**
- **Bases.** scout

### 0306 &mdash; E2123338

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 2.** 2000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 2.** 20%
- **Population 3.** 1,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 3.** Weapons of a strict military nature (such as machine guns or automatic rifles, though not submachine guns) are prohibited.
- **Technological index 8.**
- **Bases.** --

### 0307 &mdash; BAA46559

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size A.** 10000 miles diameter.
- **Atmosphere A.** Exotic.
- **Hydrographics 4.** 40%
- **Population 6.** 1,000,000
- **Government 5.** Feudal Technocracy. Ruling functions are performed by specific individuals for persons who agree to be ruled by them. Relationships are based on the performance of technical activities which are mutually beneficial.
- **Law level 5.** Personal concealable firearms (such as pistols and revolvers) are prohibited.
- **Technological index 9.**
- **Bases.** --

### 0401 &mdash; E3438436

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 4.** Thin, tainted.
- **Hydrographics 3.** 30%
- **Population 8.** 100,000,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 3.** Weapons of a strict military nature (such as machine guns or automatic rifles, though not submachine guns) are prohibited.
- **Technological index 6.**
- **Bases.** --

### 0406 &mdash; E9774352

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 9.** 9000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 7.** 70%
- **Population 4.** 10,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 5.** Personal concealable firearms (such as pistols and revolvers) are prohibited.
- **Technological index 2.**
- **Bases.** --

### 0505 &mdash; B67579BA

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 5.** 50%
- **Population 7.** 10,000,000
- **Government 9.** Impersonal Bureaucracy. Ruling functions are performed by agencies which have become insulated from the governed citizens.
- **Law level B.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index A.**
- **Bases.** --

### 0506 &mdash; A97C120B

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 9.** 9000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics C.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Population 1.** 10
- **Government 2.** Participating Democracy. Ruling function decisions are reached by the advice and consent of the citizenry directly.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index B.**
- **Bases.** --

### 0509 &mdash; C6A4312A

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere A.** Exotic.
- **Hydrographics 4.** 40%
- **Population 3.** 1,000
- **Government 1.** Company/Corporation. Ruling functions are assumed by a company managerial elite, and most citizenry are company employees or dependents.
- **Law level 2.** Portable energy weapons, such as laser rifles or carbines are prohibited. Ship's gunnery is not affected.
- **Technological index A.**
- **Bases.** --

### 0602 &mdash; B100551E

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 1.** 1000 miles diameter.
- **Atmosphere 0.** No atmosphere.
- **Hydrographics 0.** No free standing water.
- **Population 5.** 100,000
- **Government 5.** Feudal Technocracy. Ruling functions are performed by specific individuals for persons who agree to be ruled by them. Relationships are based on the performance of technical activities which are mutually beneficial.
- **Law level 1.** Certain weapons are prohibited, including specifically 1) body pistols which are undetectable by standard detectors, 2) explosive weapons such as bombs or grenades, and 3) poison gas.
- **Technological index E.**
- **Bases.** --

### 0605 &mdash; B7B2204D

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 7.** 7000 miles diameter.
- **Atmosphere B.** Corrosive.
- **Hydrographics 2.** 20%
- **Population 2.** 100
- **Government 0.** No government structure. In many cases, family bonds will predominate.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index D.**
- **Bases.** naval, scout
- **Clamped.** government threw -1 and is recorded as 0.

### 0607 &mdash; C4237ABA

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 3.** 30%
- **Population 7.** 10,000,000
- **Government A.** Charismatic Dictator. Ruling functions are performed by agencies directed by a single leader who enjoys the overwhelming confidence of the citizens.
- **Law level B.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index A.**
- **Bases.** --

### 0702 &mdash; E3105376

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 0.** No free standing water.
- **Population 5.** 100,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 7.** Shotguns are prohibited.
- **Technological index 6.**
- **Bases.** --
- **Clamped.** hydrographics threw -3 and is recorded as 0.

### 0706 &mdash; C6613314

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics 1.** 10%
- **Population 3.** 1,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 1.** Certain weapons are prohibited, including specifically 1) body pistols which are undetectable by standard detectors, 2) explosive weapons such as bombs or grenades, and 3) poison gas.
- **Technological index 4.**
- **Bases.** --

### 0801 &mdash; B86B8CC5

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 8.** 8000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics B.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Population 8.** 100,000,000
- **Government C.** Charismatic Oligarchy. Ruling functions are performed by a select group of members of an organization or class which enjoys the overwhelming confidence of the citizenry.
- **Law level C.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index 5.**
- **Bases.** scout

### 0806 &mdash; DAC46646

- **Starport D.** Poor quality installation. Only unrefined fuel available. No repair or shipyard facilities present.
- **Size A.** 10000 miles diameter.
- **Atmosphere C.** Insidious.
- **Hydrographics 4.** 40%
- **Population 6.** 1,000,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index 6.**
- **Bases.** --

### 0807 &mdash; B6747A9A

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 4.** 40%
- **Population 7.** 10,000,000
- **Government A.** Charismatic Dictator. Ruling functions are performed by agencies directed by a single leader who enjoys the overwhelming confidence of the citizens.
- **Law level 9.** Possession of any weapon outside of one's home is prohibited.
- **Technological index A.**
- **Bases.** --

### 0810 &mdash; A777447A

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 7.** 7000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 7.** 70%
- **Population 4.** 10,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 7.** Shotguns are prohibited.
- **Technological index A.**
- **Bases.** --


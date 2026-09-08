# Aramis

38 worlds, 91 routes, 53 drawn. Generated from seed 1 at occurrence DM 0. Broad areas: 0101-0805 at -1, 0106-0810 at +1, each overriding that DM where it lies (ERRATA E012).

## The map

The p. 3 sub-sector hex grid. The odd-numbered columns sit high and the even-numbered ones half a hex below them, which is how the page prints it. A world carries the letter of its starport -- p. 1 marks the hex with the letter the starports table gives -- and a hex with no world is left blank, which is what p. 1 says to leave it. P. 2 also asks for a line drawn between the worlds a route joins; a monospace grid has nowhere to put one, so this map draws none and the route table below carries them instead. `render --format pdf` draws them.

```text
0101 B    0301      0501      0701
     0201      0401 A    0601      0801 B
0102      0302      0502      0702 C
     0202 E    0402      0602 A    0802
0103 C    0303 B    0503      0703
     0203      0403      0603      0803
0104 B    0304      0504      0704
     0204      0404      0604      0804
0105      0305      0505 C    0705
     0205      0405      0605 B    0805
0106 C    0306 E    0506 E    0706 B
     0206 D    0406 C    0606 B    0806 A
0107 A    0307 E    0507 C    0707 E
     0207      0407 B    0607 C    0807 X
0108      0308      0508 B    0708 B
     0208      0408      0608 E    0808
0109 C    0309 B    0509 D    0709 C
     0209 E    0409 B    0609      0809
0110      0310      0510 B    0710 B
     0210      0410      0610      0810 E
```

## Worlds

| Hex | Name | Digits | Bases |
| --- | --- | --- | --- |
| 0101 |  | B360644A | naval, scout |
| 0103 |  | C9B33446 | -- |
| 0104 |  | B444127B | -- |
| 0106 |  | C4109CA8 | -- |
| 0107 |  | A467674B | -- |
| 0109 |  | C33197B8 | -- |
| 0202 |  | E0003315 | -- |
| 0206 |  | D86B8CC1 | scout |
| 0209 |  | EAC46646 | -- |
| 0303 |  | B6747A9A | -- |
| 0306 |  | E7774474 | -- |
| 0307 |  | E8892203 | -- |
| 0309 |  | B5269CCC | -- |
| 0401 |  | A6A23419 | -- |
| 0406 |  | C1505657 | -- |
| 0407 |  | B1008658 | -- |
| 0409 |  | B4108619 | naval, scout |
| 0505 |  | C521557A | -- |
| 0506 |  | E6495335 | -- |
| 0507 |  | C3215207 | -- |
| 0508 |  | B8B6589B | scout |
| 0509 |  | D4446455 | -- |
| 0510 |  | B5528CCA | -- |
| 0602 |  | A520976F | -- |
| 0605 |  | B449763A | naval |
| 0606 |  | B2551207 | scout |
| 0607 |  | C5443005 | -- |
| 0608 |  | E2627315 | -- |
| 0702 |  | C6274375 | -- |
| 0706 |  | B300531A | -- |
| 0707 |  | E4615648 | -- |
| 0708 |  | B224337B | naval |
| 0709 |  | C2364378 | scout |
| 0710 |  | B674111B | naval, scout |
| 0801 |  | B6673347 | -- |
| 0806 |  | A676536D | -- |
| 0807 |  | X4103001 | -- |
| 0810 |  | E24088B2 | -- |

## Routes

38 of these 91 lanes are not listed: each joins two worlds already joined by shorter lanes, which p. 2 says may be ignored in the drawing (ERRATA E007). The record carries every one of them, and `render --lanes=all` lists them.

| From | To | Parsecs |
| --- | --- | --- |
| 0101 | 0103 | 2 |
| 0101 | 0401 | 3 |
| 0103 | 0104 | 1 |
| 0103 | 0202 | 1 |
| 0103 | 0401 | 3 |
| 0104 | 0106 | 2 |
| 0106 | 0107 | 1 |
| 0107 | 0206 | 1 |
| 0109 | 0309 | 2 |
| 0202 | 0303 | 1 |
| 0206 | 0306 | 1 |
| 0303 | 0602 | 3 |
| 0307 | 0406 | 1 |
| 0307 | 0407 | 1 |
| 0309 | 0407 | 2 |
| 0309 | 0409 | 1 |
| 0401 | 0602 | 2 |
| 0406 | 0407 | 1 |
| 0406 | 0506 | 1 |
| 0406 | 0507 | 1 |
| 0407 | 0507 | 1 |
| 0407 | 0508 | 1 |
| 0409 | 0510 | 1 |
| 0505 | 0506 | 1 |
| 0505 | 0602 | 3 |
| 0505 | 0605 | 1 |
| 0506 | 0507 | 1 |
| 0506 | 0605 | 1 |
| 0507 | 0508 | 1 |
| 0507 | 0606 | 1 |
| 0507 | 0607 | 1 |
| 0508 | 0510 | 2 |
| 0508 | 0607 | 1 |
| 0509 | 0510 | 1 |
| 0509 | 0710 | 2 |
| 0510 | 0710 | 2 |
| 0602 | 0702 | 1 |
| 0605 | 0606 | 1 |
| 0605 | 0706 | 1 |
| 0606 | 0607 | 1 |
| 0606 | 0706 | 1 |
| 0606 | 0707 | 1 |
| 0607 | 0707 | 1 |
| 0607 | 0708 | 1 |
| 0608 | 0708 | 1 |
| 0702 | 0801 | 1 |
| 0706 | 0707 | 1 |
| 0706 | 0806 | 1 |
| 0707 | 0708 | 1 |
| 0707 | 0806 | 1 |
| 0708 | 0709 | 1 |
| 0709 | 0710 | 1 |
| 0710 | 0810 | 1 |

## The worlds in detail

What a technological index means is described in two halves. The first is T5's -- Core Book 2 pp. 230-232, cited for description alone and never for a throw (ERRATA E011): the band, the era it anchors the level to, and the level's energy, society and settlements. The second is pp. 10-11, read downward -- an entry printed at a level is the best of its kind until the next one (E009), and a hole is the page inviting the referee to fill it rather than an absence (E010). Where both books cover the same ground the held page speaks.

### 0101 &mdash; B360644A

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics 0.** No free standing water.
- **Population 6.** 1,000,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index A.** High Tech (T5), 2100 AD: Practical Fusion; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/4; Air/Raft; Drives H or less.
- **Bases.** naval, scout
- **Clamped.** hydrographics threw -1 and is recorded as 0.

### 0103 &mdash; C9B33446

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 9.** 9000 miles diameter.
- **Atmosphere B.** Corrosive.
- **Hydrographics 3.** 30%
- **Population 3.** 1,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index 6.** Low Tech (T5), Nuclear Age 1950 AD: Nuclear Fission; Superpowers; Suburbs. Pp. 10-11: Auto Rifle; Cloth; Model/1 bis; Rotary wing aircraft.
- **Bases.** --

### 0104 &mdash; B444127B

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 4.** Thin, tainted.
- **Hydrographics 4.** 40%
- **Population 1.** 10
- **Government 2.** Participating Democracy. Ruling function decisions are reached by the advice and consent of the citizenry directly.
- **Law level 7.** Shotguns are prohibited.
- **Technological index B.** High Tech (T5), Imperial Average Circa Year Zero: [FusionPlus]; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/5; Air/Raft; Drives K or less.
- **Bases.** --

### 0106 &mdash; C4109CA8

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 0.** No free standing water.
- **Population 9.** 1,000,000,000
- **Government C.** Charismatic Oligarchy. Ruling functions are performed by a select group of members of an organization or class which enjoys the overwhelming confidence of the citizenry.
- **Law level A.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index 8.** Mid Tech (T5), 2000 AD: Renewables; Superpowers; Suburbs. Pp. 10-11: Laser Carbine; Mesh; Model/2 bis; Air/Raft; Non-starships.
- **Bases.** --
- **Clamped.** hydrographics threw -3 and is recorded as 0.

### 0107 &mdash; A467674B

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics 7.** 70%
- **Population 6.** 1,000,000
- **Government 7.** Balkanization. No central ruling authority exists; rival governments compete for control. Law level refers to government nearest the starport.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index B.** High Tech (T5), Imperial Average Circa Year Zero: [FusionPlus]; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/5; Air/Raft; Drives K or less.
- **Bases.** --

### 0109 &mdash; C33197B8

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 3.** Very thin.
- **Hydrographics 1.** 10%
- **Population 9.** 1,000,000,000
- **Government 7.** Balkanization. No central ruling authority exists; rival governments compete for control. Law level refers to government nearest the starport.
- **Law level B.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index 8.** Mid Tech (T5), 2000 AD: Renewables; Superpowers; Suburbs. Pp. 10-11: Laser Carbine; Mesh; Model/2 bis; Air/Raft; Non-starships.
- **Bases.** --

### 0202 &mdash; E0003315

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 0.** Asteroid/Planetoid Complex.
- **Atmosphere 0.** No atmosphere.
- **Hydrographics 0.** No free standing water.
- **Population 3.** 1,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 1.** Certain weapons are prohibited, including specifically 1) body pistols which are undetectable by standard detectors, 2) explosive weapons such as bombs or grenades, and 3) poison gas.
- **Technological index 5.** Low Tech (T5), 1930 AD: Oil. Petrochemicals; Dictators; Skyscrapers. Pp. 10-11: Carbine, Rifle, Pistol, SMG; Cloth; Model/1; Fixed wing aircraft.
- **Bases.** --

### 0206 &mdash; D86B8CC1

- **Starport D.** Poor quality installation. Only unrefined fuel available. No repair or shipyard facilities present.
- **Size 8.** 8000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics B.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Population 8.** 100,000,000
- **Government C.** Charismatic Oligarchy. Ruling functions are performed by a select group of members of an organization or class which enjoys the overwhelming confidence of the citizenry.
- **Law level C.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index 1.** Vlow Tech (T5), Bronze Age 3500 BC: Water Power; Ethnic Groups; Settlement. Villages. Pp. 10-11: Dagger, pike, Sword; Jack; Abacus.
- **Bases.** scout

### 0209 &mdash; EAC46646

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size A.** 10000 miles diameter.
- **Atmosphere C.** Insidious.
- **Hydrographics 4.** 40%
- **Population 6.** 1,000,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index 6.** Low Tech (T5), Nuclear Age 1950 AD: Nuclear Fission; Superpowers; Suburbs. Pp. 10-11: Auto Rifle; Cloth; Model/1 bis; Rotary wing aircraft.
- **Bases.** --

### 0303 &mdash; B6747A9A

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 4.** 40%
- **Population 7.** 10,000,000
- **Government A.** Charismatic Dictator. Ruling functions are performed by agencies directed by a single leader who enjoys the overwhelming confidence of the citizens.
- **Law level 9.** Possession of any weapon outside of one's home is prohibited.
- **Technological index A.** High Tech (T5), 2100 AD: Practical Fusion; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/4; Air/Raft; Drives H or less.
- **Bases.** --

### 0306 &mdash; E7774474

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 7.** 7000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 7.** 70%
- **Population 4.** 10,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 7.** Shotguns are prohibited.
- **Technological index 4.** Low Tech (T5), Mechanization 1900 AD: Electricity; Democracies; Skyscrapers. Pp. 10-11: Revolver, Shotgun; Cloth; Adding Machine; Dirigibles.
- **Bases.** --

### 0307 &mdash; E8892203

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 8.** 8000 miles diameter.
- **Atmosphere 8.** Dense.
- **Hydrographics 9.** 90%
- **Population 2.** 100
- **Government 2.** Participating Democracy. Ruling function decisions are reached by the advice and consent of the citizenry directly.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index 3.** Vlow Tech (T5), Industrial Revolution 1700 AD: Coal. Steam; Democracies; Cities. Pp. 10-11: Foil, cutlass, Blade, bayonet; Jack; Abacus; Hot air balloon.
- **Bases.** --
- **Clamped.** law_level threw -3 and is recorded as 0.

### 0309 &mdash; B5269CCC

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 5.** 5000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 6.** 60%
- **Population 9.** 1,000,000,000
- **Government C.** Charismatic Oligarchy. Ruling functions are performed by a select group of members of an organization or class which enjoys the overwhelming confidence of the citizenry.
- **Law level C.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index C.** High Tech (T5), Imperial Average Circa Year Zero: [FusionPlus]; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/6; Grav belts; Drives N or less.
- **Bases.** --

### 0401 &mdash; A6A23419

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere A.** Exotic.
- **Hydrographics 2.** 20%
- **Population 3.** 1,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 1.** Certain weapons are prohibited, including specifically 1) body pistols which are undetectable by standard detectors, 2) explosive weapons such as bombs or grenades, and 3) poison gas.
- **Technological index 9.** Mid Tech (T5), 2050 AD: Early Fusion; Superpowers; Arcologies. Pp. 10-11: Laser Rifle; Ablat; Model/3; Air/Raft; Starships.
- **Bases.** --

### 0406 &mdash; C1505657

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 1.** 1000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 0.** No free standing water.
- **Population 5.** 100,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 5.** Personal concealable firearms (such as pistols and revolvers) are prohibited.
- **Technological index 7.** Mid Tech (T5), 1975 AD: Geothermal. Solar; Superpowers; Suburbs. Pp. 10-11: Body Pistol; Mesh; Model/2; Rotary wing aircraft; Non-starships.
- **Bases.** --

### 0407 &mdash; B1008658

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 1.** 1000 miles diameter.
- **Atmosphere 0.** No atmosphere.
- **Hydrographics 0.** No free standing water.
- **Population 8.** 100,000,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 5.** Personal concealable firearms (such as pistols and revolvers) are prohibited.
- **Technological index 8.** Mid Tech (T5), 2000 AD: Renewables; Superpowers; Suburbs. Pp. 10-11: Laser Carbine; Mesh; Model/2 bis; Air/Raft; Non-starships.
- **Bases.** --
- **Clamped.** atmosphere threw -2 and is recorded as 0.

### 0409 &mdash; B4108619

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 0.** No free standing water.
- **Population 8.** 100,000,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 1.** Certain weapons are prohibited, including specifically 1) body pistols which are undetectable by standard detectors, 2) explosive weapons such as bombs or grenades, and 3) poison gas.
- **Technological index 9.** Mid Tech (T5), 2050 AD: Early Fusion; Superpowers; Arcologies. Pp. 10-11: Laser Rifle; Ablat; Model/3; Air/Raft; Starships.
- **Bases.** naval, scout
- **Clamped.** hydrographics threw -4 and is recorded as 0.

### 0505 &mdash; C521557A

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 5.** 5000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 1.** 10%
- **Population 5.** 100,000
- **Government 5.** Feudal Technocracy. Ruling functions are performed by specific individuals for persons who agree to be ruled by them. Relationships are based on the performance of technical activities which are mutually beneficial.
- **Law level 7.** Shotguns are prohibited.
- **Technological index A.** High Tech (T5), 2100 AD: Practical Fusion; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/4; Air/Raft; Drives H or less.
- **Bases.** --

### 0506 &mdash; E6495335

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 4.** Thin, tainted.
- **Hydrographics 9.** 90%
- **Population 5.** 100,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 3.** Weapons of a strict military nature (such as machine guns or automatic rifles, though not submachine guns) are prohibited.
- **Technological index 5.** Low Tech (T5), 1930 AD: Oil. Petrochemicals; Dictators; Skyscrapers. Pp. 10-11: Carbine, Rifle, Pistol, SMG; Cloth; Model/1; Fixed wing aircraft.
- **Bases.** --

### 0507 &mdash; C3215207

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 1.** 10%
- **Population 5.** 100,000
- **Government 2.** Participating Democracy. Ruling function decisions are reached by the advice and consent of the citizenry directly.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index 7.** Mid Tech (T5), 1975 AD: Geothermal. Solar; Superpowers; Suburbs. Pp. 10-11: Body Pistol; Mesh; Model/2; Rotary wing aircraft; Non-starships.
- **Bases.** --
- **Clamped.** law_level threw -1 and is recorded as 0.

### 0508 &mdash; B8B6589B

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 8.** 8000 miles diameter.
- **Atmosphere B.** Corrosive.
- **Hydrographics 6.** 60%
- **Population 5.** 100,000
- **Government 8.** Civil Service Bureaucracy. Ruling functions are performed by government agencies employing individuals selected for their expertise.
- **Law level 9.** Possession of any weapon outside of one's home is prohibited.
- **Technological index B.** High Tech (T5), Imperial Average Circa Year Zero: [FusionPlus]; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/5; Air/Raft; Drives K or less.
- **Bases.** scout

### 0509 &mdash; D4446455

- **Starport D.** Poor quality installation. Only unrefined fuel available. No repair or shipyard facilities present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 4.** Thin, tainted.
- **Hydrographics 4.** 40%
- **Population 6.** 1,000,000
- **Government 4.** Representative Democracy. Ruling functions are performed by elected representatives.
- **Law level 5.** Personal concealable firearms (such as pistols and revolvers) are prohibited.
- **Technological index 5.** Low Tech (T5), 1930 AD: Oil. Petrochemicals; Dictators; Skyscrapers. Pp. 10-11: Carbine, Rifle, Pistol, SMG; Cloth; Model/1; Fixed wing aircraft.
- **Bases.** --

### 0510 &mdash; B5528CCA

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 5.** 5000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 2.** 20%
- **Population 8.** 100,000,000
- **Government C.** Charismatic Oligarchy. Ruling functions are performed by a select group of members of an organization or class which enjoys the overwhelming confidence of the citizenry.
- **Law level C.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index A.** High Tech (T5), 2100 AD: Practical Fusion; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/4; Air/Raft; Drives H or less.
- **Bases.** --

### 0602 &mdash; A520976F

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 5.** 5000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 0.** No free standing water.
- **Population 9.** 1,000,000,000
- **Government 7.** Balkanization. No central ruling authority exists; rival governments compete for control. Law level refers to government nearest the starport.
- **Law level 6.** Most firearms (all except shotguns) are prohibited. The carrying of any type of weapon openly is discouraged.
- **Technological index F.** Vhigh Tech (T5), Imperial Maximum Circa 1107: Exotics. Collectors; Mindwipe; Arcologies. Pp. 10-11: Laser Rifle; Battle Dress; Model/7; Grav belts; All drives.
- **Bases.** --

### 0605 &mdash; B449763A

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 4.** Thin, tainted.
- **Hydrographics 9.** 90%
- **Population 7.** 10,000,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 3.** Weapons of a strict military nature (such as machine guns or automatic rifles, though not submachine guns) are prohibited.
- **Technological index A.** High Tech (T5), 2100 AD: Practical Fusion; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/4; Air/Raft; Drives H or less.
- **Bases.** naval

### 0606 &mdash; B2551207

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 2.** 2000 miles diameter.
- **Atmosphere 5.** Thin.
- **Hydrographics 5.** 50%
- **Population 1.** 10
- **Government 2.** Participating Democracy. Ruling function decisions are reached by the advice and consent of the citizenry directly.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index 7.** Mid Tech (T5), 1975 AD: Geothermal. Solar; Superpowers; Suburbs. Pp. 10-11: Body Pistol; Mesh; Model/2; Rotary wing aircraft; Non-starships.
- **Bases.** scout

### 0607 &mdash; C5443005

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 5.** 5000 miles diameter.
- **Atmosphere 4.** Thin, tainted.
- **Hydrographics 4.** 40%
- **Population 3.** 1,000
- **Government 0.** No government structure. In many cases, family bonds will predominate.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index 5.** Low Tech (T5), 1930 AD: Oil. Petrochemicals; Dictators; Skyscrapers. Pp. 10-11: Carbine, Rifle, Pistol, SMG; Cloth; Model/1; Fixed wing aircraft.
- **Bases.** --
- **Clamped.** government threw -2 and is recorded as 0.
- **Clamped.** law_level threw -1 and is recorded as 0.

### 0608 &mdash; E2627315

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 2.** 2000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics 2.** 20%
- **Population 7.** 10,000,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 1.** Certain weapons are prohibited, including specifically 1) body pistols which are undetectable by standard detectors, 2) explosive weapons such as bombs or grenades, and 3) poison gas.
- **Technological index 5.** Low Tech (T5), 1930 AD: Oil. Petrochemicals; Dictators; Skyscrapers. Pp. 10-11: Carbine, Rifle, Pistol, SMG; Cloth; Model/1; Fixed wing aircraft.
- **Bases.** --

### 0702 &mdash; C6274375

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 7.** 70%
- **Population 4.** 10,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 7.** Shotguns are prohibited.
- **Technological index 5.** Low Tech (T5), 1930 AD: Oil. Petrochemicals; Dictators; Skyscrapers. Pp. 10-11: Carbine, Rifle, Pistol, SMG; Cloth; Model/1; Fixed wing aircraft.
- **Bases.** --

### 0706 &mdash; B300531A

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 3.** 3000 miles diameter.
- **Atmosphere 0.** No atmosphere.
- **Hydrographics 0.** No free standing water.
- **Population 5.** 100,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 1.** Certain weapons are prohibited, including specifically 1) body pistols which are undetectable by standard detectors, 2) explosive weapons such as bombs or grenades, and 3) poison gas.
- **Technological index A.** High Tech (T5), 2100 AD: Practical Fusion; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/4; Air/Raft; Drives H or less.
- **Bases.** --
- **Clamped.** atmosphere threw -1 and is recorded as 0.
- **Clamped.** hydrographics threw -1 and is recorded as 0.

### 0707 &mdash; E4615648

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics 1.** 10%
- **Population 5.** 100,000
- **Government 6.** Captive Government. Ruling functions are performed by an imposed leadership answerable to an outside group. A colony or conquered area.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index 8.** Mid Tech (T5), 2000 AD: Renewables; Superpowers; Suburbs. Pp. 10-11: Laser Carbine; Mesh; Model/2 bis; Air/Raft; Non-starships.
- **Bases.** --

### 0708 &mdash; B224337B

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 2.** 2000 miles diameter.
- **Atmosphere 2.** Very thin, tainted.
- **Hydrographics 4.** 40%
- **Population 3.** 1,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 7.** Shotguns are prohibited.
- **Technological index B.** High Tech (T5), Imperial Average Circa Year Zero: [FusionPlus]; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/5; Air/Raft; Drives K or less.
- **Bases.** naval

### 0709 &mdash; C2364378

- **Starport C.** Routine quality installation. Only unrefined fuel available. Reasonable repair facilities are present.
- **Size 2.** 2000 miles diameter.
- **Atmosphere 3.** Very thin.
- **Hydrographics 6.** 60%
- **Population 4.** 10,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 7.** Shotguns are prohibited.
- **Technological index 8.** Mid Tech (T5), 2000 AD: Renewables; Superpowers; Suburbs. Pp. 10-11: Laser Carbine; Mesh; Model/2 bis; Air/Raft; Non-starships.
- **Bases.** scout

### 0710 &mdash; B674111B

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 4.** 40%
- **Population 1.** 10
- **Government 1.** Company/Corporation. Ruling functions are assumed by a company managerial elite, and most citizenry are company employees or dependents.
- **Law level 1.** Certain weapons are prohibited, including specifically 1) body pistols which are undetectable by standard detectors, 2) explosive weapons such as bombs or grenades, and 3) poison gas.
- **Technological index B.** High Tech (T5), Imperial Average Circa Year Zero: [FusionPlus]; Non-Geographic Communities; Arcologies. Pp. 10-11: Laser Rifle; Reflec; Model/5; Air/Raft; Drives K or less.
- **Bases.** naval, scout

### 0801 &mdash; B6673347

- **Starport B.** Good quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of constructing non-starships present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 6.** Standard.
- **Hydrographics 7.** 70%
- **Population 3.** 1,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 4.** Light assault weapons (such as submachine guns) are prohibited.
- **Technological index 7.** Mid Tech (T5), 1975 AD: Geothermal. Solar; Superpowers; Suburbs. Pp. 10-11: Body Pistol; Mesh; Model/2; Rotary wing aircraft; Non-starships.
- **Bases.** --

### 0806 &mdash; A676536D

- **Starport A.** Excellent quality installation. Refined fuel available. Annual maintenance overhaul available. Shipyard capable of both starship and non-starship construction present.
- **Size 6.** 6000 miles diameter.
- **Atmosphere 7.** Standard, tainted.
- **Hydrographics 6.** 60%
- **Population 5.** 100,000
- **Government 3.** Self-Perpetuating Oligarchy. Ruling functions are performed by a restricted minority, with little or no input from the mass of citizenry.
- **Law level 6.** Most firearms (all except shotguns) are prohibited. The carrying of any type of weapon openly is discouraged.
- **Technological index D.** Vhigh Tech (T5), Imperial Maximum Circa 550: [FusionPlus]; Robots; Arcologies. Pp. 10-11: Laser Rifle; Battle Dress; Model/7; Grav belts; Drives O or less.
- **Bases.** --

### 0807 &mdash; X4103001

- **Starport X.** No starport. No provision is made for any starship landings.
- **Size 4.** 4000 miles diameter.
- **Atmosphere 1.** Trace.
- **Hydrographics 0.** No free standing water.
- **Population 3.** 1,000
- **Government 0.** No government structure. In many cases, family bonds will predominate.
- **Law level 0.** No laws affecting weapons possession or weapons ownership.
- **Technological index 1.** Vlow Tech (T5), Bronze Age 3500 BC: Water Power; Ethnic Groups; Settlement. Villages. Pp. 10-11: Dagger, pike, Sword; Jack; Abacus.
- **Bases.** --
- **Clamped.** hydrographics threw -2 and is recorded as 0.
- **Clamped.** law_level threw -4 and is recorded as 0.

### 0810 &mdash; E24088B2

- **Starport E.** Frontier installation. Essentially a bare spot of bedrock with no fuel, facilities, or bases present.
- **Size 2.** 2000 miles diameter.
- **Atmosphere 4.** Thin, tainted.
- **Hydrographics 0.** No free standing water.
- **Population 8.** 100,000,000
- **Government 8.** Civil Service Bureaucracy. Ruling functions are performed by government agencies employing individuals selected for their expertise.
- **Law level B.** Above the last row its table prints; p. 8 leaves the description to the referee, to explain or to replace (ERRATA E004).
- **Technological index 2.** Vlow Tech (T5), Age Of Sail 1500 AD: Wind. Sail; Nations; Cities. Pp. 10-11: Halberd, Broadsword; Jack; Abacus.
- **Bases.** --
- **Clamped.** hydrographics threw -2 and is recorded as 0.


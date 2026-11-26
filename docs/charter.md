# ColdTrace charter

## Purpose

Help cold-chain teams notice temperature excursions while a shipment is still moving, and give quality reviewers a complete, reproducible evidence file when they have to decide what happens to the product.

ColdTrace computes facts (episodes, gaps, time out of range, mean kinetic temperature) with transparent, versioned rules. People make the decisions. Drafted text never decides whether product is released, quarantined or discarded.

## Scope of the MVP

- Organizations, sites, zones and role-based access
- Device registry with per-device broker credentials and calibration due dates
- Versioned product profiles; each shipment keeps the profile version it started with
- Shipments, legs and time-bounded sensor assignments that never overlap for one device
- Idempotent ingestion from MQTT and gateway batch backfill, tolerant of duplicates and reordering
- A pure rule engine with stable rule IDs (see [rule catalog](rules/catalog.md))
- One alert per shipment, rule and episode, with acknowledgment and escalation
- A live dispatcher board over Server-Sent Events
- Case files with an immutable, hashed evidence snapshot
- Drafted incident summaries and CAPA reports for human sign-off
- QA disposition records and shipment reports

## Out of scope

- Validated GxP operation and any regulatory compliance claim (for example 21 CFR Part 11 or GDP)
- Real devices, firmware, shipments, products or patient data: every sensor is simulated
- Automated disposition of product
- Precise GPS tracking (route legs and sites only)

## Targets

These are portfolio and test targets, measured locally on documented hardware. They are not claims about production performance.

| Target | How it is checked |
|---|---|
| Replaying the same fixture produces identical excursions | Golden replay tests |
| Duplicate or out-of-order readings never create duplicate alerts | Ingestion and alert tests |
| Sustained local ingest rate of 5k readings/s | Benchmark with hardware details |
| Alert latency P95 under 5 s from broker receipt at that load | Load test |
| Zero dispositions decided by drafted text | Disposition endpoint accepts human input only |

## Personas

| Persona | Goal | Typical actions | Must not |
|---|---|---|---|
| **Dispatcher** (Dana) | Keep live shipments in range | Watches the board, acknowledges alerts, calls the driver | Record a disposition |
| **Quality reviewer** (Quinn, QA) | Make a defensible product decision | Opens cases, reviews evidence, edits drafts, records the disposition | Change raw readings or rule results |
| **Site manager** (Sam) | Keep warehouse zones healthy | Reviews zone sensors and recurring excursions | See other organizations' data |
| **Auditor** (Ari) | Verify what happened and who decided | Reads cases, evidence hashes and the audit trail | Change anything |
| **Device admin** (Devi) | Keep the sensor fleet trustworthy | Registers and revokes devices, tracks calibration | Read shipment cases beyond device context |

## Glossary

| Term | Meaning in ColdTrace |
|---|---|
| **Cold chain** | Keeping temperature-sensitive goods inside a set range from origin to destination |
| **Reefer** | A refrigerated truck or container |
| **Sensor / data logger** | A device that measures temperature (and door state, battery) at a fixed interval |
| **Gateway** | A device on a truck or site that relays sensor readings and buffers them while offline |
| **Reading** | One measurement: device, sequence number, `measured_at`, temperature in milli-degrees Celsius, door state, battery |
| **Sequence number (`seq`)** | A per-device counter; with `measured_at` it identifies a reading for deduplication |
| **Backfill** | A gateway sending buffered readings after a connectivity loss, often out of order |
| **Product profile** | The allowed range and limits for a product: min/max temperature, max single excursion, max cumulative time out of range, MKT limit |
| **Excursion** | Temperature outside the profile range |
| **Episode** | A contiguous run of out-of-range readings, from the first out-of-range reading to the first reading back in range |
| **Gap** | A period with no readings longer than the expected interval allows; "unknown", never "in range" |
| **Indeterminate** | A rule result that cannot be decided because a gap overlaps the period that matters |
| **MKT (mean kinetic temperature)** | A single temperature that expresses the cumulative thermal stress of a fluctuating temperature history |
| **Watermark** | The point up to which the evaluator treats data as complete; later readings trigger recomputation |
| **Alert** | A notification raised once per shipment, rule and episode |
| **Case** | A review file opened from one or more alerts, with a frozen evidence snapshot |
| **Evidence snapshot** | The readings, gaps, door events and rule results frozen when a case opens, identified by a SHA-256 hash |
| **Disposition** | The human QA decision for the affected product (for example release, quarantine, discard) |
| **CAPA** | Corrective and preventive action: what is fixed now and what changes to avoid a repeat |
| **QoS 1** | MQTT "at least once" delivery; the same reading can arrive more than once |

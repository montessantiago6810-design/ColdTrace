# Rule catalog (draft)

Status: **draft**. The rule engine is not implemented yet; this catalog fixes the IDs and semantics the implementation and its golden tests will follow. Changes to a rule's meaning bump the engine version.

## Conventions

- **Units.** Temperatures are integer milli-degrees Celsius (`8420` = 8.42 °C). Durations are measured from `measured_at` timestamps, never by counting readings. All timestamps are UTC.
- **Range is inclusive.** A reading exactly at `min` or `max` is in range.
- **Ordering.** Rules run over readings sorted by `measured_at`, then `seq`. Duplicates `(device_id, seq, measured_at)` are removed before evaluation.
- **Purity.** Rules are pure functions of (profile version, readings in the window, window bounds, engine version). No clock, no I/O, no randomness.
- **Results.** Every rule returns one of `within_limit`, `breached` or `indeterminate`, plus evidence (episode bounds, sequence range, values used).
- **Rule IDs.** Stable, uppercase, with the profile parameter as a suffix where it matters, for example `R-EXC-DUR-30` for a 30-minute limit.
- **Door events** are context on the evidence. They never excuse or suppress a result.

## Episodes

An episode starts at the `measured_at` of the first out-of-range reading and ends at the `measured_at` of the first reading back in range. If the window ends while still out of range, the episode is open and its duration runs to the window end. Highs (above `max`) and lows (below `min`) are separate episodes.

## Rules

| ID | Name | Breached when | Indeterminate when | Profile fields |
|---|---|---|---|---|
| `R-THR` | Threshold | Any reading is outside `[min, max]` | — | `min_c_milli`, `max_c_milli` |
| `R-EXC-DUR-{n}` | Single excursion duration | One episode lasts longer than `n` = `max_excursion_min` minutes | A gap overlaps an episode boundary and the known part is not already over the limit | `max_excursion_min` |
| `R-EXC-CUM-{n}` | Cumulative time out of range | Total episode time across the shipment exceeds `n` = `max_cumulative_min` minutes | Gaps could push the total over the limit | `max_cumulative_min` |
| `R-GAP-{n}` | Data gap | Time between consecutive readings exceeds `n` minutes (expected interval × k) | — | expected interval, `k` |
| `R-MKT` | Mean kinetic temperature | Time-weighted MKT exceeds `mkt_limit_c_milli` | Gaps cover more than the allowed share of the shipment | `mkt_limit_c_milli`, `activation_energy_j` |

`R-THR` is informational: it marks readings for the chart and does not open an alert on its own. Alerts come from `R-EXC-DUR`, `R-EXC-CUM`, `R-GAP` and `R-MKT`.

## Mean kinetic temperature

```text
T_mk = (ΔH / R) / −ln( Σ wᵢ · e^(−ΔH / (R · Tᵢ)) )
```

- `Tᵢ` in kelvin (°C + 273.15)
- `wᵢ` proportional to the time each reading covers, summing to 1 (uneven sampling is weighted, not averaged)
- `R` = 8.314 J/(mol·K); default `ΔH` = 83.144 kJ/mol, so `ΔH/R` ≈ 10,000 K
- `ΔH` is stored on each profile and printed in reports

## Late and suspicious data

- **Late readings.** The evaluator keeps a watermark (`max(measured_at)` minus an allowed lateness) and recomputes affected windows when late readings arrive before the shipment closes.
- **Post-close data.** Readings for a closed shipment are stored and flagged for QA; results are not silently changed.
- **Clock skew.** Readings with `measured_at` more than a tolerance ahead of `received_at` are flagged, and per-device drift is recorded.
- **Uncalibrated devices.** Readings from devices past `calibration_due_on` are flagged on the evidence.

## Open questions

- Default `k` for `R-GAP` (2× or 3× the expected interval)
- Allowed share of gap time before `R-MKT` becomes indeterminate
- Whether an open episode at shipment close counts toward `R-EXC-CUM` up to `actual_end` or to the last reading

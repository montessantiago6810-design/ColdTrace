# ADR 0001: One Go module with several binaries, PostgreSQL as the source of truth

- Status: accepted
- Date: 2026-11-19

## Context

ColdTrace needs an HTTP API with SSE, an MQTT ingestor, a rule evaluator, a background worker and a sensor simulator. They share the domain model (devices, shipments, profiles, readings, alerts), the rule engine and the database queries. Readings arrive at least once over MQTT (QoS 1) and out of order through gateway backfill. The project is built by a small team, and the local setup has to stay one command.

## Decision

1. **One Go module, several binaries.** `cmd/api`, `cmd/ingestor`, `cmd/evaluator`, `cmd/worker` and `cmd/simulator` share code under `internal/`. Each binary can be deployed on its own, and all can run together locally. No microservice split until a measured need appears.
2. **PostgreSQL with TimescaleDB is the source of truth.** MQTT is transport only. Every reading is deduplicated on insert by `(device_id, seq, measured_at)`. If TimescaleDB is unavailable on a target platform, native partitioning and scheduled rollups sit behind the same store interface.
3. **The rule engine is pure.** `internal/rules` and `internal/mkt` take readings and a profile version and return results. They do no I/O and read no clock, so replays are deterministic and golden tests are possible.
4. **Typed SQL.** pgx with sqlc-generated queries, goose migrations. Parameterized queries only.
5. **SvelteKit is UI only.** It calls the Go API and never touches the database.
6. **Platform basics from day one.** Configuration comes from `COLDTRACE_*` environment variables, validated at start-up with every error reported at once; logs are structured JSON via `log/slog`; every request gets a request ID that is echoed in the response and logged.
7. **HTTP routing.** The skeleton uses the standard library `ServeMux` with method patterns (Go 1.22+), so the module has no third-party dependencies yet. chi is added when the session and RBAC middleware land.

## Consequences

- Shared `internal` packages keep the domain model in one place; binaries cannot drift apart on types or rule semantics.
- Everything must handle duplicates and reordering, because the broker may redeliver and gateways backfill late.
- A single module means one CI pipeline (`go vet`, lint, `go test -race`) covers every binary.
- If one binary needs very different scaling (for example several ingestor replicas), it is deployed separately without changing the code layout.

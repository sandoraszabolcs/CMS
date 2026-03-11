# Architecture Decision Records

## ADR-001: sqlx over sqlc

**Context:** We need a database access layer for PostgreSQL. The two main candidates are sqlx (runtime query execution with struct scanning) and sqlc (compile-time SQL-to-Go code generation).

**Decision:** Use sqlx with named queries and StructScan.

**Consequences:**
- (+) Faster iteration — no code generation step during development
- (+) Named parameters (`:card_id`) make queries self-documenting
- (+) Direct struct mapping with `db:` tags keeps domain types clean
- (-) No compile-time query validation — SQL errors surface at runtime
- (-) Manual mapping between Go types and SQL types

**Migration path:** In production, migrate to sqlc for compile-time safety. The repository interface abstraction means only implementation files change.

## ADR-002: Redis pub/sub over SSE

**Context:** The frontend needs real-time event updates. Options: Server-Sent Events (SSE), WebSocket with Redis pub/sub, or direct WebSocket broadcasting.

**Decision:** WebSocket connections with Redis pub/sub as the event bus.

**Consequences:**
- (+) Redis decouples event producers (API handlers, simulator) from consumers (WebSocket hub)
- (+) Horizontal scalability — multiple backend instances can share the same Redis channel
- (+) Bidirectional communication available if needed later
- (-) Additional infrastructure dependency (Redis)
- (-) WebSocket requires upgrade handling and reconnection logic on client

**Alternative considered:** SSE would be simpler for one-way events but doesn't support the pub/sub pattern across multiple server instances.

## ADR-003: Materialized View for OD Matrix

**Context:** The OD matrix requires joining each checkin with its chronologically next checkout for the same passenger, then aggregating by origin-destination pair. This is expensive on every request.

**Decision:** PostgreSQL materialized view with `REFRESH MATERIALIZED VIEW CONCURRENTLY` every 30 seconds.

**Consequences:**
- (+) OD matrix reads are instant (pre-computed)
- (+) `CONCURRENTLY` allows reads during refresh (requires unique index)
- (+) Single SQL definition, no application-level aggregation code
- (-) Data is up to 30 seconds stale
- (-) Refresh locks a brief exclusive lock on the unique index

**Production alternative:** For real-time OD computation at scale, use an incremental aggregation pipeline (e.g., Apache Flink or Materialize) or maintain a running count table updated via triggers.

## ADR-004: Single HTML file frontend

**Context:** This is a POC. The frontend needs a map, OD heatmap, and live event feed.

**Decision:** Single `index.html` with vanilla JS IIFE modules, Leaflet from CDN, no build step.

**Consequences:**
- (+) Zero frontend toolchain — no Node.js, no bundler, no npm
- (+) Served directly by the Go backend as a static file
- (+) Easy to understand and modify for non-frontend developers
- (-) No TypeScript, no component reuse, no tree-shaking
- (-) All code in one file becomes unwieldy beyond ~500 lines

**Production alternative:** React/Vue SPA with TypeScript, component library, proper state management, and a build pipeline. The API contract (REST + WebSocket) remains identical.

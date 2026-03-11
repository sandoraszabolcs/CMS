# CMS — Cardul de Mobilitate Socială

## Project

POC for Romania's Social Climate Fund (FSC) transport component. Digitizes public transport benefits via CMS cards, tracks check-in/checkout events, and generates origin-destination matrices.

## Tech Stack

- **Backend:** Go 1.23, Gin, sqlx, gorilla/websocket, go-redis
- **Database:** PostgreSQL 16 (materialized view for OD matrix)
- **Cache/PubSub:** Redis 7
- **Frontend:** Single `index.html`, vanilla JS (IIFE modules), Leaflet.js
- **Infra:** Docker Compose

## Project Layout

```
cmd/server/main.go              - entrypoint, DI wiring, errgroup
internal/domain/                - pure types + sentinel errors
internal/repository/            - DB interfaces + postgres implementations
internal/service/               - business logic
internal/simulator/             - background event generator
internal/transport/http/        - gin handlers + middleware
internal/transport/ws/          - websocket hub
internal/infrastructure/        - postgres, redis, config
migrations/                     - SQL migrations (run by postgres entrypoint)
frontend/                       - single-page UI
```

## Conventions

- Dependency injection — no global state, no `init()` side effects
- Repository pattern — handlers never touch SQL
- Structured logging with `log/slog`
- sqlx with named queries and `db:` struct tags
- Domain sentinel errors (`ErrNotFound`, `ErrPassengerInactive`)
- Repository translates `sql.ErrNoRows` → `domain.ErrNotFound`
- Consistent error responses: `{ "error": { "code": "...", "message": "..." } }`

## Running

```bash
docker-compose up --build
# Open http://localhost:8080
```

## Testing

```bash
cd backend && go test ./...
```

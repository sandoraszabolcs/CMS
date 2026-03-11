# FSC CMS — Cardul de Mobilitate Socială

Proof of concept for Romania's Social Climate Fund (FSC) transport component. The system digitizes public transport benefits via the CMS (Cardul de Mobilitate Socială), tracking passenger check-in/checkout events on bus routes and generating origin-destination (OD) matrices for transport planning. This POC simulates Bucharest bus line 41 with real stop coordinates.

## Prerequisites

- Docker & Docker Compose v2+
- Port 8080 available

## Quick Start

```bash
git clone <repo-url> && cd CMS
docker-compose up --build
# Open http://localhost:8080
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────┐ │
│  │  Leaflet  │  │ OD Matrix│  │Event Feed│  │  WebSocket  │ │
│  │   Map     │  │ Heatmap  │  │  (live)  │  │   Client    │ │
│  └─────┬────┘  └────┬─────┘  └────┬─────┘  └──────┬──────┘ │
└────────┼────────────┼─────────────┼────────────────┼────────┘
         │  REST API  │             │         WS /ws │
─────────┴────────────┴─────────────┴────────────────┴─────────
┌─────────────────────────────────────────────────────────────┐
│                     Go Backend (Gin)                         │
│                                                              │
│  cmd/server/main.go ─── wires everything with errgroup       │
│                                                              │
│  ┌─────────────┐  ┌────────────┐  ┌──────────────────────┐  │
│  │  HTTP       │  │  WebSocket │  │  Simulator Worker    │  │
│  │  Handlers   │  │  Hub       │  │  (2s interval)       │  │
│  └──────┬──────┘  └─────┬──────┘  └──────────┬───────────┘  │
│         │               │                     │              │
│  ┌──────┴───────────────┴─────────────────────┴───────────┐  │
│  │              Service Layer (business logic)             │  │
│  └──────────────────────┬─────────────────────────────────┘  │
│  ┌──────────────────────┴─────────────────────────────────┐  │
│  │           Repository Layer (interfaces + sqlx)          │  │
│  └──────────┬─────────────────────────────┬───────────────┘  │
└─────────────┼─────────────────────────────┼──────────────────┘
              │                             │
    ┌─────────┴────────┐          ┌─────────┴─────────┐
    │  PostgreSQL 16   │          │  Redis 7 (pub/sub) │
    │  + materialized  │          │  validation_events │
    │    OD matrix     │          │  channel            │
    └──────────────────┘          └─────────────────────┘
```

## API Reference

| Method | Endpoint               | Description                        |
|--------|------------------------|------------------------------------|
| POST   | `/api/v1/checkin`      | Check in passenger at stop         |
| POST   | `/api/v1/checkout`     | Check out passenger at stop        |
| GET    | `/api/v1/od-matrix`    | OD matrix (materialized view)      |
| GET    | `/api/v1/vehicles`     | Current vehicle positions          |
| GET    | `/api/v1/stops`        | All route stops                    |
| GET    | `/api/v1/events/recent`| Last 20 events with passenger info |
| GET    | `/api/v1/stats`        | Dashboard statistics               |
| WS     | `/ws`                  | Real-time event stream             |

### Request/Response Examples

**POST /api/v1/checkin**
```json
{
  "card_id": "CMS-001",
  "vehicle_id": "BUS-101",
  "stop_id": "S1",
  "lat": 44.4268,
  "lng": 26.1025
}
```

**Error response shape**
```json
{
  "error": {
    "code": "PASSENGER_NOT_FOUND",
    "message": "passenger not found"
  }
}
```

## Passenger Categories (CMS)

| Category    | Legislation      | Benefit       |
|-------------|------------------|---------------|
| student     | Legea 198/2023   | Free travel   |
| university  | Legea 199/2023   | 90% discount  |
| pensioner   | Legea 147/2000   | Limited trips |
| disabled    | Legea 448/2006   | Free travel   |
| veteran     | Legea 44/1994    | Free travel   |

## Environment Variables

| Variable             | Required | Default | Description          |
|----------------------|----------|---------|----------------------|
| `DB_URL`             | Yes      | —       | PostgreSQL connection |
| `REDIS_ADDR`         | Yes      | —       | Redis host:port      |
| `HTTP_PORT`          | No       | 8080    | Server port          |
| `LOG_LEVEL`          | No       | info    | debug/info/warn/error|
| `SIMULATOR_INTERVAL` | No       | 2s      | Event generation rate |
| `OD_REFRESH_INTERVAL`| No       | 30s     | MV refresh interval  |

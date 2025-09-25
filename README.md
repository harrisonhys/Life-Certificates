# Life Certificate Service (LCS)

Go microservice that manages digital life certificates for pension participants. It registers participants against the Face Recognition (FR) Core, performs periodic verification with selfie uploads, and exposes status APIs for downstream channels.

## Features
- Participant registration with FR Core `/upload` integration
- Life certificate verification with liveness stub and FR Core `/recognize`
- SQLite-backed persistence using GORM ORM
- Relies on FR Core for selfie storage; LCS stores only metadata
- JSON REST API with health endpoint and standardized envelopes

## Requirements
- Go 1.21+
- PostgreSQL database (default `postgres://postgres:postgres@localhost:5432/mydb`)
- (Optional) Running FR Core instance reachable from the service

## Configuration
Environment variables (see `.env.example`):

| Variable | Default | Description |
| --- | --- | --- |
| `HTTP_HOST` | `0.0.0.0` | Bind address |
| `HTTP_PORT` | `8080` | Port |
| `DATABASE_DSN` | `postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable` | PostgreSQL DSN |
| `BASIC_AUTH_USERNAME` | `admin` | Username for HTTP Basic Auth |
| `BASIC_AUTH_PASSWORD` | `admin` | Password for HTTP Basic Auth |
| `FRCORE_BASE_URL` | `http://localhost:9000` | FR Core base URL |
| `FRCORE_UPLOAD_API_KEY` | _required_ | API key for `/upload` |
| `FRCORE_RECOGNIZE_API_KEY` | _required_ | API key for `/recognize` |
| `FRCORE_TENANT_ID` | _(empty)_ | Optional tenant header |
| `FRCORE_TIMEOUT_SECONDS` | `10` | HTTP timeout |
| `VERIFICATION_DISTANCE_THRESHOLD` | `0.6` | Distance threshold for match |
| `VERIFICATION_SIMILARITY_THRESHOLD` | `75` | Similarity fallback threshold |
| `LIVENESS_ENABLED` | `true` | Toggle noop liveness checker |

## Running Locally
```bash
cp .env.example .env # edit as needed
GOCACHE=$(pwd)/.gocache go build ./...
go run ./cmd/server
```

The server automatically loads `.env` when present, so any `FRCORE_*` keys defined there are forwarded on each request to FR Core.

The service listens on `http://localhost:8080` by default.

All API calls (except `GET /health`) require HTTP Basic authentication using the credentials defined in `BASIC_AUTH_USERNAME` / `BASIC_AUTH_PASSWORD`.

## API Overview

Swagger UI is available at `GET /swagger/index.html` (requires Basic Auth).

To regenerate the OpenAPI documentation after changing handlers or annotations, run:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.8.12 init \
  --generalInfo cmd/server/main.go \
  --output docs \
  --parseDependency
```

### `POST /participants/register`
Registers a participant with initial selfie via `multipart/form-data`. The service forwards the selfie to FR Core using a UUID label and your `participant_id` as the FR `external_ref`. Both identifiers are persisted for later verification.

Form fields:
- `nik` (text)
- `name` (text)
- `image` (file upload)

Response:
```json
{
  "status": "success",
  "data": {
    "participant_id": "uuid",
    "fr_ref": "fr-label",
    "fr_external_ref": "participant_id"
  }
}
```

### `POST /life-certificate/verify`
Multipart form fields: `participant_id`, `image` file. Returns current verification status (`VALID`, `INVALID`, `REVIEW`) plus similarity/distance metadata when available.

### `GET /life-certificate/status/{participant_id}`
Returns the most recent verification result for the participant, including `last_status`, `similarity`, `distance`, and `verified_at` when present.

### `GET /participants`
Returns the list of participants ordered by most recent creation.

### `GET /participants/{participant_id}`
Returns metadata for a specific participant.

### `PUT /participants/{participant_id}`
Updates participant name and/or NIK using a JSON payload `{ "nik": "", "name": "" }`.

### `DELETE /participants/{participant_id}`
Deletes a participant and related verification records.

### `GET /health`
Basic health probe.

## Project Layout
- `cmd/server` – program entrypoint
- `internal/config` – environment configuration loader
- `internal/database` – GORM/SQLite wiring and migrations
- `internal/domain` – domain models and constants
- `internal/frcore` – HTTP client for FR Core integrations
- `internal/liveness` – stubbed liveness checker
- `internal/repository` – persistence layer abstractions
- `internal/service` – business logic for registration/verification
- `internal/http` – router, handlers, and response helpers

## Testing & Validation
- `GOCACHE=$(pwd)/.gocache go build ./...`
- Additional tests can be added under `internal/...` as the service evolves.

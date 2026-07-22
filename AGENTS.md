# PitchOnDB Backend Agent Guide (Orthogonal Design)

This guide is for AI/code agents working on the backend only (`apps/api` + `db`).
The frontend (`apps/web`) is intentionally out of scope.

---

## 1) Backend Snapshot

- Language/runtime: Go 1.26
- API transport: Gin + Huma (`humagin` adapter)
- CLI/runtime config: `humacli` + Cobra commands
- Database: PostgreSQL (driver: `pgx` stdlib)
- Query/codegen: sqlc
- Migrations: Goose
- Logging: structured JSON via `log/slog`

Key entrypoint:
- `apps/api/cmd/service/main.go`

---

## 2) Orthogonal Design (What Changed)

The backend now follows two independent (orthogonal) axes:

1. **Business feature axis** (`internal/<feature>`, currently `internal/pigeon`)
   - Domain model
   - Use-case services
   - Ports (interfaces)
2. **Platform axis** (`internal/platform/*`)
   - Inbound adapters (HTTP API)
   - Outbound adapters (Postgres/sqlc)

This keeps business rules isolated from transport and persistence concerns.

### Current Request Flow

1. HTTP request enters Gin/Huma router (`internal/platform/httpapi`)
2. Handler calls feature port (`pigeon.Service` or `pigeon.QueryService`)
3. Port is implemented by Postgres adapter (`internal/platform/postgres`)
4. Adapter uses generated sqlc queries (`internal/platform/postgres/sqlc`)
5. Response is mapped back to API DTOs

---

## 3) Command/Query Split (CQRS-lite)

The `pigeon` feature separates write and read paths:

- **Command side**
  - `pigeon.Service`
  - `pigeon.Repository`
  - Aggregate/entity: `pigeon.Pigeon`
- **Query side**
  - `pigeon.QueryService`
  - Read model: `pigeon.Summary`
  - Backed by `pigeon_summaries` SQL view

Keep this split when adding features: avoid mixing read-model logic into aggregate persistence code.

---

## 4) Package Map

### Composition root
- `apps/api/cmd/service/main.go`
  - Initializes logging/config
  - Connects DB
  - Wires adapters to ports
  - Registers routes
  - Handles graceful shutdown
  - Defines CLI subcommands:
    - `openapi` (generate OpenAPI spec)
    - `healthcheck` (poll `/health`)

### Feature package (domain + ports + use cases)
- `apps/api/internal/pigeon/domain.go`  
  Aggregate with private fields, behavior methods (`Rename`, `DetermineSex`, etc.), snapshot export.
- `apps/api/internal/pigeon/service.go`  
  Command-side use cases (`Add`, `UpdatePigeonDetails`).
- `apps/api/internal/pigeon/repository.go`  
  Write-side persistence port.
- `apps/api/internal/pigeon/query.go`  
  Read-side query port.
- `apps/api/internal/pigeon/domain_enumer.go`  
  Generated enum helpers (from `enumer`).
- `apps/api/internal/pigeon/mocks/*`  
  Generated test mocks for ports.

### Platform adapters
- `apps/api/internal/platform/httpapi/*`
  - Router composition
  - Health and pigeon handlers
  - API DTO schemas and mapping
  - Request/response middleware
- `apps/api/internal/platform/postgres/*`
  - DB connection
  - Repository adapter (`pigeon.Repository`)
  - Query adapter (`pigeon.QueryService`)
  - Null and conversion helpers
  - Generated sqlc package

### Cross-cutting support
- `apps/api/internal/config/options.go` (runtime options)
- `apps/api/internal/telemetry/logging.go` (slog setup)
- `apps/api/internal/utils/*` (small generic helpers)

---

## 5) HTTP/API Conventions

- API versioning is route-group based (`/v1`).
- Health endpoint is top-level (`/health`).
- `requestid` middleware generates request IDs.
- Request ID is bridged into Huma context for consistent logs.
- Request logging runs as Huma middleware at group level.
- OpenAPI is generated from handler/types metadata.

When adding endpoints:
1. Add handler method + I/O types in `internal/platform/httpapi`
2. Register route in `With<Feature>Routes(...)`
3. Regenerate spec (`make -C apps/api generate-spec` or root equivalent flow)

---

## 6) Persistence and Database Conventions

DB artifacts live in root `db/`:
- Migrations: `db/migrations/*.sql`
- Queries: `db/queries/*.sql`
- sqlc config: `db/sqlc.yml`

sqlc output target:
- `apps/api/internal/platform/postgres/sqlc`

Current schema highlights:
- `pigeons` table
- `pairings` table
- `pigeon_summaries` view (read model used for list queries)

Rules:
- Never edit generated sqlc files manually.
- Add/modify SQL in `db/queries` and regenerate.
- Add schema changes via new Goose migrations only.

---

## 7) Build, Run, Test, and Codegen

### Root-level shortcuts
- `make tools` (install sqlc + goose from `tools` module)
- `make generate-sqlc`
- `make migrate` / `make rollback`
- `make api.build` / `make api.run` / `make api.test` / `make api.healthcheck`
- `make docker-up` / `make docker-down`

### API module commands
- `make -C apps/api build`
- `make -C apps/api run`
- `make -C apps/api dev`
- `make -C apps/api test`
- `make -C apps/api lint`
- `make -C apps/api generate`
- `make -C apps/api generate-mocks`
- `make -C apps/api generate-spec`

---

## 8) Coding Rules for Agents (Backend)

1. Keep business logic in feature packages (`internal/<feature>`), not in adapters.
2. Treat `internal/platform/httpapi` and `internal/platform/postgres` as adapter layers only.
3. Do not import platform packages into feature domain code.
4. Keep mapping/transformation logic at adapter boundaries.
5. Preserve command/query separation for new capabilities.
6. Regenerate generated artifacts instead of hand-editing them:
   - sqlc output
   - enum output (`domain_enumer.go`)
   - mockery output
7. When interface contracts change, regenerate and update mocks.
8. Wrap errors with context (`fmt.Errorf("context: %w", err)`), keep sentinel domain errors meaningful (for example `pigeon.ErrNotFound`).

---

## 9) How to Add a New Backend Feature (Orthogonal Way)

1. Create `internal/<feature>/domain.go` (entity/aggregate + behavior).
2. Define ports in the feature package:
   - command service interface
   - repository interface
   - query service interface (if needed)
3. Implement use cases in feature service(s).
4. Add outbound Postgres adapter(s) in `internal/platform/postgres`.
5. Add SQL queries/migrations in root `db/`, regenerate sqlc.
6. Add inbound HTTP handler/types/routes in `internal/platform/httpapi`.
7. Wire everything in `cmd/service/main.go`.
8. Generate/update:
   - sqlc (`make generate-sqlc`)
   - mocks (`make -C apps/api generate-mocks`) when interfaces change
   - OpenAPI (`make -C apps/api generate-spec`) when API changes
9. Run backend tests/lint before finalizing.

---

## 10) Generated Files and Sources of Truth

- Generated:
  - `apps/api/internal/platform/postgres/sqlc/*`
  - `apps/api/internal/pigeon/domain_enumer.go`
  - `apps/api/internal/pigeon/mocks/*`
- Source of truth:
  - SQL files in `db/queries` and `db/migrations`
  - Enum declarations in `internal/pigeon/domain.go`
  - Interfaces in feature packages

If generated output and handwritten code diverge, update the source-of-truth files and regenerate.

# Repository Guidelines

## Project Structure & Module Organization

This is an npm-workspaces monorepo with a Go backend:

- `apps/web/`: public React/Vite site for encyclopedia search and recommendations.
- `apps/admin/`: authenticated React/Vite operations console.
- `apps/api/`: Gin/pgx API. Entrypoints live in `cmd/`; domain packages live in `internal/`; shared infrastructure lives in `pkg/`.
- `apps/api/migrations/`: ordered PostgreSQL migrations embedded and applied automatically at API startup.
- `packages/`: shared UI, TypeScript utilities, and configuration.
- `docs/`: product, architecture, database, and development documentation.
- `scripts/`: database and development startup helpers.

Place Go tests beside implementation files as `*_test.go`.

## Build, Test, and Development Commands

Run commands from the repository root:

```bash
npm install          # install workspace dependencies
./scripts/dev.sh     # start database, API, web, and admin apps
npm run dev:api      # Go API on :8080
npm run dev:web      # public site on :5173
npm run dev:admin    # admin console on :5174
npm run test:api     # run all Go tests
npm run test:admin   # run admin Vitest component tests
npm run build:api    # compile the API server
npm run build:web    # type-check and build the public app
npm run build:admin  # type-check and build the admin app
```

## Coding Style & Naming Conventions

Format Go with `gofmt`; use lowercase packages and exported `PascalCase` identifiers. Keep domain logic under `apps/api/internal/<domain>`. TypeScript uses two-space indentation, `PascalCase` components/types, and `camelCase` functions and state. Prefer focused components.

Create migrations with increasing names such as `006_add_feedback_index.sql`. Never modify an applied migration.

## Testing Guidelines

Go uses `testing` plus `httptest` for handlers. Name tests `TestBehaviorBeingVerified`. Admin UI tests use Vitest and Testing Library; name them `*.test.tsx` beside the feature or app shell. Cover validation, authorization, parsers, scoring, migrations, and role-sensitive UI. Before submitting, run all builds, `npm run test:api`, and `npm run test:admin`.

## Commit & Pull Request Guidelines

History primarily uses Conventional Commit style, for example `feat(auth): implement user authentication`. Prefer `feat:`, `fix:`, `docs:`, `test:`, or `refactor:` with an optional scope. Keep commits focused.

Pull requests should include a problem/solution description, verification commands, migration notes, and linked issues. Include screenshots for UI changes and document new environment variables or API endpoints.

## Security & Configuration

Copy values from `.env.example`; never commit secrets. Replace development `JWT_SECRET` and administrator credentials outside local environments. Preserve authentication and published-data checks on all `/api/admin/*` and public relation routes.

## Local Database

Local PostgreSQL uses `localhost:55432`, database `fungi_wiki`, username `fungi`, and password `fungi`:

```text
postgres://fungi:fungi@localhost:55432/fungi_wiki?sslmode=disable
```

Start it with `./scripts/db-up.sh`. The Go API applies pending migrations automatically at startup.

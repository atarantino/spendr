# AGENTS Guide for Spendr
1. Work from repo root on Go 1.25 using modules.
2. Build pipeline: `make build` (templ generate, tailwind compile, go build).
3. Quick run via `make run`; clean artifacts with `make clean`.
4. Live reload using `make watch` (auto-installs `air` on first run).
5. Start PostgreSQL with `make docker-run`; stop using `make docker-down`.
6. Database config comes from `BLUEPRINT_DB_*`; `.env` loads automatically via godotenv.
7. Run migrations with `make migrate-up` / `make migrate-down`; create via `make migrate-create`.
8. Regenerate templ/sqlc outputs using `make build` or `make sqlc-generate`.
9. Run all tests using `make test` (`go test ./... -v`).
10. Single test example: `go test ./internal/... -run '^TestName$' -v`.
11. Integration DB tests: `make itest` (testcontainers; Docker required).
12. Format with `go fmt ./...` and ensure goimports-style import grouping.
13. Keep generated `*_templ.go` untouched; edit `.templ` files then rerun `templ generate`.
14. Use constructors to inject dependencies across handlers, services, and database layers.
15. Thread `context.Context` from handlers into services and sqlc queries.
16. Handle errors early with explicit status codes and wrap via `fmt.Errorf("context: %w", err)`.
17. Avoid panics in HTTP handlers; prefer logging and safe fallbacks.
18. JSON/HTMX responses should set `Content-Type`, use `json.NewEncoder`, and leverage `HX-Redirect`.
19. Sessions rely on scs; secure routes live in `internal/server/routes.go` using `auth.RequireAuth` and `database.Service`.
20. No Cursor or Copilot instruction files detected.

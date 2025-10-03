# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Spendr is a Go web application built with:
- **chi** for HTTP routing
- **templ** for type-safe HTML templating
- **htmx** for dynamic frontend interactions
- **Tailwind CSS** for styling
- **PostgreSQL** with pgx driver for database
- **sqlc** for type-safe SQL code generation
- **scs** (alexedwards/scs/v2) for session management with PostgreSQL backend
- **bcrypt** for password hashing

## Build System

The project uses a Makefile with automatic tool installation:

- `make build` - Installs templ, downloads tailwindcss binary, generates templ files, compiles CSS, and builds the Go binary to `./main`
- `make run` - Run the application directly with `go run`
- `make watch` - Live reload development using air (auto-installs if missing)
- `make test` - Run all tests
- `make itest` - Run integration tests (database tests in `internal/database`)
- `make docker-run` - Start PostgreSQL container via docker compose
- `make docker-down` - Stop PostgreSQL container
- `make clean` - Remove build artifacts

### Database Migration Commands
- `make migrate-install` - Install golang-migrate CLI tool
- `make migrate-create` - Create a new migration (prompts for name)
- `make migrate-up` - Run all pending migrations
- `make migrate-down` - Rollback the last migration
- `make migrate-force` - Force set migration version (for fixing errors)

### SQLC Commands
- `make sqlc-install` - Install sqlc CLI tool
- `make sqlc-generate` - Generate Go code from SQL queries in `internal/database/queries/`

## Architecture

### Application Entry Point
- `cmd/api/main.go` - Main entry point with graceful shutdown handling (5s timeout on SIGINT/SIGTERM)

### Server Layer
- `internal/server/server.go` - HTTP server configuration, creates database service singleton, initializes session manager with PostgreSQL store, and auth service
- `internal/server/routes.go` - Chi router setup with middleware (logger, session management, CORS), route registration for public/protected routes
- Server uses chi router with CORS enabled for all origins, supports WebSocket connections
- Session middleware (`sessionManager.LoadAndSave`) applied globally to all routes

### Authentication & Authorization
- `internal/auth/service.go` - Authentication service with user registration and login
  - `Register()` - Creates new user with bcrypt password hashing
  - `Login()` - Validates credentials using bcrypt comparison
- `internal/auth/middleware.go` - Authentication middleware
  - `RequireAuth()` - Protects routes by checking session for userID, redirects to `/login` if unauthorized
  - Adds userID to request context for authenticated requests
- Session management via scs with PostgreSQL backend (pgxstore)
- Sessions stored in `sessions` table with 24-hour lifetime

### Handler Layer
- `internal/handlers/auth.go` - Authentication handlers (login, register, logout pages and form submissions)
  - Uses htmx redirects (`HX-Redirect` header) for seamless SPA-like navigation
  - Stores userID in session after successful auth
- `internal/handlers/dashboard.go` - Protected dashboard handler (requires authentication)
- `internal/handlers/health.go` - Health check endpoint with database connection stats
- `internal/handlers/websocket.go` - WebSocket handler for real-time communication

### Database Layer
- `internal/database/database.go` - Database service interface and implementation
- Uses singleton pattern (`dbInstance`) to reuse database connections
- Connection configured via environment variables: `BLUEPRINT_DB_*` (HOST, PORT, DATABASE, USERNAME, PASSWORD, SCHEMA)
- Includes health check endpoint with detailed connection pool statistics
- Exposes `GetPool()` for session store and `GetQueries()` for sqlc queries

### Database Migrations
- `internal/database/migrations/000001_create_users_tabl.up.sql` - Users table with email, password_hash, timestamps
- `internal/database/migrations/000002_create_sessions.up.sql` - Sessions table for scs session management
- Migrations managed via golang-migrate

### Database Queries (SQLC)
- `internal/database/queries/users.sql` - User-related queries:
  - `CreateUser` - Insert new user with name, email, password_hash
  - `GetUserByEmail` - Retrieve user for login authentication
  - `GetUserByID` - Retrieve user profile by ID
- Generated code in `internal/database/sqlc/` provides type-safe query functions

### Frontend Layer
- `cmd/web/` - Web UI components using templ templates
- `cmd/web/efs.go` - Embeds static assets (CSS, JS) via go:embed
- `cmd/web/base.templ` - Base HTML layout template
- `cmd/web/login.templ` - Login and registration page templates
- Uses htmx for dynamic form interactions without full page reloads
- Tailwind input CSS: `cmd/web/styles/input.css` → compiled to `cmd/web/assets/css/output.css`

### SQLC Integration
- Configuration: `sqlc.yaml` at project root
- Queries directory: `internal/database/queries/` - SQL queries with sqlc annotations
- Migrations directory: `internal/database/migrations/` - Database schema migrations
- Generated code output: `internal/database/sqlc/` with pgx/v5 driver
- Run `make sqlc-generate` to regenerate Go code from SQL queries

## Development Workflow

1. Start PostgreSQL: `make docker-run`
2. Run migrations: `make migrate-up`
3. Generate sqlc code: `make sqlc-generate` (after modifying queries)
4. For development with live reload: `make watch`
5. For production build: `make build` then `./main`

Environment variables are loaded automatically via godotenv from `.env` file.

## Testing

- Integration tests use testcontainers-go to spin up PostgreSQL instances
- Run all tests: `make test`
- Run only database integration tests: `make itest`

## Key Dependencies

- **github.com/a-h/templ** - Type-safe Go templating
- **github.com/go-chi/chi/v5** - Lightweight HTTP router
- **github.com/jackc/pgx/v5** - PostgreSQL driver
- **github.com/alexedwards/scs/v2** - Session management library
- **github.com/alexedwards/scs/pgxstore** - PostgreSQL session store
- **golang.org/x/crypto/bcrypt** - Password hashing
- **github.com/coder/websocket** - WebSocket support
- **github.com/testcontainers/testcontainers-go** - Integration testing
- **github.com/golang-migrate/migrate/v4** - Database migrations

## Frontend Stack

Templates are written in templ syntax and compiled to Go code. The build process:
1. `templ generate` creates `*_templ.go` files
2. `tailwindcss` CLI compiles CSS
3. `go build` compiles everything including generated code

Static assets are embedded in the binary via `embed.FS` and served through chi's FileServer.

## Route Structure

### Public Routes
- `GET /` - Hello World API endpoint
- `GET /health` - Database health check
- `GET /websocket` - WebSocket connection
- `GET /login` - Login page
- `POST /login` - Login form submission
- `GET /register` - Registration page
- `POST /register` - Registration form submission
- `POST /logout` - Logout (destroys session)
- `/assets/*` - Static files (CSS, JS)

### Protected Routes (require authentication)
- `GET /dashboard` - User dashboard (shows userID from context)

## External API Documentation

### Plaid API
When working with the Plaid API, refer to the official documentation at https://plaid.com/docs/llms.txt for API reference, endpoints, parameters, and integration guidance.
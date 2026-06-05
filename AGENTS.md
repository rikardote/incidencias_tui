# Repository Guidelines

## Project Structure & Module Organization

This is a Go terminal UI application built with Bubble Tea, Bubbles, and Lip Gloss. `main.go` owns top-level screen routing and shared state.

- `internal/api/`: HTTP API client and request helpers.
- `internal/config/`: config loading/saving; defaults to `http://localhost:8190` and writes `~/.config/incidencias-tui/config.json`.
- `internal/models/`: shared data structures for API and UI state.
- `internal/views/`: Bubble Tea models for login, menu, employees, codes, capture, reports, and biometric screens.
- `internal/styles/`: reusable Lip Gloss styles.
- `Makefile`, `Dockerfile`: local build/run and container workflows.

No test directory exists yet; add tests beside the package they cover as `*_test.go`.

## Build, Test, and Development Commands

- `make build`: compiles `./incidencias-tui` with stripped symbols.
- `make run`: builds and runs the local binary.
- `go test ./...`: runs all package tests.
- `go fmt ./...`: formats Go files using standard Go formatting.
- `go vet ./...`: runs basic static checks.
- `make clean`: removes generated binaries.
- `make docker-build`: builds the Docker image `incidencias-tui`.
- `make docker-run`: runs the image interactively with config mounted and host networking.

## Coding Style & Naming Conventions

Use standard Go formatting with tabs via `gofmt`. Keep package names short and lowercase. Add comments for exported package-boundary types and methods, following patterns such as `Client`, `Config`, and `Login`.

Prefer small Bubble Tea models per screen under `internal/views/`. Keep API DTOs and shared structs in `internal/models/`. Use Spanish user-facing messages where the existing UI already does.

## Testing Guidelines

Use Go’s built-in `testing` package unless a stronger need appears. Name tests by behavior, for example `TestLoadReturnsDefaultConfigWhenMissing`. For API code, prefer `httptest.Server` over real network calls.

Run `go test ./...` plus `go vet ./...` after logic, HTTP, or config changes.

## Commit & Pull Request Guidelines

The current history only uses `first commit`, so there is no strict convention yet. Use short imperative subjects such as `Add employee search view` or `Fix API error handling`; include a body when behavior needs context.

Pull requests should include a summary, testing performed, linked issue or task when available, and screenshots or terminal recordings for UI changes. Note API endpoint or config changes explicitly.

## Security & Configuration Tips

Do not commit real tokens or personal config files. The app persists auth tokens in `~/.config/incidencias-tui/config.json`; treat that file as local-only. Avoid hard-coding environment-specific credentials.

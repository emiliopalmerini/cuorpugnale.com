# Repository Guidelines

## Project Structure & Module Organization

- `cmd/server/main.go` starts the HTTP server, reads `ADDR`, and handles graceful shutdown.
- `internal/web/web.go` defines routes, security headers, static caching, and embedded assets.
- `internal/web/templates/` contains HTML templates embedded with `go:embed`.
- `internal/web/static/` contains CSS and image assets served under `/static/`.
- `internal/web/web_test.go` contains HTTP and rendering tests.
- `Dockerfile`, `docker-compose.yml`, and `Makefile` support local and container workflows.

## Build, Test, and Development Commands

- `make help` lists available project targets.
- `make build` runs `fmt` and `vet`, then builds `bin/server`.
- `make run` builds and starts the server. Default address is `:8080`; override with `ADDR=:3000 make run`.
- `make dev` starts hot reload with `air` using `.air.toml`; install `air` first if it is missing.
- `make test` formats, vets, and runs `go test -v ./...`.
- `make test-cover` writes `coverage.out` and `coverage.html`.
- `make lint` runs `golangci-lint run ./...`.
- `make check` runs formatting, vetting, linting, and tests.
- `make docker-up` and `make docker-down` manage the Compose environment.

## Coding Style & Naming Conventions

Use standard Go formatting: run `go fmt ./...` or `make fmt` before committing. Keep package names short and lowercase, and export names only when the API crosses package boundaries. Tests should use Go’s `TestNameBehavior` style, as in `TestHomePageUsesOptimizedHeroImage`.

Keep HTML template changes in `internal/web/templates/` and static asset changes in `internal/web/static/`. When adding served files, confirm they are included by the existing embedded `static` directory.

## Testing Guidelines

Use the standard Go `testing` package and `httptest` for route behavior. Add focused tests when changing headers, cache behavior, template output, environment-driven URLs, or static assets. Run `make test` for normal verification and `make test-cover` when changing broader behavior.

## Commit & Pull Request Guidelines

Recent history uses concise Conventional Commit-style subjects, often with scopes: `feat(web): add social links`, `fix(site): update teaser trailer copy`, `perf(web): optimize landing page first render`. Follow that pattern when practical.

Pull requests should include a short description, the user-visible impact, tests run, and screenshots for visual template or CSS changes. Link related issues when available and call out any configuration changes such as new environment variables.

## Security & Configuration Tips

`SPOTIFY_URL` customizes the outbound Spotify link, and `ADDR` controls the listen address. Keep the Content Security Policy in `internal/web/web.go` aligned with any new external embeds or asset origins, and add tests for CSP changes.

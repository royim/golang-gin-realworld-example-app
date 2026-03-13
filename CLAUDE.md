# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go REST API implementing the [RealWorld](https://github.com/gothinkster/realworld) spec ("Conduit" blogging platform) using Gin v1.10, GORM v2 with SQLite, and JWT v5 authentication. Requires Go 1.21+ and cgo (for SQLite driver).

## Commands

```bash
# Run server (default port 8080)
go run hello.go
PORT=3000 go run hello.go

# Run all tests
go test ./...

# Run tests for a single package
go test ./articles
go test ./users
go test ./common

# Run a specific test function
go test ./articles -run TestArticleCreate

# Test with coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Build
go build ./...

# Tidy dependencies
go mod tidy
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `GIN_MODE` | `debug` | Gin mode (`debug` or `release`) |
| `DB_PATH` | `./data/gorm.db` | SQLite database path |
| `TEST_DB_PATH` | `./data/gorm_test.db` | Test database path |

## Architecture

### Domain-based package structure

Each domain package (`users/`, `articles/`) follows the same 5-file convention:

| File | Purpose |
|------|---------|
| `models.go` | GORM models, DB operations (find, save, update, delete) |
| `routers.go` | HTTP handlers + route registration functions |
| `serializers.go` | Response DTOs and JSON formatting logic |
| `validators.go` | Request binding/validation structs |
| `unit_test.go` | Integration tests using real SQLite DB |

The `common/` package provides shared infrastructure: database init (`database.go`), JWT/error utilities (`utils.go`), and test helpers (`test_helpers.go`). The `users/` package also has `middlewares.go` for auth middleware.

### Request flow

1. `hello.go` — entrypoint, sets up Gin router with route groups under `/api`
2. Routes are registered via `XxxRegister()` functions from each package
3. `AuthMiddleware(auto401 bool)` — if `auto401=false`, allows anonymous access but still parses token if present; if `true`, returns 401 when no valid token
4. Handler calls validator's `Bind()` → model DB operation → serializer `Response()`

### Route registration pattern (hello.go)

Anonymous routes (no auth required) are registered first with `AuthMiddleware(false)`, then authenticated routes use `AuthMiddleware(true)`. The middleware stores `my_user_model` and `my_user_id` in gin.Context, accessed via `c.MustGet("my_user_model")`. Note: `r.RedirectTrailingSlash = false` is set to prevent POST body loss during redirects.

### Key model relationships

- `UserModel` ↔ `FollowModel` — many-to-many self-referential (following/followers)
- `ArticleUserModel` — bridge between `users.UserModel` and article domain; created via `GetArticleUserModel()` using `FirstOrCreate`
- `ArticleModel` → `TagModel` (many2many), `CommentModel` (has-many), `FavoriteModel` (many-to-many via `ArticleUserModel`)

### Auth scheme

- JWT tokens use `Authorization: Token <jwt>` header format (not `Bearer`)
- Token can also be passed via `access_token` query parameter
- Token contains `id` (user ID) and `exp` (24h expiry) claims, signed with HS256
- `common.GenToken(id)` generates tokens; `common.HeaderTokenMock(req, id)` adds auth headers in tests

### Database

- SQLite via GORM v2 with `gorm.io/driver/sqlite` (requires cgo)
- Global DB singleton: `common.DB`, accessed via `common.GetDB()`
- `Migrate()` in `hello.go` runs auto-migration for all models on startup
- Test DB: `common.TestDBInit()` creates isolated test DB, `common.TestDBFree()` cleans up

### N+1 query optimization

List endpoints use batch operations: `BatchGetFavoriteCounts()`, `BatchGetFavoriteStatus()` for favorites, and `Preload("Author.UserModel").Preload("Tags")` for eager loading. The `ArticlesSerializer.Response()` uses `ResponseWithPreloaded()` to avoid per-article DB queries.

### Testing pattern

Tests are integration-style using a real SQLite database. Each test file calls `TestDBInit()` in setup and `TestDBFree()` in teardown. Tests exercise the full HTTP handler stack via `httptest.NewRecorder()` with Gin's test mode. Validation uses `common.Bind()` (wraps `ShouldBindWith`) instead of Gin's default `MustBindWith` to avoid automatic 400 responses.

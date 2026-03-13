# Project Overview

## Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.21+ |
| Web Framework | Gin | v1.10 |
| ORM | GORM v2 | gorm.io/gorm |
| Database | SQLite | gorm.io/driver/sqlite (cgo required) |
| Auth | JWT | golang-jwt/jwt/v5 |
| Password | bcrypt | golang.org/x/crypto/bcrypt |

## Architecture

Domain-based package structure under `/api` prefix:

```
hello.go          — entrypoint, router setup, migration
common/
  database.go     — DB init, singleton, test helpers
  utils.go        — JWT, random, error formatting, bind helper
  test_helpers.go — shared test utilities
users/
  models.go       — UserModel, FollowModel, DB operations
  routers.go      — auth/profile HTTP handlers
  serializers.go  — UserResponse, ProfileResponse
  validators.go   — registration, login, update validators
  middlewares.go   — AuthMiddleware
  unit_test.go    — integration tests
articles/
  models.go       — Article, Tag, Comment, Favorite models + batch queries
  routers.go      — article/comment/tag HTTP handlers
  serializers.go  — Article, Comment, Tag responses
  validators.go   — article, comment validators
  unit_test.go    — integration tests
```

## Request Flow

```
Client → Gin Router → AuthMiddleware → Handler
  Handler: Validator.Bind(c) → Model DB op → Serializer.Response() → c.JSON()
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server listen port |
| `GIN_MODE` | `debug` | `debug` or `release` |
| `DB_PATH` | `./data/gorm.db` | SQLite database file path |
| `TEST_DB_PATH` | `./data/gorm_test.db` | Test database file path |

## Route Registration (hello.go)

Two middleware groups applied in order:

1. **Anonymous group** — `AuthMiddleware(false)`: parses token if present, allows anonymous
2. **Authenticated group** — `AuthMiddleware(true)`: requires valid token, returns 401 otherwise

Setting: `r.RedirectTrailingSlash = false` to prevent POST body loss on redirect.

## Error Response Format

All errors use a consistent JSON structure:

```json
{
  "errors": {
    "<field>": "<message>"
  }
}
```

### HTTP Status Codes

| Code | Usage |
|------|-------|
| 200 | Success |
| 201 | Resource created (registration, article create, comment create) |
| 400 | Validation error |
| 401 | Authentication required or invalid token |
| 403 | Forbidden (not the author) |
| 404 | Resource not found |
| 422 | Unprocessable entity (DB error, duplicate data) |

## Database Migration

`Migrate()` in hello.go runs AutoMigrate for all models on startup:

```
UserModel → FollowModel → ArticleModel → TagModel
→ FavoriteModel → ArticleUserModel → CommentModel
```

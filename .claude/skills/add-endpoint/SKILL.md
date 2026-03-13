---
name: add-endpoint
description: Generate a new API endpoint following the project's 5-file convention (models, routers, serializers, validators, test). Use when adding new routes or extending existing domain packages.
argument-hint: [endpoint description, e.g. "GET /api/health healthcheck"]
---

Generate a new API endpoint following this project's conventions.

**Input**: $ARGUMENTS — a description of the endpoint to create (HTTP method, path, purpose).

If no input is provided, ask what endpoint the user wants to add.

---

## Reference Specs

Before generating code, read these project specs for context:

1. `openspec/specs/overview.md` — architecture, request flow, error format, status codes
2. `openspec/specs/models.md` — all models, relationships, DB functions
3. `openspec/specs/middleware.md` — auth middleware modes (auto401 true/false)
4. The relevant domain spec (`openspec/specs/auth.md`, `openspec/specs/articles.md`, etc.) for existing patterns

## Steps

1. **Determine target package**
   - If the endpoint fits an existing domain (users/, articles/), extend that package
   - If it's a new domain, create a new package directory

2. **Read existing files** in the target package to match patterns exactly:
   - `models.go` — model struct style, DB operation patterns
   - `routers.go` — handler function style, route registration pattern
   - `serializers.go` — response struct and Response() method pattern
   - `validators.go` — validator struct and Bind() method pattern
   - `unit_test.go` — test setup/teardown, httptest patterns

3. **Generate code** following the 5-file convention:

   ### models.go
   - GORM model struct with proper tags (`gorm:"column:...;size:..."`)
   - DB operation functions: `FindOne*()`, `SaveOne()`, `Delete*()`
   - Method receivers for model-specific operations
   - Use `common.GetDB()` for database access

   ### routers.go
   - Handler function: `func EndpointName(c *gin.Context)`
   - Route registration function: `func DomainRegister(router *gin.RouterGroup)`
   - Follow pattern: validate → model op → serialize → c.JSON()
   - Error responses use `common.NewError(key, err)` format
   - Status codes: 200 (success), 201 (created), 400 (validation), 401 (auth), 403 (forbidden), 404 (not found), 422 (DB error)

   ### serializers.go
   - Response struct with `json:"camelCase"` tags
   - Serializer struct: `type XxxSerializer struct { C *gin.Context; XxxModel }`
   - `Response()` method returning the response struct
   - Timestamps as `time.RFC3339Nano` format

   ### validators.go
   - Validator struct with nested request body: `json:"<root>"` + `binding:"required"`
   - `Bind(c *gin.Context) error` method using `common.Bind(c, &validator)`
   - `New*Validator()` constructor
   - `New*ValidatorFillWith(model)` for update endpoints

   ### unit_test.go
   - `TestMain(m *testing.M)` with `common.TestDBInit()` / `TestDBFree()`
   - At least: success case, auth failure (401), validation error (400)
   - Use `httptest.NewRecorder()` + `gin.SetMode(gin.TestMode)`
   - Use `common.HeaderTokenMock(req, id)` for authenticated requests

4. **Update hello.go** if needed:
   - Add route registration call in the appropriate middleware group
   - Anonymous endpoints → `AuthMiddleware(false)` group
   - Authenticated endpoints → `AuthMiddleware(true)` group
   - Add `AutoMigrate()` call in `Migrate()` if new model created

5. **Verify** — run `go build ./...` and `go test ./<package>/` to confirm

## Output

After generating, summarize:
- Files created/modified
- Route path and HTTP method
- Auth requirement
- How to test: `curl` or `go test` command

## Guardrails
- NEVER modify existing tests — only add new test functions
- ALWAYS match the exact coding patterns in existing files
- ALWAYS include `trailing slash` duplicate route (e.g., both `/articles` and `/articles/`)
- Use `Preload()` for relationships to avoid N+1 queries
- Set `r.RedirectTrailingSlash = false` is already configured — don't change it

---
name: add-test
description: Generate integration tests for existing handlers or model functions. Creates tests with real SQLite DB using httptest, covering success and error cases. Use when adding test coverage to existing code.
argument-hint: [package/function, e.g. "articles/ArticleCreate" or "users/"]
---

Generate integration tests for existing handlers or model functions.

**Input**: $ARGUMENTS — target package, handler, or function name.

If no input is provided, ask what to test.

---

## Reference Specs

Read these before generating tests:

1. `openspec/specs/common.md` — **test patterns section** (TestDBInit, TestDBFree, HeaderTokenMock, test structure)
2. The relevant domain spec for request/response schemas:
   - `openspec/specs/auth.md` — registration, login, user CRUD
   - `openspec/specs/profiles.md` — profile, follow/unfollow
   - `openspec/specs/articles.md` — article CRUD, feed, list
   - `openspec/specs/comments.md` — comment CRUD
   - `openspec/specs/favorites.md` — favorite/unfavorite
3. `openspec/specs/middleware.md` — auth behavior for both modes

## Steps

1. **Read the target source code** to understand:
   - Function signature and behavior
   - Database operations performed
   - Error conditions and response codes
   - Auth requirements (check route group in hello.go)

2. **Read existing tests** in the target package (`unit_test.go`) to match patterns

3. **Generate test functions** covering these cases:

   ### For HTTP Handlers
   ```
   TestXxxSuccess          — happy path, verify response body and status
   TestXxxUnauthorized     — no/invalid token → 401 (if auth required)
   TestXxxForbidden        — wrong user → 403 (if ownership check)
   TestXxxValidationError  — invalid input → 400
   TestXxxNotFound         — missing resource → 404
   ```

   ### For Model Functions
   ```
   TestXxxCreate           — successful creation
   TestXxxFind             — successful query
   TestXxxUpdate           — field updates
   TestXxxDelete           — deletion and verification
   TestXxxEdgeCases        — empty inputs, duplicates, constraints
   ```

4. **Test structure** (must match existing pattern):

   ```go
   func TestHandlerName(t *testing.T) {
       // Setup: create test data
       // Build request
       req, _ := http.NewRequest("METHOD", "/api/path", bytes.NewBufferString(`{...}`))
       req.Header.Set("Content-Type", "application/json")
       common.HeaderTokenMock(req, userID) // if auth needed

       // Execute
       w := httptest.NewRecorder()
       router.ServeHTTP(w, req)

       // Assert
       assert.Equal(t, expectedStatus, w.Code)
       // Parse and verify response body
   }
   ```

5. **Test data setup**:
   - Create required users/articles/etc. directly via model functions
   - Do NOT rely on data from other tests (each test self-contained)
   - Clean up test data if it could affect other tests

6. **Verify** — run `go test -v ./<package>/ -run TestNewFunction`

## Output

After generating, show:
- Number of test functions created
- Cases covered (success, auth, validation, etc.)
- How to run: `go test -v ./<package>/ -run <pattern>`
- Current coverage: `go test -cover ./<package>/`

## Guardrails
- NEVER modify existing test functions — only add new ones
- ALWAYS use `common.TestDBInit()` / `TestDBFree()` pattern (check if TestMain exists)
- ALWAYS use real DB, never mock the database
- ALWAYS use `common.HeaderTokenMock()` for auth, never construct JWT manually
- ALWAYS use `common.Bind()` pattern (not Gin's default MustBindWith)
- Test response JSON structure matches RealWorld spec (check domain spec files)
- Use `testify/assert` for assertions, not manual if/t.Error

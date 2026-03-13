# Common Package Spec

## Database (common/database.go)

### Global Singleton

```go
var DB *gorm.DB
```

Accessed via `common.GetDB()` throughout the application.

### Init Functions

```go
func Init() *gorm.DB
```
- Path: `DB_PATH` env var or `./data/gorm.db`
- Auto-creates directory if not exists
- Max idle connections: 10
- Opens SQLite via `gorm.io/driver/sqlite`

```go
func TestDBInit() *gorm.DB
```
- Path: `TEST_DB_PATH` env var or `./data/gorm_test.db`
- Max idle connections: 3
- Logger: Info level enabled
- Used in test setup

```go
func TestDBFree(test_db *gorm.DB) error
```
- Closes DB connection via `sqlDB.Close()`
- Deletes test DB file via `os.Remove()`
- Used in test teardown

### Path Helpers

```go
func GetDBPath() string      // returns DB_PATH or default
func GetTestDBPath() string  // returns TEST_DB_PATH or default
```

---

## Utilities (common/utils.go)

### Constants

```go
const JWTSecret = "A String Very Very Very Strong!!@##$!@#$"
const RandomPassword = "A String Very Very Very Random!!@##$!@#4"
```

### JWT

```go
func GenToken(id uint) string
```
- Signs with HS256 using `JWTSecret`
- Claims: `id` (user ID), `exp` (now + 24h)
- Returns empty string on failure

### Random

```go
func RandString(n int) string  // alphanumeric, length n
func RandInt() int             // 0-999999
```

### Error Formatting

```go
type CommonError struct {
    Errors map[string]interface{} `json:"errors"`
}

func NewValidatorError(err error) CommonError
```
- Converts `validator.ValidationErrors` to readable map
- Maps each field name to its validation tag
- Example output: `{"errors": {"username": "required"}}`

```go
func NewError(key string, err error) CommonError
```
- Creates error with custom key
- Example: `NewError("article", err)` → `{"errors": {"article": "not found"}}`

### Request Binding

```go
func Bind(c *gin.Context, obj interface{}) error
```
- Wraps `c.ShouldBindWith()` instead of Gin's default `MustBindWith`
- Prevents automatic 400 response on bind failure
- Allows manual error handling in handlers
- Content-type detection: JSON body → `binding.JSON`, otherwise `binding.Form`

---

## Test Helpers (common/test_helpers.go)

```go
func HeaderTokenMock(req *http.Request, id uint)
```
- Generates JWT for given user ID via `GenToken(id)`
- Sets `Authorization: Token <jwt>` header on request
- Used in integration tests to authenticate requests

---

## Testing Pattern

All test files follow this pattern:

```go
func TestMain(m *testing.M) {
    // Setup
    testDB := common.TestDBInit()
    common.DB = testDB
    // Run migrations
    AutoMigrate()

    // Run tests
    code := m.Run()

    // Teardown
    common.TestDBFree(testDB)
    os.Exit(code)
}

func TestSomeHandler(t *testing.T) {
    // Setup Gin test mode
    r := gin.New()
    // Register routes
    // Create httptest.NewRecorder()
    // Make request
    // Assert response
}
```

### Key Points
- Real SQLite DB (not mocked)
- Each test file has isolated DB lifecycle
- `httptest.NewRecorder()` for HTTP testing
- `gin.SetMode(gin.TestMode)` for test mode
- `common.HeaderTokenMock(req, userID)` for authenticated requests

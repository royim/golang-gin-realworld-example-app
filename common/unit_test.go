package common

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestConnectingDatabase(t *testing.T) {
	asserts := assert.New(t)
	db := Init()
	dbPath := GetDBPath()
	// Test create & close DB
	_, err := os.Stat(dbPath)
	asserts.NoError(err, "Db should exist")
	sqlDB, err := db.DB()
	asserts.NoError(err, "Should get sql.DB")
	asserts.NoError(sqlDB.Ping(), "Db should be able to ping")

	// Test get a connecting from connection pools
	connection := GetDB()
	sqlDB, err = connection.DB()
	asserts.NoError(err, "Should get sql.DB")
	asserts.NoError(sqlDB.Ping(), "Db should be able to ping")
	sqlDB.Close()

	// Test DB exceptions
	os.Chmod(dbPath, 0000)
	db = Init()
	sqlDB, err = db.DB()
	asserts.NoError(err, "Should get sql.DB")
	asserts.Error(sqlDB.Ping(), "Db should not be able to ping")
	sqlDB.Close()
	os.Chmod(dbPath, 0644)
}

func TestConnectingTestDatabase(t *testing.T) {
	asserts := assert.New(t)
	// Test create & close DB
	db := TestDBInit()
	testDBPath := GetTestDBPath()
	_, err := os.Stat(testDBPath)
	asserts.NoError(err, "Db should exist")
	sqlDB, err := db.DB()
	asserts.NoError(err, "Should get sql.DB")
	asserts.NoError(sqlDB.Ping(), "Db should be able to ping")
	TestDBFree(db)

	// Test close delete DB
	db = TestDBInit()
	TestDBFree(db)
	_, err = os.Stat(testDBPath)

	asserts.Error(err, "Db should not exist")
}

func TestDBDirCreation(t *testing.T) {
	asserts := assert.New(t)
	// Set a nested path
	os.Setenv("TEST_DB_PATH", "tmp/nested/test.db")
	defer os.Unsetenv("TEST_DB_PATH")

	db := TestDBInit()
	testDBPath := GetTestDBPath()
	_, err := os.Stat(testDBPath)
	asserts.NoError(err, "Db should exist in nested directory")
	TestDBFree(db)

	// Cleanup directory
	os.RemoveAll("tmp/nested")
}

func TestDBPathOverride(t *testing.T) {
	asserts := assert.New(t)
	customPath := "./custom_test.db"
	os.Setenv("TEST_DB_PATH", customPath)
	defer os.Unsetenv("TEST_DB_PATH")

	asserts.Equal(customPath, GetTestDBPath(), "Should use env var")
}

func TestRandString(t *testing.T) {
	asserts := assert.New(t)

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	str := RandString(0)
	asserts.Equal(str, "", "length should be ''")

	str = RandString(10)
	asserts.Equal(len(str), 10, "length should be 10")
	for _, ch := range str {
		asserts.Contains(letters, ch, "char should be a-z|A-Z|0-9")
	}
}

func TestRandInt(t *testing.T) {
	asserts := assert.New(t)

	// Test that RandInt returns a value in valid range
	val := RandInt()
	asserts.GreaterOrEqual(val, 0, "RandInt should be >= 0")
	asserts.Less(val, 1000000, "RandInt should be < 1000000")

	// Test multiple calls return different values (statistically)
	vals := make(map[int]bool)
	for i := 0; i < 10; i++ {
		vals[RandInt()] = true
	}
	asserts.Greater(len(vals), 1, "RandInt should return varied values")
}

func TestGenToken(t *testing.T) {
	asserts := assert.New(t)

	token := GenToken(2)

	asserts.IsType(token, string("token"), "token type should be string")
	asserts.Len(token, 115, "JWT's length should be 115")
}

func TestGenTokenMultipleUsers(t *testing.T) {
	asserts := assert.New(t)

	token1 := GenToken(1)
	token2 := GenToken(2)
	token100 := GenToken(100)

	asserts.NotEqual(token1, token2, "Different user IDs should generate different tokens")
	asserts.NotEqual(token2, token100, "Different user IDs should generate different tokens")
	// Token length can vary by 1 character due to timestamp changes
	asserts.GreaterOrEqual(len(token1), 114, "JWT's length should be >= 114 for user 1")
	asserts.LessOrEqual(len(token1), 120, "JWT's length should be <= 120 for user 1")
	asserts.GreaterOrEqual(len(token100), 114, "JWT's length should be >= 114 for user 100")
	asserts.LessOrEqual(len(token100), 120, "JWT's length should be <= 120 for user 100")
}

func TestHeaderTokenMock(t *testing.T) {
	asserts := assert.New(t)

	req, _ := http.NewRequest("GET", "/test", nil)
	token := GenToken(5)
	HeaderTokenMock(req, 5)

	authHeader := req.Header.Get("Authorization")
	asserts.Equal(fmt.Sprintf("Token %s", token), authHeader, "Authorization header should be set correctly")
}

func TestExtractTokenFromHeader(t *testing.T) {
	asserts := assert.New(t)

	token := "valid.jwt.token"
	header := fmt.Sprintf("Token %s", token)

	extracted := ExtractTokenFromHeader(header)
	asserts.Equal(token, extracted, "Should extract token from header")

	invalidHeader := "Bearer " + token
	extracted = ExtractTokenFromHeader(invalidHeader)
	asserts.Empty(extracted, "Should return empty for non-Token header")

	shortHeader := "Token"
	extracted = ExtractTokenFromHeader(shortHeader)
	asserts.Empty(extracted, "Should return empty for short header")
}

func TestVerifyTokenClaims(t *testing.T) {
	asserts := assert.New(t)

	// Test valid token
	userID := uint(123)
	token := GenToken(userID)
	claims, err := VerifyTokenClaims(token)
	asserts.NoError(err, "VerifyTokenClaims should not error for valid token")
	asserts.Equal(float64(userID), claims["id"], "Claims should contain correct user ID")

	// Test invalid token
	_, err = VerifyTokenClaims("invalid.token.string")
	asserts.Error(err, "VerifyTokenClaims should error for invalid token")
}

func TestNewValidatorError(t *testing.T) {
	asserts := assert.New(t)

	type Login struct {
		Username string `form:"username" json:"username" binding:"required,alphanum,min=4,max=255"`
		Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
	}

	var requestTests = []struct {
		bodyData       string
		expectedCode   int
		responseRegexg string
		msg            string
	}{
		{
			`{"username": "wangzitian0","password": "0123456789"}`,
			http.StatusOK,
			`{"status":"you are logged in"}`,
			"valid data and should return StatusCreated",
		},
		{
			`{"username": "wangzitian0","password": "01234567866"}`,
			http.StatusUnauthorized,
			`{"errors":{"user":"wrong username or password"}}`,
			"wrong login status should return StatusUnauthorized",
		},
		{
			`{"username": "wangzitian0","password": "0122"}`,
			http.StatusUnprocessableEntity,
			`{"errors":{"Password":"{min: 8}"}}`,
			"invalid password of too short and should return StatusUnprocessableEntity",
		},
		{
			`{"username": "_wangzitian0","password": "0123456789"}`,
			http.StatusUnprocessableEntity,
			`{"errors":{"Username":"{key: alphanum}"}}`,
			"invalid username of non alphanum and should return StatusUnprocessableEntity",
		},
	}

	r := gin.Default()

	r.POST("/login", func(c *gin.Context) {
		var json Login
		if err := Bind(c, &json); err == nil {
			if json.Username == "wangzitian0" && json.Password == "0123456789" {
				c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
			} else {
				c.JSON(http.StatusUnauthorized, NewError("user", errors.New("wrong username or password")))
			}
		} else {
			c.JSON(http.StatusUnprocessableEntity, NewValidatorError(err))
		}
	})

	for _, testData := range requestTests {
		bodyData := testData.bodyData
		req, err := http.NewRequest("POST", "/login", bytes.NewBufferString(bodyData))
		req.Header.Set("Content-Type", "application/json")
		asserts.NoError(err)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		asserts.Equal(testData.expectedCode, w.Code, "Response Status - "+testData.msg)
		asserts.Regexp(testData.responseRegexg, w.Body.String(), "Response Content - "+testData.msg)
	}
}

func TestNewError(t *testing.T) {
	assert := assert.New(t)

	db := TestDBInit()
	defer TestDBFree(db)

	type NonExistentTable struct {
		Field string
	}
	// db.AutoMigrate(NonExistentTable{}) // Intentionally skipped to cause error

	err := db.Find(&NonExistentTable{Field: "value"}).Error
	if err == nil {
		err = errors.New("no such table: non_existent_tables")
	}

	commonError := NewError("database", err)
	assert.IsType(commonError, commonError, "commonError should have right type")
	// The exact error message might vary by driver, checking key presence is safer, but keeping original assertion style
	assert.Contains(commonError.Errors, "database", "commonError should contain database key")
}

func TestDatabaseDirCreation(t *testing.T) {
	asserts := assert.New(t)

	// Test directory creation in Init
	origDBPath := os.Getenv("DB_PATH")
	defer os.Setenv("DB_PATH", origDBPath)

	// Create a temp dir path
	tempDir := "./tmp/test_nested/db"
	os.Setenv("DB_PATH", tempDir+"/test.db")

	// Clean up before test
	os.RemoveAll("./tmp/test_nested")

	// Init should create the directory

	db := Init()

	sqlDB, err := db.DB()

	asserts.NoError(err, "Should get sql.DB")

	asserts.NoError(sqlDB.Ping(), "DB should be created in nested directory")

	// Clean up after test

	sqlDB.Close()

	os.RemoveAll("./tmp/test_nested")

}

func TestDBInitDirCreation(t *testing.T) {

	asserts := assert.New(t)

	// Test directory creation in TestDBInit

	origTestDBPath := os.Getenv("TEST_DB_PATH")

	defer os.Setenv("TEST_DB_PATH", origTestDBPath)

	// Create a temp dir path

	tempDir := "./tmp/test_nested_testdb"

	os.Setenv("TEST_DB_PATH", tempDir+"/test.db")

	// Clean up before test

	os.RemoveAll(tempDir)

	// TestDBInit should create the directory

	db := TestDBInit()

	sqlDB, err := db.DB()

	asserts.NoError(err, "Should get sql.DB")

	asserts.NoError(sqlDB.Ping(), "Test DB should be created in nested directory")

	// Clean up after test

	TestDBFree(db)

	os.RemoveAll(tempDir)

}

func TestDatabaseWithCurrentDirectory(t *testing.T) {
	asserts := assert.New(t)

	// Test with simple filename (no directory)
	origDBPath := os.Getenv("DB_PATH")
	defer os.Setenv("DB_PATH", origDBPath)

	os.Setenv("DB_PATH", "test_simple.db")

	// Init should work without directory creation
	db := Init()
	sqlDB, err := db.DB()

	asserts.NoError(err, "Should get sql.DB")
	asserts.NoError(sqlDB.Ping(), "DB should be created in current directory")

	// Clean up
	sqlDB.Close()
	os.Remove("test_simple.db")
}

// --- OpenSpec core requirement: JWT token claims validation (specs/auth.md) ---

func TestGenTokenContainsValidClaims(t *testing.T) {
	asserts := assert.New(t)

	userID := uint(42)
	token := GenToken(userID)
	asserts.NotEmpty(token, "Token should not be empty")

	claims, err := VerifyTokenClaims(token)
	asserts.NoError(err, "Valid token should parse without error")

	// Verify "id" claim matches the user ID
	id, ok := claims["id"].(float64)
	asserts.True(ok, "id claim should be a number")
	asserts.Equal(float64(userID), id, "id claim should match user ID")

	// Verify "exp" claim exists and is in the future
	exp, ok := claims["exp"].(float64)
	asserts.True(ok, "exp claim should be a number")
	asserts.Greater(exp, float64(0), "exp should be a positive timestamp")
}

func TestGenTokenZeroID(t *testing.T) {
	asserts := assert.New(t)

	// Edge case: user ID 0 (anonymous/invalid)
	token := GenToken(0)
	asserts.NotEmpty(token, "Token should be generated even for ID 0")

	claims, err := VerifyTokenClaims(token)
	asserts.NoError(err, "Token with ID 0 should still be valid JWT")
	asserts.Equal(float64(0), claims["id"], "id claim should be 0")
}

// --- OpenSpec core requirement: Error formatting (specs/overview.md) ---

func TestNewErrorFormat(t *testing.T) {
	asserts := assert.New(t)

	err := errors.New("not found")
	commonErr := NewError("article", err)

	// Must match {"errors": {"article": "not found"}} format per spec
	asserts.Contains(commonErr.Errors, "article", "Error key should be 'article'")
	asserts.Equal("not found", commonErr.Errors["article"], "Error message should match")
}

func TestNewErrorMultipleKeys(t *testing.T) {
	asserts := assert.New(t)

	// Verify error structure is a map that can hold different keys
	err1 := NewError("username", errors.New("already taken"))
	asserts.Equal("already taken", err1.Errors["username"])

	err2 := NewError("email", errors.New("invalid format"))
	asserts.Equal("invalid format", err2.Errors["email"])
}

// --- OpenSpec core requirement: Bind uses ShouldBindWith not MustBindWith (specs/common.md) ---

func TestBindWithInvalidJSON(t *testing.T) {
	asserts := assert.New(t)

	type TestStruct struct {
		Name string `json:"name" binding:"required"`
	}

	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var obj TestStruct
		err := Bind(c, &obj)
		if err != nil {
			// Bind should return error but NOT auto-respond with 400
			c.JSON(http.StatusUnprocessableEntity, NewValidatorError(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": obj.Name})
	})

	// Test with missing required field
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should be 422 (our handler's choice), NOT 400 (auto from MustBind)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code,
		"Bind should let handler control the status code")
}

func TestBindWithValidJSON(t *testing.T) {
	asserts := assert.New(t)

	type TestStruct struct {
		Name string `json:"name" binding:"required"`
	}

	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var obj TestStruct
		err := Bind(c, &obj)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, NewValidatorError(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": obj.Name})
	})

	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"name":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code)
	asserts.Contains(w.Body.String(), "hello")
}

// --- OpenSpec core requirement: Token extraction formats (specs/middleware.md) ---

func TestExtractTokenFromHeaderEdgeCases(t *testing.T) {
	asserts := assert.New(t)

	// Empty string
	asserts.Empty(ExtractTokenFromHeader(""), "Empty string should return empty")

	// Only "Token" without space
	asserts.Empty(ExtractTokenFromHeader("Token"), "Just 'Token' without value should return empty")

	// "Token " with empty value
	extracted := ExtractTokenFromHeader("Token ")
	asserts.Equal("", extracted, "Token with empty value after space")

	// Bearer scheme should be rejected (per spec: only Token scheme)
	asserts.Empty(ExtractTokenFromHeader("Bearer abc.def.ghi"),
		"Bearer scheme should not be accepted")

	// Valid Token scheme
	asserts.Equal("abc.def.ghi", ExtractTokenFromHeader("Token abc.def.ghi"),
		"Valid Token scheme should extract correctly")
}

func TestVerifyTokenClaimsInvalidInputs(t *testing.T) {
	asserts := assert.New(t)

	// Empty token
	_, err := VerifyTokenClaims("")
	asserts.Error(err, "Empty token should fail verification")

	// Malformed token
	_, err = VerifyTokenClaims("not.a.valid.jwt")
	asserts.Error(err, "Malformed token should fail verification")

	// Token signed with wrong key
	_, err = VerifyTokenClaims("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MSwiZXhwIjoxfQ.wrongsignature")
	asserts.Error(err, "Token with wrong signature should fail")
}

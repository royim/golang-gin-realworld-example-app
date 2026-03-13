package users

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/gothinkster/golang-gin-realworld-example-app/common"
)

var image_url = "https://golang.org/doc/gopher/frontpage.png"
var test_db *gorm.DB

func newUserModel() UserModel {
	return UserModel{
		ID:           2,
		Username:     "asd123!@#ASD",
		Email:        "wzt@g.cn",
		Bio:          "heheda",
		Image:        &image_url,
		PasswordHash: "",
	}
}

func userModelMocker(n int) []UserModel {
	var offset int64
	test_db.Model(&UserModel{}).Count(&offset)
	var ret []UserModel
	for i := int(offset) + 1; i <= int(offset)+n; i++ {
		image := fmt.Sprintf("http://image/%v.jpg", i)
		userModel := UserModel{
			Username: fmt.Sprintf("user%v", i),
			Email:    fmt.Sprintf("user%v@linkedin.com", i),
			Bio:      fmt.Sprintf("bio%v", i),
			Image:    &image,
		}
		userModel.setPassword("password123")
		test_db.Create(&userModel)
		ret = append(ret, userModel)
	}
	return ret
}

func TestUserModel(t *testing.T) {
	asserts := assert.New(t)

	//Testing UserModel's password feature
	userModel := newUserModel()
	err := userModel.checkPassword("")
	asserts.Error(err, "empty password should return err")

	userModel = newUserModel()
	err = userModel.setPassword("")
	asserts.Error(err, "empty password can not be set null")

	userModel = newUserModel()
	err = userModel.setPassword("asd123!@#ASD")
	asserts.NoError(err, "password should be set successful")
	asserts.Len(userModel.PasswordHash, 60, "password hash length should be 60")

	err = userModel.checkPassword("sd123!@#ASD")
	asserts.Error(err, "password should be checked and not validated")

	err = userModel.checkPassword("asd123!@#ASD")
	asserts.NoError(err, "password should be checked and validated")

	//Testing the following relationship between users
	users := userModelMocker(3)
	a := users[0]
	b := users[1]
	c := users[2]
	asserts.Equal(0, len(a.GetFollowings()), "GetFollowings should be right before following")
	asserts.Equal(false, a.isFollowing(b), "isFollowing relationship should be right at init")
	a.following(b)
	asserts.Equal(1, len(a.GetFollowings()), "GetFollowings should be right after a following b")
	asserts.Equal(true, a.isFollowing(b), "isFollowing should be right after a following b")
	a.following(c)
	asserts.Equal(2, len(a.GetFollowings()), "GetFollowings be right after a following c")
	asserts.EqualValues(b, a.GetFollowings()[0], "GetFollowings should be right")
	asserts.EqualValues(c, a.GetFollowings()[1], "GetFollowings should be right")
	a.unFollowing(b)
	asserts.Equal(1, len(a.GetFollowings()), "GetFollowings should be right after a unFollowing b")
	asserts.EqualValues(c, a.GetFollowings()[0], "GetFollowings should be right after a unFollowing b")
	asserts.Equal(false, a.isFollowing(b), "isFollowing should be right after a unFollowing b")
}

// Reset test DB and create new one with mock data
func resetDBWithMock() {
	common.TestDBFree(test_db)
	test_db = common.TestDBInit()
	AutoMigrate()
	userModelMocker(3)
}

// You could write the init logic like reset database code here
var unauthRequestTests = []struct {
	init           func(*http.Request)
	url            string
	method         string
	bodyData       string
	expectedCode   int
	responseRegexg string
	msg            string
}{
	//Testing will run one by one, so you can combine it to a user story till another init().
	//And you can modified the header or body in the func(req *http.Request) {}

	//---------------------   Testing for user register   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
		},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusCreated,
		`{"user":{"username":"wangzitian0","email":"wzt@gg.cn","bio":"","image":"","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"valid data and should return StatusCreated",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"database":"UNIQUE constraint failed: user_models.email"}}`,
		"duplicated data and should return StatusUnprocessableEntity",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "u","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Username":"{min: 4}"}}`,
		"too short username should return error",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "j"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Password":"{min: 8}"}}`,
		"too short password should return error",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wztgg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Email":"{key: email}"}}`,
		"email invalid should return error",
	},

	//---------------------   Testing for user login   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
		},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "password123"}}`,
		http.StatusOK,
		`{"user":{"username":"user1","email":"user1@linkedin.com","bio":"bio1","image":"http://image/1.jpg","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"right info login should return user",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user112312312@linkedin.com","password": "password123"}}`,
		http.StatusUnauthorized,
		`{"errors":{"login":"Not Registered email or invalid password"}}`,
		"email not exist should return error info",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "password126"}}`,
		http.StatusUnauthorized,
		`{"errors":{"login":"Not Registered email or invalid password"}}`,
		"password error should return error info",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "passw"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Password":"{min: 8}"}}`,
		"password too short should return error info",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "passw"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Password":"{min: 8}"}}`,
		"password too short should return error info",
	},

	//---------------------   Testing for self info get & auth module  ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
		},
		"/user/",
		"GET",
		``,
		http.StatusUnauthorized,
		``,
		"request should return 401 without token",
	},
	{
		func(req *http.Request) {
			req.Header.Set("Authorization", fmt.Sprintf("Tokee %v", common.GenToken(1)))
		},
		"/user/",
		"GET",
		``,
		http.StatusUnauthorized,
		``,
		"wrong token should return 401",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 1)
		},
		"/user/",
		"GET",
		``,
		http.StatusOK,
		`{"user":{"username":"user1","email":"user1@linkedin.com","bio":"bio1","image":"http://image/1.jpg","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"request should return current user with token",
	},

	//---------------------   Testing for users' profile get   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":false}}`,
		"anonymous request should return profile with following=false",
	},
	{
		func(req *http.Request) {
			resetDBWithMock()
			common.HeaderTokenMock(req, 1)
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":false}}`,
		"request should return self profile",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":false}}`,
		"request should return correct other's profile",
	},

	//---------------------   Testing for users' profile update   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
			common.HeaderTokenMock(req, 1)
		},
		"/profiles/user123",
		"GET",
		``,
		http.StatusNotFound,
		``,
		"user should not exist profile before changed",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 1)
		},
		"/user/",
		"PUT",
		`{"user":{"username":"user123","password": "password126","email":"user123@linkedin.com","bio":"bio123","image":"http://hehe/123.jpg"}}`,
		http.StatusOK,
		`{"user":{"username":"user123","email":"user123@linkedin.com","bio":"bio123","image":"http://hehe/123.jpg","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"current user profile should be changed",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 1)
		},
		"/profiles/user123",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user123","bio":"bio123","image":"http://hehe/123.jpg","following":false}}`,
		"request should return self profile after changed",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user123@linkedin.com","password": "password126"}}`,
		http.StatusOK,
		`{"user":{"username":"user123","email":"user123@linkedin.com","bio":"bio123","image":"http://hehe/123.jpg","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"user should login using new password after changed",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/user/",
		"PUT",
		`{"user":{"password": "pas"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Password":"{min: 8}"}}`,
		"current user profile should not be changed with error user info",
	},

	//---------------------   Testing for db errors   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
			common.HeaderTokenMock(req, 4)
		},
		"/user/",
		"PUT",
		`{"password": "password321"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Email":"{key: required}","Username":"{key: required}"}}`,
		"test database pk error for user update",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 0)
		},
		"/user/",
		"PUT",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"database":"WHERE conditions required"}}`,
		"cheat validator and test database connecting error for user update",
	},
	{
		func(req *http.Request) {
			common.TestDBFree(test_db)
			test_db = common.TestDBInit()

			test_db.AutoMigrate(&UserModel{})
			userModelMocker(3)
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"POST",
		``,
		http.StatusUnprocessableEntity,
		`{"errors":{"database":"no such table: follow_models"}}`,
		"test database error for following",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"DELETE",
		``,
		http.StatusUnprocessableEntity,
		`{"errors":{"database":"no such table: follow_models"}}`,
		"test database error for canceling following",
	},
	{
		func(req *http.Request) {
			resetDBWithMock()
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user666/follow",
		"POST",
		``,
		http.StatusNotFound,
		`{"errors":{"profile":"Invalid username"}}`,
		"following wrong user name should return errors",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user666/follow",
		"DELETE",
		``,
		http.StatusNotFound,
		`{"errors":{"profile":"Invalid username"}}`,
		"cancel following wrong user name should return errors",
	},

	//---------------------   Testing for user following   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"POST",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":true}}`,
		"user follow another should work",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":true}}`,
		"user follow another should make sure database changed",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"DELETE",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":false}}`,
		"user cancel follow another should work",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":false}}`,
		"user cancel follow another should make sure database changed",
	},
}

func TestWithoutAuth(t *testing.T) {
	asserts := assert.New(t)
	//You could write the reset database code here if you want to create a database for this block
	//resetDB()

	r := gin.New()
	UsersRegister(r.Group("/users"))
	r.Use(AuthMiddleware(false))
	ProfileRetrieveRegister(r.Group("/profiles"))
	r.Use(AuthMiddleware(true))
	UserRegister(r.Group("/user"))
	ProfileRegister(r.Group("/profiles"))
	for _, testData := range unauthRequestTests {
		bodyData := testData.bodyData
		req, err := http.NewRequest(testData.method, testData.url, bytes.NewBufferString(bodyData))
		req.Header.Set("Content-Type", "application/json")
		asserts.NoError(err)

		testData.init(req)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		asserts.Equal(testData.expectedCode, w.Code, "Response Status - "+testData.msg)
		asserts.Regexp(testData.responseRegexg, w.Body.String(), "Response Content - "+testData.msg)
	}
}

func TestExtractTokenFromQueryParameter(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(AuthMiddleware(false))
	r.GET("/test", func(c *gin.Context) {
		userID := c.MustGet("my_user_id").(uint)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	resetDBWithMock()

	// Test with access_token query parameter
	token := common.GenToken(1)
	req, _ := http.NewRequest("GET", "/test?access_token="+token, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "Request with query token should succeed")
	asserts.Contains(w.Body.String(), `"user_id":1`, "User ID should be 1")
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(AuthMiddleware(true))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Test with invalid JWT token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Token invalid.jwt.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusUnauthorized, w.Code, "Invalid token should return 401")
}

func TestAuthMiddlewareNoToken(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(AuthMiddleware(false))
	r.GET("/test", func(c *gin.Context) {
		userID := c.MustGet("my_user_id").(uint)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// Test with no token (auto401=false should still proceed)
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "No token with auto401=false should proceed")
	asserts.Contains(w.Body.String(), `"user_id":0`, "User ID should be 0")
}

// --- OpenSpec core requirement: User model DB operations (specs/models.md) ---

func TestFindOneUserByUsername(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	// Find existing user
	user, err := FindOneUser(&UserModel{Username: "user1"})
	asserts.NoError(err, "Should find existing user by username")
	asserts.Equal("user1", user.Username)
	asserts.Equal("user1@linkedin.com", user.Email)

	// Find non-existing user
	_, err = FindOneUser(&UserModel{Username: "nonexistent"})
	asserts.Error(err, "Should return error for non-existing user")
}

func TestFindOneUserByEmail(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	user, err := FindOneUser(&UserModel{Email: "user2@linkedin.com"})
	asserts.NoError(err, "Should find existing user by email")
	asserts.Equal("user2", user.Username)
}

func TestUserModelUpdate(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	user, _ := FindOneUser(&UserModel{Username: "user1"})

	// Update bio field
	err := user.Update(UserModel{Bio: "updated bio"})
	asserts.NoError(err, "Update should succeed")

	// Verify update persisted
	updatedUser, _ := FindOneUser(&UserModel{ID: user.ID})
	asserts.Equal("updated bio", updatedUser.Bio, "Bio should be updated")
	asserts.Equal("user1", updatedUser.Username, "Username should remain unchanged")
}

func TestUserModelSaveOne(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	newUser := UserModel{
		Username: "newuser",
		Email:    "new@example.com",
		Bio:      "new bio",
	}
	newUser.setPassword("password123")

	err := SaveOne(&newUser)
	asserts.NoError(err, "SaveOne should succeed")
	asserts.NotEqual(uint(0), newUser.ID, "ID should be assigned after save")

	// Verify persisted
	found, err := FindOneUser(&UserModel{Username: "newuser"})
	asserts.NoError(err)
	asserts.Equal("new@example.com", found.Email)
}

// --- OpenSpec core requirement: Password security (specs/auth.md) ---

func TestPasswordHashIsDifferentFromInput(t *testing.T) {
	asserts := assert.New(t)

	user := newUserModel()
	user.setPassword("mypassword123")

	asserts.NotEqual("mypassword123", user.PasswordHash,
		"Password hash should differ from plaintext")
	asserts.Equal(60, len(user.PasswordHash),
		"bcrypt hash should be 60 characters")
}

func TestPasswordVerificationAfterSave(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	// User from mock has password "password123"
	user, _ := FindOneUser(&UserModel{Username: "user1"})

	asserts.NoError(user.checkPassword("password123"),
		"Correct password should verify")
	asserts.Error(user.checkPassword("wrongpassword"),
		"Wrong password should fail")
	asserts.Error(user.checkPassword(""),
		"Empty password should fail")
}

// --- OpenSpec core requirement: Follow self-referential M:N (specs/profiles.md, specs/models.md) ---

func TestFollowSelfIsAllowed(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	users := userModelMocker(1)
	a := users[0]

	// Following yourself should not error (spec doesn't explicitly forbid it)
	err := a.following(a)
	asserts.NoError(err, "Following self should not error")
	asserts.True(a.isFollowing(a), "isFollowing self should return true")
}

func TestFollowIdempotent(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	users := userModelMocker(2)
	a := users[0]
	b := users[1]

	// Follow twice should not create duplicate
	a.following(b)
	a.following(b)
	asserts.Equal(1, len(a.GetFollowings()),
		"Following same user twice should still result in 1 following")
}

func TestUnfollowNonFollowed(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	users := userModelMocker(2)
	a := users[0]
	b := users[1]

	// Unfollow someone not followed should not error
	err := a.unFollowing(b)
	asserts.NoError(err, "Unfollowing non-followed user should not error")
	asserts.False(a.isFollowing(b))
}

func TestGetFollowingsEmpty(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	users := userModelMocker(1)
	a := users[0]

	followings := a.GetFollowings()
	asserts.Equal(0, len(followings), "New user should have no followings")
}

// --- OpenSpec core requirement: Validator boundary values (specs/auth.md) ---

func TestUserRegistrationValidatorBoundary(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	r := gin.New()
	UsersRegister(r.Group("/users"))

	tests := []struct {
		body         string
		expectedCode int
		msg          string
	}{
		// Username min=4
		{
			`{"user":{"username":"abc","email":"t@t.com","password":"password1"}}`,
			http.StatusUnprocessableEntity,
			"username 3 chars should fail (min=4)",
		},
		{
			`{"user":{"username":"abcd","email":"bound@t.com","password":"password1"}}`,
			http.StatusCreated,
			"username 4 chars should pass (min=4)",
		},
		// Password min=8
		{
			`{"user":{"username":"passtest","email":"p@t.com","password":"1234567"}}`,
			http.StatusUnprocessableEntity,
			"password 7 chars should fail (min=8)",
		},
		{
			`{"user":{"username":"passtest2","email":"p2@t.com","password":"12345678"}}`,
			http.StatusCreated,
			"password 8 chars should pass (min=8)",
		},
		// Email format
		{
			`{"user":{"username":"emailtest","email":"notanemail","password":"password1"}}`,
			http.StatusUnprocessableEntity,
			"invalid email format should fail",
		},
		// Missing required fields
		{
			`{"user":{"email":"m@t.com","password":"password1"}}`,
			http.StatusUnprocessableEntity,
			"missing username should fail",
		},
		{
			`{"user":{"username":"missemail","password":"password1"}}`,
			http.StatusUnprocessableEntity,
			"missing email should fail",
		},
		{
			`{"user":{"username":"misspass","email":"mp@t.com"}}`,
			http.StatusUnprocessableEntity,
			"missing password should fail",
		},
	}

	for _, tc := range tests {
		req, _ := http.NewRequest("POST", "/users/", bytes.NewBufferString(tc.body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		asserts.Equal(tc.expectedCode, w.Code, tc.msg)
	}
}

// --- OpenSpec core requirement: Login error cases (specs/auth.md) ---

func TestLoginValidatorBoundary(t *testing.T) {
	asserts := assert.New(t)
	resetDBWithMock()

	r := gin.New()
	UsersRegister(r.Group("/users"))

	tests := []struct {
		body         string
		expectedCode int
		msg          string
	}{
		// Empty body
		{
			`{}`,
			http.StatusUnprocessableEntity,
			"empty body should fail",
		},
		// Missing email
		{
			`{"user":{"password":"password123"}}`,
			http.StatusUnprocessableEntity,
			"missing email should fail",
		},
		// Missing password
		{
			`{"user":{"email":"user1@linkedin.com"}}`,
			http.StatusUnprocessableEntity,
			"missing password should fail",
		},
		// Invalid email format
		{
			`{"user":{"email":"notanemail","password":"password123"}}`,
			http.StatusUnprocessableEntity,
			"invalid email format should fail",
		},
	}

	for _, tc := range tests {
		req, _ := http.NewRequest("POST", "/users/login", bytes.NewBufferString(tc.body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		asserts.Equal(tc.expectedCode, w.Code, tc.msg)
	}
}

// --- OpenSpec core requirement: NewUserModelValidatorFillWith preserves data (specs/auth.md PUT /api/user) ---

func TestNewUserModelValidatorFillWith(t *testing.T) {
	asserts := assert.New(t)

	image := "https://example.com/photo.jpg"
	user := UserModel{
		Username: "testuser",
		Email:    "test@example.com",
		Bio:      "my bio",
		Image:    &image,
	}

	validator := NewUserModelValidatorFillWith(user)
	asserts.Equal("testuser", validator.User.Username)
	asserts.Equal("test@example.com", validator.User.Email)
	asserts.Equal("my bio", validator.User.Bio)
	asserts.Equal("https://example.com/photo.jpg", validator.User.Image)
	// Password should be set to RandomPassword to bypass re-hashing
	asserts.Equal(common.RandomPassword, validator.User.Password,
		"Password should be RandomPassword to skip re-hashing on update")
}

func TestNewUserModelValidatorFillWithNilImage(t *testing.T) {
	asserts := assert.New(t)

	user := UserModel{
		Username: "testuser",
		Email:    "test@example.com",
		Image:    nil,
	}

	validator := NewUserModelValidatorFillWith(user)
	asserts.Equal("", validator.User.Image,
		"Nil image should result in empty string in validator")
}

// This is a hack way to add test database for each case, as whole test will just share one database.
// You can read TestWithoutAuth's comment to know how to not share database each case.
func TestMain(m *testing.M) {
	test_db = common.TestDBInit()
	AutoMigrate()
	exitVal := m.Run()
	common.TestDBFree(test_db)
	os.Exit(exitVal)
}

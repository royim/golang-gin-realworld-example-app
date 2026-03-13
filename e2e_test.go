package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/gothinkster/golang-gin-realworld-example-app/articles"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/gothinkster/golang-gin-realworld-example-app/users"
)

// ── Shared state across sequential scenarios ──

var (
	e2eRouter *gin.Engine
	e2eDB     *gorm.DB

	// User A (primary actor)
	userAUsername = "alice_e2e"
	userAEmail    = "alice_e2e@example.com"
	userAPassword = "password1234"
	userAToken    string

	// User B (secondary actor for authorization tests)
	userBUsername = "bob_e2e"
	userBEmail    = "bob_e2e@example.com"
	userBPassword = "password5678"
	userBToken    string

	// Article state shared between scenarios
	articleSlug string
	commentID   float64
)

func setupE2ERouter() *gin.Engine {
	r := gin.New()
	r.RedirectTrailingSlash = false

	v1 := r.Group("/api")
	users.UsersRegister(v1.Group("/users"))
	v1.Use(users.AuthMiddleware(false))
	articles.ArticlesAnonymousRegister(v1.Group("/articles"))
	articles.TagsAnonymousRegister(v1.Group("/tags"))
	users.ProfileRetrieveRegister(v1.Group("/profiles"))

	v1.Use(users.AuthMiddleware(true))
	users.UserRegister(v1.Group("/user"))
	users.ProfileRegister(v1.Group("/profiles"))
	articles.ArticlesRegister(v1.Group("/articles"))

	return r
}

// doRequest is a helper that creates and executes an HTTP request, returning the recorder.
func doRequest(method, path, body, token string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req, _ = http.NewRequest(method, path, bytes.NewBufferString(body))
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}
	w := httptest.NewRecorder()
	e2eRouter.ServeHTTP(w, req)
	return w
}

// parseJSON parses the response body into a map.
func parseJSON(w *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if result == nil {
		result = make(map[string]interface{})
	}
	return result
}

// getObj safely extracts a nested object from parsed JSON.
func getObj(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok && v != nil {
		if obj, ok := v.(map[string]interface{}); ok {
			return obj
		}
	}
	return make(map[string]interface{})
}

// getArr safely extracts a nested array from parsed JSON.
func getArr(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key]; ok && v != nil {
		if arr, ok := v.([]interface{}); ok {
			return arr
		}
	}
	return []interface{}{}
}

// getStr safely extracts a string value from a map.
func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getFloat safely extracts a float64 value from a map.
func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok && v != nil {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

// getBool safely extracts a bool value from a map.
func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// ════════════════════════════════════════════════
// Scenario 1: User Registration
// ════════════════════════════════════════════════

func TestE2E_01_Registration(t *testing.T) {
	asserts := assert.New(t)

	// Register User A
	body := fmt.Sprintf(`{"user":{"username":"%s","email":"%s","password":"%s"}}`,
		userAUsername, userAEmail, userAPassword)
	w := doRequest("POST", "/api/users/", body, "")

	asserts.Equal(http.StatusCreated, w.Code, "Registration should return 201")

	resp := parseJSON(w)
	user := getObj(resp, "user")
	asserts.Equal(userAUsername, getStr(user, "username"), "Username should match")
	asserts.Equal(userAEmail, getStr(user, "email"), "Email should match")
	asserts.NotEmpty(getStr(user, "token"), "Token should be returned")
	asserts.Equal("", getStr(user, "bio"), "Bio should default to empty string")

	userAToken = getStr(user, "token")

	// Register User B
	body = fmt.Sprintf(`{"user":{"username":"%s","email":"%s","password":"%s"}}`,
		userBUsername, userBEmail, userBPassword)
	w = doRequest("POST", "/api/users/", body, "")

	asserts.Equal(http.StatusCreated, w.Code, "User B registration should return 201")
	resp = parseJSON(w)
	userBToken = getStr(getObj(resp, "user"), "token")
}

func TestE2E_01b_RegistrationDuplicate(t *testing.T) {
	asserts := assert.New(t)

	// Duplicate username/email should fail
	body := fmt.Sprintf(`{"user":{"username":"%s","email":"%s","password":"%s"}}`,
		userAUsername, userAEmail, userAPassword)
	w := doRequest("POST", "/api/users/", body, "")

	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "Duplicate registration should return 422")
}

// ════════════════════════════════════════════════
// Scenario 2: Login + Token + User Update
// ════════════════════════════════════════════════

func TestE2E_02_LoginAndToken(t *testing.T) {
	asserts := assert.New(t)

	// Login with User A
	body := fmt.Sprintf(`{"user":{"email":"%s","password":"%s"}}`, userAEmail, userAPassword)
	w := doRequest("POST", "/api/users/login", body, "")

	asserts.Equal(http.StatusOK, w.Code, "Login should return 200")
	resp := parseJSON(w)
	user := getObj(resp, "user")
	asserts.Equal(userAUsername, getStr(user, "username"), "Username should match")
	asserts.Equal(userAEmail, getStr(user, "email"), "Email should match")
	asserts.NotEmpty(getStr(user, "token"), "Token should be returned on login")

	// Update the token from login response (fresher)
	userAToken = getStr(user, "token")
}

func TestE2E_02b_LoginInvalidPassword(t *testing.T) {
	asserts := assert.New(t)

	body := fmt.Sprintf(`{"user":{"email":"%s","password":"wrongpassword"}}`, userAEmail)
	w := doRequest("POST", "/api/users/login", body, "")

	asserts.Equal(http.StatusUnauthorized, w.Code, "Invalid password should return 401")
}

func TestE2E_02c_GetCurrentUser(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("GET", "/api/user", "", userAToken)

	asserts.Equal(http.StatusOK, w.Code, "Get current user should return 200")
	resp := parseJSON(w)
	user := getObj(resp, "user")
	asserts.Equal(userAUsername, getStr(user, "username"))
	asserts.Equal(userAEmail, getStr(user, "email"))
}

func TestE2E_02d_GetCurrentUserWithoutAuth(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("GET", "/api/user", "", "")
	asserts.Equal(http.StatusUnauthorized, w.Code, "No auth should return 401")
}

func TestE2E_02e_UpdateUser(t *testing.T) {
	asserts := assert.New(t)

	body := `{"user":{"bio":"Updated bio from e2e test"}}`
	w := doRequest("PUT", "/api/user", body, userAToken)

	asserts.Equal(http.StatusOK, w.Code, "User update should return 200")
	resp := parseJSON(w)
	user := getObj(resp, "user")
	asserts.Equal("Updated bio from e2e test", getStr(user, "bio"), "Bio should be updated")
	asserts.Equal(userAUsername, getStr(user, "username"), "Username should remain")

	// Verify by re-fetching
	w = doRequest("GET", "/api/user", "", userAToken)
	resp = parseJSON(w)
	user = getObj(resp, "user")
	asserts.Equal("Updated bio from e2e test", getStr(user, "bio"), "Bio should persist after re-fetch")
}

// ════════════════════════════════════════════════
// Scenario 3: Profile + Follow/Unfollow
// ════════════════════════════════════════════════

func TestE2E_03_ProfileRetrieve(t *testing.T) {
	asserts := assert.New(t)

	// Anonymous profile retrieval
	w := doRequest("GET", "/api/profiles/"+userBUsername, "", "")

	asserts.Equal(http.StatusOK, w.Code, "Profile retrieval should return 200")
	resp := parseJSON(w)
	profile := getObj(resp, "profile")
	asserts.Equal(userBUsername, getStr(profile, "username"), "Username should match")
	asserts.Equal(false, getBool(profile, "following"), "Anonymous should not follow anyone")
}

func TestE2E_03b_ProfileFollow(t *testing.T) {
	asserts := assert.New(t)

	// User A follows User B
	w := doRequest("POST", "/api/profiles/"+userBUsername+"/follow", "", userAToken)

	asserts.Equal(http.StatusOK, w.Code, "Follow should return 200")
	resp := parseJSON(w)
	profile := getObj(resp, "profile")
	asserts.Equal(userBUsername, getStr(profile, "username"))
	asserts.Equal(true, getBool(profile, "following"), "Following should be true after follow")

	// Verify by re-fetching profile with auth
	w = doRequest("GET", "/api/profiles/"+userBUsername, "", userAToken)
	resp = parseJSON(w)
	profile = getObj(resp, "profile")
	asserts.Equal(true, getBool(profile, "following"), "Following should persist on re-fetch")
}

func TestE2E_03c_ProfileUnfollow(t *testing.T) {
	asserts := assert.New(t)

	// User A unfollows User B
	w := doRequest("DELETE", "/api/profiles/"+userBUsername+"/follow", "", userAToken)

	asserts.Equal(http.StatusOK, w.Code, "Unfollow should return 200")
	resp := parseJSON(w)
	profile := getObj(resp, "profile")
	asserts.Equal(false, getBool(profile, "following"), "Following should be false after unfollow")
}

// ════════════════════════════════════════════════
// Scenario 4: Article CRUD
// ════════════════════════════════════════════════

func TestE2E_04_ArticleCreate(t *testing.T) {
	asserts := assert.New(t)

	body := `{"article":{"title":"E2E Test Article","description":"E2E Description","body":"E2E Body Content","tagList":["e2e","testing"]}}`
	w := doRequest("POST", "/api/articles/", body, userAToken)

	asserts.Equal(http.StatusCreated, w.Code, "Article creation should return 201")

	resp := parseJSON(w)
	article := getObj(resp, "article")
	asserts.Equal("E2E Test Article", getStr(article, "title"))
	asserts.Equal("E2E Description", getStr(article, "description"))
	asserts.Equal("E2E Body Content", getStr(article, "body"))
	asserts.NotEmpty(getStr(article, "slug"), "Slug should be generated")
	asserts.NotEmpty(getStr(article, "createdAt"), "CreatedAt should be present")
	asserts.NotEmpty(getStr(article, "updatedAt"), "UpdatedAt should be present")

	// Verify author info
	author := getObj(article, "author")
	asserts.Equal(userAUsername, getStr(author, "username"), "Author should be user A")

	// Verify tags are sorted
	tagList := getArr(article, "tagList")
	asserts.Equal(2, len(tagList), "Should have 2 tags")
	if len(tagList) == 2 {
		asserts.Equal("e2e", tagList[0], "Tags should be sorted: e2e first")
		asserts.Equal("testing", tagList[1], "Tags should be sorted: testing second")
	}

	// Verify favorites default
	asserts.Equal(false, getBool(article, "favorited"), "New article should not be favorited")
	asserts.Equal(float64(0), getFloat(article, "favoritesCount"), "New article should have 0 favorites")

	articleSlug = getStr(article, "slug")
}

func TestE2E_04b_ArticleRetrieve(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("GET", "/api/articles/"+articleSlug, "", "")

	asserts.Equal(http.StatusOK, w.Code, "Article retrieval should return 200")
	resp := parseJSON(w)
	article := getObj(resp, "article")
	asserts.Equal("E2E Test Article", getStr(article, "title"))
	asserts.Equal(articleSlug, getStr(article, "slug"))
}

func TestE2E_04c_ArticleUpdate(t *testing.T) {
	asserts := assert.New(t)

	// Keep the same title to avoid slug change in DB (ArticleUpdate handler
	// regenerates slug from title via slug.Make, which would break downstream tests)
	body := `{"article":{"title":"E2E Test Article","description":"Updated Description","body":"Updated Body"}}`
	w := doRequest("PUT", "/api/articles/"+articleSlug, body, userAToken)

	asserts.Equal(http.StatusOK, w.Code, "Article update should return 200")
	resp := parseJSON(w)
	article := getObj(resp, "article")
	asserts.Equal("E2E Test Article", getStr(article, "title"))
	asserts.Equal("Updated Description", getStr(article, "description"))
	asserts.Equal("Updated Body", getStr(article, "body"))
}

func TestE2E_04d_ArticleList(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("GET", "/api/articles", "", "")

	asserts.Equal(http.StatusOK, w.Code, "Article list should return 200")
	resp := parseJSON(w)
	articlesList := getArr(resp, "articles")
	asserts.GreaterOrEqual(len(articlesList), 1, "Should have at least 1 article")
	asserts.GreaterOrEqual(getFloat(resp, "articlesCount"), float64(1), "articlesCount should be >= 1")
}

func TestE2E_04e_ArticleListByAuthor(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("GET", "/api/articles?author="+userAUsername, "", "")

	asserts.Equal(http.StatusOK, w.Code)
	resp := parseJSON(w)
	articlesList := getArr(resp, "articles")
	asserts.GreaterOrEqual(len(articlesList), 1, "User A should have at least 1 article")

	// All articles should be authored by user A
	for _, a := range articlesList {
		if art, ok := a.(map[string]interface{}); ok {
			author := getObj(art, "author")
			asserts.Equal(userAUsername, getStr(author, "username"))
		}
	}
}

// ════════════════════════════════════════════════
// Scenario 5: Tags
// ════════════════════════════════════════════════

func TestE2E_05_TagList(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("GET", "/api/tags", "", "")

	asserts.Equal(http.StatusOK, w.Code, "Tag list should return 200")
	resp := parseJSON(w)
	tags := getArr(resp, "tags")
	asserts.GreaterOrEqual(len(tags), 2, "Should have at least 2 tags (e2e, testing)")

	// Verify our tags exist
	tagStrings := make([]string, len(tags))
	for i, tag := range tags {
		if s, ok := tag.(string); ok {
			tagStrings[i] = s
		}
	}
	asserts.Contains(tagStrings, "e2e")
	asserts.Contains(tagStrings, "testing")
}

func TestE2E_05b_ArticleFilterByTag(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("GET", "/api/articles?tag=e2e", "", "")

	asserts.Equal(http.StatusOK, w.Code, "Filter by tag should return 200")
	resp := parseJSON(w)
	articlesList := getArr(resp, "articles")
	asserts.GreaterOrEqual(len(articlesList), 1, "Should find articles with 'e2e' tag")

	// Verify each article has the tag
	for _, a := range articlesList {
		if art, ok := a.(map[string]interface{}); ok {
			tagList := getArr(art, "tagList")
			found := false
			for _, tag := range tagList {
				if s, ok := tag.(string); ok && s == "e2e" {
					found = true
					break
				}
			}
			asserts.True(found, "Each article should have the 'e2e' tag")
		}
	}
}

// ════════════════════════════════════════════════
// Scenario 6: Comment CRUD
// ════════════════════════════════════════════════

func TestE2E_06_CommentCreate(t *testing.T) {
	asserts := assert.New(t)

	body := `{"comment":{"body":"E2E test comment"}}`
	w := doRequest("POST", "/api/articles/"+articleSlug+"/comments", body, userAToken)

	asserts.Equal(http.StatusCreated, w.Code, "Comment creation should return 201")

	resp := parseJSON(w)
	comment := getObj(resp, "comment")
	asserts.Equal("E2E test comment", getStr(comment, "body"))
	asserts.NotEmpty(getStr(comment, "createdAt"))
	asserts.NotEmpty(getStr(comment, "updatedAt"))

	author := getObj(comment, "author")
	asserts.Equal(userAUsername, getStr(author, "username"))

	commentID = getFloat(comment, "id")
}

func TestE2E_06b_CommentList(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("GET", "/api/articles/"+articleSlug+"/comments", "", "")

	asserts.Equal(http.StatusOK, w.Code, "Comment list should return 200")
	resp := parseJSON(w)
	comments := getArr(resp, "comments")
	asserts.GreaterOrEqual(len(comments), 1, "Should have at least 1 comment")

	// Find our comment
	found := false
	for _, c := range comments {
		if cm, ok := c.(map[string]interface{}); ok {
			if getStr(cm, "body") == "E2E test comment" {
				found = true
				break
			}
		}
	}
	asserts.True(found, "Our comment should be in the list")
}

func TestE2E_06c_CommentDelete(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("DELETE", fmt.Sprintf("/api/articles/%s/comments/%d", articleSlug, int(commentID)), "", userAToken)

	asserts.Equal(http.StatusOK, w.Code, "Comment deletion should return 200")

	// Verify comment is gone
	w = doRequest("GET", "/api/articles/"+articleSlug+"/comments", "", "")
	resp := parseJSON(w)
	comments := getArr(resp, "comments")
	for _, c := range comments {
		if cm, ok := c.(map[string]interface{}); ok {
			asserts.NotEqual(commentID, getFloat(cm, "id"), "Deleted comment should not appear in list")
		}
	}
}

// ════════════════════════════════════════════════
// Scenario 7: Favorite + Feed
// ════════════════════════════════════════════════

func TestE2E_07_Favorite(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("POST", "/api/articles/"+articleSlug+"/favorite", "", userBToken)

	asserts.Equal(http.StatusOK, w.Code, "Favorite should return 200")
	resp := parseJSON(w)
	article := getObj(resp, "article")
	asserts.Equal(float64(1), getFloat(article, "favoritesCount"), "Favorites count should be 1")

	// Verify via article retrieve (with auth)
	w = doRequest("GET", "/api/articles/"+articleSlug, "", userBToken)
	resp = parseJSON(w)
	article = getObj(resp, "article")
	asserts.Equal(true, getBool(article, "favorited"), "Should show as favorited for user B")
	asserts.Equal(float64(1), getFloat(article, "favoritesCount"))
}

func TestE2E_07b_ArticleListFilterByFavorited(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("GET", "/api/articles?favorited="+userBUsername, "", "")

	asserts.Equal(http.StatusOK, w.Code)
	resp := parseJSON(w)
	articlesList := getArr(resp, "articles")
	asserts.GreaterOrEqual(len(articlesList), 1, "User B should have at least 1 favorited article")
}

func TestE2E_07c_Feed(t *testing.T) {
	asserts := assert.New(t)

	// User B follows User A first
	doRequest("POST", "/api/profiles/"+userAUsername+"/follow", "", userBToken)

	// User B's feed should contain User A's articles
	w := doRequest("GET", "/api/articles/feed", "", userBToken)

	asserts.Equal(http.StatusOK, w.Code, "Feed should return 200")
	resp := parseJSON(w)
	articlesList := getArr(resp, "articles")
	asserts.GreaterOrEqual(len(articlesList), 1, "Feed should contain user A's articles")
	asserts.GreaterOrEqual(getFloat(resp, "articlesCount"), float64(1))

	// Cleanup: unfollow
	doRequest("DELETE", "/api/profiles/"+userAUsername+"/follow", "", userBToken)
}

func TestE2E_07d_Unfavorite(t *testing.T) {
	asserts := assert.New(t)

	w := doRequest("DELETE", "/api/articles/"+articleSlug+"/favorite", "", userBToken)

	asserts.Equal(http.StatusOK, w.Code, "Unfavorite should return 200")
	resp := parseJSON(w)
	article := getObj(resp, "article")
	asserts.Equal(float64(0), getFloat(article, "favoritesCount"), "Favorites count should be 0 after unfavorite")

	// Verify
	w = doRequest("GET", "/api/articles/"+articleSlug, "", userBToken)
	resp = parseJSON(w)
	article = getObj(resp, "article")
	asserts.Equal(false, getBool(article, "favorited"), "Should not be favorited after unfavorite")
}

// ════════════════════════════════════════════════
// Scenario 8: Authorization — Cross-user operations
// ════════════════════════════════════════════════

func TestE2E_08_UpdateOtherUserArticle(t *testing.T) {
	asserts := assert.New(t)

	// User B tries to update User A's article → 403
	body := `{"article":{"title":"Hijacked Title","description":"Hijacked","body":"Hijacked"}}`
	w := doRequest("PUT", "/api/articles/"+articleSlug, body, userBToken)

	asserts.Equal(http.StatusForbidden, w.Code, "Updating other user's article should return 403")

	// Verify article unchanged
	w = doRequest("GET", "/api/articles/"+articleSlug, "", "")
	resp := parseJSON(w)
	article := getObj(resp, "article")
	asserts.NotEqual("Hijacked Title", getStr(article, "title"), "Title should not be changed by unauthorized user")
}

func TestE2E_08b_DeleteOtherUserArticle(t *testing.T) {
	asserts := assert.New(t)

	// User B tries to delete User A's article → 403
	w := doRequest("DELETE", "/api/articles/"+articleSlug, "", userBToken)

	asserts.Equal(http.StatusForbidden, w.Code, "Deleting other user's article should return 403")

	// Verify article still exists
	w = doRequest("GET", "/api/articles/"+articleSlug, "", "")
	asserts.Equal(http.StatusOK, w.Code, "Article should still exist after unauthorized delete attempt")
}

func TestE2E_08c_DeleteOtherUserComment(t *testing.T) {
	asserts := assert.New(t)

	// User A creates a comment
	body := `{"comment":{"body":"Auth test comment"}}`
	w := doRequest("POST", "/api/articles/"+articleSlug+"/comments", body, userAToken)
	asserts.Equal(http.StatusCreated, w.Code)
	resp := parseJSON(w)
	cmtID := getFloat(getObj(resp, "comment"), "id")

	// User B tries to delete User A's comment → 403
	w = doRequest("DELETE", fmt.Sprintf("/api/articles/%s/comments/%d", articleSlug, int(cmtID)), "", userBToken)
	asserts.Equal(http.StatusForbidden, w.Code, "Deleting other user's comment should return 403")

	// Verify comment still exists
	w = doRequest("GET", "/api/articles/"+articleSlug+"/comments", "", "")
	resp = parseJSON(w)
	comments := getArr(resp, "comments")
	found := false
	for _, c := range comments {
		if cm, ok := c.(map[string]interface{}); ok {
			if getFloat(cm, "id") == cmtID {
				found = true
				break
			}
		}
	}
	asserts.True(found, "Comment should still exist after unauthorized delete")
}

// ════════════════════════════════════════════════
// Scenario 9: Article Delete by Owner (run last)
// ════════════════════════════════════════════════

func TestE2E_09_ArticleDeleteByOwner(t *testing.T) {
	asserts := assert.New(t)

	// Owner (User A) deletes the article
	w := doRequest("DELETE", "/api/articles/"+articleSlug, "", userAToken)
	asserts.Equal(http.StatusOK, w.Code, "Owner delete should return 200")

	// Verify article is gone
	w = doRequest("GET", "/api/articles/"+articleSlug, "", "")
	asserts.Equal(http.StatusNotFound, w.Code, "Deleted article should return 404")
}

// ════════════════════════════════════════════════
// TestMain — DB setup/teardown
// ════════════════════════════════════════════════

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// Remove stale test DB to ensure clean state
	testDBPath := common.GetTestDBPath()
	os.Remove(testDBPath)

	e2eDB = common.TestDBInit()
	Migrate(e2eDB)

	e2eRouter = setupE2ERouter()

	exitVal := m.Run()
	common.TestDBFree(e2eDB)
	os.Exit(exitVal)
}

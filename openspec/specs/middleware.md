# Auth Middleware Spec

## Overview

`users/middlewares.go` provides JWT-based authentication middleware for Gin routes.

## Token Extraction

```go
func extractToken(c *gin.Context) string
```

Extraction priority:
1. `Authorization` header with `Token ` prefix → strip prefix, return token
2. `access_token` query parameter → return value
3. Neither found → return empty string

**Important**: Uses `Token` scheme, NOT `Bearer`.

```
# Valid
Authorization: Token eyJhbGciOiJIUzI1NiIs...

# Also valid
GET /api/user?access_token=eyJhbGciOiJIUzI1NiIs...

# NOT supported
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

## AuthMiddleware

```go
func AuthMiddleware(auto401 bool) gin.HandlerFunc
```

### Mode: auto401 = false (Anonymous Allowed)

Used for public routes that optionally benefit from auth context.

**Flow**:
1. Extract token from request
2. If token exists:
   - Parse JWT with HS256 and `common.JWTSecret`
   - Extract `id` claim (user ID)
   - Call `UpdateContextUserModel(c, id)` to set user in context
3. If no token or invalid token:
   - Set user ID to 0 in context (anonymous)
   - Continue without error
4. Call `c.Next()`

**Applied to**: Article list, article retrieve, profile retrieve, comments list, tags

### Mode: auto401 = true (Authentication Required)

Used for protected routes.

**Flow**:
1. Extract token from request
2. If no token → `c.AbortWithStatusJSON(401, ...)`
3. Parse JWT:
   - Invalid signature → `c.AbortWithStatusJSON(401, ...)`
   - Expired token → `c.AbortWithStatusJSON(401, ...)`
4. Extract `id` claim
5. Call `UpdateContextUserModel(c, id)`
6. Call `c.Next()`

**Applied to**: Current user, user update, follow/unfollow, article CRUD, comments CRUD, favorites

## Context Update

```go
func UpdateContextUserModel(c *gin.Context, my_user_id uint)
```

Sets two values in Gin context:
- `"my_user_id"` (uint) — user ID (0 if anonymous)
- `"my_user_model"` (UserModel) — full user model (empty if anonymous)

**Access in handlers**:
```go
myUserModel := c.MustGet("my_user_model").(users.UserModel)
myUserId := c.MustGet("my_user_id").(uint)
```

## JWT Token Structure

**Generation** (`common/utils.go`):
```go
func GenToken(id uint) string
// Algorithm: HS256
// Secret: common.JWTSecret
// Claims:
//   "id":  uint (user ID)
//   "exp": int64 (Unix timestamp, now + 24 hours)
```

**Validation in middleware**:
```go
token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method")
    }
    return []byte(common.JWTSecret), nil
})
```

## Route Groups in hello.go

```go
// Group 1: Anonymous routes
v1 := r.Group("/api")
v1.Use(AuthMiddleware(false))
// → UsersRegister, ArticlesAnonymousRegister, TagsAnonymousRegister, ProfileRetrieveRegister

// Group 2: Authenticated routes
v1.Use(AuthMiddleware(true))
// → UserRegister, ProfileRegister, ArticlesRegister
```

## Test Helper

```go
// common/test_helpers.go
func HeaderTokenMock(req *http.Request, id uint)
// Sets "Authorization: Token <generated-jwt>" header for test requests
```

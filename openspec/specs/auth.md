# Auth API Spec

## Authentication Scheme

- Header format: `Authorization: Token <jwt>`
- Alternative: `access_token` query parameter
- JWT signed with HS256, claims: `id` (user ID), `exp` (24h expiry)
- Token generation: `common.GenToken(id uint) string`

---

## POST /api/users — Registration

**Auth**: Not required

**Request**:
```json
{
  "user": {
    "username": "string (required, 4-255 chars)",
    "email": "string (required, email format)",
    "password": "string (required, 8-255 chars)",
    "bio": "string (optional, max 1024)",
    "image": "string (optional, URL format)"
  }
}
```

**Response 201**:
```json
{
  "user": {
    "username": "jacob",
    "email": "jake@jake.jake",
    "bio": "",
    "image": null,
    "token": "<jwt>"
  }
}
```

**Errors**:
- 400: Validation failure (missing fields, format errors)
- 422: Duplicate email or username

**Implementation**: `users/routers.go:UsersRegistration`
- Validates via `UserModelValidator.Bind(c)`
- Password hashed with bcrypt via `setPassword()`
- Saves user, generates JWT token

---

## POST /api/users/login — Login

**Auth**: Not required

**Request**:
```json
{
  "user": {
    "email": "string (required, email format)",
    "password": "string (required, 8-255 chars)"
  }
}
```

**Response 200**:
```json
{
  "user": {
    "username": "jacob",
    "email": "jake@jake.jake",
    "bio": "I work at statefarm",
    "image": null,
    "token": "<jwt>"
  }
}
```

**Errors**:
- 400: Validation failure
- 403: Wrong email/password (`"login": "not Found or wrong password"`)

**Implementation**: `users/routers.go:UsersLogin`
- Finds user by email via `FindOneUser()`
- Verifies password via `checkPassword()`

---

## GET /api/user — Current User

**Auth**: Required

**Response 200**:
```json
{
  "user": {
    "username": "jacob",
    "email": "jake@jake.jake",
    "bio": "I work at statefarm",
    "image": "https://example.com/photo.jpg",
    "token": "<jwt>"
  }
}
```

**Errors**:
- 401: No valid token

**Implementation**: `users/routers.go:UserRetrieve`
- Gets user from context via `c.MustGet("my_user_model")`

---

## PUT /api/user — Update User

**Auth**: Required

**Request**:
```json
{
  "user": {
    "username": "string (optional, 4-255)",
    "email": "string (optional, email format)",
    "password": "string (optional, 8-255)",
    "bio": "string (optional, max 1024)",
    "image": "string (optional, URL format)"
  }
}
```

**Response 200**: Same as GET /api/user

**Errors**:
- 400: Validation failure
- 401: No valid token
- 422: Duplicate email/username on update

**Implementation**: `users/routers.go:UserUpdate`
- Prefills validator with current data via `NewUserModelValidatorFillWith()`
- Updates only changed fields via `userModel.Update()`
- Re-hashes password if changed

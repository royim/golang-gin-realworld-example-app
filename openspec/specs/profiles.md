# Profiles API Spec

## GET /api/profiles/:username — Get Profile

**Auth**: Optional (affects `following` field)

**Path Parameters**:
- `username`: target user's username

**Response 200**:
```json
{
  "profile": {
    "username": "jake",
    "bio": "I work at statefarm",
    "image": "https://example.com/photo.jpg",
    "following": false
  }
}
```

**`following` field logic**:
- If authenticated: checks `FollowModel` for current user → target user relationship
- If anonymous: always `false`

**Errors**:
- 404: `{"errors": {"profile": "Invalid username"}}`

**Implementation**: `users/routers.go:ProfileRetrieve`
- Finds user by username via `FindOneUser(&UserModel{Username: username})`
- Uses `ProfileSerializer` with context user for `following` status

---

## POST /api/profiles/:username/follow — Follow User

**Auth**: Required

**Path Parameters**:
- `username`: user to follow

**Response 200**:
```json
{
  "profile": {
    "username": "jake",
    "bio": "...",
    "image": "...",
    "following": true
  }
}
```

**Errors**:
- 401: No valid token
- 404: `{"errors": {"profile": "Invalid username"}}`
- 422: Follow failed (DB error)

**Implementation**: `users/routers.go:ProfileFollow`
- Gets current user from context via `c.MustGet("my_user_model")`
- Finds target user by username
- Creates follow relationship via `currentUser.following(targetUser)`
- Returns profile with `following: true`

**Model operation** (`users/models.go`):
```go
func (u UserModel) following(v UserModel) error
// Creates FollowModel{FollowingID: v.ID, FollowedByID: u.ID}
```

---

## DELETE /api/profiles/:username/follow — Unfollow User

**Auth**: Required

**Path Parameters**:
- `username`: user to unfollow

**Response 200**:
```json
{
  "profile": {
    "username": "jake",
    "bio": "...",
    "image": "...",
    "following": false
  }
}
```

**Errors**:
- 401: No valid token
- 404: `{"errors": {"profile": "Invalid username"}}`
- 422: Unfollow failed (DB error)

**Implementation**: `users/routers.go:ProfileUnfollow`
- Deletes follow relationship via `currentUser.unFollowing(targetUser)`
- Returns profile with `following: false`

**Model operation** (`users/models.go`):
```go
func (u UserModel) unFollowing(v UserModel) error
// Hard-deletes FollowModel where FollowingID=v.ID AND FollowedByID=u.ID
// Uses Unscoped().Delete() for permanent removal
```

---

## Data Flow

```
ProfileRetrieve:
  FindOneUser(username) → ProfileSerializer{C, UserModel}.Response()

ProfileFollow:
  c.MustGet("my_user_model") → FindOneUser(username) → following() → ProfileSerializer

ProfileUnfollow:
  c.MustGet("my_user_model") → FindOneUser(username) → unFollowing() → ProfileSerializer
```

## Follow Model (self-referential many-to-many)

```
UserModel ←──── FollowModel ────→ UserModel
              FollowedByID          FollowingID
              (the follower)        (being followed)
```

- `GetFollowings()` returns users that this user follows
- `isFollowing(v)` checks if this user follows v

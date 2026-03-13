# Comments API Spec

## GET /api/articles/:slug/comments — List Comments

**Auth**: Optional (affects `author.following` field)

**Path Parameters**:
- `slug`: article slug

**Response 200**:
```json
{
  "comments": [
    {
      "id": 1,
      "body": "It takes a Jacobian",
      "createdAt": "2016-02-18T03:22:56.637Z",
      "updatedAt": "2016-02-18T03:22:56.637Z",
      "author": {
        "username": "jake",
        "bio": "...",
        "image": "...",
        "following": false
      }
    }
  ]
}
```

**Errors**:
- 404: Invalid slug — `{"errors": {"articles": "Invalid slug"}}`

**Implementation**: `articles/routers.go:ArticleCommentList`
- Finds article by slug via `FindOneArticle()`
- Loads comments via `articleModel.getComments()`
- Serializes via `CommentsSerializer.Response()`

**Comment loading** (`articles/models.go`):
```go
func (model *ArticleModel) getComments() error
// Preloads Comments with Author.UserModel
// Loads into model.Comments slice
```

---

## POST /api/articles/:slug/comments — Create Comment

**Auth**: Required

**Path Parameters**:
- `slug`: article slug

**Request**:
```json
{
  "comment": {
    "body": "string (required, max 2048 chars)"
  }
}
```

**Response 201**:
```json
{
  "comment": {
    "id": 1,
    "body": "His name was my name too.",
    "createdAt": "2016-02-18T03:22:56.637Z",
    "updatedAt": "2016-02-18T03:22:56.637Z",
    "author": {
      "username": "jake",
      "bio": "...",
      "image": "...",
      "following": false
    }
  }
}
```

**Errors**:
- 400: Validation failure (empty body, exceeds 2048 chars)
- 401: No valid token
- 404: Invalid slug
- 422: DB save error

**Implementation**: `articles/routers.go:ArticleCommentCreate`
- Finds article by slug
- Binds via `CommentModelValidator.Bind(c)`
- Sets ArticleID and AuthorID (via `GetArticleUserModel`)
- Saves via `SaveOne()`

---

## DELETE /api/articles/:slug/comments/:id — Delete Comment

**Auth**: Required (comment author only)

**Path Parameters**:
- `slug`: article slug
- `id`: comment ID

**Response 200**:
```json
{
  "comment": "delete success"
}
```

**Errors**:
- 401: No valid token
- 403: `{"errors": {"comment": "you are not the author"}}`
- 404: Invalid slug or comment ID

**Authorization check**:
```go
myArticleUserModel := GetArticleUserModel(myUserModel)
if commentModel.AuthorID != myArticleUserModel.ID {
    c.JSON(403, ...)
}
```

**Implementation**: `articles/routers.go:ArticleCommentDelete`
- Finds article by slug, then finds comment by ID within article
- Checks author ownership
- Deletes via `DeleteCommentModel()`

---

## CommentModel Schema

```go
type CommentModel struct {
    gorm.Model          // ID, CreatedAt, UpdatedAt, DeletedAt
    Article   ArticleModel
    ArticleID uint      // belongs to article
    Author    ArticleUserModel
    AuthorID  uint      // belongs to article user
    Body      string    // max 2048 chars
}
```

**Relationships**:
- CommentModel → ArticleModel (belongs to, via ArticleID)
- CommentModel → ArticleUserModel (belongs to, via AuthorID)
- ArticleModel → []CommentModel (has many, ForeignKey: ArticleID)

## Serializer

```go
type CommentSerializer struct {
    C *gin.Context
    CommentModel
}
// Response() returns CommentResponse with author profile
// Timestamps formatted as time.RFC3339Nano
```

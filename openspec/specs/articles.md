# Articles API Spec

## GET /api/articles — List Articles

**Auth**: Optional (affects `favorited` and `author.following` fields)

**Query Parameters**:

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `tag` | string | — | Filter by tag |
| `author` | string | — | Filter by author username |
| `favorited` | string | — | Filter by user who favorited |
| `limit` | int | 20 | Number of articles |
| `offset` | int | 0 | Pagination offset |

**Response 200**:
```json
{
  "articles": [
    {
      "slug": "how-to-train-your-dragon",
      "title": "How to train your dragon",
      "description": "Ever wonder how?",
      "body": "You have to believe",
      "tagList": ["dragons", "training"],
      "createdAt": "2016-02-18T03:22:56.637Z",
      "updatedAt": "2016-02-18T03:22:56.637Z",
      "favorited": false,
      "favoritesCount": 0,
      "author": {
        "username": "jake",
        "bio": "...",
        "image": "...",
        "following": false
      }
    }
  ],
  "articlesCount": 1
}
```

**Implementation**: `articles/routers.go:ArticleList`
- Uses `FindManyArticle(tag, author, limit, offset, favorited)`
- N+1 optimized via `ArticlesSerializer.Response()` with batch queries

---

## GET /api/articles/feed — Feed Articles

**Auth**: Required

**Query Parameters**: `limit` (default 20), `offset` (default 0)

**Response 200**: Same structure as List Articles (only articles from followed users)

**Implementation**: `articles/routers.go:ArticleFeed`
- Gets `ArticleUserModel` via `GetArticleUserModel(myUserModel)`
- Calls `articleUserModel.GetArticleFeed(limit, offset)`
- Returns articles from users the current user follows

---

## GET /api/articles/:slug — Get Article

**Auth**: Optional

**Response 200**:
```json
{
  "article": { ...single article object... }
}
```

**Errors**: 404 — `{"errors": {"articles": "Invalid slug"}}`

**Implementation**: `articles/routers.go:ArticleRetrieve`
- Uses `FindOneArticle(&ArticleModel{Slug: slug})` with Preload

---

## POST /api/articles — Create Article

**Auth**: Required

**Request**:
```json
{
  "article": {
    "title": "string (required, min 4 chars)",
    "description": "string (required, max 2048)",
    "body": "string (required, max 2048)",
    "tagList": ["tag1", "tag2"]
  }
}
```

**Response 201**: Single article object

**Errors**: 400 (validation), 401 (no token), 422 (DB error)

**Implementation**: `articles/routers.go:ArticleCreate`
- `ArticleModelValidator.Bind(c)` generates slug from title, sets author, creates tags
- Slug format: `<title-slugified>-<random-string>`
- Tags created via `setTags()` with race condition handling (FirstOrCreate)

---

## PUT /api/articles/:slug — Update Article

**Auth**: Required (author only)

**Request**:
```json
{
  "article": {
    "title": "string (optional)",
    "description": "string (optional)",
    "body": "string (optional)",
    "tagList": ["tag1"]
  }
}
```

**Response 200**: Updated article object

**Errors**:
- 401: No token
- 403: `{"errors": {"article": "you are not the author"}}`
- 404: Invalid slug

**Authorization check**:
```go
if articleModel.AuthorID != GetArticleUserModel(myUserModel).ID {
    c.JSON(403, ...)
}
```

---

## DELETE /api/articles/:slug — Delete Article

**Auth**: Required (author only)

**Response 200**: `{"article": "delete success"}`

**Errors**: 401, 403 (not author), 404 (invalid slug)

**Implementation**: `articles/routers.go:ArticleDelete`
- Same authorization check as update
- Uses `DeleteArticleModel(&ArticleModel{Slug: slug})`

---

## N+1 Optimization

List/Feed endpoints use batch operations:
- `BatchGetFavoriteCounts(articleIDs)` → `map[uint]uint`
- `BatchGetFavoriteStatus(articleIDs, userID)` → `map[uint]bool`
- `Preload("Author.UserModel").Preload("Tags")` for eager loading

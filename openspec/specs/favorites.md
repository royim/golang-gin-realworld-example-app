# Favorites API Spec

## POST /api/articles/:slug/favorite — Favorite Article

**Auth**: Required

**Path Parameters**:
- `slug`: article slug

**Request**: No body required

**Response 200**:
```json
{
  "article": {
    "slug": "how-to-train-your-dragon",
    "title": "How to train your dragon",
    "description": "...",
    "body": "...",
    "tagList": ["dragons"],
    "createdAt": "2016-02-18T03:22:56.637Z",
    "updatedAt": "2016-02-18T03:22:56.637Z",
    "favorited": true,
    "favoritesCount": 1,
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
- 401: No valid token
- 404: Invalid slug — `{"errors": {"articles": "Invalid slug"}}`
- 422: Favorite failed (DB error)

**Implementation**: `articles/routers.go:ArticleFavorite`
- Finds article by slug via `FindOneArticle()`
- Gets `ArticleUserModel` via `GetArticleUserModel(myUserModel)`
- Creates favorite via `articleModel.favoriteBy(articleUserModel)`
- Returns article with updated `favorited: true` and `favoritesCount`

---

## DELETE /api/articles/:slug/favorite — Unfavorite Article

**Auth**: Required

**Path Parameters**:
- `slug`: article slug

**Request**: No body required

**Response 200**:
```json
{
  "article": {
    "slug": "how-to-train-your-dragon",
    "...": "...",
    "favorited": false,
    "favoritesCount": 0
  }
}
```

**Errors**:
- 401: No valid token
- 404: Invalid slug
- 422: Unfavorite failed (DB error)

**Implementation**: `articles/routers.go:ArticleUnfavorite`
- Finds article by slug
- Removes favorite via `articleModel.unFavoriteBy(articleUserModel)`
- Returns article with updated status

---

## FavoriteModel Schema

```go
type FavoriteModel struct {
    gorm.Model
    Favorite     ArticleModel
    FavoriteID   uint             // the article being favorited
    FavoriteBy   ArticleUserModel
    FavoriteByID uint             // the user who favorited
}
```

**Relationships**:
```
ArticleModel ←── FavoriteModel ──→ ArticleUserModel
              FavoriteID            FavoriteByID
```

## Model Operations

```go
// Add favorite
func (model ArticleModel) favoriteBy(user ArticleUserModel) error
// Creates FavoriteModel{FavoriteID: model.ID, FavoriteByID: user.ID}

// Remove favorite
func (model ArticleModel) unFavoriteBy(user ArticleUserModel) error
// Hard-deletes FavoriteModel (Unscoped)

// Check if favorited
func (model ArticleModel) isFavoriteBy(user ArticleUserModel) bool
// Counts FavoriteModel where FavoriteID=model.ID AND FavoriteByID=user.ID

// Count favorites
func (model ArticleModel) favoritesCount() uint
// Counts all FavoriteModel where FavoriteID=model.ID
```

## Batch Operations (N+1 Optimization)

Used by list/feed endpoints to avoid per-article queries:

```go
// Get favorite counts for multiple articles at once
func BatchGetFavoriteCounts(articleIDs []uint) map[uint]uint
// Single SQL: SELECT favorite_id, COUNT(*) FROM favorite_models
//             WHERE favorite_id IN (?) GROUP BY favorite_id

// Get favorite status for multiple articles at once
func BatchGetFavoriteStatus(articleIDs []uint, userID uint) map[uint]bool
// Single SQL: SELECT favorite_id FROM favorite_models
//             WHERE favorite_id IN (?) AND favorite_by_id = ?
```

These batch functions are called in `ArticlesSerializer.Response()` to optimize
list rendering with O(2) queries instead of O(N).

# Tags API Spec

## GET /api/tags — List Tags

**Auth**: Not required

**Request**: No parameters

**Response 200**:
```json
{
  "tags": ["reactjs", "angularjs", "dragons"]
}
```

**Implementation**: `articles/routers.go:TagList`
- Queries all `TagModel` records from database
- Serializes via `TagsSerializer.Response()`
- Returns flat array of tag strings

---

## TagModel Schema

```go
type TagModel struct {
    gorm.Model          // ID, CreatedAt, UpdatedAt, DeletedAt
    Tag           string `gorm:"uniqueIndex"` // unique tag name
    ArticleModels []ArticleModel `gorm:"many2many:article_tags"`
}
```

**Relationships**:
```
ArticleModel ←──many2many──→ TagModel
              (article_tags join table)
```

## Tag Creation

Tags are created when articles are created/updated, not via a dedicated endpoint.

```go
// articles/models.go
func (model *ArticleModel) setTags(tags []string) error
```

**Race condition handling**:
- Uses `FirstOrCreate` to find or create each tag
- If `FirstOrCreate` creates a new tag but unique constraint fails,
  falls back to `Find` to get the existing tag
- This handles concurrent article creation with the same new tag

**Flow**:
1. Article creation: `ArticleModelValidator.Bind()` calls `setTags()`
2. For each tag string, `FirstOrCreate` finds/creates `TagModel`
3. Tags are associated with the article via GORM `many2many` join table

## Serializers

```go
type TagSerializer struct {
    C *gin.Context
    TagModel
}
// Response() string — returns the tag string

type TagsSerializer struct {
    C    *gin.Context
    Tags []TagModel
}
// Response() []string — returns array of tag strings
```

## Tag Filtering

Tags are used as a filter parameter in article list:

```
GET /api/articles?tag=dragons
```

**Implementation** in `FindManyArticle()`:
- Joins `article_tags` table when `tag` parameter is provided
- Filters articles that have the specified tag

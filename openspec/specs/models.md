# Data Models & Relationships

## Entity Relationship Diagram

```
┌──────────┐     FollowModel      ┌──────────┐
│ UserModel │◄──── (self M:N) ────►│ UserModel │
└─────┬─────┘   FollowedByID       └──────────┘
      │          FollowingID
      │
      │ 1:1 (FirstOrCreate)
      ▼
┌─────────────────┐
│ ArticleUserModel │──── 1:N ────► ArticleModel
└────────┬────────┘                  │    │
         │                           │    │
         │ 1:N                       │    │ M:N (article_tags)
         ▼                           │    ▼
    FavoriteModel ◄── N:1 ──────────┘  TagModel
                                      │
                                      │ 1:N (ForeignKey: ArticleID)
                                      ▼
                                 CommentModel ◄── N:1 ── ArticleUserModel
```

## UserModel (users/models.go)

```go
type UserModel struct {
    ID           uint
    Username     string  `gorm:"column:username;unique"`
    Email        string  `gorm:"column:email;uniqueIndex"`
    Bio          string  `gorm:"column:bio;size:1024"`
    Image        *string `gorm:"column:image"`
    PasswordHash string  `gorm:"column:password;not null"`
}
```

| Method | Description |
|--------|-------------|
| `setPassword(password)` | bcrypt hash and store |
| `checkPassword(password)` | bcrypt compare |
| `Update(data)` | GORM Updates |
| `following(user)` | Create FollowModel |
| `unFollowing(user)` | Hard-delete FollowModel |
| `isFollowing(user)` | Check follow exists |
| `GetFollowings()` | List followed users |

## FollowModel (users/models.go)

```go
type FollowModel struct {
    gorm.Model
    Following    UserModel
    FollowingID  uint  // user being followed
    FollowedBy   UserModel
    FollowedByID uint  // user who follows
}
```

Self-referential many-to-many via bridge table.

## ArticleUserModel (articles/models.go)

Bridge between `users.UserModel` and the articles domain:

```go
type ArticleUserModel struct {
    gorm.Model
    UserModel      users.UserModel
    UserModelID    uint
    ArticleModels  []ArticleModel  `gorm:"ForeignKey:AuthorID"`
    FavoriteModels []FavoriteModel `gorm:"ForeignKey:FavoriteByID"`
}
```

Created via `GetArticleUserModel(userModel)` using `FirstOrCreate`.

## ArticleModel (articles/models.go)

```go
type ArticleModel struct {
    gorm.Model
    Slug        string           `gorm:"uniqueIndex"`
    Title       string
    Description string           `gorm:"size:2048"`
    Body        string           `gorm:"size:2048"`
    Author      ArticleUserModel
    AuthorID    uint
    Tags        []TagModel       `gorm:"many2many:article_tags"`
    Comments    []CommentModel   `gorm:"ForeignKey:ArticleID"`
}
```

| Method | Description |
|--------|-------------|
| `setTags(tags)` | FirstOrCreate tags, handles race condition |
| `Update(data)` | GORM Updates |
| `favoriteBy(user)` | Create FavoriteModel |
| `unFavoriteBy(user)` | Hard-delete FavoriteModel |
| `isFavoriteBy(user)` | Check favorite exists |
| `favoritesCount()` | Count favorites |
| `getComments()` | Preload comments |

## TagModel (articles/models.go)

```go
type TagModel struct {
    gorm.Model
    Tag           string         `gorm:"uniqueIndex"`
    ArticleModels []ArticleModel `gorm:"many2many:article_tags"`
}
```

## FavoriteModel (articles/models.go)

```go
type FavoriteModel struct {
    gorm.Model
    Favorite     ArticleModel
    FavoriteID   uint
    FavoriteBy   ArticleUserModel
    FavoriteByID uint
}
```

## CommentModel (articles/models.go)

```go
type CommentModel struct {
    gorm.Model
    Article   ArticleModel
    ArticleID uint
    Author    ArticleUserModel
    AuthorID  uint
    Body      string `gorm:"size:2048"`
}
```

## Key DB Functions

| Function | Description |
|----------|-------------|
| `FindOneUser(condition)` | Find user by condition |
| `FindOneArticle(condition)` | Find article with Preloads |
| `FindOneComment(condition)` | Find comment with Preloads |
| `FindManyArticle(tag,author,limit,offset,favorited)` | Filtered list |
| `GetArticleUserModel(userModel)` | FirstOrCreate bridge |
| `BatchGetFavoriteCounts(ids)` | Batch favorite counts |
| `BatchGetFavoriteStatus(ids, uid)` | Batch favorite status |
| `SaveOne(data)` | Generic save |
| `DeleteArticleModel(condition)` | Delete article |
| `DeleteCommentModel(condition)` | Delete comment |

# Repository Layer

This folder contains all GORM-backed repository implementations.
Repositories are the **only layer allowed to touch the database directly**.
Usecases call repos. Handlers never call repos.

---

## How it works

`base_repo.go` exports a generic `BaseRepository[T]` struct.
Embed it in your concrete repository and you get ~18 standard methods for free.
The concrete repo only needs to define methods that are genuinely domain-specific.

```
base_repo.go       ← generic read/write methods for any entity
user_repo.go       ← embeds BaseRepository[entity.User] + user-specific queries
admin_repo.go      ← embeds BaseRepository[entity.Admin] + admin-specific queries
wallet_repo.go     ← embeds BaseRepository[entity.Wallet] + wallet-specific queries
```

---

## Defining a concrete repository

```go
type UserRepository struct {
    repo.BaseRepository[entity.User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{
        BaseRepository: repo.BaseRepository[entity.User]{DB: db},
    }
}
```

That's it. The repo now has all base methods automatically.
Add domain-specific methods below as plain methods on `*UserRepository`.

---

## Base method reference

### Writes

| Method | Description |
|--------|-------------|
| `Create(ctx, *T)` | Insert a new record |
| `CreateTx(tx, *T)` | Insert inside an existing transaction |
| `UpdateFields(ctx, id, map)` | Partial update by ID |
| `UpdateFieldsTx(tx, id, map)` | Partial update inside a transaction |
| `Delete(ctx, id)` | Delete by ID (soft-delete if `DeletedAt` field exists) |
| `DeleteTx(tx, id)` | Delete inside a transaction |

### Reads — AND conditions (all conditions must match)

| Method | Description |
|--------|-------------|
| `Get(ctx, conditions, fields...)` | First match |
| `GetAll(ctx, conditions, fields...)` | All matches, no pagination |
| `GetAllPaged(ctx, conditions, page, limit, order, fields...)` | Paginated |
| `GetWithPreloads(ctx, conditions, preloads, fields...)` | First match with associations |
| `GetAllWithPreloads(ctx, conditions, preloads, fields...)` | All matches with associations |
| `GetAllPagedWithPreloads(ctx, conditions, page, limit, order, preloads, fields...)` | Paginated with associations |
| `Exists(ctx, conditions)` | True if any row matches |

### Reads — OR conditions (any condition can match)

| Method | Description |
|--------|-------------|
| `GetAny(ctx, conditions, fields...)` | First match of any condition |
| `GetAllAny(ctx, conditions, fields...)` | All matches of any condition |
| `GetAllAnyPaged(ctx, conditions, page, limit, order, fields...)` | Paginated |
| `ExistsAny(ctx, conditions)` | True if any row matches any condition |

---

## Field selection (`fields ...string`)

All read methods accept an optional trailing `fields ...string`.
When supplied, GORM emits `SELECT field1, field2, ...` instead of `SELECT *`.
Omit it to fetch all columns — existing call sites need no changes.

**When to use it:**
- Auth: only fetch the columns needed to verify a password or check status.
- List endpoints: only fetch the columns rendered in the table row.
- Exists-like checks: fetch only `id` when you just need to confirm a record exists.

**When NOT to use it:**
- When you need the full object for further processing.
- When you will pass the result to a function that reads other fields.
- On preloaded associations — select applies to the root entity only.

---

## Examples

### 1. Fetch a user by ID (all columns)

```go
user, err := r.Get(ctx, map[string]any{"id": userID})
if errors.Is(err, gorm.ErrRecordNotFound) {
    return apperror.NewNotFoundErr("User not found")
}
```

---

### 2. Fetch only the columns needed for login

```go
user, err := r.Get(ctx,
    map[string]any{"email": email},
    "id", "email", "password", "status",
)
```

GORM emits: `SELECT id, email, password, status FROM users WHERE email = ? LIMIT 1`

---

### 3. Check a unique constraint before creating

```go
exists, err := r.Exists(ctx, map[string]any{"email": email})
if exists {
    return apperror.NewAlreadyExistsErr("Email already registered")
}
```

---

### 4. Paginated list with column projection

```go
// Handler parses: ?page=1&limit=20
users, total, err := r.GetAllPaged(ctx,
    map[string]any{"status": "active"},
    page, limit, "created_at DESC",
    "id", "name", "email", "status", "created_at",
)
```

---

### 5. OR lookup (find user by email OR phone)

```go
user, err := r.GetAny(ctx, map[string]any{
    "email": identifier,
    "phone": identifier,
})
```

GORM emits: `SELECT * FROM users WHERE email = ? OR phone = ? LIMIT 1`

---

### 6. Fetch with a preloaded association

```go
user, err := r.GetWithPreloads(ctx,
    map[string]any{"id": userID},
    []repo.Preload{
        {Entity: "Wallet"},
        {Entity: "Notifications", Limit: 10, Order: "created_at DESC"},
    },
)
```

The `fields` parameter applies to the root `users` table only.
Preloaded associations always fetch all their own columns.

---

### 7. Update specific fields

```go
err := r.UpdateFields(ctx, userID, map[string]any{
    "status":     "suspended",
    "updated_at": time.Now(),
})
```

Only the listed columns are written to the database.

---

### 8. Delete (soft-delete if entity has `DeletedAt`)

```go
err := r.Delete(ctx, userID)
if errors.Is(err, gorm.ErrRecordNotFound) {
    return apperror.NewNotFoundErr("User not found")
}
```

If the entity embeds `gorm.Model` or has a `DeletedAt *time.Time` field,
GORM soft-deletes automatically. Otherwise it performs a hard delete.

---

### 9. Create inside a DB transaction

```go
err := db.Transaction(func(tx *gorm.DB) error {
    if err := userRepo.CreateTx(tx, user); err != nil {
        return err
    }
    if err := walletRepo.CreateTx(tx, wallet); err != nil {
        return err
    }
    return nil
})
```

Always use the `Tx` variants inside `db.Transaction(...)` callbacks.
Never mix `ctx`-based methods with an explicit `tx` — only the `Tx` variant
actually uses the transaction connection.

---

### 10. Domain-specific method on a concrete repo

When the base methods cannot express the query (e.g. a JOIN, a subquery,
a raw SQL expression), add a plain method to the concrete repo:

```go
// user_repo.go

// FindByEmailOrPhone does a case-insensitive lookup on both columns.
// Used at login — email and phone share the same login form field.
func (r *UserRepository) FindByEmailOrPhone(ctx context.Context, identifier string) (*entity.User, error) {
    identifier = strings.ToLower(strings.TrimSpace(identifier))
    var user entity.User
    err := r.DB.WithContext(ctx).
        Where("LOWER(email) = ? OR LOWER(phone) = ?", identifier, identifier).
        First(&user).Error
    return &user, err
}
```

Concrete methods live on the concrete repo type, not on `BaseRepository`.
Keep them in the same file as the repo struct they belong to.

---

## Pagination convention

`GetAllPaged` and `GetAllAnyPaged` normalize inputs before querying:

| Input | Clamped to |
|-------|-----------|
| `page < 1` | `1` |
| `limit < 1` | `50` |
| `limit > 100` | `100` |
| `order == ""` | `"created_at DESC"` |

Handlers should pass `page` and `limit` directly from query params after
parsing — no need to guard against zero or negative values before calling.

---

## Error handling

Base methods return raw GORM errors. Usecases translate these into
`*apperror.AppError` before returning to the handler.

```go
// In the usecase — not in the repo:
user, err := r.UserRepo.Get(ctx, map[string]any{"id": id})
if errors.Is(err, gorm.ErrRecordNotFound) {
    return apperror.NewNotFoundErr("User not found")
}
if err != nil {
    return apperror.NewDatabaseErr("Failed to fetch user", err.Error())
}
```

Repos never import `apperror`. That dependency only flows one way: upward.

# CLAUDE.md — GoFiber Boilerplate

Standing instructions for working on any project derived from this boilerplate.

---

## Module name

This template ships with module name `boilerplate`. When starting a new project, rename it:

```bash
find . -type f -name "*.go" -exec sed -i 's|"boilerplate/|"yourproject/|g' {} \;
sed -i 's/^module boilerplate/module yourproject/' go.mod
```

---

## Architecture layers

```
entity   →  repo   →  usecase  →  handler
          (DB)     (business)   (HTTP)
```

- **Handlers** are thin. No business logic. Parse → call usecase → map response.
- **Usecases** hold all business logic. They call repos and service infra only.
- **Repos** are the only layer that touches GORM / the database.
- **Service** is the dependency hub. Inject into usecases via their constructors.

---

## Ground rules

- Run `go build ./...` after every batch of backend changes. Never stack broken changes.
- All JSON keys are camelCase. No exceptions. Query params too (`?userId=`, not `?user_id=`).
- All monetary values use `entity.Money` (int64 kobo). Never floats.
- IDs use ULID. No auto-increment integers.
- After refactoring a layer, compile before moving to the next.

---

## Typed-nil rule

Never return `*apperror.AppError` directly as `error`.

```go
// WRONG
func Foo() error {
    return someHelper()  // someHelper returns *apperror.AppError
}

// CORRECT
func Foo() error {
    if err := someHelper(); err != nil {
        return err
    }
    return nil
}
```

Constructors (`apperror.NewXxx(...)`) are always non-nil and safe to return directly.

---

## Error code → HTTP status mapping

```
ErrCodeValidation        → 400
ErrCodeInvalidCredentials → 400
ErrCodeRecordNotFound    → 404
ErrCodeForbidden         → 403
ErrCodeAlreadyExists     → 409
ErrCodeUserConflict      → 409
ErrCodeDatabase          → 500
ErrCodeInsufficientFunds → 402
default                  → 500
```

---

## Repo layer convention

Every concrete repo embeds `BaseRepository[T]` from `base_repo.go`.
The base provides ~16 standard read/write methods.
Add only genuinely domain-specific methods on the concrete repo.

---

## Service layer

`service.go` is the single place that assembles all infrastructure.
Usecases never import infrastructure packages directly — they go through `*service.Service`.

`service_cron.go` contains all periodic background jobs.
Register new ones in `StartCronJobs()` only — never scatter `go func()` calls in `main.go`.

---

## Cron jobs (default)

- Cache entry expiry cleaner — every 5 minutes
- Refresh token DB cleanup — every 24 hours

---

## Protected patterns

- Never break existing endpoint signatures without an explicit instruction.
- Prefer small, targeted changes. Compile after each layer.
- Do not add error handling for scenarios that cannot happen.
- No comments unless the WHY is non-obvious.

# DEV.md — Developer Guide

Technical conventions and rules for working on this codebase.
Read this before writing any code. These are not suggestions.

---

## Architecture overview

```
Handler  →  Usecase  →  Repo  →  DB
              ↓
           Service (dependency hub)
              ↓
         Infra packages (DB, JWT, Mailer, Cache, Telegram, Paystack, ...)
```

**The dependency rule is strictly one-directional:**

| Layer | May import | May NOT import |
|-------|-----------|----------------|
| Handler | Usecase, Response, AppError, Service | Repo, Infra directly |
| Usecase | Repo, Service, AppError, Entity | Handler, Response |
| Repo | Entity, GORM | Usecase, Handler, Service, AppError |
| Service | All Infra packages | Usecase, Handler |
| Infra packages | Standard library, third-party | Any internal layer |

If you feel an infra package needs to call back into the application, use a
function-pointer hook registered from the service layer instead (the same
pattern used by `initGlobalHooks` in `service.go`).

---

## Compile check

Run after every batch of changes. Never stack broken changes.

```bash
go build ./...
# or
make build
```

---

## Naming conventions

- All JSON keys: **camelCase**. No exceptions. This includes request bodies,
  response bodies, and query parameters (`?userId=`, not `?user_id=`).
- IDs: **ULID** (`ulid.Make().String()`). No auto-increment integers.
- Money: **`entity.Money` (int64 — smallest currency unit)**. Never `float64`.
- DB column names: snake_case (GORM default from struct field names).

---

## Error handling — AppError

`AppError` is the structured error type for the entire application.
See `internal/http/apperror/README.md` for the full explanation.

**Short version:**

Usecases return errors through the `error` interface.
When the error is application-level (not a raw Go error), use a typed constructor:

```go
// In a usecase
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, apperror.NewNotFoundErr("user id=" + id)
}
if err != nil {
    return nil, apperror.NewDatabaseErr(err.Error())
}
```

Handlers detect the `*AppError` and delegate to `response.MapError`:

```go
// In a handler
if err != nil {
    var appErr *apperror.AppError
    if errors.As(err, &appErr) {
        return response.MapError(c, appErr)
    }
    return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
}
```

**Typed-nil rule — critical:**
Never declare a function that returns `*apperror.AppError` as its concrete type.
Always return the `error` interface. Returning a nil `*AppError` from a function
typed `error` produces a non-nil interface — `err != nil` is `true` even though
the pointer is nil. This causes phantom errors and silent failures.

```go
// WRONG
func (uc *FooUsecase) doThing() error {
    return someHelper()  // someHelper() returns *apperror.AppError
}

// CORRECT
func (uc *FooUsecase) doThing() error {
    if err := someHelper(); err != nil {
        return err
    }
    return nil  // untyped nil — safe
}
```

Constructors (`apperror.NewXxx(...)`) always return non-nil and are safe to
return directly: `return apperror.NewValidationErr("...")` ✅

---

## Service layer

`internal/service/service.go` is the **dependency hub**.

- It holds every infrastructure instance (DB, Logger, Mailer, Cache, Telegram, Paystack, Queue, ...).
- Usecases receive a `*service.Service` via their constructor and access infra through it.
- Usecases never import infra packages directly.

```go
// Constructor pattern for every usecase
type UserUsecase struct {
    Service  *service.Service
    UserRepo *repo.UserRepository
}

func NewUserUsecase(svc *service.Service, userRepo *repo.UserRepository) *UserUsecase {
    return &UserUsecase{Service: svc, UserRepo: userRepo}
}

// Accessing infra from a usecase — always go through Service
func (uc *UserUsecase) SendWelcome(ctx context.Context, user *entity.User) error {
    uc.Service.Queue.Push(queue.NewMailJob(&mail.Email{...}))
    return nil
}
```

**Why**: this avoids scattering import paths across the domain layer, makes
testing easier (one mock struct instead of many), and means adding a new infra
dependency is a one-line change in `service.go` with no ripple.

---

## Cron jobs

All periodic background jobs live in `internal/service/service_cron.go`.

```
StartCronJobs()          ← single entry point, called once from main.go
  StartCacheCleaner()    ← evicts expired in-memory cache entries every 5 min
  StartRefreshTokenCleaner()  ← deletes expired DB refresh tokens every 24 h
```

**Rules:**
- Register every new cron job inside `StartCronJobs()` only.
- Do not scatter `go func()` background calls anywhere in `main.go`.
- Every cron goroutine must listen on `s.Context.Done()` for clean shutdown.
- All timing constants go at the top of `service_cron.go` as named `const` values.

```go
// Correct pattern for a new cron job
func (s *Service) StartMyJob() {
    go func() {
        ticker := time.NewTicker(myJobInterval)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                s.runMyJob()
            case <-s.Context.Done():
                s.Logger.Info("My job stopped")
                return
            }
        }
    }()
    s.Logger.Info("My job started", zap.Duration("interval", myJobInterval))
}
```

---

## Middleware

Middleware lives in `internal/http/middleware/`.

| File | Purpose |
|------|---------|
| `auth.go` | JWT validation, populates `c.Locals` with identity fields |
| `permissions.go` | Actor-type gate (`ActorTypeRequired`) and superadmin gate |
| `extras.go` | Input sanitization, misc request middleware |

**What JWT auth middleware does:**

1. Reads `Authorization: Bearer <token>` from the request header.
2. Validates the JWT signature and expiry.
3. Extracts claims and writes them to `c.Locals`:
   - `actor_id` — the authenticated user/admin ID
   - `actor_type` — `"user"` or `"admin"`
   - `is_superadmin` — bool
   - `status` — actor's current status

**Reading identity in a handler:**

```go
// handler/helpers.go provides typed helpers — always use these:
id   := actorID(c)        // c.Locals("actor_id").(string)
kind := actorType(c)      // c.Locals("actor_type").(string)
su   := isSuperAdmin(c)   // c.Locals("is_superadmin").(bool)
```

**Route-level protection:**

```go
// Require a specific actor type on a group
userGroup := api.Group("/user", middleware.ActorTypeRequired("user"))
adminGroup := api.Group("/admin", middleware.ActorTypeRequired("admin"))

// Superadmin-only group
superGroup := api.Group("/superadmin", middleware.SuperAdminOnly())
```

Public routes (register, login, refresh-token) go outside any protected group.

---

## Repository layer

Concrete repos embed `BaseRepository[T]` from `internal/domain/repo/base_repo.go`.
The base provides ~18 standard methods. Only add domain-specific methods to the
concrete repo. See `internal/domain/repo/README.md` for the full guide.

```go
type UserRepository struct {
    repo.BaseRepository[entity.User]
}
```

**Error handling in repos:** return raw GORM errors. Never import `apperror`.
Error translation happens in the usecase layer.

---

## Background job queue

The durable queue (`internal/infra/queue/`) processes jobs in background goroutines.
Jobs are defined as typed structs in `queue/jobs.go`. Push from usecases, dispatch
in `service.processJob`.

```go
// Push a mail job from a usecase
uc.Service.Queue.Push(queue.NewMailJob(&mail.Email{
    To:       []string{user.Email},
    Subject:  "Welcome",
    Template: "welcome.html",
    Data:     map[string]any{"Name": user.Name},
}))
```

Add new job types by:
1. Adding a type constant in `queue/interface.go`
2. Adding a job data struct in `queue/jobs.go`
3. Adding a `case` in `service.processJob`

---

## Response envelope

All API responses follow the same shape:

```json
{
  "message": "Success",
  "success": true,
  "data": { ... },
  "error": {}
}
```

On error:

```json
{
  "message": "We couldn't find what you're looking for",
  "success": false,
  "data": {},
  "error": {
    "code": "RECORD_NOT_FOUND",
    "message": "We couldn't find what you're looking for",
    "details": "user id=01JX..."
  }
}
```

Handlers always use the helpers in `internal/http/response/` — never hand-build
JSON maps in handlers.

---

## Renaming the module

This project ships as `module boilerplate`. To rename it for a real project:

```bash
find . -type f -name "*.go" -exec sed -i 's|"boilerplate/|"yourproject/|g' {} \;
sed -i 's/^module boilerplate/module yourproject/' go.mod
go mod tidy
```

---

## Adding a new feature (checklist)

1. Add or update entity in `internal/domain/entity/`
2. Add migration in `internal/infra/db/migrations.go` → `AutoMigrateAll`
3. Add repo in `internal/domain/repo/` (embed `BaseRepository[T]`)
4. Add usecase in `internal/domain/usecase/`
5. Add handler in `internal/http/handler/`
6. Add response mapper in `internal/http/response/`
7. Register route in `main.go` → `registerRoutes`
8. `go build ./...` — confirm zero errors before continuing
9. Add an integration test in `tests/integration/`

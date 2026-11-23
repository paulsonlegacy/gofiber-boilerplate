# AppError Package

Structured error type for the application layer.
Usecases return `*AppError` through the `error` interface.
Handlers detect it, switch on its `Code`, and produce the right HTTP response.

---

## Why not use `fmt.Errorf` or `errors.New`?

Standard errors carry a single string. That is not enough for an API:

- The **user** needs a friendly, safe message: *"We couldn't find that booking."*
- The **developer** needs the raw cause for logs: *"gorm: record not found (id=01JX...)"*
- The **handler** needs a machine-readable code to pick the right HTTP status: `RECORD_NOT_FOUND → 404`

`AppError` carries all three fields in one value.

---

## The struct

```go
type AppError struct {
    Code    string  // machine-readable, e.g. RECORD_NOT_FOUND
    Message string  // user-facing, safe to send in the JSON response
    Details string  // developer-facing, logged but never sent to clients in production
}
```

`AppError` satisfies the `error` interface via its `Error() string` method,
which returns `Details`. This is the critical detail explained below.

---

## Why `error` is an interface — and why that matters

In Go, `error` is not a concrete type. It is an interface:

```go
type error interface {
    Error() string
}
```

Any type that has an `Error() string` method satisfies it. `*AppError` does, so
you can return a `*AppError` wherever an `error` return type is declared.

This is what lets handlers use `errors.As` to pull the `*AppError` back out:

```go
var appErr *apperror.AppError
if errors.As(err, &appErr) {
    return response.MapError(c, appErr)
}
```

---

## The typed-nil trap — read this carefully

Go interfaces hold two internal slots: **TYPE** and **VALUE**.
An interface is `nil` only when **both slots are empty**.

When a function whose return type is `*AppError` returns `nil`, Go sets the
VALUE slot to nil but the TYPE slot to `*AppError` — the interface is **not nil**.

```go
// WRONG — if someHelper() returns a nil *AppError,
// the returned error interface is non-nil, and err != nil is TRUE.
func doSomething() error {
    return someHelper()  // someHelper() returns *apperror.AppError
}

// CORRECT — explicitly return nil so both interface slots are empty.
func doSomething() error {
    if err := someHelper(); err != nil {
        return err
    }
    return nil  // untyped nil — both slots empty — err == nil is TRUE
}
```

**Rule:** Never declare a function that returns `*apperror.AppError` as its type.
All functions must return `error` (the interface).

The constructors (`apperror.New`, `apperror.NewXxx(...)`) always return a non-nil
pointer and are safe to return directly:

```go
return apperror.NewNotFoundErr("User not found")  // ✅ always non-nil
```

---

## Constructors

| Constructor | Code | Use for |
|-------------|------|---------|
| `NewValidationErr(message)` | `VALIDATION_ERROR` | Bad input, failed rule, invalid state |
| `NewNotFoundErr(details)` | `RECORD_NOT_FOUND` | Missing record |
| `NewNotFoundErrWithMessage(msg, details)` | `RECORD_NOT_FOUND` | Missing with custom message |
| `NewAlreadyExistsErr(message)` | `ALREADY_EXISTS` | Duplicate email, phone, etc. |
| `NewDatabaseErr(details)` | `DATABASE_ERROR` | Raw DB errors |
| `NewUnauthorizedErr(details)` | `UNAUTHORIZED` | Missing/invalid token |
| `NewForbiddenErr(details)` | `FORBIDDEN` | Authenticated but not permitted |
| `NewInsufficientFundsErr(message)` | `INSUFFICIENT_FUNDS` | Wallet balance too low |
| `NewServerErr(details)` | `INTERNAL_SERVER_ERROR` | Unexpected failures |
| `New(code, message, details)` | custom | When no typed constructor fits |

**Convention on `message` vs `details`:**
- `message` — safe to show users. Keep it generic. Never put raw DB text here.
- `details` — developer context. The raw `err.Error()` string goes here.

```go
// Good
return apperror.NewDatabaseErr(err.Error())
//                             ^^^^^^^^^^^ raw error goes in Details, not Message

// Bad — raw DB text exposed to the client
return apperror.New("DATABASE_ERROR", err.Error(), "")
```

---

## Usage in usecases

```go
func (uc *UserUsecase) GetProfile(ctx context.Context, userID string) (*entity.User, error) {
    user, err := uc.UserRepo.Get(ctx, map[string]any{"id": userID})
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperror.NewNotFoundErr("user id=" + userID)
    }
    if err != nil {
        return nil, apperror.NewDatabaseErr(err.Error())
    }
    return user, nil
}
```

---

## Usage in handlers

Handlers never build their own error messages. They receive an `error`,
check if it is an `*AppError`, and pass it to `response.MapError`:

```go
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
    user, err := h.UserUC.GetProfile(c.Context(), actorID(c))
    if err != nil {
        var appErr *apperror.AppError
        if errors.As(err, &appErr) {
            return response.MapError(c, appErr)  // picks the right HTTP status
        }
        return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
    }
    return response.Success(c, "Profile fetched", response.ToUserResponse(user))
}
```

`response.MapError` switches on `appErr.Code` and calls the matching response
helper. See `internal/http/response/base_response.go` for the full mapping.

---

## Code → HTTP status mapping

| Code | HTTP |
|------|------|
| `RECORD_NOT_FOUND` | 404 |
| `VALIDATION_ERROR`, `INVALID_CREDENTIALS`, `MISSING_FIELD`, `INVALID_FORMAT` | 400 |
| `UNAUTHORIZED`, `TOKEN_EXPIRED`, `TOKEN_INVALID`, `ACCOUNT_DELETED` | 401 |
| `FORBIDDEN` | 403 |
| `ALREADY_EXISTS`, `USER_CONFLICT`, `DUPLICATE_KEY` | 409 |
| `INSUFFICIENT_FUNDS` | 402 |
| everything else | 500 |

# GoFiber Boilerplate

A production-ready Go/Fiber/GORM starter built on clean architecture.
Ships with JWT auth, role-based access, mailer, cache, Telegram alerts, Paystack payments,
a durable job queue, and a full DDD layer structure.

---

## Stack

| Concern | Tech |
|---------|------|
| Language | Go |
| Framework | GoFiber v2 |
| ORM | GORM |
| DB (dev) | SQLite (zero config) |
| DB (prod) | PostgreSQL |
| Auth | JWT + Argon2id passwords |
| IDs | ULID |
| Logger | Zap + rotating log files |
| Cache | In-memory (Valkey/Redis-ready interface) |
| Mail | SMTP with multi-provider fallback |
| Payments | Paystack |
| Alerts | Telegram Bot API |
| Validation | go-playground/validator v10 |
| Config | .env + typed struct |

---

## Quick start

```bash
# 1. Copy .env and fill in your values
cp .env.example .env

# 2. Run (SQLite auto-created when DB_URL is empty)
make run

# 3. Build a production binary
make build

# 4. Run tests
make test
```

---

## Actors

The boilerplate ships with two actors:

| Actor | Description |
|-------|-------------|
| `user` | End-user account with wallet |
| `admin` | Platform admin — superadmin flag for elevated access |

Add more actor types by following the same pattern across `entity/`, `repo/`, `usecase/`, and `handler/`.

---

## Project structure

```
boilerplate/
├── main.go                       ← server bootstrap and route registration
├── templates/                    ← HTML email templates
├── internal/
│   ├── config/                   ← typed .env config
│   ├── domain/
│   │   ├── entity/               ← GORM models
│   │   ├── filters/              ← query filter structs
│   │   ├── repo/                 ← database access layer
│   │   └── usecase/              ← business logic
│   ├── http/
│   │   ├── apperror/             ← structured error type
│   │   ├── handler/              ← HTTP handlers
│   │   ├── middleware/           ← JWT auth, permissions, sanitization
│   │   └── response/             ← JSON response helpers
│   ├── infra/
│   │   ├── cache/                ← in-memory TTL cache
│   │   ├── db/                   ← GORM init + auto-migrate
│   │   ├── jwt/                  ← token generation + password hashing
│   │   ├── logger/               ← Zap logger
│   │   ├── mail/                 ← SMTP mailer
│   │   ├── paystack/             ← Paystack API client
│   │   ├── queue/                ← durable background job queue
│   │   ├── storage/              ← Cloudinary file storage
│   │   └── telegram/             ← Telegram alert bot
│   └── service/
│       ├── service.go            ← dependency hub
│       ├── service_cron.go       ← all background/cron jobs
│       └── wallet_service.go     ← atomic wallet operations
└── tests/
    ├── unit/
    └── integration/
```

---

## Rename module

Replace `boilerplate` with your project name before writing any code:

```bash
find . -type f -name "*.go" -exec sed -i 's|"boilerplate/|"yourproject/|g' {} \;
sed -i 's/^module boilerplate/module yourproject/' go.mod
go mod tidy
```

---

## Developer guide

See **[DEV.md](DEV.md)** for architecture rules, error handling conventions,
service layer usage, cron job patterns, middleware, and the new-feature checklist.

# Starterpack — Architectural Change Plan

Derived from the **Uber Go Style Guide** cross-referenced with the current
codebase state. The guide's essence is captured in
`.agents/skills/starterpack/references/good-practices.md`. Changes are grouped by priority so they can be tackled
in logical waves without breaking the running app between each step.

---

## Wave 1 — Critical correctness & safety

### 1. Eliminate `init()` from `internal/platform/validator`

**File**: `apps/api/internal/platform/validator/validator.go`

**Problem**: The package initialises a global `*validator.Validate` and a
compiled `*regexp.Regexp` in an `init()` function. The Uber guide explicitly
prohibits `init()` that accesses global state or performs side-effects, because
it is order-dependent, hard to test, and impossible to mock.

**Change**: Replace the two package-level globals and their `init()` with a
`Validator` struct and a `New()` constructor:

```go
type Validator struct {
    v *validator.Validate
}

func New() *Validator {
    v := validator.New()
    _ = v.RegisterValidation("username", validateUsername)
    return &Validator{v: v}
}

func (vl *Validator) ValidateAndMap(entity string, s any) error { ... }
```

Instantiate once in `cmd/api/main.go` and pass to every service that needs it.

---

### 2. `main.go` — Extract `run()` function (Exit Once pattern)

**File**: `apps/api/cmd/api/main.go`

**Problem**: `main()` contains multiple `log.Fatal()` calls scattered across
the function body. The Uber guide mandates that `os.Exit` / `log.Fatal` appear
**at most once**, in `main()`, by extracting a `run() error` function.

Additionally, the HTTP server goroutine is fire-and-forget: there is no
`WaitGroup` or `done` channel to confirm it has exited before the process
terminates. The Uber guide requires every goroutine to have a predictable stop
mechanism.

**Change**:

```go
func main() {
    if err := run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func run() error {
    _ = godotenv.Load(".env.local", ".env")
    cfg := config.Load()
    log := logger.New(cfg.LogLevel, cfg.LogPretty)
    // ... setup ...

    serverErr := make(chan error, 1)
    go func() { serverErr <- srv.ListenAndServe() }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    select {
    case err := <-serverErr:
        if !errors.Is(err, http.ErrServerClosed) {
            return fmt.Errorf("server: %w", err)
        }
    case <-quit:
    }
    // graceful shutdown ...
    return nil
}
```

---

### 3. `domain/errors.go` — Fix zero-value enum trap

**File**: `apps/api/internal/domain/errors.go`

**Problem**: `KindNotFound ErrorKind = iota` gives `KindNotFound` the value 0,
which is the Go zero-value for `int`. A freshly allocated `Error{}` struct would
silently be treated as "not found" rather than as uninitialized. The Uber guide
requires enums to start at 1.

**Change**:

```go
const (
    KindUnknown      ErrorKind = iota // 0 — zero value, uninitialized
    KindNotFound                      // 1 — 404
    KindAlreadyExists                 // 2 — 409
    KindValidation                    // 3 — 422
)
```

Update `mapDomainError` in `response.go` to handle `KindUnknown` as a 500.

---

## Wave 2 — Interface hygiene & missing wiring

### 4. Add compile-time interface compliance checks

**Files**: All `*_repository.go` files and HTTP handler files.

**Problem**: None of the port implementations carry a compile-time assertion.
If a method is renamed or dropped from the `Repository` interface, the error
surfaces only at link time — often far from the breaking change.

**Change**: Add at the top of every adapter file:

```go
// Compile-time interface check.
var _ userdomain.Repository = (*UserRepository)(nil)
```

Apply to:
- `adapters/persistence/postgres/user_repository.go`
- `adapters/persistence/memory/user_repository.go`
- `adapters/persistence/postgres/todo_repository.go` (once created)
- `adapters/persistence/memory/todo_repository.go` (once created)

---

### 5. Define service interfaces at the HTTP adapter boundary

**Problem**: `UserHandler` depends on `*userapp.Service` (a concrete type). The
Uber guide and hexagonal architecture both require the adapters to depend on
abstractions, not implementations. This also makes the handler trivially testable
with a mock.

**Change**: In `internal/application/user/service.go` (or a new `port.go`):

```go
// UserService is the inbound port consumed by the HTTP adapter.
type UserService interface {
    Create(ctx context.Context, in CreateInput) (*userdomain.User, error)
    Get(ctx context.Context, id string) (*userdomain.User, error)
    List(ctx context.Context, limit, offset int32) ([]*userdomain.User, error)
}
```

Update `UserHandler` to take `UserService` instead of `*Service`. Repeat for
every domain service.

---

### 6. Wire the Todo HTTP adapter (TodoHandler is missing)

**Problem**: `internal/domain/todo` and `internal/application/todo/service.go`
both exist, but there is no `internal/adapters/http/todo_handler.go` or
`todo_dto.go`. The todo domain is invisible from the API.

**Change**: Create the following (following the user pattern exactly):
- `internal/adapters/http/todo_handler.go` — CRUD handlers + swag annotations
- `internal/adapters/http/todo_dto.go` — DTOs with only `json:` tags
- Register the todo repository + service + handler in `cmd/api/main.go`
- Add the persistence adapters:
  - `adapters/persistence/postgres/todo_repository.go`
  - `adapters/persistence/memory/todo_repository.go`
- Add DB migration + sqlc query for todos (already partially present in the
  sqlc-generated files).

---

## Wave 3 — Style & performance

### 7. Enforce 3-group import ordering across all Go files

**Problem**: Several files mix stdlib, external, and internal imports without
blank-line separation. The Uber guide (and `goimports`) prescribes three groups:
1. Standard library
2. External packages
3. Internal packages (`github.com/starterpack/api/...`)

**Change**: Run `goimports -w ./...` from `apps/api/`, and add a golangci-lint
rule (`goimports` linter) to enforce this on CI.

---

### 8. Pre-allocate slices and maps with known capacity

**Files**: `user_handler.go` (list endpoint), memory repositories.

**Problem**: `make([]userResponse, 0)` without a capacity hint causes
reallocation as the slice grows. The Uber guide requires capacity hints wherever
the size is known.

**Change**:

```go
// Before
out := make([]userResponse, 0)

// After
out := make([]userResponse, 0, len(users))
```

Apply the same fix in any `for range` loop that appends to a slice.

---

### 9. Avoid embedding types in public structs

**Problem**: If any struct ever embeds another exported type (e.g. embedding
`gin.Context` or a logger), that leaks implementation details into the public
API. Currently not present, but must be enforced as a code-review rule going
forward.

**Rule**: Never embed exported types in exported structs. Compose via a private
field and delegate explicitly.

---

### 10. Use `time.Time` and `time.Duration` everywhere (no raw ints for time)

**Problem**: Any API request/response field representing a timestamp must use
`time.Time` (marshalled as RFC 3339 by `encoding/json`). Any duration must use
`time.Duration`, not an `int` labelled `"seconds"`.

**Current state**: Already correct in `userResponse` (`CreatedAt time.Time`).
Must be enforced for all future DTOs and config fields (e.g. `ReadTimeout`
should stay as `time.Duration`, not `int`).

---

### 11. Add `.golangci.yml` for the Go API

**File**: `apps/api/.golangci.yml`

**Change**: Add a golangci-lint config enabling the linters recommended by the
Uber guide and Go best practices:

```yaml
linters:
  enable:
    - goimports        # import grouping
    - govet            # suspicious constructs
    - errcheck         # unchecked errors
    - staticcheck      # comprehensive static analysis
    - revive           # style (replaces golint)
    - gosimple         # simplifications
    - ineffassign      # dead assignments
    - unused           # unused code
    - misspell         # spelling
    - prealloc         # slice pre-allocation hints
    - bodyclose        # response body close
    - exhaustive       # switch exhaustiveness on enums
    - noctx            # http requests without context
    - godot            # comment formatting
    - gocritic         # bug patterns
```

Add `make lint-go` target to Makefile that runs `golangci-lint run ./...` inside
`apps/api/`.

---

## Wave 4 — Testing

### 12. Add table-driven tests

**Problem**: There are no Go tests in the codebase. The Uber guide mandates
table-driven tests for any function with more than one input case.

**Change**: Add the following test files:

- `internal/application/user/service_test.go` — table-driven tests for
  `Create`, `Get`, `List` using an in-memory repository stub.
- `internal/adapters/http/user_handler_test.go` — HTTP tests using
  `httptest.NewRecorder()` and a mock `UserService` interface.
- `internal/domain/errors_test.go` — tests for `errors.Is()` / `errors.As()`
  chains on the structured `Error` type.

Use `testify/assert` (already idiomatic in Go) or the stdlib `testing` package.

Example pattern:

```go
func TestServiceCreate(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateInput
        wantErr bool
    }{
        {"valid user", CreateInput{Username: "alice", Email: "a@b.com"}, false},
        {"short username", CreateInput{Username: "x", Email: "a@b.com"}, true},
        {"invalid email", CreateInput{Username: "alice", Email: "notanemail"}, true},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            svc := NewService(memory.NewUserRepository())
            _, err := svc.Create(context.Background(), tc.input)
            if (err != nil) != tc.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
            }
        })
    }
}
```

---

## Wave 5 — Future-proofing & observability

### 13. Add request-ID middleware

**Problem**: The zerolog request logger does not attach a request ID. This makes
tracing a single request through logs impossible in production.

**Change**: Add `middleware/requestid.go` that generates a UUID per request,
stores it in context, and adds it to the zerolog logger stored in context. All
subsequent log lines from that request automatically carry `"request_id"`.

---

### 14. Use `go.uber.org/atomic` for any concurrent shared state

**Problem**: Not currently present, but as concurrency is added (background
workers, metrics counters), raw `sync/atomic` operations on primitive types are
error-prone. The Uber guide requires `go.uber.org/atomic` typed wrappers.

**Rule**: When adding shared mutable state accessed from multiple goroutines, use
`atomic.Bool`, `atomic.Int64`, etc. from `go.uber.org/atomic`. Do not use
`sync/atomic` directly on raw primitives.

---

### 15. Functional Options for complex constructors

**Problem**: `NewService(repo)` is fine today, but as services grow they will
need optional dependencies (cache, event bus, metrics). Without the functional
options pattern, every new optional dep breaks every call site.

**Change**: Introduce the pattern for services that are likely to grow:

```go
type ServiceOption func(*Service)

func WithCache(c Cache) ServiceOption {
    return func(s *Service) { s.cache = c }
}

func NewService(repo Repository, opts ...ServiceOption) *Service {
    s := &Service{repo: repo}
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

---

### 16. Observability: Prometheus metrics endpoint

**Problem**: The API has no metrics exposition. For a production SaaS codebase
this is a gap.

**Change**: Add `GET /metrics` behind the Prometheus client using an
`ObservabilityConfig` feature toggle (requires `METRICS_ENABLED=true`). Expose:
- HTTP request count and latency histograms (via Gin middleware)
- Go runtime metrics (via the default Prometheus collector)

---

## Summary table

| # | Wave | File(s) | Uber Guide Rule | Effort |
|---|------|---------|-----------------|--------|
| 1 | 1 | `platform/validator/validator.go` | Avoid `init()` | M |
| 2 | 1 | `cmd/api/main.go` | Exit Once; goroutine lifecycle | M |
| 3 | 1 | `domain/errors.go` | Start Enums at One | S |
| 4 | 2 | `*_repository.go` | Verify Interface Compliance | S |
| 5 | 2 | `application/*/service.go` | Interface over concrete type | M |
| 6 | 2 | `adapters/http/todo_*.go` | (missing wiring) | L |
| 7 | 3 | all `.go` files | Import Group Ordering | S |
| 8 | 3 | `user_handler.go`, memory repos | Container Capacity | S |
| 9 | 3 | code-review rule | Avoid Embedding in Public Structs | XS |
| 10 | 3 | DTOs, config | Use `time.Time`/`time.Duration` | XS |
| 11 | 3 | `apps/api/.golangci.yml` | Linting | S |
| 12 | 4 | `*_test.go` | Table-driven Tests | L |
| 13 | 5 | `middleware/requestid.go` | (observability) | M |
| 14 | 5 | future code | Use `go.uber.org/atomic` | XS |
| 15 | 5 | complex services | Functional Options | M |
| 16 | 5 | `adapters/http/metrics.go` | (observability) | L |

Effort key: XS < S < M < L (relative to this repo's size).

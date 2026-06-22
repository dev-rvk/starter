# Good Practices — Go Backend (Uber Style Guide)

This reference distils the **Uber Go Style Guide** into the conventions that
every agent or engineer must apply when working on the `apps/api` backend.

---

## 1. Errors

### 1.1 Structured domain errors (`domain.Error`)

The project uses a typed `*domain.Error` with a `Kind` (NotFound / AlreadyExists
/ Validation) and wraps sentinel errors so both `errors.Is()` and `errors.As()`
work:

```go
// Constructors in internal/domain/errors.go
domain.NotFound("user")        // Kind=KindNotFound, wraps ErrNotFound
domain.AlreadyExists("user")   // Kind=KindAlreadyExists
domain.ValidationError("user", "Email", "invalid format")
```

`response.go` (`mapDomainError`) maps these to HTTP status codes via
`errors.As(err, &domErr)`. **Never** add per-domain error imports to
`response.go`; the shared sentinels + structured type are sufficient.

### 1.2 Error wrapping — use `%w`, not `%v` ([Error Wrapping])

Wrap errors with `fmt.Errorf("context: %w", err)` when the caller might need to
match the cause. Use `%v` only to intentionally hide it.

```go
// Good
return fmt.Errorf("user repo: create: %w", err)

// Bad — caller cannot errors.Is against the cause
return fmt.Errorf("user repo: create: %v", err)
```

Keep context prefixes terse: **no "failed to"** prefix — just the noun + verb.

### 1.3 Handle errors once ([Handle Errors Once])

Do **not** log an error AND return it. Either:
- Wrap and return (let the boundary log it), or
- Log and degrade gracefully (for non-critical operations).

```go
// Bad
log.Error().Err(err).Msg("get user")
return err

// Good — wrap and return
return fmt.Errorf("get user %q: %w", id, err)
```

### 1.4 Error naming ([Error Naming])

- Exported sentinel vars: `ErrNotFound`, `ErrAlreadyExists`, `ErrValidation`
- Unexported sentinels: `errSomething` (lower-case `err` prefix)
- Custom error types: `NotFoundError`, `ValidationError` (suffix `Error`)

### 1.5 Enum starts at 1, not 0 ([Start Enums at One])

The `ErrorKind` const block must have an explicit `KindUnknown = iota` zero
value so a freshly allocated struct is not accidentally treated as a "not found"
error:

```go
const (
    KindUnknown      ErrorKind = iota // 0 — zero value / uninitialized
    KindNotFound                      // 1 — 404
    KindAlreadyExists                 // 2 — 409
    KindValidation                    // 3 — 422
)
```

---

## 2. Interfaces

### 2.1 Never take a pointer to an interface ([Pointers to Interfaces])

Pass interfaces as values. The concrete type behind the interface can still be a
pointer:

```go
// Bad
func NewService(repo *userdomain.Repository) ...

// Good
func NewService(repo userdomain.Repository) ...
```

### 2.2 Compile-time interface compliance ([Verify Interface Compliance])

Every port implementation must carry a compile-time assertion:

```go
var _ userdomain.Repository = (*UserRepository)(nil)
```

Place this line immediately after the `type` declaration in the adapter file.

### 2.3 Handler dependencies use interfaces, not concrete types

HTTP handlers must depend on the application-layer **interface**, not on
`*userapp.Service`:

```go
// In internal/application/user/service.go (or port.go)
type UserService interface {
    Create(ctx context.Context, in CreateInput) (*userdomain.User, error)
    Get(ctx context.Context, id string) (*userdomain.User, error)
    List(ctx context.Context, limit, offset int32) ([]*userdomain.User, error)
}

// In internal/adapters/http/user_handler.go
type UserHandler struct {
    svc userapp.UserService  // interface, not *userapp.Service
}
```

---

## 3. Avoiding `init()` ([Avoid init()])

**Do not use `init()`** in new code. Move setup into explicit constructors:

```go
// Bad — package global initialised in init()
var validate *validator.Validate
func init() { validate = validator.New() }

// Good — what internal/platform/validator/validator.go actually does:
// a struct with a New() constructor, injected from cmd/api/main.go
type Validator struct { v *validator.Validate }
func New() *Validator {
    v := validator.New()
    _ = v.RegisterValidation("username", validateUsername)
    return &Validator{v: v}
}
```

Instantiate the validator once in `cmd/api/main.go` and inject it into every
service that needs it.

The one accepted use of `init()` in this codebase is the Swagger docs blank
import (`_ "github.com/starterpack/api/docs"`) — this is an idiomatic registry
pattern and is exempt.

---

## 4. Exit in Main — Extract `run()` ([Exit in Main], [Exit Once])

`os.Exit` / `log.Fatal` must appear **at most once** and only in `main()`.
Extract all startup logic into `run() error`:

```go
func main() {
    if err := run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func run() error {
    // ... all startup, server start, graceful shutdown ...
    return nil
}
```

This keeps `main()` as the single exit point and makes `run()` independently
testable.

---

## 5. Goroutine lifecycle ([Don't fire-and-forget goroutines])

Every goroutine must have a predictable stop mechanism. For the HTTP server:

```go
// Good — server errors are surfaced; caller can select on both
serverErr := make(chan error, 1)
go func() { serverErr <- srv.ListenAndServe() }()

select {
case err := <-serverErr:
    if !errors.Is(err, http.ErrServerClosed) {
        return fmt.Errorf("server: %w", err)
    }
case <-quit:
    // Graceful shutdown path
}
```

Never spawn goroutines in `init()`.

---

## 6. Avoid mutable globals ([Avoid Mutable Globals])

Use dependency injection instead of package-level variables. Global validators,
loggers, and DB pools must be passed as constructor arguments, not accessed as
package globals. The only exception is immutable package-level values (compiled
regexes, constant lookup tables).

---

## 7. Performance

### 7.1 Pre-allocate slices and maps ([Prefer Specifying Container Capacity])

When the output size is known before the loop, provide the capacity hint:

```go
// Bad
out := make([]userResponse, 0)
for _, u := range users { out = append(out, toUserResponse(u)) }

// Good
out := make([]userResponse, 0, len(users))
for _, u := range users { out = append(out, toUserResponse(u)) }
```

### 7.2 `strconv` over `fmt` for primitive conversion ([Prefer strconv over fmt])

```go
// Bad
s := fmt.Sprintf("%d", n)

// Good
s := strconv.Itoa(n)
```

`strconv` is already used in `user_handler.go` for `parseInt32`. Keep it
consistent everywhere.

---

## 8. Style

### 8.1 Import groups ([Import Group Ordering])

Three groups, each separated by a blank line:

```go
import (
    // 1. Standard library
    "context"
    "net/http"

    // 2. External packages
    "github.com/gin-gonic/gin"
    "github.com/rs/zerolog"

    // 3. Internal packages
    "github.com/starterpack/api/internal/domain"
    "github.com/starterpack/api/internal/application/user"
)
```

Run `goimports -w ./...` or enforce with the `goimports` golangci-lint linter.

### 8.2 Reduce nesting ([Reduce Nesting], [Unnecessary Else])

Handle errors and return early; the happy path should be left-aligned.

```go
// Bad
if err == nil {
    // ... many lines
} else {
    return err
}

// Good
if err != nil {
    return err
}
// ... happy path
```

### 8.3 Struct initialisation — always use field names ([Use Field Names to Initialize Structs])

```go
// Bad
return &User{"id", "alice", "a@b.com", time.Now(), time.Now()}

// Good
return &User{
    ID:        "id",
    Username:  "alice",
    Email:     "a@b.com",
    CreatedAt: now,
    UpdatedAt: now,
}
```

### 8.4 Avoid embedding in public structs ([Avoid Embedding Types in Public Structs])

Never embed an exported type in an exported struct. Compose via a private field
and delegate explicitly.

### 8.5 Use field tags in marshalled structs ([Use field tags in marshaled structs])

Every JSON-serialised struct field must have an explicit `json:` tag. Relying on
Go's default capitalisation rules makes refactoring break the contract silently.

### 8.6 Do not shadow built-in names ([Avoid Using Built-In Names])

Never use `error`, `string`, `len`, `new`, etc. as variable or field names.

---

## 9. Testing

### 9.1 Table-driven tests ([Test Tables])

Every function with more than one meaningful input case must have a table-driven
test:

```go
tests := []struct {
    name    string
    input   CreateInput
    wantErr bool
}{
    {"valid",         CreateInput{Username: "alice", Email: "a@b.com"}, false},
    {"short username", CreateInput{Username: "x", Email: "a@b.com"}, true},
}
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) { ... })
}
```

### 9.2 Use `t.Fatal` / `t.FailNow` in tests, never `panic` ([Don't Panic])

```go
// Bad
if err != nil { panic("setup failed") }

// Good
if err != nil { t.Fatal("setup failed:", err) }
```

---

## 10. Functional Options ([Functional Options])

Use the functional options pattern for constructors that have more than ~2
optional dependencies, or are likely to grow:

```go
type ServiceOption func(*Service)

func WithCache(c Cache) ServiceOption {
    return func(s *Service) { s.cache = c }
}

func NewService(repo Repository, opts ...ServiceOption) *Service {
    s := &Service{repo: repo}
    for _, opt := range opts { opt(s) }
    return s
}
```

---

## 11. Type assertions — always use comma-ok ([Handle Type Assertion Failures])

```go
// Bad — panics if type is wrong
val := iface.(string)

// Good
val, ok := iface.(string)
if !ok {
    return fmt.Errorf("unexpected type %T", iface)
}
```

---

## 12. Time handling ([Use `"time"` to handle time])

- Use `time.Time` for instants; use `time.Duration` for periods.
- Never use raw `int` / `int64` for time values in struct fields or function
  parameters. If you must use a primitive (e.g. for JSON compat), include the
  unit in the name (`IntervalMillis int`).
- Always work in UTC: `time.Now().UTC()`.

---

## 13. Linting

What actually runs today (from the root `Makefile`):

```bash
make lint       # bun run check (ultracite/Biome on JS/TS) + `go vet ./...` in apps/api
make lint-fix   # bun run fix  (ultracite auto-fix for JS/TS)
```

There is **no** `golangci-lint` config in the repo today — Go linting is `go vet`
only. The style rules in this document are enforced by convention and review, not
by a linter. If you add `golangci-lint`, wire it as a new Makefile target (e.g.
`lint-go: cd apps/api && golangci-lint run ./...`) and fold it into `lint`. A
sensible enabled set that maps to the rules above: `goimports`, `govet`,
`errcheck`, `staticcheck`, `revive`, `ineffassign`, `unused`, `misspell`,
`prealloc`, `bodyclose`, `exhaustive`, `noctx`, `gocritic`.

---

## 14. Adding a new domain — checklist

Follow this checklist every time (see `references/customization.md` for the
full recipe):

- [ ] `internal/domain/<x>/`: entity with `validate:` tags, `Repository` port interface
- [ ] `internal/domain/<x>/errors.go`: wraps shared sentinels via `domain.NotFound("x")`
- [ ] `internal/application/<x>/service.go`: use cases, depends only on the port
- [ ] `internal/application/<x>/service.go` (or `port.go`): define `XService` interface
- [ ] `var _ userdomain.Repository = (*XRepository)(nil)` in each adapter
- [ ] `var _ XService = (*Service)(nil)` in the service file
- [ ] `internal/adapters/http/{x}_handler.go` + `{x}_dto.go`: thin handlers, no business logic
- [ ] DTOs have only `json:` tags, no `binding:` tags
- [ ] `internal/adapters/persistence/postgres/{x}_repository.go`
- [ ] `internal/adapters/persistence/memory/{x}_repository.go`
- [ ] `db/queries/{x}.sql` + `make sqlc`
- [ ] Wire in `cmd/api/main.go`
- [ ] `make openapi && make gen-client`
- [ ] Table-driven tests for service and handler

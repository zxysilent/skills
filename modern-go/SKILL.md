---
name: modern-go
description: Use when writing or reviewing Go code and encountering legacy patterns (interface{}, manual loops, io/ioutil, pre-generics operations) — replace with modern stdlib APIs based on the project's go.mod Go version
---

# Modernizing Go Code

## Overview

Go's standard library has steadily absorbed common boilerplate into clean,
safe, often zero-allocation APIs. **Core principle: prefer the newest stdlib
idiom that the project's Go version supports** — it is shorter, safer, and
frequently faster than the hand-rolled legacy version.

**Always check `go.mod` for the project's Go version before applying a rule.**
Do not introduce an API newer than the declared `go` directive.

## How to Use

1. Read the target Go version from `go.mod`.
2. Scan the code for legacy patterns in the lookup table below.
3. Replace with the modern equivalent **only if the version supports it**.
4. Build and test (`go build ./... && go test ./...`) after changes.

## Quick Lookup Table

| Legacy pattern | Modern API | Since |
|----------------|-----------|-------|
| `time.Now().Sub(start)` | `time.Since(start)` | 1.0 |
| `deadline.Sub(time.Now())` | `time.Until(deadline)` | 1.8 |
| `err == ErrFoo` | `errors.Is(err, ErrFoo)` | 1.13 |
| `target, ok := err.(*T); ok` | `errors.As(err, &target)` | 1.13 |
| `fmt.Errorf("... %s", err)` (no unwrap) | `fmt.Errorf("... %w", err)` | 1.13 |
| `interface{}` | `any` | 1.18 |
| `strings.Index` / `strings.SplitN` + slicing | `strings.Cut` / `bytes.Cut` | 1.18 |
| `append(buf, []byte(fmt.Sprintf(...))...)` | `fmt.Appendf(buf, ...)` | 1.19 |
| `atomic.StoreInt32` / `atomic.Value` | `atomic.Bool` / `atomic.Int64` / `atomic.Pointer[T]` | 1.19 |
| `string([]byte(s[:n]))` | `strings.Clone` / `bytes.Clone` | 1.20 |
| `io/ioutil.ReadAll` | `io.ReadAll` | 1.16 |
| `io/ioutil.ReadFile` | `os.ReadFile` | 1.16 |
| `io/ioutil.WriteFile` | `os.WriteFile` | 1.16 |
| `io/ioutil.ReadDir` | `os.ReadDir` | 1.16 |
| `io/ioutil.TempFile` | `os.CreateTemp` | 1.16 |
| `io/ioutil.TempDir` | `os.MkdirTemp` | 1.16 |
| `io/ioutil.NopCloser` | `io.NopCloser` | 1.16 |
| `io/ioutil.Discard` | `io.Discard` | 1.16 |
| `net.IP` (mutable, alloc-heavy) | `net/netip.Addr` | 1.18 |
| `*net.UDPAddr` | `net/netip.AddrPort` | 1.18 |
| `sync.Mutex` manual try-lock pattern | `sync.Mutex.TryLock` | 1.18 |
| `context.WithCancel` (no reason) | `context.WithCancelCause` + `context.Cause` | 1.20 |
| `context.WithTimeout` (no reason) | `context.WithTimeoutCause` + `context.Cause` | 1.21 |
| `context.WithDeadline` (no reason) | `context.WithDeadlineCause` + `context.Cause` | 1.21 |
| `errors.New` + `fmt.Errorf` to combine | `errors.Join` | 1.20 |
| `strings.TrimPrefix(s, pre); found := s != orig` | `after, found := strings.CutPrefix(s, pre)` | 1.20 |
| `strings.TrimSuffix(s, suf); found := s != orig` | `after, found := strings.CutSuffix(s, suf)` | 1.20 |
| `time.Now().Format("2006-01-02")` | `time.DateOnly` | 1.20 |
| `time.Now().Format("15:04:05")` | `time.TimeOnly` | 1.20 |
| `time.Now().Format("2006-01-02 15:04:05")` | `time.DateTime` | 1.20 |
| `t1.Before(t2) \|\| t1.Equal(t2)` | `t1.Compare(t2) >= 0` | 1.20 |
| `math/rand.Seed` + global funcs | local `rand.New(rand.NewSource(seed))` | 1.20 |
| manual min/max if-else | `min(a, b)` / `max(a, b)` | 1.21 |
| `for k := range m { delete(m, k) }` | `clear(m)` | 1.21 |
| manual `contains` loop | `slices.Contains` / `slices.IndexFunc` | 1.21 |
| `sort.Slice` | `slices.Sort` / `slices.SortFunc` | 1.21 |
| manual dedupe/clip | `slices.Compact` / `slices.Clip` | 1.21 |
| manual map copy/filter | `maps.Clone` / `maps.DeleteFunc` | 1.21 |
| `sync.Once` + outer var | `sync.OnceValue` / `sync.OnceFunc` | 1.21 |
| manual structured logging | `log/slog` | 1.21 |
| `encoding/binary.BigEndian` / LittleEndian check | `encoding/binary.NativeEndian` | 1.21 |
| custom "unsupported" error | `errors.ErrUnsupported` | 1.21 |
| `for i := 0; i < n; i++` | `for i := range n` | 1.22 |
| `if x == "" { x = def }` / `if x == 0 { x = def }` | `cmp.Or(x, def)` | 1.22 |
| 3rd-party router for method/path params | `http.ServeMux` w/ `"POST /x/{id}"` + `PathValue` | 1.22 |
| `math/rand` global funcs | `math/rand/v2` (ChaCha8/PCG) | 1.22 |
| `database/sql.NullString` / `NullInt64` etc. | `database/sql.Null[T]` | 1.22 |
| manual collect keys + sort | `slices.Collect` / `slices.Sorted` + `maps.Keys` | 1.23 |
| reverse slice loop for backwards | `slices.Backward` | 1.23 |
| manual chunk loop | `slices.Chunk` | 1.23 |
| `for k, v := range m` (iterator form) | `maps.All` / `maps.Keys` / `maps.Values` | 1.23 |
| `sync.Map` clear loop | `sync.Map.Clear` | 1.23 |
| `sync/atomic` manual bitwise CAS loops | `sync/atomic.Or` / `sync/atomic.And` | 1.23 |
| manual struct host-layout annotation | `structs.HostLayout` | 1.23 |
| `ctx, cancel := WithCancel` in tests | `t.Context()` | 1.24 |
| `for i := 0; i < b.N; i++` | `for b.Loop()` | 1.24 |
| `omitempty` on `time.Time` / struct | `omitzero` | 1.24 |
| `range strings.Split(s, ",")` (alloc) | `range strings.SplitSeq(s, ",")` | 1.24 |
| `runtime.SetFinalizer` | `runtime.AddCleanup` | 1.24 |
| filesystem sandbox by hand | `os.Root` / `os.OpenRoot` | 1.24 |
| `golang.org/x/crypto/hkdf` | `crypto/hkdf` | 1.24 |
| `golang.org/x/crypto/pbkdf2` | `crypto/pbkdf2` | 1.24 |
| `golang.org/x/crypto/sha3` | `crypto/sha3` | 1.24 |
| `wg.Add(1)` + `go func(){ defer wg.Done() }` | `wg.Go(func(){ ... })` | 1.25 |
| manual hash clone | `hash.Cloner` (all stdlib Hashes) | 1.25 |
| concurrent test with fake time by hand | `testing/synctest.Test` + `synctest.Wait` | 1.25 |
| `x := v; p := &x` | `new(v)` (expression form) | 1.26 |
| `errors.As(err, &target)` | `errors.AsType[*T](err)` | 1.26 |
| `io.ReadAll` (suboptimal) | improved `io.ReadAll` (half mem, 2x speed) | 1.26 |

## Phase 1 — Early Cleanups (Go 1.0–1.19)

### Time arithmetic
```go
// Legacy
elapsed := time.Now().Sub(start)
remaining := deadline.Sub(time.Now())
// Modern
elapsed := time.Since(start)          // 1.0
remaining := time.Until(deadline)     // 1.8
```

### Error wrapping (1.13)
`==` breaks once an error is wrapped with `%w`. Use `errors.Is` and `errors.As`.
```go
// Legacy
if err == sql.ErrNoRows { ... }
target, ok := err.(*os.PathError)

// Modern
if errors.Is(err, sql.ErrNoRows) { ... }           // 1.13
var pathErr *os.PathError
if errors.As(err, &pathErr) { ... }                 // 1.13
```

### Wrapping errors with context (1.13)
```go
// Legacy
return fmt.Errorf("open %s: %s", path, err)

// Modern — preserves the original error for Is/As checks
return fmt.Errorf("open %s: %w", path, err)         // 1.13
```

### `any` alias (1.18)
```go
// Legacy
func PrintAll(vals []interface{}) { ... }

// Modern
func PrintAll(vals []any) { ... }                    // 1.18
```

### Safe key/value split (1.18)
`strings.Cut` avoids `slice bounds out of range` panics.
```go
// Legacy
idx := strings.Index(header, ":")
if idx < 0 { return }
key, value := header[:idx], header[idx+1:]

// Modern
if key, value, found := strings.Cut(header, ":"); found { ... }  // 1.18
```

### Zero-alloc formatted append (1.19)
```go
// Legacy
buf = append(buf, []byte(fmt.Sprintf("user_id=%d", id))...)

// Modern
buf = fmt.Appendf(buf, "user_id=%d", id)             // 1.19
```

### Typed atomics (1.19)
```go
// Legacy
var flag int32
atomic.StoreInt32(&flag, 1)
if atomic.LoadInt32(&flag) == 1 { ... }

var cfg atomic.Value
cfg.Store(&Config{})

// Modern
var flag atomic.Bool                              // 1.19
flag.Store(true)
if flag.Load() { ... }

var cfg atomic.Pointer[Config]                      // 1.19
cfg.Store(&Config{})
```

### io/ioutil migration (1.16)
All `io/ioutil` functions have moved to `io` or `os`. The old package still
works but new code should use the new locations.
```go
// Legacy
data, _ := ioutil.ReadAll(r)
body, _ := ioutil.ReadFile("config.json")
ioutil.WriteFile("out.txt", data, 0644)
files, _ := ioutil.ReadDir(".")
f, _ := ioutil.TempFile("", "prefix-*")
d, _ := ioutil.TempDir("", "prefix-*")
var r io.ReadCloser = ioutil.NopCloser(buf)
io.Copy(ioutil.Discard, data)

// Modern
data, _ := io.ReadAll(r)                           // 1.16
body, _ := os.ReadFile("config.json")               // 1.16
os.WriteFile("out.txt", data, 0644)                 // 1.16
files, _ := os.ReadDir(".")                         // 1.16 (returns []os.DirEntry)
f, _ := os.CreateTemp("", "prefix-*")               // 1.16
d, _ := os.MkdirTemp("", "prefix-*")                // 1.16
var r io.ReadCloser = io.NopCloser(buf)             // 1.16
io.Copy(io.Discard, data)                           // 1.16
```

### net/netip — efficient IP types (1.18)
```go
// Legacy (heap-allocated, mutable, not comparable)
ip := net.ParseIP("192.0.2.1")
udpAddr := &net.UDPAddr{IP: ip, Port: 53}

// Modern (value type, comparable, zero-allocation parse)
ip, _ := netip.ParseAddr("192.0.2.1")               // 1.18
ap := netip.AddrPortFrom(ip, 53)                     // 1.18
// ip and ap can be map keys
```

### Mutex TryLock (1.18)
```go
// Legacy — no non-blocking lock attempt; must use channel or atomic
if mu.TryLock() {                                  // 1.18
    defer mu.Unlock()
    // ...
}
```

## Phase 2 — Generic Renaissance (Go 1.20–1.21)

### Explicit clone (1.20)
Avoid pinning a huge backing array when keeping a small slice.
```go
// Legacy — small slice keeps huge backing array alive
copied := make([]byte, 10)
copy(copied, huge[:10])

// Modern
copied := strings.Clone(huge[:10])                  // 1.20
copiedBytes := bytes.Clone(hugeBytes[:10])          // 1.20
```

### Cancel cause (1.20; `WithTimeoutCause` in 1.21)
```go
// Legacy
ctx, cancel := context.WithCancel(parent)
cancel() // no reason

// Modern
ctx, cancel := context.WithCancelCause(parent)      // 1.20
cancel(fmt.Errorf("db connection lost"))
err := context.Cause(ctx) // "db connection lost"

// For timeout/deadline
ctx, cancel := context.WithTimeoutCause(parent, 5*time.Second, errTimeout)  // 1.21
```

### Join multiple errors (1.20)
```go
// Legacy
var errs []error
errs = append(errs, err1, err2)
// ...custom joining logic...

// Modern
err := errors.Join(err1, err2)                      // 1.20
// errors.Is and errors.As traverse all joined errors
```

### CutPrefix / CutSuffix (1.20)
```go
// Legacy
if strings.HasPrefix(s, prefix) {
    result = s[len(prefix):]
}

// Modern
if result, ok := strings.CutPrefix(s, prefix); ok { ... }   // 1.20
if result, ok := strings.CutSuffix(s, suffix); ok { ... }   // 1.20
```

### Time layout constants (1.20)
```go
// Legacy
t.Format("2006-01-02")
t.Format("15:04:05")
t.Format("2006-01-02 15:04:05")

// Modern
t.Format(time.DateOnly)                             // 1.20
t.Format(time.TimeOnly)                             // 1.20
t.Format(time.DateTime)                             // 1.20
```

### Time comparison (1.20)
```go
// Legacy
if t1.Before(t2) || t1.Equal(t2) { ... }
if t1.After(t2) { ... }

// Modern
if t1.Compare(t2) >= 0 { ... }                      // 1.20
if t1.Compare(t2) > 0 { ... }                       // 1.20
```

### math/rand — seeded source, no global Seed (1.20)
```go
// Legacy
rand.Seed(time.Now().UnixNano())
n := rand.Intn(100)

// Modern
rng := rand.New(rand.NewSource(time.Now().UnixNano()))   // 1.20
n := rng.IntN(100)
```

### Builtins (1.21)
```go
m := max(a, b)                                      // 1.21
clear(myMap)                                        // 1.21 — keeps capacity
clear(mySlice)                                      // 1.21 — zeroes elements
```

### slices / maps (1.21)
```go
// slices
found := slices.Contains(items, target)             // 1.21
idx := slices.IndexFunc(users, func(u User) bool { return u.ID == 42 })  // 1.21
slices.Sort(ints)                                   // 1.21
slices.SortFunc(users, func(a, b User) int { return cmp.Compare(a.Age, b.Age) })
items = slices.Compact(items)                       // 1.21 — consecutive only
items = slices.Clip(items)                          // 1.21

// maps
cloned := maps.Clone(original)                      // 1.21
maps.DeleteFunc(m, func(k string, v int) bool { return v < 0 })  // 1.21
```

### Once with value (1.21)
```go
// Legacy
var (
    config *Config
    once   sync.Once
)
func GetConfig() *Config {
    once.Do(func() { config = loadConfig() })
    return config
}

// Modern
var GetConfig = sync.OnceValue(func() *Config {     // 1.21
    return loadConfig()
})
```

### Structured logging (1.21)
```go
// Legacy
log.Printf("request %s took %v", path, elapsed)

// Modern
slog.Info("request", "path", path, "elapsed", elapsed)        // 1.21
logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))       // 1.21
```

### Native endianness (1.21)
```go
// Legacy
var order binary.ByteOrder = binary.BigEndian
// ...or hardcode LittleEndian...

// Modern
order := binary.NativeEndian                       // 1.21
```

### Standard "unsupported" error (1.21)
```go
// Legacy
return fmt.Errorf("operation unsupported")

// Modern
return fmt.Errorf("%w: feature X", errors.ErrUnsupported)  // 1.21
// errors.Is(err, errors.ErrUnsupported) returns true
```

## Phase 3 — Syntax & Routing (Go 1.22)

### Integer range & cmp.Or
```go
// Legacy
for i := 0; i < 10; i++ { ... }

// Modern
for i := range 10 { ... }                           // 1.22

// Legacy
port := os.Getenv("PORT")
if port == "" { port = "8080" }

// Modern
port := cmp.Or(os.Getenv("PORT"), "8080")           // 1.22
```

### Enhanced HTTP routing
```go
// Legacy — need chi, gorilla/mux, or manual method/path parsing
mux := http.NewServeMux()
mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
    id := strings.TrimPrefix(r.URL.Path, "/api/users/")
    // ...
})

// Modern — stdlib supports method and path wildcards
mux := http.NewServeMux()
mux.HandleFunc("GET /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("id")                     // 1.22
})
mux.HandleFunc("POST /api/users", createUser)       // 1.22 — method-specific
mux.HandleFunc("/files/{path...}", serveFile)       // 1.22 — catch-all wildcard
```

### math/rand/v2 (1.22)
```go
// Legacy
import "math/rand"
rand.Seed(42)
n := rand.Intn(100)

// Modern
import "math/rand/v2"
// No global Seed — top-level is auto-seeded with ChaCha8
n := rand.IntN(100)                                 // 1.22
dur := rand.N(5 * time.Minute)                      // 1.22 — generic N
// For reproducible sequences:
rng := rand.New(rand.NewPCG(42, 0))                 // 1.22
```

### Generic nullable types in database/sql (1.22)
```go
// Legacy
var name sql.NullString
var age  sql.NullInt64
db.QueryRow("...").Scan(&name, &age)

// Modern
var name sql.Null[string]                           // 1.22
var age  sql.Null[int64]                            // 1.22
db.QueryRow("...").Scan(&name, &age)
```

### slices.Concat (1.22)
```go
// Legacy
combined := append(append([]int{}, a...), b...)
combined = append(combined, c...)

// Modern
combined := slices.Concat(a, b, c)                  // 1.22
```

## Phase 4 — Iterators (Go 1.23)

### range-over-func
Go 1.23 lets `for-range` iterate over functions of type `func(func(K) bool)` or `func(func(K, V) bool)`.
```go
// Custom iterators
func BackwardSlice[T any](s []T) func(func(int, T) bool) {
    return func(yield func(int, T) bool) {
        for i := len(s) - 1; i >= 0; i-- {
            if !yield(i, s[i]) { return }
        }
    }
}
for i, v := range BackwardSlice(items) { ... }      // 1.23
```

### slices / maps iterator functions (1.23)
```go
// slices
for i, v := range slices.All(items) { ... }         // 1.23
for v := range slices.Values(items) { ... }         // 1.23
for i, v := range slices.Backward(items) { ... }    // 1.23
for chunk := range slices.Chunk(items, 3) { ... }   // 1.23

collected := slices.Collect(iter)                   // 1.23
slices.AppendSeq(&dst, iter)                        // 1.23
sorted := slices.Sorted(iter)                       // 1.23

// maps
for k, v := range maps.All(m) { ... }               // 1.23
for k := range maps.Keys(m) { ... }                 // 1.23
for v := range maps.Values(m) { ... }               // 1.23
maps.Insert(dst, iter)                              // 1.23 — insert from iterator
collected := maps.Collect(iter)                     // 1.23
```

### sync.Map.Clear (1.23)
```go
// Legacy
m.Range(func(k, v any) bool { m.Delete(k); return true })

// Modern
m.Clear()                                           // 1.23
```

### sync/atomic bitwise operations (1.23)
```go
// Legacy — manual CAS loop
old := atomic.LoadUint32(&flags)
for !atomic.CompareAndSwapUint32(&flags, old, old|mask) {
    old = atomic.LoadUint32(&flags)
}

// Modern
old := atomic.OrUint32(&flags, mask)                // 1.23
old := atomic.AndUint32(&flags, ^mask)              // 1.23
```

### CopyFS (1.23)
```go
// Legacy — manual walk + copy
filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
    // ...create dirs, copy files...
})

// Modern
os.CopyFS(dstDir, os.DirFS(srcDir))                 // 1.23
```

### Generic type aliases (1.23 — preview, 1.24 GA)
```go
// Legacy — type aliases couldn't be parameterized
//  (work around with thin wrapper types)

// Modern (Go 1.24+)
type Set[T comparable] = map[T]struct{}             // 1.24
```

## Phase 5 — Quality of Life (Go 1.24)

### Better benchmarks & tests
```go
// Legacy
func TestFoo(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    doSomething(ctx)
}

func BenchmarkBar(b *testing.B) {
    for i := 0; i < b.N; i++ { doWork() }
}

// Modern
func TestFoo(t *testing.T) {
    ctx := t.Context()                              // 1.24 — auto-cancelled
    doSomething(ctx)
}

func BenchmarkBar(b *testing.B) {
    for b.Loop() { doWork() }                       // 1.24 — per-count, no inlining loss
}
```

### omitzero (1.24)
```go
// Legacy — time.Time{} is never "empty" for omitempty
type User struct {
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at,omitempty"` // doesn't omit zero time
}

// Modern
type User struct {
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at,omitzero"`  // 1.24 — omits zero time
}
// If both omitempty and omitzero are specified, the field is omitted if either is true.
// omitzero works with any type that has an IsZero() bool method.
```

### Zero-allocation SplitSeq (1.24)
```go
// Legacy — allocates a slice
for _, part := range strings.Split(s, ",") { ... }

// Modern — zero-allocation iterator
for part := range strings.SplitSeq(s, ",") { ... }  // 1.24
for part := range strings.SplitAfterSeq(s, ",") { ... }  // 1.24
for line := range strings.Lines(s) { ... }          // 1.24
for field := range strings.FieldsSeq(s) { ... }     // 1.24
// Same for bytes.SplitSeq, bytes.Lines, etc.
```

### os.Root — directory-limited filesystem access (1.24)
```go
// Legacy
f, err := os.Open(filepath.Join(root, userPath))
// Must manually check for path traversal attacks

// Modern
r, err := os.OpenRoot("/var/data")                  // 1.24
f, err := r.Open(userPath)                          // 1.24 — paths cannot escape root
r.Create("subdir/file.txt")                         // 1.24
r.Mkdir("newdir", 0755)                             // 1.24
r.Stat("existing")                                  // 1.24
```

### runtime.AddCleanup (1.24)
```go
// Legacy
runtime.SetFinalizer(obj, func(o *Obj) { cleanup(o) })
// Problems: only one finalizer, must be pointer to allocated object,
// can cause leaks in cycles, delays freeing.

// Modern
runtime.AddCleanup(obj, cleanup, obj)               // 1.24
// Multiple cleanups, works on interior pointers, no cycle leaks.
```

### New crypto packages promoted from x/crypto (1.24)
```go
// Legacy
import "golang.org/x/crypto/hkdf"
import "golang.org/x/crypto/pbkdf2"
import "golang.org/x/crypto/sha3"

// Modern
import "crypto/hkdf"                                // 1.24
import "crypto/pbkdf2"                              // 1.24
import "crypto/sha3"                                // 1.24
```

### weak package (1.24)
```go
// Modern — weak pointers for caches and canonicalization
var cache sync.Map // map[weak.Pointer[Key]]*Value
ptr := weak.Make(&key)                              // 1.24
cache.Store(ptr, value)
// When key becomes unreachable, ptr can be collected.
```

### FIPS 140-3 compliance (1.24)
```go
// Build with GOFIPS140=latest for FIPS 140-3 mode.
// Runtime: GODEBUG=fips140=on (or =only for strict enforcement).
// crypto/fips140.Enforced() reports current status.
```

## Phase 6 — Present Future (Go 1.25–1.26)

### WaitGroup.Go (1.25)
```go
// Legacy
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(item Item) {
        defer wg.Done()
        process(item)
    }(item)
}
wg.Wait()

// Modern
var wg sync.WaitGroup
for _, item := range items {
    wg.Go(func() { process(item) })                 // 1.25
}
wg.Wait()
```

### new(expr) (1.26)
```go
// Legacy
cfg := Config{
    Timeout: func() *int { v := 30; return &v }(),
    Debug:   func() *bool { v := true; return &v }(),
    Role:    func() *string { v := "admin"; return &v }(),
}

// Modern
cfg := Config{                                      // 1.26
    Timeout: new(30),
    Debug:   new(true),
    Role:    new("admin"),
}
```

### errors.AsType (1.26)
```go
// Legacy — verbose, error-prone (forget to pass pointer)
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    handle(pathErr)
}

// Modern — type-safe, no &&-panic risk
if pathErr, ok := errors.AsType[*os.PathError](err); ok {   // 1.26
    handle(pathErr)
}
```

### hash.Cloner (1.25)
```go
// Legacy — no way to clone a hash state
// (had to keep the original bytes and re-hash)

// Modern
if cloner, ok := h.(hash.Cloner); ok {              // 1.25
    clone := cloner.Clone()
}
// All stdlib Hash implementations now implement hash.Cloner.
```

### testing/synctest (1.25)
```go
// Legacy — testing time-dependent concurrent code required
// sleeping, mock clocks, or complex test harnesses.

// Modern — isolated bubble with virtual time
func TestConcurrent(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        // Inside this bubble, time is virtual.
        // The clock advances only when all goroutines block.
        // time.Sleep, time.Ticker, context.WithTimeout all work.
        synctest.Wait() // waits for all goroutines in bubble to block
    })
}
```

### crypto.MessageSigner (1.25)
```go
// Legacy — signing required crypto.Signer which hashes separately
// Modern — "one-shot" signing where the signer does the hashing
if ms, ok := key.(crypto.MessageSigner); ok {       // 1.25
    sig, err := ms.SignMessage(nil, digest, opts)
}
```

### crypto.Encapsulator / Decapsulator (1.26)
```go
// Modern — abstract KEM operations
func ExchangeKey(decapsulator crypto.Decapsulator) ([]byte, error) {  // 1.26
    enc, shared, err := decapsulator.Encapsulator().Encapsulate()
    // ...
}
```

### Testing artifacts (1.26)
```go
// Legacy — write test outputs to os.TempDir() or hardcoded paths
// Modern
dir := t.ArtifactDir()                              // 1.26
os.WriteFile(filepath.Join(dir, "output.json"), data, 0644)
// Run with: go test -artifacts -outputdir=./testdata
```

## Guardrails

- **Never** introduce an API newer than the `go` directive in `go.mod`.
- `clear(map)` empties; `clear(slice)` zeroes elements (does not shrink).
- `slices.Compact` only removes **consecutive** duplicates — sort first if you
  need full dedupe.
- `omitzero` and `omitempty` differ: `omitzero` drops zero-valued structs/times;
  keep `omitempty` for slices/maps where empty-but-non-nil matters.
- `slices.Chunk(n)` returns an iterator, not a `[][]T`. Use `slices.Collect`
  if you need a materialized slice of chunks.
- `os.Root` prevents path traversal: all paths are resolved within the root.
  Methods like `ReadFile`, `WriteFile`, `Readlink`, `MkdirAll`, etc., are
  progressively added across Go 1.24–1.26.
- `sync.WaitGroup.Go` panics if called after `Wait` (same as manual `Add`).
- `errors.AsType` is a generic type-safe alternative to `errors.As`. Use it
  for simpler, safer error type assertions on Go ≥1.26.
- `testing/synctest` creates an isolated bubble — synctest.Wait() blocks
  until all goroutines in the bubble are blocked. Virtual time advances
  automatically only when all goroutines are blocked.
- `hash/maphash.Comparable` (1.24) lets you hash any comparable value
  (including structs) for use in custom hash tables.
- After modernizing, run `gofmt -l .`, `go build ./...`, and `go test ./...`.

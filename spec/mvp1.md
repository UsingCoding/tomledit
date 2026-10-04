# Go TOML Edit Library via Rust `toml_edit` + WebAssembly

## 1. Goal

Implement a Go library that provides format-preserving TOML editing by embedding Rust's `toml_edit` library as WebAssembly.

Primary goals:

- Public API is pure Go.
- Final consumer applications remain single executable binaries.
- No CGO.
- No external Rust runtime.
- No external `.wasm` file at runtime.
- No Rust installation required for consumers of the Go module.
- Preserve existing TOML formatting, comments, whitespace, quoting, and relative item ordering as supported by `toml_edit`.
- Support modification of:
  - scalar values;
  - tables;
  - arrays;
  - arrays of tables;
  - raw TOML values;
  - multiple operations atomically.
- Support insertion into the middle of `[[array.of.tables]]`.
- Keep all Rust-specific implementation details internal.

Architecture:

```text
Application
    │
    │ Go API
    ▼
┌─────────────────────────────┐
│ tomledit Go package         │
│                             │
│ Apply(input, operations...) │
└──────────────┬──────────────┘
               │
               │ JSON protocol
               ▼
┌─────────────────────────────┐
│ wazero                      │
│                             │
│ embedded tomledit.wasm      │
└──────────────┬──────────────┘
               │
               │ WASM ABI
               ▼
┌─────────────────────────────┐
│ Rust shim                   │
│                             │
│ protocol decode             │
│ path resolution             │
│ operation dispatch          │
└──────────────┬──────────────┘
               │
               ▼
┌─────────────────────────────┐
│ toml_edit                   │
│                             │
│ DocumentMut                 │
└─────────────────────────────┘
```

The Rust layer MUST NOT expose persistent `DocumentMut`, `Table`, or other Rust objects to Go.

Each call operates as:

```text
TOML bytes
+
operations
    │
    ▼
one WASM invocation
    │
    ▼
parse DocumentMut once
    │
    ├── operation 1
    ├── operation 2
    ├── operation 3
    └── ...
    │
    ▼
serialize once
    │
    ▼
modified TOML bytes
```

---

# 2. Technology Stack

## Go

Use:

```text
Go 1.27.1
```

Go 1.27 was released in August 2026 and 1.27.1 is the current patch release used for this project baseline.

Runtime dependency:

```text
github.com/tetratelabs/wazero v1.12.0
```

`wazero` is used to execute the embedded WebAssembly module without CGO.

No other runtime Go dependencies should be required initially.

Standard library should be preferred for:

- JSON encoding;
- error handling;
- embedding;
- synchronization;
- context;
- testing.

## Rust

Use:

```text
Rust 1.99.0
edition = "2024"
```

Rust 1.99.0 was released October 1, 2026.

Compile target:

```text
wasm32-unknown-unknown
```

WASI SHOULD NOT be required.

The Rust code:

- does not access files;
- does not access network;
- does not read environment variables;
- does not require clocks;
- does not require OS APIs.

This allows the Rust shim to compile to a plain WebAssembly module with a minimal ABI.

## TOML implementation

Use:

```text
toml_edit = "0.25.15"
```

`toml_edit` specifically supports parsing and modifying TOML while preserving comments, spaces, and relative ordering.

Known upstream formatting limitations MUST be documented:

- ordering of dotted keys is not fully preserved;
- a missing trailing newline may not be preserved.

These are upstream `toml_edit` limitations rather than limitations introduced by the Go/WASM wrapper.

---

# 3. Toolchain

Toolchain management MUST use `mise`.

Root:

```text
mise.toml
```

Recommended configuration:

```toml
[tools]
go = "1.27.1"
rust = "1.99.0"

[tasks.bootstrap]
description = "Install additional development toolchain components"
run = """
rustup target add wasm32-unknown-unknown
"""

[tasks."rust:fmt"]
description = "Format Rust shim"
dir = "rust"
run = "cargo fmt"

[tasks."rust:lint"]
description = "Lint Rust shim"
dir = "rust"
run = "cargo clippy --target wasm32-unknown-unknown --all-targets -- -D warnings"

[tasks."rust:test"]
description = "Run native Rust unit tests"
dir = "rust"
run = "cargo test"

[tasks."wasm:build"]
description = "Build TOML editor WebAssembly module"
depends = ["bootstrap"]
run = """
cargo build \
  --manifest-path rust/Cargo.toml \
  --target wasm32-unknown-unknown \
  --release

cp \
  rust/target/wasm32-unknown-unknown/release/tomledit_wasm.wasm \
  internal/wasm/assets/tomledit.wasm
"""

[tasks."go:fmt"]
description = "Format Go code"
run = "gofmt -w ."

[tasks."go:test"]
description = "Run Go tests"
run = "go test ./..."

[tasks.test]
description = "Run complete test suite"
depends = [
  "rust:test",
  "go:test",
]

[tasks.build]
description = "Build library examples"
depends = ["wasm:build"]
run = """
CGO_ENABLED=0 go build ./...
"""

[tasks.check]
description = "Run formatting, linting and tests"
depends = [
  "rust:lint",
  "rust:test",
  "go:test",
]
```

The exact task implementation may evolve, but the canonical developer workflow MUST remain:

```bash
mise install
mise run bootstrap
mise run wasm:build
mise run test
```

Normal Go consumers MUST NOT need to execute these commands.

---

# 4. Generated WASM Policy

The built WebAssembly module MUST be committed into the Go repository:

```text
internal/wasm/assets/tomledit.wasm
```

Reason:

A user should be able to:

```bash
go get github.com/<org>/tomledit
```

and then:

```bash
CGO_ENABLED=0 go build
```

without:

- Cargo;
- Rust;
- rustup;
- a WASM compiler;
- network access during compilation beyond normal Go module resolution.

The Rust source remains the source of truth.

CI SHOULD verify that the committed WASM corresponds to the current Rust source.

Conceptually:

```text
build Rust to temporary WASM
            │
            ▼
compare
            │
            ▼
internal/wasm/assets/tomledit.wasm
```

If different, CI fails with an instruction to run:

```bash
mise run wasm:build
```

---

# 5. Project Structure

Use the following structure:

```text
.
├── go.mod
├── go.sum
├── mise.toml
├── README.md
├── LICENSE
│
├── tomledit.go
├── operation.go
├── path.go
├── value.go
├── field.go
├── errors.go
│
├── internal/
│   ├── engine/
│   │   ├── engine.go
│   │   ├── protocol.go
│   │   └── errors.go
│   │
│   └── wasm/
│       ├── embed.go
│       └── assets/
│           └── tomledit.wasm
│
├── rust/
│   ├── Cargo.toml
│   ├── Cargo.lock
│   │
│   ├── src/
│   │   ├── lib.rs
│   │   ├── protocol.rs
│   │   ├── apply.rs
│   │   ├── path.rs
│   │   ├── value.rs
│   │   └── error.rs
│   │
│   └── tests/
│       ├── set.rs
│       ├── array.rs
│       ├── array_of_tables.rs
│       └── formatting.rs
│
├── examples/
│   ├── set/
│   │   └── main.go
│   ├── array/
│   │   └── main.go
│   ├── array_of_tables/
│   │   └── main.go
│   └── batch/
│       └── main.go
│
└── testdata/
    ├── formatting.toml
    ├── arrays.toml
    └── belt.toml
```

Responsibilities:

```text
public root package
    Go-facing API only

internal/engine
    wazero execution
    JSON protocol
    WASM lifecycle

internal/wasm
    embedded artifact

rust/
    Rust shim
    toml_edit integration

examples/
    executable examples of supported API

testdata/
    integration/format preservation fixtures
```

The public Go package MUST NOT expose anything from:

```text
wazero
WASM ABI
Rust
serde
toml_edit
```

---

# 6. Rust Project

`rust/Cargo.toml`:

```toml
[package]
name = "tomledit-wasm"
version = "0.1.0"
edition = "2024"

[lib]
crate-type = ["cdylib"]

[dependencies]
toml_edit = "0.25.15"

serde = { version = "1", features = ["derive"] }
serde_json = "1"
```

`Cargo.lock` MUST be committed.

---

# 7. WASM ABI

Keep the low-level ABI deliberately tiny.

Rust MUST export:

```text
alloc
dealloc
apply
```

Conceptually:

```rust
extern "C" fn alloc(len: u32) -> u32

extern "C" fn dealloc(
    ptr: u32,
    len: u32,
)

extern "C" fn apply(
    ptr: u32,
    len: u32,
) -> u64
```

`apply()` returns:

```text
high 32 bits = output pointer
low  32 bits = output length
```

The flow is:

```text
Go
 │
 ├─ JSON marshal request
 │
 ├─ WASM alloc(request length)
 │
 ├─ write request into WASM memory
 │
 ├─ apply(ptr, len)
 │
 ├─ unpack result ptr/len
 │
 ├─ read result bytes
 │
 ├─ dealloc input
 │
 └─ dealloc output
```

No host callbacks should be required by the Rust module.

---

# 8. Internal Protocol

Protocol MUST be explicitly versioned.

Example request:

```json
{
  "version": 1,
  "toml": "[[belt.actions]]\n...",
  "operations": [
    {
      "op": "aot_insert",
      "path": ["belt", "actions"],
      "index": 2,
      "fields": [
        {
          "key": "name",
          "value": {
            "type": "string",
            "value": "go-tidy"
          }
        }
      ]
    }
  ]
}
```

Success response:

```json
{
  "version": 1,
  "ok": true,
  "toml": "[[belt.actions]]\n..."
}
```

Error response:

```json
{
  "version": 1,
  "ok": false,
  "error": {
    "code": "index_out_of_range",
    "message": "index 5 is outside belt.actions",
    "path": [
      "belt",
      "actions"
    ]
  }
}
```

Operations MUST be applied atomically from the caller's perspective.

If operation 3 fails:

```text
operation 1 → success
operation 2 → success
operation 3 → error
```

the API MUST return an error and MUST NOT return a partially edited document.

---

# 9. Public Go API

Primary API:

```go
type Editor struct {
    // internal
}

func New(ctx context.Context) (*Editor, error)

func (e *Editor) Close(ctx context.Context) error

func (e *Editor) Apply(
    ctx context.Context,
    input []byte,
    operations ...Operation,
) ([]byte, error)
```

`Editor` MUST be safe for normal concurrent use.

Implementation may either:

- serialize access to a WASM module instance; or
- instantiate lightweight module instances from a compiled wazero module.

The concurrency strategy MUST remain internal.

---

# 10. Paths

Define:

```go
type Path []string
```

Helper:

```go
func P(parts ...string) Path
```

Example:

```go
tomledit.P(
    "belt",
    "actions",
)
```

Avoid APIs based on parsing dotted strings such as:

```go
"belt.actions"
```

because TOML keys themselves may contain dots.

For example:

```toml
["foo.bar"]
value = 1
```

must be distinguishable from:

```toml
[foo.bar]
value = 1
```

Therefore:

```go
P("foo.bar", "value")
```

and:

```go
P("foo", "bar", "value")
```

MUST represent different paths.

---

# 11. Values

Define a Go `Value` abstraction.

Required constructors:

```go
func String(v string) Value

func Integer(v int64) Value

func Float(v float64) Value

func Boolean(v bool) Value

func Array(values ...Value) Value

func Raw(v string) Value
```

Potential later additions:

```go
func DateTime(v time.Time) Value

func LocalDate(...)
func LocalTime(...)
func InlineTable(fields ...Field) Value
```

## Raw values

`Raw()` is an intentional advanced escape hatch.

Example:

```go
tomledit.Raw("'hello'")
```

produces the TOML representation:

```toml
'hello'
```

instead of allowing `toml_edit` to select its default representation.

Another example:

```go
tomledit.Raw(`{ name = "foo", enabled = true }`)
```

Rust MUST parse raw values using TOML syntax before inserting them.

Invalid raw values MUST return:

```text
invalid_raw_value
```

rather than injecting invalid TOML.

---

# 12. Ordered Fields

Do not represent newly created table fields with:

```go
map[string]Value
```

because output ordering matters.

Use:

```go
type Field struct {
    Key   string
    Value Value
}
```

Helper:

```go
func F(
    key string,
    value Value,
) Field
```

Usage:

```go
tomledit.F("name", tomledit.String("go-tidy"))
```

The following order:

```go
tomledit.F("name", ...),
tomledit.F("action", ...),
tomledit.F("stage", ...),
```

MUST produce:

```toml
name = ...
action = ...
stage = ...
```

in that order.

---

# 13. MVP Operations

## Set

```go
func Set(
    path Path,
    value Value,
) Operation
```

Example:

```go
tomledit.Set(
    tomledit.P("package", "version"),
    tomledit.String("1.2.3"),
)
```

Input:

```toml
[package]
name = "demo"
version = "1.0.0"
```

Output:

```toml
[package]
name = "demo"
version = "1.2.3"
```

Existing surrounding formatting MUST remain intact.

---

## Delete

```go
func Delete(
    path Path,
) Operation
```

Example:

```go
tomledit.Delete(
    tomledit.P("package", "deprecated"),
)
```

Input:

```toml
[package]
name = "demo"
deprecated = true
```

Output:

```toml
[package]
name = "demo"
```

---

# 14. Array Operations

Required:

```go
func ArrayAppend(
    path Path,
    value Value,
) Operation

func ArrayInsert(
    path Path,
    index int,
    value Value,
) Operation

func ArrayReplace(
    path Path,
    index int,
    value Value,
) Operation

func ArrayRemove(
    path Path,
    index int,
) Operation
```

Example:

```go
out, err := editor.Apply(
    ctx,
    input,

    tomledit.ArrayInsert(
        tomledit.P("tools"),
        1,
        tomledit.String("golangci-lint"),
    ),
)
```

Input:

```toml
tools = ["go", "goreleaser"]
```

Output:

```toml
tools = ["go", "golangci-lint", "goreleaser"]
```

Formatting of the existing array SHOULD be preserved according to `toml_edit` behavior.

---

# 15. Array-of-Tables Operations

These operations are a primary feature of the library.

Required:

```go
func AOTAppend(
    path Path,
    fields ...Field,
) Operation

func AOTInsert(
    path Path,
    index int,
    fields ...Field,
) Operation

func AOTReplace(
    path Path,
    index int,
    fields ...Field,
) Operation

func AOTRemove(
    path Path,
    index int,
) Operation
```

`AOT` means:

```toml
[[array.of.tables]]
```

---

# 16. Core Example: Insert into `[[belt.actions]]`

This MUST be one of the primary integration tests and README examples.

Input:

```toml
[[belt.actions]]
name = "goblin-check"
action = "goblin:check"
stage = "validate"

[[belt.actions]]
name = "goblin-install"
action = "goblin:install"
stage = "generate"

[[belt.actions]]
name = "go-build"
action = "golang:build"
stage = "build"

[[belt.actions]]
name = "go-test"
action = "golang:test"
stage = "test"
```

Go:

```go
output, err := editor.Apply(
    ctx,
    input,

    tomledit.AOTInsert(
        tomledit.P("belt", "actions"),
        2,

        tomledit.F(
            "name",
            tomledit.String("go-tidy"),
        ),

        tomledit.F(
            "action",
            tomledit.String("golang:tidy"),
        ),

        tomledit.F(
            "stage",
            tomledit.String("build"),
        ),
    ),
)
```

Expected output:

```toml
[[belt.actions]]
name = "goblin-check"
action = "goblin:check"
stage = "validate"

[[belt.actions]]
name = "goblin-install"
action = "goblin:install"
stage = "generate"

[[belt.actions]]
name = "go-tidy"
action = "golang:tidy"
stage = "build"

[[belt.actions]]
name = "go-build"
action = "golang:build"
stage = "build"

[[belt.actions]]
name = "go-test"
action = "golang:test"
stage = "test"
```

Rust implementation eventually maps to:

```rust
let actions = item
    .as_array_of_tables_mut()
    .ok_or(...)?;

actions.insert(index, table);
```

This exact use case is one of the reasons for using `toml_edit`.

---

# 17. Core Example: Append Array-of-Tables Entry

Go:

```go
output, err := editor.Apply(
    ctx,
    input,

    tomledit.AOTAppend(
        tomledit.P("belt", "actions"),

        tomledit.F(
            "name",
            tomledit.String("go-tidy"),
        ),

        tomledit.F(
            "action",
            tomledit.String("golang:tidy"),
        ),

        tomledit.F(
            "stage",
            tomledit.String("build"),
        ),
    ),
)
```

Equivalent Rust operation:

```rust
actions.push(table);
```

---

# 18. Core Example: Modify Existing AOT Element

Support modifying nested content via normal paths including array indices.

The public representation should avoid encoding indexes as strings.

One possible extension to `Path` is:

```go
type PathElement struct {
    // internal
}

func Key(v string) PathElement
func Index(v int) PathElement
```

which enables:

```go
tomledit.Set(
    tomledit.Path(
        tomledit.Key("belt"),
        tomledit.Key("actions"),
        tomledit.Index(2),
        tomledit.Key("stage"),
    ),
    tomledit.String("validate"),
)
```

This is preferred long term over:

```go
[]string{"belt", "actions", "2", "stage"}
```

because numeric TOML keys are legal.

For MVP, this typed path model SHOULD be implemented from the beginning if practical.

Recommended final API:

```go
tomledit.P(
    tomledit.K("belt"),
    tomledit.K("actions"),
    tomledit.I(2),
    tomledit.K("stage"),
)
```

---

# 19. Core Example: Raw Formatting

Input:

```toml
channel = "stable"
```

Go:

```go
output, err := editor.Apply(
    ctx,
    input,

    tomledit.Set(
        tomledit.P("channel"),
        tomledit.Raw("'nightly'"),
    ),
)
```

Output:

```toml
channel = 'nightly'
```

This demonstrates that the caller can intentionally control TOML representation where necessary.

---

# 20. Core Example: Batch Modification

Batching SHOULD be the preferred API.

Input:

```toml
version = "1"

plugins = ["git"]

[[belt.actions]]
name = "build"
action = "golang:build"
stage = "build"
```

Go:

```go
output, err := editor.Apply(
    ctx,
    input,

    tomledit.Set(
        tomledit.P("version"),
        tomledit.String("2"),
    ),

    tomledit.ArrayAppend(
        tomledit.P("plugins"),
        tomledit.String("golang"),
    ),

    tomledit.AOTInsert(
        tomledit.P("belt", "actions"),
        0,

        tomledit.F(
            "name",
            tomledit.String("tidy"),
        ),

        tomledit.F(
            "action",
            tomledit.String("golang:tidy"),
        ),

        tomledit.F(
            "stage",
            tomledit.String("build"),
        ),
    ),
)
```

All operations execute against a single parsed `DocumentMut`.

If any operation fails, `Apply` returns an error.

---

# 21. Rust Operation Model

Rust protocol:

```rust
#[derive(Deserialize)]
#[serde(tag = "op", rename_all = "snake_case")]
enum Operation {
    Set {
        path: Vec<PathElement>,
        value: TomlValue,
    },

    Delete {
        path: Vec<PathElement>,
    },

    ArrayAppend {
        path: Vec<PathElement>,
        value: TomlValue,
    },

    ArrayInsert {
        path: Vec<PathElement>,
        index: usize,
        value: TomlValue,
    },

    ArrayReplace {
        path: Vec<PathElement>,
        index: usize,
        value: TomlValue,
    },

    ArrayRemove {
        path: Vec<PathElement>,
        index: usize,
    },

    AotAppend {
        path: Vec<PathElement>,
        fields: Vec<Field>,
    },

    AotInsert {
        path: Vec<PathElement>,
        index: usize,
        fields: Vec<Field>,
    },

    AotReplace {
        path: Vec<PathElement>,
        index: usize,
        fields: Vec<Field>,
    },

    AotRemove {
        path: Vec<PathElement>,
        index: usize,
    },
}
```

---

# 22. Typed Path Protocol

Prefer:

```rust
#[derive(Deserialize)]
#[serde(tag = "type", content = "value")]
enum PathElement {
    #[serde(rename = "key")]
    Key(String),

    #[serde(rename = "index")]
    Index(usize),
}
```

JSON:

```json
[
  {
    "type": "key",
    "value": "belt"
  },
  {
    "type": "key",
    "value": "actions"
  },
  {
    "type": "index",
    "value": 2
  },
  {
    "type": "key",
    "value": "stage"
  }
]
```

This avoids ambiguity between:

```text
array index 2
```

and:

```toml
"2" = "some TOML key"
```

---

# 23. Error Model

Public Go errors SHOULD expose typed error codes.

Example:

```go
type ErrorCode string

const (
    ErrInvalidTOML      ErrorCode = "invalid_toml"
    ErrPathNotFound     ErrorCode = "path_not_found"
    ErrTypeMismatch     ErrorCode = "type_mismatch"
    ErrIndexOutOfRange  ErrorCode = "index_out_of_range"
    ErrInvalidRawValue  ErrorCode = "invalid_raw_value"
    ErrProtocol         ErrorCode = "protocol_error"
    ErrRuntime          ErrorCode = "wasm_runtime_error"
)
```

Error:

```go
type Error struct {
    Code    ErrorCode
    Message string
    Path    Path
}
```

Allow:

```go
var editErr *tomledit.Error

if errors.As(err, &editErr) {
    switch editErr.Code {
    case tomledit.ErrPathNotFound:
        // ...
    }
}
```

Do not require users to parse error strings.

---

# 24. Embedded Runtime

Use:

```go
//go:embed assets/tomledit.wasm
var moduleBytes []byte
```

The Go runtime layer should:

1. create a wazero runtime;
2. compile the embedded module;
3. retain the compiled module;
4. invoke it through the internal engine;
5. clean up runtime resources in `Close`.

The raw WASM module MUST NOT be exposed publicly.

---

# 25. Formatting Guarantees

Primary contract:

> Existing TOML syntax not involved in an edit should be preserved as faithfully as `toml_edit` permits.

The library SHOULD preserve:

- comments;
- blank lines;
- indentation;
- spacing;
- existing quote styles;
- relative ordering;
- formatting of unrelated tables;
- formatting of unrelated arrays;
- inline comments.

Example test fixture:

```toml
# project configuration

name    =  'demo'   # important

[foo] # foo config
enabled=true


[[belt.actions]]
name   = "one"
action = "one:run"
stage  = "build"

# keep me
[[belt.actions]]
name = 'two'
action = "two:run"
stage = "test"

[unrelated]
value={foo="bar"}
```

After inserting another action, tests MUST verify that unrelated source fragments remain unchanged.

Do not claim byte-for-byte preservation beyond what `toml_edit` itself supports.

---

# 26. Testing Strategy

## Rust unit tests

Directly test:

- TOML parsing;
- path resolution;
- value conversion;
- scalar Set;
- Delete;
- array operations;
- AOT operations;
- raw value parsing;
- error conversion.

Rust tests SHOULD invoke the internal operation engine directly without going through WASM.

---

## Go integration tests

Go tests MUST execute the actual embedded WASM through wazero.

Required tests:

```text
TestSet
TestDelete

TestArrayAppend
TestArrayInsert
TestArrayReplace
TestArrayRemove

TestAOTAppend
TestAOTInsert
TestAOTReplace
TestAOTRemove

TestRawValue

TestBatch

TestInvalidTOML
TestPathNotFound
TestTypeMismatch
TestIndexOutOfRange

TestFormattingPreserved

TestConcurrentApply
```

The AOT insertion test MUST include the `belt.actions` example.

---

# 27. Concurrency

Public `Editor` MUST be safe to call concurrently:

```go
go editor.Apply(...)
go editor.Apply(...)
go editor.Apply(...)
```

A single WASM instance MUST NOT be used concurrently unless synchronization is provided.

Possible implementation:

```text
wazero Runtime
     │
     ▼
CompiledModule
     │
     ├── instantiate per Apply
     ├── execute
     └── close instance
```

This is attractive because parsing TOML dominates neither runtime nor build performance for normal configuration files and it avoids shared linear-memory synchronization.

If benchmarks later show significant overhead, an instance pool may be introduced without changing the public API.

---

# 28. Performance

This library is optimized for correctness and source preservation rather than high-throughput TOML processing.

Expected workloads:

```text
1 KB
10 KB
100 KB
1 MB
```

configuration files.

Add benchmarks for:

```text
10 KB / one operation
10 KB / ten operations

100 KB / one operation
100 KB / ten operations

1 MB / one operation
```

Batching multiple operations in one `Apply` call SHOULD be significantly preferred to repeatedly reparsing the same document.

---

# 29. Security / Isolation

The embedded WASM module:

- MUST NOT receive filesystem access;
- MUST NOT receive environment access;
- MUST NOT receive networking;
- MUST NOT use WASI unless a future feature explicitly requires it.

The module operates only on bytes provided by Go.

This gives the architecture:

```text
input bytes
     │
     ▼
isolated WASM memory
     │
     ▼
output bytes
```

---

# 30. MVP Scope

MVP MUST include:

### Runtime

- embedded WASM;
- wazero runner;
- single executable compatibility;
- `CGO_ENABLED=0`.

### Values

- string;
- integer;
- float;
- boolean;
- array;
- raw TOML.

### Operations

- Set;
- Delete;

- ArrayAppend;
- ArrayInsert;
- ArrayReplace;
- ArrayRemove;

- AOTAppend;
- AOTInsert;
- AOTReplace;
- AOTRemove.

### Other

- ordered table fields;
- typed paths;
- batch operations;
- structured errors;
- formatting-preservation tests;
- concurrent Go usage.

---

# 31. Explicit Non-Goals for MVP

Do not initially implement:

- persistent editable document handles;
- Go wrappers around Rust objects;
- callbacks from Rust to Go;
- WASI;
- filesystem access from Rust;
- schema validation;
- TOML diff generation;
- comment editing API;
- arbitrary formatting/decor manipulation;
- streaming TOML processing;
- CGO;
- dynamic libraries.

These may be considered later.

---

# 32. Future API Candidates

Potential future additions:

```go
Get(...)
Exists(...)

Rename(...)

TableInsert(...)
TableDelete(...)

InlineTableSet(...)
InlineTableDelete(...)

CommentBefore(...)
CommentAfter(...)

SetPrefix(...)
SetSuffix(...)

Format(...)
```

Also potentially:

```go
doc, err := tomledit.Parse(input)

doc.Apply(...)
doc.Bytes()
```

but a stateful document API should only be introduced if there is a concrete performance or ergonomics need.

The initial API should remain operation-based.

---

# 33. README Minimal Example

The README SHOULD lead with the main use case:

```go
package main

import (
    "context"
    "fmt"

    "github.com/<org>/tomledit"
)

func main() {
    ctx := context.Background()

    editor, err := tomledit.New(ctx)
    if err != nil {
        panic(err)
    }
    defer editor.Close(ctx)

    input := []byte(`
[[belt.actions]]
name = "goblin-check"
action = "goblin:check"
stage = "validate"

[[belt.actions]]
name = "goblin-install"
action = "goblin:install"
stage = "generate"

[[belt.actions]]
name = "go-build"
action = "golang:build"
stage = "build"
`)

    output, err := editor.Apply(
        ctx,
        input,

        tomledit.AOTInsert(
            tomledit.P(
                tomledit.K("belt"),
                tomledit.K("actions"),
            ),
            2,

            tomledit.F(
                "name",
                tomledit.String("go-tidy"),
            ),
            tomledit.F(
                "action",
                tomledit.String("golang:tidy"),
            ),
            tomledit.F(
                "stage",
                tomledit.String("build"),
            ),
        ),
    )
    if err != nil {
        panic(err)
    }

    fmt.Print(string(output))
}
```

No part of this example should mention:

```text
Rust
Cargo
WASM memory
JSON protocol
toml_edit
```

Those are implementation details.

---

# 34. Build and Distribution Acceptance Criteria

The project is considered successfully implemented when all of the following work.

A clean Go consumer with no Rust installed can run:

```bash
go get github.com/<org>/tomledit
```

and:

```bash
CGO_ENABLED=0 go build
```

successfully.

Result:

```text
single executable
```

There MUST NOT be runtime dependencies on:

```text
tomledit.wasm
cargo
rustc
libc via CGO
shared Rust libraries
external subprocesses
```

The WASM module MUST be part of the Go executable via `go:embed`.

Contributor workflow MUST support:

```bash
mise install
mise run wasm:build
mise run test
mise run build
```

The `belt.actions` insertion example MUST preserve existing comments and formatting while inserting:

```toml
[[belt.actions]]
name = "go-tidy"
action = "golang:tidy"
stage = "build"
```

at index `2`.

---

# 35. Design Principle

The main boundary of the project should remain:

```text
Go decides WHAT to change.

Rust/toml_edit decides HOW to safely modify TOML source.
```

The Go API is the stable product.

The JSON/WASM protocol and Rust implementation are private implementation details.

This allows:

- changing `toml_edit` versions;
- changing the Rust shim structure;
- changing memory management;
- changing serialization protocol;
- replacing wazero internals;

without breaking users of the Go library.

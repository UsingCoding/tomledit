---
name: tomledit
description: Use github.com/usingcoding/tomledit to atomically edit TOML while preserving unrelated source formatting where supported.
---

# tomledit

Use `github.com/usingcoding/tomledit` for format-preserving TOML mutations. The public API is pure Go and embeds its editor; consumers do not need CGO, Rust, or a runtime asset file.

## Lifecycle

Create one editor, reuse it across independent concurrent calls, and close it when finished:

```go
editor, err := tomledit.New(ctx)
if err != nil {
    return err
}
defer editor.Close(ctx)

output, err := editor.Apply(ctx, input, operations...)
if err != nil {
    return err
}
```

`Apply` parses once, applies every supplied operation in order, and returns output only when the full batch succeeds. On failure, it returns no partially edited document.

## Paths

Paths are typed. Use `K` for TOML keys and `I` for array or array-of-tables indices; never encode an index as a string.

```go
path := tomledit.P(
    tomledit.K("foo"),
    tomledit.K("items"),
    tomledit.I(2),
    tomledit.K("group"),
)
```

This keeps a literal numeric TOML key distinct from an array index.

## Values and operations

Create values with `String`, `Integer`, `Float`, `Boolean`, `Array`, or `Raw`. `Raw` is parsed as one TOML value before insertion; invalid or multi-item input returns `ErrInvalidRawValue`.

```go
output, err := editor.Apply(ctx, input,
    tomledit.Set(
        tomledit.P(tomledit.K("version")),
        tomledit.String("2"),
    ),
    tomledit.ArrayAppend(
        tomledit.P(tomledit.K("tools")),
        tomledit.String("bar"),
    ),
)
```

Available mutations:

- `Set`, `Delete`
- `ArrayAppend`, `ArrayInsert`, `ArrayReplace`, `ArrayRemove`
- `AOTAppend`, `AOTInsert`, `AOTReplace`, `AOTRemove`

Array-of-tables fields are ordered. Build them with `F`; duplicate field keys are a protocol error.

```go
tomledit.AOTInsert(
    tomledit.P(tomledit.K("foo"), tomledit.K("items")),
    2,
    tomledit.F("name", tomledit.String("qux")),
    tomledit.F("value", tomledit.String("qux")),
    tomledit.F("group", tomledit.String("foo")),
)
```

## Errors

Every editor failure is a `*tomledit.Error`. Branch on `Code`, not error text.

```go
var editErr *tomledit.Error
if errors.As(err, &editErr) {
    switch editErr.Code {
    case tomledit.ErrPathNotFound:
        // Missing intermediate key or terminal key.
    case tomledit.ErrTypeMismatch:
        // A path target has the wrong TOML container type.
    case tomledit.ErrIndexOutOfRange:
        // Negative or invalid array/AOT index.
    }
}
```

Other codes are `ErrInvalidTOML`, `ErrInvalidRawValue`, `ErrProtocol`, and `ErrRuntime`.

## Preservation expectations

Do not promise byte-for-byte preservation. Unrelated comments, blank lines, spacing, quote styles, inline comments, and relative ordering are retained as faithfully as the TOML editor permits. Dotted-key ordering and a missing trailing newline can change.

## Development workflow

For development on a new machine:

```sh
mise install
mise run bootstrap
mise run wasm:build
mise run test
```

`bootstrap` installs the local Rust `wasm32-unknown-unknown` target used by linting. `wasm:build` requires Docker and runs in a pinned Linux AMD64 Rust image, so its generated artifact is reproducible on developer machines and GitHub Actions. Normal Go consumers only need `CGO_ENABLED=0 go build ./...`.

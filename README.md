# tomledit

Edit TOML while preserving existing source formatting where the TOML editor supports it.

```go
editor, err := tomledit.New(context.Background())
if err != nil {
    return err
}
defer editor.Close(context.Background())

output, err := editor.Apply(ctx, input,
    tomledit.AOTInsert(
        tomledit.P(tomledit.K("foo"), tomledit.K("items")),
        2,
        tomledit.F("name", tomledit.String("qux")),
        tomledit.F("value", tomledit.String("qux")),
        tomledit.F("group", tomledit.String("foo")),
    ),
)
if err != nil {
    return err
}
```

The entry is inserted between the existing `bar` and `baz` entries in `[[foo.items]]`.

## Lifecycle

Create one `Editor` with `New`, share it across concurrent `Apply` calls, then call `Close` when finished. Each `Apply` receives input bytes plus one or more operations and returns output only when the complete batch succeeds. A failed batch returns no partial document.

## API

Paths are typed: `K("name")` selects a TOML key and `I(2)` selects an array or array-of-tables entry. This keeps a numeric key distinct from an index.

- `Set`, `Delete`
- `ArrayAppend`, `ArrayInsert`, `ArrayReplace`, `ArrayRemove`
- `AOTAppend`, `AOTInsert`, `AOTReplace`, `AOTRemove`
- `String`, `Integer`, `Float`, `Boolean`, `Array`, and validated `Raw` values
- Ordered array-of-tables fields via `F`

`Raw` parses its supplied source as one TOML value before mutation. For example, `Raw("'nightly'")` preserves single quotes; malformed or multi-item raw input is rejected.

## Errors

All failures are `*tomledit.Error`. Branch on `Error.Code` with `errors.As`; never parse error text. Codes are `ErrInvalidTOML`, `ErrPathNotFound`, `ErrTypeMismatch`, `ErrIndexOutOfRange`, `ErrInvalidRawValue`, `ErrProtocol`, and `ErrRuntime`.

## Preservation limits

Unrelated comments, blank lines, spacing, quote styles, inline comments, and relative ordering are retained as faithfully as the underlying TOML editor permits. Upstream behavior can reorder dotted keys and can add a trailing newline when the original source omitted one.

## Development

```sh
mise install
mise run bootstrap
mise run wasm:build
mise run test
```

`wasm:build` requires Docker and always builds in a pinned Linux AMD64 Rust image. This makes the committed artifact reproducible on developer machines and GitHub Actions.

The generated editor artifact is committed so consumers build with `CGO_ENABLED=0 go build ./...` and do not need a native toolchain at build or runtime.

## License

[MIT](LICENSE)

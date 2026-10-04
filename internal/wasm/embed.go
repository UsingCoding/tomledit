// Package wasm embeds the TOML editing WebAssembly module.
package wasm

import _ "embed"

// Bytes is the embedded tomledit WebAssembly module.
//
//go:embed assets/tomledit.wasm
var Bytes []byte

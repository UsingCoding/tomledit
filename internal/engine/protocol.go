// Package engine owns the private JSON and WebAssembly boundaries.
package engine

const ProtocolVersion = 1

type Request struct {
	Version    int         `json:"version"`
	TOML       string      `json:"toml"`
	Operations []Operation `json:"operations"`
}

type PathElement struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type Value struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type Field struct {
	Key   string `json:"key"`
	Value Value  `json:"value"`
}

type Operation struct {
	Op     string        `json:"op"`
	Path   []PathElement `json:"path"`
	Value  *Value        `json:"value,omitempty"`
	Index  *int          `json:"index,omitempty"`
	Fields []Field       `json:"fields,omitempty"`
}

type Response struct {
	Version int            `json:"version"`
	OK      bool           `json:"ok"`
	TOML    *string        `json:"toml,omitempty"`
	Error   *ProtocolError `json:"error,omitempty"`
}

type ProtocolError struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Path    []PathElement `json:"path"`
}

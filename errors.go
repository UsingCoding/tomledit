package tomledit

import "fmt"

// ErrorCode identifies a stable TOML-editing failure category.
type ErrorCode string

const (
	ErrInvalidTOML     ErrorCode = "invalid_toml"
	ErrPathNotFound    ErrorCode = "path_not_found"
	ErrTypeMismatch    ErrorCode = "type_mismatch"
	ErrIndexOutOfRange ErrorCode = "index_out_of_range"
	ErrInvalidRawValue ErrorCode = "invalid_raw_value"
	ErrProtocol        ErrorCode = "protocol_error"
	ErrRuntime         ErrorCode = "wasm_runtime_error"
)

// Error is returned for all editor and protocol failures.
type Error struct {
	Code    ErrorCode
	Message string
	Path    Path
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Message == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func protocolError(message string) *Error {
	return &Error{Code: ErrProtocol, Message: message}
}

func runtimeError(message string) *Error {
	return &Error{Code: ErrRuntime, Message: message}
}

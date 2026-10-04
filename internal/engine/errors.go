package engine

import "fmt"

// Error describes a private ABI or response-validation failure.
type Error struct {
	Protocol bool
	Message  string
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Message
}

func runtimeError(format string, args ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, args...)}
}

func protocolError(format string, args ...any) *Error {
	return &Error{Protocol: true, Message: fmt.Sprintf(format, args...)}
}

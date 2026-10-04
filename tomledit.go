package tomledit

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/usingcoding/tomledit/internal/engine"
	"github.com/usingcoding/tomledit/internal/wasm"
)

// Editor applies TOML mutations through the embedded WebAssembly editor.
type Editor struct {
	mu       sync.RWMutex
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
	closed   bool
}

// New creates an Editor that owns one compiled embedded WebAssembly module.
func New(ctx context.Context) (*Editor, error) {
	runtime := wazero.NewRuntime(ctx)
	compiled, err := runtime.CompileModule(ctx, wasm.Bytes)
	if err != nil {
		_ = runtime.Close(ctx)
		return nil, runtimeError(fmt.Sprintf("compile embedded WebAssembly module: %v", err))
	}
	return &Editor{runtime: runtime, compiled: compiled}, nil
}

// Apply applies operations atomically to input.
func (e *Editor) Apply(ctx context.Context, input []byte, operations ...Operation) ([]byte, error) {
	if e == nil {
		return nil, runtimeError("editor is nil")
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed {
		return nil, runtimeError("editor is closed")
	}

	request, err := makeRequest(input, operations)
	if err != nil {
		return nil, err
	}
	response, err := engine.Apply(ctx, e.runtime, e.compiled, request)
	if err != nil {
		return nil, fromEngineError(err)
	}
	if response.OK {
		return []byte(*response.TOML), nil
	}
	return nil, fromProtocolError(*response.Error)
}

// Close releases the compiled module and WebAssembly runtime. It is idempotent.
func (e *Editor) Close(ctx context.Context) error {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	e.closed = true
	if err := e.compiled.Close(ctx); err != nil {
		_ = e.runtime.Close(ctx)
		return runtimeError(fmt.Sprintf("close compiled WebAssembly module: %v", err))
	}
	if err := e.runtime.Close(ctx); err != nil {
		return runtimeError(fmt.Sprintf("close WebAssembly runtime: %v", err))
	}
	return nil
}

func makeRequest(input []byte, operations []Operation) (engine.Request, error) {
	converted := make([]engine.Operation, len(operations))
	for i, operation := range operations {
		convertedOperation, err := convertOperation(operation)
		if err != nil {
			return engine.Request{}, err
		}
		converted[i] = convertedOperation
	}
	return engine.Request{Version: engine.ProtocolVersion, TOML: string(input), Operations: converted}, nil
}

func convertOperation(operation Operation) (engine.Operation, error) {
	path, err := convertPath(operation.path)
	if err != nil {
		return engine.Operation{}, err
	}
	result := engine.Operation{Path: path}
	withValue := func(op string) (engine.Operation, error) {
		value, valueErr := convertValue(operation.value)
		if valueErr != nil {
			return engine.Operation{}, valueErr
		}
		result.Op, result.Value = op, &value
		return result, nil
	}
	withIndex := func(op string, needsValue bool) (engine.Operation, error) {
		index := operation.index
		result.Op, result.Index = op, &index
		if !needsValue {
			return result, nil
		}
		value, valueErr := convertValue(operation.value)
		if valueErr != nil {
			return engine.Operation{}, valueErr
		}
		result.Value = &value
		return result, nil
	}
	withFields := func(op string, indexed bool) (engine.Operation, error) {
		fields, fieldsErr := convertFields(operation.fields)
		if fieldsErr != nil {
			return engine.Operation{}, fieldsErr
		}
		result.Op, result.Fields = op, fields
		if indexed {
			index := operation.index
			result.Index = &index
		}
		return result, nil
	}

	switch operation.kind {
	case opSet:
		return withValue("set")
	case opDelete:
		result.Op = "delete"
		return result, nil
	case opArrayAppend:
		return withValue("array_append")
	case opArrayInsert:
		return withIndex("array_insert", true)
	case opArrayReplace:
		return withIndex("array_replace", true)
	case opArrayRemove:
		return withIndex("array_remove", false)
	case opAOTAppend:
		return withFields("aot_append", false)
	case opAOTInsert:
		return withFields("aot_insert", true)
	case opAOTReplace:
		return withFields("aot_replace", true)
	case opAOTRemove:
		return withIndex("aot_remove", false)
	default:
		return engine.Operation{}, protocolError("unknown operation")
	}
}

func convertPath(path Path) ([]engine.PathElement, error) {
	if len(path) == 0 {
		return nil, protocolError("path is empty")
	}
	converted := make([]engine.PathElement, len(path))
	for i, element := range path {
		switch element.kind {
		case pathKey:
			converted[i] = engine.PathElement{Type: "key", Value: element.key}
		case pathIndex:
			converted[i] = engine.PathElement{Type: "index", Value: element.index}
		default:
			return nil, protocolError("path contains an invalid element")
		}
	}
	return converted, nil
}

func convertValue(value Value) (engine.Value, error) {
	switch value.kind {
	case valueString:
		return engine.Value{Type: "string", Value: value.string}, nil
	case valueInteger:
		return engine.Value{Type: "integer", Value: value.integer}, nil
	case valueFloat:
		if math.IsNaN(value.float) || math.IsInf(value.float, 0) {
			return engine.Value{}, protocolError("float value must be finite")
		}
		return engine.Value{Type: "float", Value: value.float}, nil
	case valueBoolean:
		return engine.Value{Type: "boolean", Value: value.boolean}, nil
	case valueArray:
		values := make([]engine.Value, len(value.array))
		for i, element := range value.array {
			converted, err := convertValue(element)
			if err != nil {
				return engine.Value{}, err
			}
			values[i] = converted
		}
		return engine.Value{Type: "array", Value: values}, nil
	case valueRaw:
		return engine.Value{Type: "raw", Value: value.string}, nil
	default:
		return engine.Value{}, protocolError("unknown value")
	}
}

func convertFields(fields []Field) ([]engine.Field, error) {
	converted := make([]engine.Field, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for i, field := range fields {
		if _, exists := seen[field.Key]; exists {
			return nil, protocolError("array-of-tables fields contain duplicate key " + field.Key)
		}
		seen[field.Key] = struct{}{}
		value, err := convertValue(field.Value)
		if err != nil {
			return nil, err
		}
		converted[i] = engine.Field{Key: field.Key, Value: value}
	}
	return converted, nil
}

func fromEngineError(err error) *Error {
	if engineError, ok := err.(*engine.Error); ok && engineError.Protocol {
		return protocolError(engineError.Message)
	}
	return runtimeError(err.Error())
}

func fromProtocolError(protocol engine.ProtocolError) *Error {
	code := ErrorCode(protocol.Code)
	switch code {
	case ErrInvalidTOML, ErrPathNotFound, ErrTypeMismatch, ErrIndexOutOfRange, ErrInvalidRawValue, ErrProtocol, ErrRuntime:
	default:
		return protocolError("unknown protocol error code " + protocol.Code)
	}
	path, err := pathFromProtocol(protocol.Path)
	if err != nil {
		return err
	}
	return &Error{Code: code, Message: protocol.Message, Path: path}
}

func pathFromProtocol(path []engine.PathElement) (Path, *Error) {
	converted := make(Path, len(path))
	for i, element := range path {
		switch element.Type {
		case "key":
			value, ok := element.Value.(string)
			if !ok {
				return nil, protocolError("protocol key path value is invalid")
			}
			converted[i] = K(value)
		case "index":
			value, ok := element.Value.(float64)
			if !ok || math.Trunc(value) != value || float64(int(value)) != value {
				return nil, protocolError("protocol index path value is invalid")
			}
			converted[i] = I(int(value))
		default:
			return nil, protocolError("protocol path element type is invalid")
		}
	}
	return converted, nil
}

package engine

import (
	"context"
	"encoding/json"
	"math"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Apply instantiates a private WebAssembly module and executes one request.
func Apply(ctx context.Context, runtime wazero.Runtime, compiled wazero.CompiledModule, request Request) (response Response, err error) {
	input, marshalErr := json.Marshal(request)
	if marshalErr != nil {
		return Response{}, protocolError("marshal request: %v", marshalErr)
	}
	if uint64(len(input)) > math.MaxUint32 {
		return Response{}, protocolError("request exceeds WebAssembly ABI length limit")
	}

	module, instantiateErr := runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName(""))
	if instantiateErr != nil {
		return Response{}, runtimeError("instantiate WebAssembly module: %v", instantiateErr)
	}
	defer func() {
		if closeErr := module.Close(ctx); err == nil && closeErr != nil {
			err = runtimeError("close WebAssembly module: %v", closeErr)
		}
	}()

	memory := module.Memory()
	if memory == nil {
		return Response{}, runtimeError("WebAssembly module has no exported memory")
	}
	alloc := module.ExportedFunction("alloc")
	dealloc := module.ExportedFunction("dealloc")
	apply := module.ExportedFunction("apply")
	if alloc == nil || dealloc == nil || apply == nil {
		return Response{}, runtimeError("WebAssembly module is missing required ABI exports")
	}

	inputResults, callErr := alloc.Call(ctx, uint64(len(input)))
	if callErr != nil || len(inputResults) != 1 || inputResults[0] > math.MaxUint32 {
		return Response{}, abiCallError("allocate input", callErr, len(inputResults))
	}
	inputPointer := uint32(inputResults[0])
	inputAllocated := true
	defer func() {
		if !inputAllocated {
			return
		}
		if freeErr := deallocate(ctx, dealloc, inputPointer, uint32(len(input))); err == nil && freeErr != nil {
			err = freeErr
		}
	}()

	if !memory.Write(inputPointer, input) {
		return Response{}, runtimeError("write input to WebAssembly memory")
	}
	result, callErr := apply.Call(ctx, uint64(inputPointer), uint64(len(input)))
	if callErr != nil || len(result) != 1 {
		return Response{}, abiCallError("apply request", callErr, len(result))
	}
	outputPointer := uint32(result[0] >> 32)
	outputLength := uint32(result[0])
	output, ok := memory.Read(outputPointer, outputLength)
	if !ok {
		return Response{}, runtimeError("read output from WebAssembly memory")
	}
	outputCopy := append([]byte(nil), output...)
	if freeErr := deallocate(ctx, dealloc, outputPointer, outputLength); freeErr != nil {
		return Response{}, freeErr
	}

	if unmarshalErr := json.Unmarshal(outputCopy, &response); unmarshalErr != nil {
		return Response{}, protocolError("decode response: %v", unmarshalErr)
	}
	if response.Version != ProtocolVersion {
		return Response{}, protocolError("unsupported response version %d", response.Version)
	}
	if response.OK {
		if response.TOML == nil || response.Error != nil {
			return Response{}, protocolError("invalid success response envelope")
		}
		return response, nil
	}
	if response.TOML != nil || response.Error == nil {
		return Response{}, protocolError("invalid error response envelope")
	}
	return response, nil
}

func abiCallError(action string, callErr error, resultCount int) *Error {
	if callErr != nil {
		return runtimeError("%s: %v", action, callErr)
	}
	return runtimeError("%s returned %d results", action, resultCount)
}

func deallocate(ctx context.Context, dealloc api.Function, pointer, length uint32) *Error {
	results, err := dealloc.Call(ctx, uint64(pointer), uint64(length))
	if err != nil {
		return runtimeError("deallocate WebAssembly buffer: %v", err)
	}
	if len(results) != 0 {
		return runtimeError("deallocate WebAssembly buffer returned %d results", len(results))
	}
	return nil
}

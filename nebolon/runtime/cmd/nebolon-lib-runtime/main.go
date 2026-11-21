// Command nebolon-lib-runtime is the NEBOLON Library Runtime.
// It runs as a standalone child process of nebolon-runtime and communicates
// with it over JSON Lines on stdin/stdout.
//
// Each line received on stdin is a protocol.Request.
// Each line written to stdout is a protocol.Response.
// Diagnostic messages are written to stderr only.
//
// The process exits cleanly when stdin is closed (EOF).
//
// Usage:
//
//	nebolon-lib-runtime --block-library <path>
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"nebolon/runtime/internal/blocks"
	"nebolon/runtime/internal/protocol"
)

func main() {
	fs := flag.NewFlagSet("nebolon-lib-runtime", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	libraryRoot := fs.String("block-library", "", "Path to the block library directory")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "nebolon-lib-runtime: flag error: %v\n", err)
		os.Exit(1)
	}

	resolvedLibRoot := blocks.ResolveLibraryRoot(*libraryRoot)
	if err := serve(os.Stdin, os.Stdout, resolvedLibRoot); err != nil {
		fmt.Fprintf(os.Stderr, "nebolon-lib-runtime: fatal: %v\n", err)
		os.Exit(1)
	}
}

// serve reads JSON Lines from r, dispatches each request, and writes responses
// to w until r reaches EOF.
func serve(r *os.File, w *os.File, libraryRoot string) error {
	scanner := bufio.NewScanner(r)
	encoder := json.NewEncoder(w)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		resp := dispatch(line, libraryRoot)

		if err := encoder.Encode(resp); err != nil {
			// Writing to stdout failed — nothing useful we can do.
			return fmt.Errorf("failed to write response: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stdin read error: %w", err)
	}

	return nil
}

// dispatch parses a single request line and returns the appropriate response.
// All errors produce a protocol-level error response rather than crashing.
func dispatch(line []byte, libraryRoot string) protocol.Response {
	var req protocol.Request
	if err := json.Unmarshal(line, &req); err != nil {
		// No ID available; use an empty string.
		return errResponse("", "INVALID_REQUEST", fmt.Sprintf("failed to parse request: %v", err))
	}

	if req.Protocol != protocol.ProtocolName {
		return errResponse(req.ID, "WRONG_PROTOCOL",
			fmt.Sprintf("expected protocol %q, got %q", protocol.ProtocolName, req.Protocol))
	}
	if req.Version != protocol.Version {
		return errResponse(req.ID, "UNSUPPORTED_VERSION",
			fmt.Sprintf("expected version %q, got %q", protocol.Version, req.Version))
	}

	switch req.Method {
	case "execute":
		return handleExecute(req, libraryRoot)
	case "get-definition":
		return handleGetDefinition(req, libraryRoot)
	default:
		return errResponse(req.ID, "UNKNOWN_METHOD",
			fmt.Sprintf("unknown method %q", req.Method))
	}
}

// handleExecute handles an "execute" request.
// If the block implementation is available, it executes it and returns the result.
// If the implementation is missing or not yet implemented, it returns
// BLOCK_EXECUTION_NOT_IMPLEMENTED.
func handleExecute(req protocol.Request, libraryRoot string) protocol.Response {
	var params protocol.ExecuteParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errResponse(req.ID, "INVALID_PARAMS",
			fmt.Sprintf("failed to parse execute params: %v", err))
	}

	if params.BlockType == "" {
		return errResponse(req.ID, "INVALID_PARAMS", "blockType is required")
	}

	def, err := blocks.Load(libraryRoot, params.BlockType)
	if err != nil {
		return errResponse(req.ID, "BLOCK_NOT_FOUND",
			fmt.Sprintf("block %q not found in library: %v", params.BlockType, err))
	}

	switch def.Implementation.Language {
	case "go":
		return executeGoBlock(req, libraryRoot, def, params)
	default:
		return errResponse(req.ID, "BLOCK_EXECUTION_NOT_IMPLEMENTED",
			fmt.Sprintf("execution of block %q with language %q is not yet implemented", params.BlockType, def.Implementation.Language))
	}
}

// executeGoBlock executes a Go-based block either via its compiled binary or main.go.
func executeGoBlock(req protocol.Request, libraryRoot string, def blocks.BlockDefinition, params protocol.ExecuteParams) protocol.Response {
	execTarget, isSource, found := resolveGoExecutable(libraryRoot, params.BlockType, def)
	if !found {
		return errResponse(req.ID, "BLOCK_EXECUTION_NOT_IMPLEMENTED",
			fmt.Sprintf("execution of block %q is not yet implemented (no binary or main.go found)", params.BlockType))
	}

	var cmd *exec.Cmd
	if isSource {
		cmd = exec.Command("go", "run", ".")
		cmd.Dir = execTarget
	} else {
		cmd = exec.Command(execTarget)
	}

	inputPayload := map[string]any{
		"config": params.Config,
	}
	inputBytes, err := json.Marshal(inputPayload)
	if err != nil {
		return errResponse(req.ID, "INTERNAL_ERROR", fmt.Sprintf("failed to encode input: %v", err))
	}
	cmd.Stdin = bytes.NewReader(inputBytes)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		execResult := protocol.ExecuteResult{
			Status: "failed",
			Outputs: map[string]any{
				"success": false,
				"fail":    true,
				"error": map[string]any{
					"code":    "BLOCK_EXECUTION_FAILED",
					"message": errMsg,
				},
			},
		}
		resJSON, _ := json.Marshal(execResult)
		return protocol.Response{
			Protocol: protocol.ProtocolName,
			Version:  protocol.Version,
			ID:       req.ID,
			Result:   resJSON,
		}
	}

	var rawOut struct {
		Outputs map[string]any `json:"outputs"`
		Error   any            `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &rawOut); err != nil {
		var flatOut map[string]any
		if err2 := json.Unmarshal(stdout.Bytes(), &flatOut); err2 == nil {
			rawOut.Outputs = flatOut
		} else {
			return errResponse(req.ID, "BLOCK_MALFORMED_OUTPUT",
				fmt.Sprintf("block produced invalid output: %v", err))
		}
	}

	outputs := map[string]any{
		"success": true,
		"fail":    false,
	}
	for k, v := range rawOut.Outputs {
		outputs[k] = v
	}

	execResult := protocol.ExecuteResult{
		Status:  "succeeded",
		Outputs: outputs,
	}
	resJSON, err := json.Marshal(execResult)
	if err != nil {
		return errResponse(req.ID, "INTERNAL_ERROR", fmt.Sprintf("failed to marshal result: %v", err))
	}

	return protocol.Response{
		Protocol: protocol.ProtocolName,
		Version:  protocol.Version,
		ID:       req.ID,
		Result:   resJSON,
	}
}

// resolveGoExecutable checks for an executable binary or a main.go file for the block.
func resolveGoExecutable(libraryRoot, blockType string, def blocks.BlockDefinition) (target string, isSource bool, found bool) {
	candidates := []string{
		filepath.Join(libraryRoot, blockType),
	}
	if def.Implementation.Path != "" {
		candidates = append(candidates,
			def.Implementation.Path,
			filepath.Join(libraryRoot, "..", def.Implementation.Path),
		)
	}

	exeName := blockType
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}

	// 1. Look for a compiled binary.
	for _, dir := range candidates {
		bin := filepath.Join(dir, exeName)
		if fi, err := os.Stat(bin); err == nil && !fi.IsDir() && (fi.Mode()&0111 != 0 || runtime.GOOS == "windows") {
			return bin, false, true
		}
	}

	// 2. Look for main.go source file.
	for _, dir := range candidates {
		src := filepath.Join(dir, "main.go")
		if fi, err := os.Stat(src); err == nil && !fi.IsDir() {
			return dir, true, true
		}
	}

	return "", false, false
}

// handleGetDefinition handles a "get-definition" request.
// It loads the block definition from disk and returns it as raw JSON.
func handleGetDefinition(req protocol.Request, libraryRoot string) protocol.Response {
	var params protocol.GetDefinitionParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errResponse(req.ID, "INVALID_PARAMS",
			fmt.Sprintf("failed to parse get-definition params: %v", err))
	}

	if params.BlockType == "" {
		return errResponse(req.ID, "INVALID_PARAMS", "blockType is required")
	}

	def, err := blocks.Load(libraryRoot, params.BlockType)
	if err != nil {
		return errResponse(req.ID, "BLOCK_NOT_FOUND",
			fmt.Sprintf("block %q not found: %v", params.BlockType, err))
	}

	defJSON, err := json.Marshal(def)
	if err != nil {
		return errResponse(req.ID, "INTERNAL_ERROR",
			fmt.Sprintf("failed to serialise block definition: %v", err))
	}

	result := protocol.GetDefinitionResult{Definition: defJSON}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return errResponse(req.ID, "INTERNAL_ERROR",
			fmt.Sprintf("failed to serialise result: %v", err))
	}

	return protocol.Response{
		Protocol: protocol.ProtocolName,
		Version:  protocol.Version,
		ID:       req.ID,
		Result:   resultJSON,
	}
}

// errResponse constructs a protocol-level error response.
func errResponse(id, code, message string) protocol.Response {
	return protocol.Response{
		Protocol: protocol.ProtocolName,
		Version:  protocol.Version,
		ID:       id,
		Error: &protocol.ResponseError{
			Code:    code,
			Message: message,
		},
	}
}

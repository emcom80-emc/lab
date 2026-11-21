package execution

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync/atomic"

	"nebolon/runtime/internal/blocks"
	"nebolon/runtime/internal/flow"
	"nebolon/runtime/internal/protocol"
)

// LibraryRuntimeInvoker implements Invoker by communicating with a running
// nebolon-lib-runtime child process over JSON Lines on its stdin/stdout.
//
// A single long-lived process is started when the invoker is created.
// Use Close to shut it down when done.
type LibraryRuntimeInvoker struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
	nextID atomic.Uint64
}

// NewLibraryRuntimeInvoker starts the nebolon-lib-runtime binary at the given
// path and returns an invoker backed by the running process.
// The caller must call Close when finished.
func NewLibraryRuntimeInvoker(binaryPath, libraryRoot string) (*LibraryRuntimeInvoker, error) {
	cmd := exec.Command(binaryPath, "--block-library", libraryRoot)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("lib-runtime: failed to open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("lib-runtime: failed to open stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("lib-runtime: failed to start process %q: %w", binaryPath, err)
	}

	return &LibraryRuntimeInvoker{
		cmd:    cmd,
		stdin:  stdin,
		reader: bufio.NewReader(stdout),
	}, nil
}

// Close shuts down the lib-runtime process gracefully by closing its stdin,
// then waiting for it to exit.
func (l *LibraryRuntimeInvoker) Close() error {
	if err := l.stdin.Close(); err != nil {
		return fmt.Errorf("lib-runtime: failed to close stdin: %w", err)
	}
	return l.cmd.Wait()
}

// Invoke sends an "execute" request to the lib-runtime process and translates
// the response into an execution.Result.
func (l *LibraryRuntimeInvoker) Invoke(def blocks.BlockDefinition, node flow.Node) Result {
	params := protocol.ExecuteParams{
		BlockType: def.ID,
		Config:    node.Config,
	}

	resp, err := l.call("execute", params)
	if err != nil {
		return errorResult(fmt.Sprintf("lib-runtime communication error: %v", err))
	}

	// A protocol-level error (e.g. BLOCK_EXECUTION_NOT_IMPLEMENTED) is surfaced
	// directly as a failed execution result so the Runtime can record it normally.
	if resp.Error != nil {
		return Result{
			Status: "failed",
			Outputs: map[string]any{
				"success": false,
				"fail":    true,
				"error": map[string]any{
					"code":    resp.Error.Code,
					"message": resp.Error.Message,
				},
			},
		}
	}

	// Decode the successful execution result.
	var execResult protocol.ExecuteResult
	if err := json.Unmarshal(resp.Result, &execResult); err != nil {
		return errorResult(fmt.Sprintf("lib-runtime: malformed execute result: %v", err))
	}

	return Result{
		Status:  execResult.Status,
		Outputs: execResult.Outputs,
	}
}

// call sends a JSON Lines request and reads one JSON Lines response.
func (l *LibraryRuntimeInvoker) call(method string, params any) (protocol.Response, error) {
	id := fmt.Sprintf("req-%d", l.nextID.Add(1))

	rawParams, err := json.Marshal(params)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("failed to marshal params: %w", err)
	}

	req := protocol.Request{
		Protocol: protocol.ProtocolName,
		Version:  protocol.Version,
		ID:       id,
		Method:   method,
		Params:   rawParams,
	}

	line, err := json.Marshal(req)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("failed to marshal request: %w", err)
	}
	line = append(line, '\n')

	if _, err := l.stdin.Write(line); err != nil {
		return protocol.Response{}, fmt.Errorf("failed to write request: %w", err)
	}

	respLine, err := l.reader.ReadString('\n')
	if err != nil {
		return protocol.Response{}, fmt.Errorf("failed to read response: %w", err)
	}

	var resp protocol.Response
	if err := json.Unmarshal([]byte(respLine), &resp); err != nil {
		return protocol.Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.ID != id {
		return protocol.Response{}, fmt.Errorf("response ID mismatch: sent %q, got %q", id, resp.ID)
	}

	return resp, nil
}

// CallForTest sends a raw protocol.Request to the lib-runtime process and
// returns the protocol.Response. It is intended for use in tests that need to
// exercise the protocol layer directly without going through Invoke.
func (l *LibraryRuntimeInvoker) CallForTest(req protocol.Request) (protocol.Response, error) {
	line, err := json.Marshal(req)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("CallForTest: failed to marshal request: %w", err)
	}
	line = append(line, '\n')

	if _, err := l.stdin.Write(line); err != nil {
		return protocol.Response{}, fmt.Errorf("CallForTest: failed to write request: %w", err)
	}

	respLine, err := l.reader.ReadString('\n')
	if err != nil {
		return protocol.Response{}, fmt.Errorf("CallForTest: failed to read response: %w", err)
	}

	var resp protocol.Response
	if err := json.Unmarshal([]byte(respLine), &resp); err != nil {
		return protocol.Response{}, fmt.Errorf("CallForTest: failed to parse response: %w", err)
	}

	return resp, nil
}

// errorResult is a convenience helper that wraps a communication-level failure
// into a standardised execution result without a known error code.
func errorResult(message string) Result {
	return Result{
		Status: "failed",
		Outputs: map[string]any{
			"success": false,
			"fail":    true,
			"error": map[string]any{
				"code":    "LIB_RUNTIME_ERROR",
				"message": message,
			},
		},
	}
}

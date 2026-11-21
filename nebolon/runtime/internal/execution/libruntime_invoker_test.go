// Package execution_test contains integration tests for LibraryRuntimeInvoker.
// These tests build the real nebolon-lib-runtime binary and communicate with
// it through the production code path: JSON Lines over stdin/stdout.
package execution_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"nebolon/runtime/internal/execution"
	"nebolon/runtime/internal/protocol"
)

// --- Helpers -----------------------------------------------------------------

// buildLibRuntime compiles nebolon-lib-runtime into a temp directory and
// returns the path to the resulting binary.
func buildLibRuntime(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "nebolon-lib-runtime")
	cmd := exec.Command("go", "build", "-o", bin, "nebolon/runtime/cmd/nebolon-lib-runtime")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build nebolon-lib-runtime: %v", err)
	}
	return bin
}

// newHelloWorldLibrary writes a minimal hello-world block definition into a
// temp directory and returns the library root path.
func newHelloWorldLibrary(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "hello-world", "hello-world.block"), `{
		"id": "hello-world",
		"name": "Hello World",
		"description": "A simple hello world block.",
		"category": "examples",
		"outputs": {
			"message": { "type": "string" }
		},
		"implementation": {
			"type": "process",
			"language": "go",
			"path": "block-library/hello-world"
		}
	}`)
	return root
}

// startInvoker builds the binary, starts it, and returns the invoker.
// t.Cleanup is used to Close the invoker after the test.
func startInvoker(t *testing.T, libraryRoot string) *execution.LibraryRuntimeInvoker {
	t.Helper()
	bin := buildLibRuntime(t)
	inv, err := execution.NewLibraryRuntimeInvoker(bin, libraryRoot)
	if err != nil {
		t.Fatalf("failed to start lib-runtime: %v", err)
	}
	t.Cleanup(func() {
		if err := inv.Close(); err != nil {
			t.Logf("lib-runtime close: %v", err)
		}
	})
	return inv
}

// --- Protocol-level tests (direct JSON Lines) --------------------------------

// TestLibRuntime_ExecuteReturnsBlockExecutionNotImplemented verifies that the
// lib-runtime process starts, accepts an execute request, identifies the block,
// and returns a BLOCK_EXECUTION_NOT_IMPLEMENTED error over the protocol.
func TestLibRuntime_ExecuteReturnsBlockExecutionNotImplemented(t *testing.T) {
	lib := newHelloWorldLibrary(t)
	inv := startInvoker(t, lib)

	req := protocol.Request{
		Protocol: protocol.ProtocolName,
		Version:  protocol.Version,
		ID:       "test-1",
		Method:   "execute",
	}
	params, _ := json.Marshal(protocol.ExecuteParams{
		BlockType: "hello-world",
		Config:    map[string]any{},
	})
	req.Params = params

	resp, err := inv.CallForTest(req)
	if err != nil {
		t.Fatalf("unexpected communication error: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected protocol error, got nil")
	}
	if resp.Error.Code != "BLOCK_EXECUTION_NOT_IMPLEMENTED" {
		t.Errorf("expected error code %q, got %q", "BLOCK_EXECUTION_NOT_IMPLEMENTED", resp.Error.Code)
	}
	if resp.ID != "test-1" {
		t.Errorf("expected response ID %q, got %q", "test-1", resp.ID)
	}
}

// TestLibRuntime_GetDefinitionReturnsBlockDefinition verifies that the
// lib-runtime returns a valid block definition when get-definition is called.
func TestLibRuntime_GetDefinitionReturnsBlockDefinition(t *testing.T) {
	lib := newHelloWorldLibrary(t)
	inv := startInvoker(t, lib)

	params, _ := json.Marshal(protocol.GetDefinitionParams{BlockType: "hello-world"})
	req := protocol.Request{
		Protocol: protocol.ProtocolName,
		Version:  protocol.Version,
		ID:       "test-2",
		Method:   "get-definition",
		Params:   params,
	}

	resp, err := inv.CallForTest(req)
	if err != nil {
		t.Fatalf("unexpected communication error: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected protocol error: %s – %s", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		t.Fatal("expected result, got nil")
	}

	var result protocol.GetDefinitionResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal get-definition result: %v", err)
	}

	// The returned definition must at least carry the correct block ID.
	var def map[string]any
	if err := json.Unmarshal(result.Definition, &def); err != nil {
		t.Fatalf("failed to unmarshal embedded definition: %v", err)
	}
	if def["id"] != "hello-world" {
		t.Errorf("expected definition id %q, got %v", "hello-world", def["id"])
	}
}

// TestLibRuntime_ExecuteUnknownBlockReturnsBlockNotFound verifies that
// requesting a block type absent from the library returns BLOCK_NOT_FOUND.
func TestLibRuntime_ExecuteUnknownBlockReturnsBlockNotFound(t *testing.T) {
	lib := newHelloWorldLibrary(t) // only contains hello-world
	inv := startInvoker(t, lib)

	params, _ := json.Marshal(protocol.ExecuteParams{BlockType: "does-not-exist"})
	req := protocol.Request{
		Protocol: protocol.ProtocolName,
		Version:  protocol.Version,
		ID:       "test-3",
		Method:   "execute",
		Params:   params,
	}

	resp, err := inv.CallForTest(req)
	if err != nil {
		t.Fatalf("unexpected communication error: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error for unknown block, got nil")
	}
	if resp.Error.Code != "BLOCK_NOT_FOUND" {
		t.Errorf("expected error code %q, got %q", "BLOCK_NOT_FOUND", resp.Error.Code)
	}
}

// TestLibRuntime_UnknownMethodReturnsError verifies that calling an
// unrecognised method returns an UNKNOWN_METHOD protocol error.
func TestLibRuntime_UnknownMethodReturnsError(t *testing.T) {
	lib := newHelloWorldLibrary(t)
	inv := startInvoker(t, lib)

	req := protocol.Request{
		Protocol: protocol.ProtocolName,
		Version:  protocol.Version,
		ID:       "test-4",
		Method:   "not-a-real-method",
		Params:   []byte(`{}`),
	}

	resp, err := inv.CallForTest(req)
	if err != nil {
		t.Fatalf("unexpected communication error: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error for unknown method, got nil")
	}
	if resp.Error.Code != "UNKNOWN_METHOD" {
		t.Errorf("expected error code %q, got %q", "UNKNOWN_METHOD", resp.Error.Code)
	}
}

// --- Invoker-level tests (via Executor) --------------------------------------

// TestExecutor_LibraryRuntimeInvoker_ProducesFailedResult verifies the full
// integration path: Executor → LibraryRuntimeInvoker → lib-runtime process →
// BLOCK_EXECUTION_NOT_IMPLEMENTED → Result recorded on node's Outputs.
func TestExecutor_LibraryRuntimeInvoker_ProducesFailedResult(t *testing.T) {
	lib := newHelloWorldLibrary(t)
	inv := startInvoker(t, lib)

	ex := execution.New(lib, inv)
	fm := singleNodeFlow("hello1", "hello-world")

	updated, err := ex.Execute(fm, "hello1")
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	node, found := updated.FindNode("hello1")
	if !found {
		t.Fatal("node hello1 missing from updated flow map")
	}

	if node.Status != "failed" {
		t.Errorf("expected status %q, got %q", "failed", node.Status)
	}
	if node.Outputs == nil {
		t.Fatal("expected Outputs to be set, got nil")
	}

	fail, _ := node.Outputs["fail"].(bool)
	if !fail {
		t.Error("expected outputs.fail to be true")
	}

	errOut, ok := node.Outputs["error"].(map[string]any)
	if !ok {
		t.Fatal("expected outputs.error to be a map")
	}
	if errOut["code"] != "BLOCK_EXECUTION_NOT_IMPLEMENTED" {
		t.Errorf("expected error code %q, got %v", "BLOCK_EXECUTION_NOT_IMPLEMENTED", errOut["code"])
	}
}

// TestExecutor_LibraryRuntimeInvoker_ExecutesRealBlockSuccessfully tests the end-to-end
// execution of the implemented hello-world block in the block library.
func TestExecutor_LibraryRuntimeInvoker_ExecutesRealBlockSuccessfully(t *testing.T) {
	libPath, err := filepath.Abs("../../../block-library")
	if err != nil {
		t.Fatalf("failed to resolve block-library path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(libPath, "hello-world", "hello-world.block")); err != nil {
		t.Skip("hello-world block not present in block-library, skipping")
	}

	inv := startInvoker(t, libPath)
	ex := execution.New(libPath, inv)
	fm := singleNodeFlow("hello1", "hello-world")

	updated, err := ex.Execute(fm, "hello1")
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	node, found := updated.FindNode("hello1")
	if !found {
		t.Fatal("node hello1 missing from updated flow map")
	}

	if node.Status != "succeeded" {
		t.Errorf("expected status %q, got %q", "succeeded", node.Status)
	}
	if node.Outputs == nil {
		t.Fatal("expected Outputs to be set, got nil")
	}

	success, _ := node.Outputs["success"].(bool)
	if !success {
		t.Error("expected outputs.success to be true")
	}

	fail, _ := node.Outputs["fail"].(bool)
	if fail {
		t.Error("expected outputs.fail to be false")
	}

	msg, _ := node.Outputs["message"].(string)
	if msg != "Hello, World!" {
		t.Errorf("expected outputs.message to be %q, got %q", "Hello, World!", msg)
	}
}

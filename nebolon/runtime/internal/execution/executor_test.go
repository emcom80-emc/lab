package execution_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"nebolon/runtime/internal/execution"
	"nebolon/runtime/internal/flow"
)

// --- Helpers -----------------------------------------------------------------

// writeFile writes content to path, creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

// newLibrary creates a temporary block library directory containing a
// hello-world block definition and returns its root path.
func newLibrary(t *testing.T) string {
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

// singleNodeFlow returns a FlowMap containing one node of the given type.
func singleNodeFlow(nodeID, blockType string) flow.FlowMap {
	return flow.FlowMap{
		ID:      "test-flow",
		Name:    "Test Flow",
		Version: "1",
		Nodes: []flow.Node{
			{
				ID:     nodeID,
				Type:   blockType,
				Config: map[string]any{},
				Status: "idle",
			},
		},
		Connections: []flow.Connection{},
	}
}

// --- Tests -------------------------------------------------------------------

// TestExecute_BlockAPIMissing verifies the full happy-path pipeline:
// a valid flow, a valid node, and a valid block definition all resolve
// correctly and produce the standardised BLOCK_API_MISSING result.
func TestExecute_BlockAPIMissing(t *testing.T) {
	lib := newLibrary(t)
	ex := execution.New(lib, execution.StubInvoker{})
	fm := singleNodeFlow("hello1", "hello-world")

	updated, err := ex.Execute(fm, "hello1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	success, _ := node.Outputs["success"].(bool)
	if success {
		t.Error("expected outputs.success to be false")
	}

	fail, _ := node.Outputs["fail"].(bool)
	if !fail {
		t.Error("expected outputs.fail to be true")
	}

	errOut, ok := node.Outputs["error"].(map[string]any)
	if !ok {
		t.Fatal("expected outputs.error to be a map, got nil or wrong type")
	}
	if errOut["code"] != "BLOCK_API_MISSING" {
		t.Errorf("expected error code %q, got %q", "BLOCK_API_MISSING", errOut["code"])
	}
}

// TestExecute_NodeNotFound verifies that requesting a node ID that does not
// exist in the flow map returns an error and no updated flow map.
func TestExecute_NodeNotFound(t *testing.T) {
	lib := newLibrary(t)
	ex := execution.New(lib, execution.StubInvoker{})
	fm := singleNodeFlow("hello1", "hello-world")

	_, err := ex.Execute(fm, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing node, got nil")
	}
}

// TestExecute_BlockDefinitionMissing verifies that a node whose block type has
// no corresponding definition file in the library produces an error.
func TestExecute_BlockDefinitionMissing(t *testing.T) {
	lib := newLibrary(t) // only contains hello-world
	ex := execution.New(lib, execution.StubInvoker{})
	fm := singleNodeFlow("node1", "unknown-block-type")

	_, err := ex.Execute(fm, "node1")
	if err == nil {
		t.Fatal("expected error for missing block definition, got nil")
	}
}

// TestExecute_MalformedBlockDefinition verifies that a block definition file
// containing invalid JSON produces an error.
func TestExecute_MalformedBlockDefinition(t *testing.T) {
	lib := t.TempDir()
	writeFile(t, filepath.Join(lib, "bad-block", "bad-block.block"), `{ not valid json `)
	ex := execution.New(lib, execution.StubInvoker{})
	fm := singleNodeFlow("node1", "bad-block")

	_, err := ex.Execute(fm, "node1")
	if err == nil {
		t.Fatal("expected error for malformed block definition, got nil")
	}
}

// TestExecute_OriginalFlowMapUnchanged verifies that Execute never mutates
// the FlowMap passed to it as input.
func TestExecute_OriginalFlowMapUnchanged(t *testing.T) {
	lib := newLibrary(t)
	ex := execution.New(lib, execution.StubInvoker{})
	fm := singleNodeFlow("hello1", "hello-world")

	// Capture a JSON snapshot of the original before execution.
	before, err := json.Marshal(fm)
	if err != nil {
		t.Fatalf("failed to marshal original flow: %v", err)
	}

	if _, err := ex.Execute(fm, "hello1"); err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	after, err := json.Marshal(fm)
	if err != nil {
		t.Fatalf("failed to marshal flow after execution: %v", err)
	}

	if string(before) != string(after) {
		t.Errorf("original FlowMap was mutated by Execute\nbefore: %s\nafter:  %s", before, after)
	}
}

// TestExecute_UpdatedFlowMapContainsResult verifies that the updated FlowMap
// returned by Execute carries the execution result on the correct node while
// all other nodes remain unchanged.
func TestExecute_UpdatedFlowMapContainsResult(t *testing.T) {
	lib := newLibrary(t)
	ex := execution.New(lib, execution.StubInvoker{})

	// Two-node flow; only hello1 is executed.
	fm := flow.FlowMap{
		ID:      "two-node-flow",
		Version: "1",
		Nodes: []flow.Node{
			{ID: "hello1", Type: "hello-world", Config: map[string]any{}, Status: "idle"},
			{ID: "hello2", Type: "hello-world", Config: map[string]any{}, Status: "idle"},
		},
		Connections: []flow.Connection{},
	}

	updated, err := ex.Execute(fm, "hello1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Executed node must have outputs set.
	n1, _ := updated.FindNode("hello1")
	if n1.Outputs == nil {
		t.Error("executed node hello1 should have Outputs set")
	}

	// Non-executed node must remain untouched.
	n2, _ := updated.FindNode("hello2")
	if n2.Outputs != nil {
		t.Error("non-executed node hello2 should not have Outputs set")
	}
	if n2.Status != "idle" {
		t.Errorf("non-executed node hello2 status should remain %q, got %q", "idle", n2.Status)
	}
}

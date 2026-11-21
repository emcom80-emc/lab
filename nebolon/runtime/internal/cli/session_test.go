package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nebolon/runtime/internal/cli"
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

// newLibrary creates a temporary block library with a hello-world block.
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

// newSession creates a Session using a temp library.
func newSession(t *testing.T) *cli.Session {
	t.Helper()
	return cli.NewSession(newLibrary(t), "")
}

// mustCommand runs a command and fails the test on error.
func mustCommand(t *testing.T, s *cli.Session, cmd string) string {
	t.Helper()
	out, err := s.Command(cmd)
	if err != nil {
		t.Fatalf("unexpected error for %q: %v", cmd, err)
	}
	return out
}

// parseFlowJSON parses JSON output from a command into a FlowMap.
func parseFlowJSON(t *testing.T, out string) flow.FlowMap {
	t.Helper()
	var fm flow.FlowMap
	if err := json.Unmarshal([]byte(out), &fm); err != nil {
		t.Fatalf("failed to parse flow JSON: %v\noutput was: %s", err, out)
	}
	return fm
}

// newFlowFile writes a minimal valid .flow file to a temp dir and returns its path.
func newFlowFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.flow")
	fm := flow.FlowMap{
		ID:          "test-flow",
		Name:        "Test Flow",
		Version:     "1",
		Nodes:       []flow.Node{},
		Connections: []flow.Connection{},
	}
	data, err := flow.Marshal(fm)
	if err != nil {
		t.Fatalf("failed to marshal flow: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write flow file: %v", err)
	}
	return path
}

// --- Tests -------------------------------------------------------------------

func TestCreate_CreatesFlowMapWithCorrectName(t *testing.T) {
	s := newSession(t)
	out := mustCommand(t, s, "/create My Flow")
	fm := parseFlowJSON(t, out)
	if fm.Name != "My Flow" {
		t.Errorf("expected name %q, got %q", "My Flow", fm.Name)
	}
	if fm.ID == "" {
		t.Error("expected non-empty ID")
	}
	if fm.Version != "1" {
		t.Errorf("expected version %q, got %q", "1", fm.Version)
	}
}

func TestCreate_ReturnsValidJSON(t *testing.T) {
	s := newSession(t)
	out := mustCommand(t, s, "/create TestFlow")
	if !json.Valid([]byte(out)) {
		t.Errorf("output is not valid JSON: %s", out)
	}
}

func TestCreate_ResetsFilePath(t *testing.T) {
	s := newSession(t)
	// Create a flow and save it somewhere.
	dir := t.TempDir()
	path := filepath.Join(dir, "test.flow")
	mustCommand(t, s, "/create InitialFlow")
	mustCommand(t, s, "/save "+path)
	// Create a new flow — file path should be cleared.
	mustCommand(t, s, "/create NewFlow")
	// Saving without a path should fail.
	_, err := s.Command("/save")
	if err == nil {
		t.Error("expected error after /create clears file path")
	}
}

func TestLoad_LoadsValidFlowFile(t *testing.T) {
	s := newSession(t)
	path := newFlowFile(t)
	out := mustCommand(t, s, "/load "+path)
	fm := parseFlowJSON(t, out)
	if fm.Name != "Test Flow" {
		t.Errorf("expected name %q, got %q", "Test Flow", fm.Name)
	}
}

func TestLoad_ErrorOnMissingFile(t *testing.T) {
	s := newSession(t)
	_, err := s.Command("/load /nonexistent/path/to/file.flow")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoad_ErrorOnMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.flow")
	writeFile(t, path, "{ not valid json")
	s := newSession(t)
	_, err := s.Command("/load " + path)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestSave_SavesCurrentFlowToPath(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create SaveTest")
	dir := t.TempDir()
	path := filepath.Join(dir, "out.flow")
	out := mustCommand(t, s, "/save "+path)
	if !strings.Contains(out, path) {
		t.Errorf("expected output to mention path %q, got %q", path, out)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	var fm flow.FlowMap
	if err := json.Unmarshal(data, &fm); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}
	if fm.Name != "SaveTest" {
		t.Errorf("expected name %q in saved file, got %q", "SaveTest", fm.Name)
	}
}

func TestSave_UpdatesFilePath(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create PathTest")
	dir := t.TempDir()
	path := filepath.Join(dir, "first.flow")
	mustCommand(t, s, "/save "+path)
	// Now /save without args should use the same path.
	out := mustCommand(t, s, "/save")
	if !strings.Contains(out, path) {
		t.Errorf("expected /save to reuse path %q, got %q", path, out)
	}
}

func TestSave_ErrorWithNoPath(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create NoPath")
	_, err := s.Command("/save")
	if err == nil {
		t.Error("expected error when saving without a path")
	}
}

func TestSave_ErrorWithNoFlowMap(t *testing.T) {
	s := newSession(t)
	_, err := s.Command("/save /tmp/x.flow")
	if err == nil {
		t.Error("expected error when no flow map is loaded")
	}
}

func TestShow_ReturnsCurrentState(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create ShowMe")
	out := mustCommand(t, s, "/show")
	fm := parseFlowJSON(t, out)
	if fm.Name != "ShowMe" {
		t.Errorf("expected name %q in /show output, got %q", "ShowMe", fm.Name)
	}
}

func TestShow_ErrorWithNoFlowMap(t *testing.T) {
	s := newSession(t)
	_, err := s.Command("/show")
	if err == nil {
		t.Error("expected error when no flow map loaded")
	}
}

func TestBlockAdd_AddsNodeToFlowMap(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create BlockTest")
	out := mustCommand(t, s, "/block-add hello-world")
	fm := parseFlowJSON(t, out)
	if len(fm.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(fm.Nodes))
	}
	node := fm.Nodes[0]
	if node.Type != "hello-world" {
		t.Errorf("expected node type %q, got %q", "hello-world", node.Type)
	}
	if !strings.HasPrefix(node.ID, "hello-world-") {
		t.Errorf("expected node ID to start with 'hello-world-', got %q", node.ID)
	}
}

func TestBlockAdd_GeneratesUniqueIDs(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create MultiNode")
	mustCommand(t, s, "/block-add hello-world")
	out := mustCommand(t, s, "/block-add hello-world")
	fm := parseFlowJSON(t, out)
	if len(fm.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(fm.Nodes))
	}
	if fm.Nodes[0].ID == fm.Nodes[1].ID {
		t.Errorf("expected unique node IDs, both are %q", fm.Nodes[0].ID)
	}
}

func TestBlockAdd_SetsDefaultPosition(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create PosTest")
	mustCommand(t, s, "/block-add hello-world")
	out := mustCommand(t, s, "/block-add hello-world")
	fm := parseFlowJSON(t, out)
	// First node: x=100, y=100
	if fm.Nodes[0].Position.X != 100 || fm.Nodes[0].Position.Y != 100 {
		t.Errorf("first node position: expected (100,100), got (%v,%v)",
			fm.Nodes[0].Position.X, fm.Nodes[0].Position.Y)
	}
	// Second node: x=320, y=100
	if fm.Nodes[1].Position.X != 320 || fm.Nodes[1].Position.Y != 100 {
		t.Errorf("second node position: expected (320,100), got (%v,%v)",
			fm.Nodes[1].Position.X, fm.Nodes[1].Position.Y)
	}
}

func TestBlockAdd_ErrorForUnknownBlock(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create BlockErr")
	_, err := s.Command("/block-add nonexistent-block")
	if err == nil {
		t.Error("expected error for unknown block type")
	}
}

func TestBlockAdd_ErrorWithNoFlowMap(t *testing.T) {
	s := newSession(t)
	_, err := s.Command("/block-add hello-world")
	if err == nil {
		t.Error("expected error when no flow map loaded")
	}
}

func TestBlocks_ListsAvailableBlockTypes(t *testing.T) {
	lib := t.TempDir()
	// Add two block types.
	writeFile(t, filepath.Join(lib, "hello-world", "hello-world.block"), `{"id":"hello-world","name":"Hello World","implementation":{"type":"process","language":"go","path":"."}}`)
	writeFile(t, filepath.Join(lib, "farewell", "farewell.block"), `{"id":"farewell","name":"Farewell","implementation":{"type":"process","language":"go","path":"."}}`)
	// A directory without a .block file should not appear.
	if err := os.MkdirAll(filepath.Join(lib, "not-a-block"), 0755); err != nil {
		t.Fatal(err)
	}

	s := cli.NewSession(lib, "")
	out := mustCommand(t, s, "/blocks")
	if !strings.Contains(out, "hello-world") {
		t.Errorf("expected 'hello-world' in /blocks output, got: %s", out)
	}
	if !strings.Contains(out, "farewell") {
		t.Errorf("expected 'farewell' in /blocks output, got: %s", out)
	}
	if strings.Contains(out, "not-a-block") {
		t.Errorf("expected 'not-a-block' to be absent from /blocks output, got: %s", out)
	}
}

func TestBlocks_ErrorOnMissingLibrary(t *testing.T) {
	s := cli.NewSession("/nonexistent/library/path", "")
	_, err := s.Command("/blocks")
	if err == nil {
		t.Error("expected error for missing library directory")
	}
}

func TestHelp_ReturnsNonEmptyOutput(t *testing.T) {
	s := newSession(t)
	out := mustCommand(t, s, "/help")
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty output from /help")
	}
}

func TestHelp_MentionsAllCommands(t *testing.T) {
	s := newSession(t)
	out := mustCommand(t, s, "/help")
	for _, cmd := range []string{"/create", "/load", "/save", "/show", "/block-add", "/blocks", "/exec", "/exit"} {
		if !strings.Contains(out, cmd) {
			t.Errorf("expected /help to mention %q, but it didn't", cmd)
		}
	}
}

func TestExit_ReturnsSentinel(t *testing.T) {
	s := newSession(t)
	out, err := s.Command("/exit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != cli.ExitSentinel {
		t.Errorf("expected ExitSentinel %q, got %q", cli.ExitSentinel, out)
	}
}

func TestUnknownCommand_ReturnsError(t *testing.T) {
	s := newSession(t)
	_, err := s.Command("/notacommand")
	if err == nil {
		t.Error("expected error for unknown command, got nil")
	}
}

func TestNonSlashInput_ReturnsError(t *testing.T) {
	s := newSession(t)
	_, err := s.Command("hello world")
	if err == nil {
		t.Error("expected error for non-slash input, got nil")
	}
}

func TestExec_AllNodes_WithStubInvoker(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create ExecTest")
	mustCommand(t, s, "/block-add hello-world")
	out := mustCommand(t, s, "/exec")
	fm := parseFlowJSON(t, out)
	if len(fm.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(fm.Nodes))
	}
	// StubInvoker always returns "failed".
	if fm.Nodes[0].Status != "failed" {
		t.Errorf("expected stub result status %q, got %q", "failed", fm.Nodes[0].Status)
	}
}

func TestExec_SpecificNode_WithStubInvoker(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create ExecOneTest")
	out1 := mustCommand(t, s, "/block-add hello-world")
	fm1 := parseFlowJSON(t, out1)
	nodeID := fm1.Nodes[0].ID

	out2 := mustCommand(t, s, "/exec "+nodeID)
	fm2 := parseFlowJSON(t, out2)
	node, found := fm2.FindNode(nodeID)
	if !found {
		t.Fatalf("node %q not found in result", nodeID)
	}
	if node.Status != "failed" {
		t.Errorf("expected stub result status %q, got %q", "failed", node.Status)
	}
}

func TestExec_ErrorForUnknownNode(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create ExecMissing")
	_, err := s.Command("/exec nonexistent-node")
	if err == nil {
		t.Error("expected error for unknown node ID")
	}
}

func TestExec_ErrorForMultipleArgs(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create ExecMultiArg")
	_, err := s.Command("/exec node1 node2")
	if err == nil {
		t.Error("expected error when /exec is given multiple args")
	}
}

func TestBlockAdd_AvoidsExistingIDs(t *testing.T) {
	s := newSession(t)
	mustCommand(t, s, "/create PreExistingID")
	// Add node hello-world-1
	mustCommand(t, s, "/block-add hello-world")
	// Artificially reset counter to 0 to simulate collision
	s.SetCounterForTest(0)
	// Next add should still generate a unique ID (e.g. hello-world-2)
	out := mustCommand(t, s, "/block-add hello-world")
	fm := parseFlowJSON(t, out)
	if len(fm.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(fm.Nodes))
	}
	if fm.Nodes[0].ID == fm.Nodes[1].ID {
		t.Errorf("expected unique IDs, got duplicate: %q", fm.Nodes[0].ID)
	}
}

func TestRunREPL_FullSession(t *testing.T) {
	s := newSession(t)
	input := "/create REPLTest\nnot-a-command\n/show\n/exit\n"
	var out strings.Builder
	err := cli.RunREPL(s, strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("RunREPL error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "REPLTest") {
		t.Errorf("expected output to contain 'REPLTest', got: %s", output)
	}
	if !strings.Contains(output, "hint: commands start with /") {
		t.Errorf("expected hint for non-slash command, got: %s", output)
	}
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("expected 'Goodbye!' on exit, got: %s", output)
	}
}

// Package cli provides the interactive session and command dispatch for the
// Nebolon Flow CLI. It contains no terminal I/O itself — only pure command
// logic — so it can be tested without a real terminal.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nebolon/runtime/internal/blocks"
	"nebolon/runtime/internal/execution"
	"nebolon/runtime/internal/flow"
)

// ExitSentinel is returned by Command when the user types /exit.
// The REPL layer should recognise this string and terminate.
const ExitSentinel = "__EXIT__"

// Session holds the state for one interactive CLI session.
type Session struct {
	// fm is the current FlowMap being worked on.
	fm flow.FlowMap
	// hasFM tracks whether a flow map has been created/loaded.
	hasFM bool
	// filePath is the current save path; empty means unsaved.
	filePath string
	// LibraryRoot is the path to the block library directory.
	LibraryRoot string
	// LibRuntimePath is the path to the nebolon-lib-runtime binary.
	// Empty means use StubInvoker.
	LibRuntimePath string
	// counter increments for each node added, used to generate unique IDs.
	counter int
}

// NewSession creates a Session with the given library configuration.
func NewSession(libraryRoot, libRuntimePath string) *Session {
	return &Session{
		LibraryRoot:    libraryRoot,
		LibRuntimePath: libRuntimePath,
	}
}

// SetCounterForTest allows tests to set the node ID counter.
func (s *Session) SetCounterForTest(val int) {
	s.counter = val
}

// Command parses a slash-command line and returns the text to display.
// All file I/O and execution happen inside this method.
func (s *Session) Command(input string) (string, error) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return "", fmt.Errorf("commands must start with /")
	}

	parts := strings.Fields(input)
	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "/create":
		return s.cmdCreate(args)
	case "/load":
		return s.cmdLoad(args)
	case "/save":
		return s.cmdSave(args)
	case "/show":
		return s.cmdShow()
	case "/block-add":
		return s.cmdBlockAdd(args)
	case "/blocks":
		return s.cmdBlocks()
	case "/exec":
		return s.cmdExec(args)
	case "/help":
		return s.cmdHelp(), nil
	case "/exit":
		return ExitSentinel, nil
	default:
		return "", fmt.Errorf("unknown command %q — type /help for available commands", cmd)
	}
}

// cmdCreate creates a new FlowMap with the given name.
func (s *Session) cmdCreate(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("usage: /create <name>")
	}
	name := strings.Join(args, " ")
	s.fm = flow.New(name)
	s.hasFM = true
	s.filePath = ""
	s.counter = 0
	return marshalFlow(s.fm)
}

// cmdLoad loads a .flow file and sets it as the current flow map.
func (s *Session) cmdLoad(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("usage: /load <file>")
	}
	path := args[0]
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %q: %w", path, err)
	}
	fm, err := flow.Parse(data)
	if err != nil {
		return "", err
	}
	s.fm = fm
	s.hasFM = true
	s.filePath = path
	return marshalFlow(s.fm)
}

// cmdSave saves the current flow map.
func (s *Session) cmdSave(args []string) (string, error) {
	if err := s.requireFlow(); err != nil {
		return "", err
	}
	if len(args) > 0 {
		path := args[0]
		if !strings.HasSuffix(strings.ToLower(path), ".flow") {
			path += ".flow"
		}
		s.filePath = path
	}
	if s.filePath == "" {
		return "", fmt.Errorf("no file path set — use /save <file>")
	}
	data, err := flow.Marshal(s.fm)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write %q: %w", s.filePath, err)
	}
	return fmt.Sprintf("saved to %s", s.filePath), nil
}

// cmdShow returns the current flow map as formatted JSON.
func (s *Session) cmdShow() (string, error) {
	if err := s.requireFlow(); err != nil {
		return "", err
	}
	return marshalFlow(s.fm)
}

// cmdBlockAdd loads a block definition and adds a node to the current flow map.
func (s *Session) cmdBlockAdd(args []string) (string, error) {
	if err := s.requireFlow(); err != nil {
		return "", err
	}
	if len(args) == 0 {
		return "", fmt.Errorf("usage: /block-add <block-id>")
	}
	blockType := args[0]

	// Verify the block definition exists through the existing blocks package.
	if _, err := blocks.Load(s.LibraryRoot, blockType); err != nil {
		return "", err
	}

	// Add node through the FlowMap domain model.
	s.fm.AddNode(blockType, nil)
	return marshalFlow(s.fm)
}

// cmdBlocks lists available block types from the library root using block discovery.
func (s *Session) cmdBlocks() (string, error) {
	types, err := blocks.Discover(s.LibraryRoot)
	if err != nil {
		return "", err
	}
	if len(types) == 0 {
		return "no block types found in " + s.LibraryRoot, nil
	}
	return "available blocks:\n  " + strings.Join(types, "\n  "), nil
}

// cmdExec executes nodes in the current flow map.
// With no args, executes all nodes in order.
// With one arg, executes that specific node.
func (s *Session) cmdExec(args []string) (string, error) {
	if err := s.requireFlow(); err != nil {
		return "", err
	}
	if len(args) > 1 {
		return "", fmt.Errorf("usage: /exec [node-id]")
	}

	invoker, cleanup, err := s.makeInvoker()
	if err != nil {
		return "", err
	}
	if cleanup != nil {
		defer cleanup()
	}

	ex := execution.New(s.LibraryRoot, invoker)

	if len(args) == 1 {
		// Execute a single named node.
		nodeID := args[0]
		updated, err := ex.Execute(s.fm, nodeID)
		if err != nil {
			return "", err
		}
		s.fm = updated
		return marshalFlow(s.fm)
	}

	// Execute the entire flow using the Execution engine.
	updated, err := ex.ExecuteFlow(s.fm)
	if err != nil {
		return "", err
	}
	s.fm = updated
	return marshalFlow(s.fm)
}

// cmdHelp returns usage text for all commands.
func (s *Session) cmdHelp() string {
	return `Nebolon Flow CLI — available commands:

  /create <name>       Create a new flow map with the given name
  /load <file>         Load a .flow file
  /save                Save current flow map to the current file path
  /save <file>         Save current flow map to <file>
  /show                Print current flow map as JSON
  /block-add <id>      Add a node of block type <id> to the current flow map
  /blocks              List available block types from the block library
  /exec                Execute all nodes in the current flow map
  /exec <node-id>      Execute a specific node by ID
  /help                Show this help text
  /exit                Exit the CLI`
}

// requireFlow returns an error if no flow map is currently loaded.
func (s *Session) requireFlow() error {
	if !s.hasFM {
		return fmt.Errorf("no flow map loaded — use /create or /load first")
	}
	return nil
}

// makeInvoker returns an Invoker (and optional cleanup func) based on session config.
func (s *Session) makeInvoker() (execution.Invoker, func(), error) {
	if s.LibRuntimePath == "" {
		return execution.StubInvoker{}, nil, nil
	}
	libInvoker, err := execution.NewLibraryRuntimeInvoker(s.LibRuntimePath, s.LibraryRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start lib-runtime: %w", err)
	}
	cleanup := func() { libInvoker.Close() }
	return libInvoker, cleanup, nil
}

// marshalFlow serialises a FlowMap to indented JSON string.
func marshalFlow(fm flow.FlowMap) (string, error) {
	data, err := flow.Marshal(fm)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// sanitizeID converts a name to a simple lowercase ID (spaces → hyphens).
func sanitizeID(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

// ResolveLibraryRoot finds the block library directory.
// ResolveLibraryRoot finds the block library directory using the centralized blocks logic.
// If explicit is non-empty, it is used verbatim. Otherwise it resolves relative to the installation.
func ResolveLibraryRoot(explicit string) string {
	return blocks.ResolveLibraryRoot(explicit)
}

// ResolveLibRuntime finds the nebolon-lib-runtime binary if available.
// If explicit is non-empty, it is used. Otherwise it checks common locations.
func ResolveLibRuntime(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if info, err := os.Stat("./nebolon-lib-runtime"); err == nil && !info.IsDir() {
		return "./nebolon-lib-runtime"
	}
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "nebolon-lib-runtime")
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

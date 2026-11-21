// Command nebolon is the NEBOLON CLI.
// It acts as a thin interface to the NEBOLON Runtime execution engine.
// All core execution logic lives in the internal packages; this file is
// responsible only for flag parsing, file I/O, and exit codes.
//
// Usage:
//
//	nebolon                                               — interactive mode
//	nebolon execute --flow <path> --node <id> [flags]     — execute a node
//	nebolon help                                          — show help
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"nebolon/runtime/internal/blocks"
	"nebolon/runtime/internal/cli"
	"nebolon/runtime/internal/execution"
	"nebolon/runtime/internal/flow"
)

const defaultLibraryRoot = blocks.DefaultLibraryRoot

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return runInteractive("", "")
	}

	switch args[0] {
	case "help", "--help", "-help", "-h":
		fmt.Print(helpText)
		return nil
	case "execute":
		return runExecute(args[1:])
	default:
		if strings.HasPrefix(args[0], "-") {
			fs := flag.NewFlagSet("nebolon", flag.ContinueOnError)
			fs.SetOutput(os.Stderr)
			blockLibrary := fs.String("block-library", "", "Path to the block library directory")
			libRuntime := fs.String("lib-runtime", "", "Path to the nebolon-lib-runtime binary")
			if err := fs.Parse(args); err == nil {
				return runInteractive(*blockLibrary, *libRuntime)
			}
		}
		return fmt.Errorf("unknown command %q. Run 'nebolon help' for usage.", args[0])
	}
}

func runInteractive(libraryRoot, libRuntime string) error {
	fmt.Println("Nebolon Flow CLI")
	fmt.Println("Type /help for available commands.")

	libRoot := cli.ResolveLibraryRoot(libraryRoot)
	libRun := cli.ResolveLibRuntime(libRuntime)

	session := cli.NewSession(libRoot, libRun)
	return cli.RunREPL(session, os.Stdin, os.Stdout)
}

const helpText = `Nebolon Runtime CLI

Usage:
  nebolon                                            Start interactive flow editor
  nebolon execute --flow <path> --node <id> [flags]  Execute a single node
  nebolon help                                       Show this help text

Execute flags:
  --flow <path>           Path to the input .flow file (required)
  --node <id>             ID of the node to execute (required)
  --block-library <path>  Path to the block library directory (default: block-library)
  --output <path>         Path to write the updated .flow file (default: stdout)
  --lib-runtime <path>    Path to the nebolon-lib-runtime binary (optional)

Interactive commands (type /help inside the CLI for details):
  /create  /load  /save  /show  /block-add  /blocks  /exec  /help  /exit
`

func runExecute(args []string) error {
	fs := flag.NewFlagSet("execute", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	flowPath := fs.String("flow", "", "Path to the input .flow file (required)")
	nodeID := fs.String("node", "", "ID of the node to execute (required)")
	libraryRoot := fs.String("block-library", "", "Path to the block library directory")
	outputPath := fs.String("output", "", "Path to write the updated .flow file (defaults to stdout)")
	libRuntimePath := fs.String("lib-runtime", "", "Path to the nebolon-lib-runtime binary (optional; uses stub invoker if not set)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *flowPath == "" {
		return fmt.Errorf("--flow is required")
	}
	if *nodeID == "" {
		return fmt.Errorf("--node is required")
	}

	resolvedLibRoot := cli.ResolveLibraryRoot(*libraryRoot)

	// Read the input flow map. File I/O is handled here, not in the engine.
	flowData, err := os.ReadFile(*flowPath)
	if err != nil {
		return fmt.Errorf("failed to read flow file %q: %w", *flowPath, err)
	}

	fm, err := flow.Parse(flowData)
	if err != nil {
		return err
	}

	// Choose invoker based on flags.
	var inv execution.Invoker = execution.StubInvoker{}
	if *libRuntimePath != "" {
		libInvoker, err := execution.NewLibraryRuntimeInvoker(*libRuntimePath, resolvedLibRoot)
		if err != nil {
			return fmt.Errorf("failed to start lib-runtime: %w", err)
		}
		defer libInvoker.Close()
		inv = libInvoker
	}

	// Execute the requested node using the Runtime engine.
	ex := execution.New(resolvedLibRoot, inv)
	updated, err := ex.Execute(fm, *nodeID)
	if err != nil {
		return err
	}

	// Serialise the updated flow map.
	out, err := flow.Marshal(updated)
	if err != nil {
		return err
	}

	// Write to the output file, or to stdout if no output path was given.
	// The original input file is never overwritten.
	if *outputPath != "" {
		if err := os.WriteFile(*outputPath, out, 0644); err != nil {
			return fmt.Errorf("failed to write output file %q: %w", *outputPath, err)
		}
		fmt.Fprintf(os.Stderr, "written to %s\n", *outputPath)
	} else {
		fmt.Println(string(out))
	}

	return nil
}

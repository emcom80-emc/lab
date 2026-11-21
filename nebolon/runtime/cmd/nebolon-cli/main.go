// Command nebolon-cli is an interactive REPL for the Nebolon visual programming platform.
// It allows users to create, edit, and execute flow maps interactively.
//
// Usage:
//
//	nebolon-cli [--block-library <path>] [--lib-runtime <path>]
package main

import (
	"flag"
	"fmt"
	"os"

	"nebolon/runtime/internal/blocks"
	"nebolon/runtime/internal/cli"
)

const defaultLibraryRoot = blocks.DefaultLibraryRoot

func main() {
	blockLibrary := flag.String("block-library", "", "Path to the block library directory")
	libRuntime := flag.String("lib-runtime", "", "Path to the nebolon-lib-runtime binary (optional)")
	flag.Parse()

	fmt.Println("Nebolon Flow CLI")
	fmt.Println("Type /help for available commands.")

	libRoot := cli.ResolveLibraryRoot(*blockLibrary)
	libRun := cli.ResolveLibRuntime(*libRuntime)

	session := cli.NewSession(libRoot, libRun)
	if err := cli.RunREPL(session, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

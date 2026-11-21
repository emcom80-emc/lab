// Package blocks provides types and utilities for loading NEBOLON Block Definitions.
// A Block Definition (.block) describes a block's identity, its block-specific
// data inputs and outputs, and its implementation reference.
//
// Standard execution ports (exec, success, fail) are part of the Runtime's block
// model and are NOT declared in individual .block files. They are provided
// centrally here so every regular block inherits them automatically.
package blocks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Port describes a single input or output port on a block.
type Port struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// Implementation holds a reference to the executable code for a block.
// The Runtime uses this to locate and invoke the block.
type Implementation struct {
	Type     string `json:"type"`
	Language string `json:"language"`
	Path     string `json:"path"`
}

// BlockDefinition describes the interface and implementation of a block type.
// Only block-specific data inputs and outputs are declared here.
// Standard execution ports are provided centrally by the Runtime (see below).
type BlockDefinition struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Category       string                 `json:"category,omitempty"`
	Inputs         map[string]Port        `json:"inputs,omitempty"`
	Outputs        map[string]Port        `json:"outputs,omitempty"`
	Implementation Implementation         `json:"implementation"`
}

// StandardInputs contains the execution input ports that every regular block
// receives from the Runtime. Block definitions must not redeclare these.
var StandardInputs = map[string]Port{
	"exec": {
		Type:        "exec",
		Description: "Triggers the block to execute.",
	},
}

// StandardOutputs contains the execution output ports that every regular block
// exposes through the Runtime. Block definitions must not redeclare these.
var StandardOutputs = map[string]Port{
	"success": {
		Type:        "exec",
		Description: "Activated when the block completes successfully.",
	},
	"fail": {
		Type:        "exec",
		Description: "Activated when the block fails.",
	},
}

// Load reads and parses the block definition for blockType from the given
// library root directory. The expected file path is:
//
//	<libraryRoot>/<blockType>/<blockType>.block
func Load(libraryRoot, blockType string) (BlockDefinition, error) {
	path := filepath.Join(libraryRoot, blockType, blockType+".block")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return BlockDefinition{}, fmt.Errorf("block definition not found for type %q (looked in %s)", blockType, path)
		}
		return BlockDefinition{}, fmt.Errorf("failed to read block definition for type %q: %w", blockType, err)
	}

	var def BlockDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return BlockDefinition{}, fmt.Errorf("invalid block definition for type %q: %w", blockType, err)
	}

	return def, nil
}

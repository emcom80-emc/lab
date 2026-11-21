// Package execution contains the NEBOLON Runtime execution engine.
// It is responsible for processing nodes within a Flow Map using an Invoker,
// and returning an updated Flow Map reflecting the execution outcome.
package execution

import (
	"nebolon/runtime/internal/blocks"
	"nebolon/runtime/internal/flow"
)

// Result is the structured outcome of a single block execution attempt.
// It is kept as a separate type from the flow map representation so that
// the invoker layer can evolve independently.
//
// Status is the execution state of the node ("idle", "running", "succeeded", "failed").
//
// Outputs holds all outputs the block produced, keyed by port name.
// This includes:
//   - "success" (bool)  – true when execution completed successfully.
//   - "fail"    (bool)  – true when execution failed.
//   - "error"   (map)   – structured error information; present only on failure.
//   - Any block-specific data outputs declared in the block definition.
//
// Outputs is nil when the block has not yet been executed.
type Result struct {
	Status  string         `json:"status"`
	Outputs map[string]any `json:"outputs"`
}

// Invoker is the interface through which the Executor triggers block execution.
// This is the primary extension point for introducing real block process
// communication in a future milestone. For now, only the stub is provided.
type Invoker interface {
	Invoke(def blocks.BlockDefinition, node flow.Node) Result
}

// StubInvoker is used in the first milestone. It always returns a standardised
// BLOCK_API_MISSING result without starting any external process.
type StubInvoker struct{}

// Invoke returns a simulated failure indicating that the block execution API
// has not been implemented yet. This result is intentional and allows the
// full execution pipeline to be validated before real block communication
// is introduced.
func (StubInvoker) Invoke(_ blocks.BlockDefinition, _ flow.Node) Result {
	return Result{
		Status: "failed",
		Outputs: map[string]any{
			"success": false,
			"fail":    true,
			"error": map[string]any{
				"code":    "BLOCK_API_MISSING",
				"message": "Block execution API is not implemented yet.",
			},
		},
	}
}

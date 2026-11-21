package execution

import (
	"encoding/json"
	"fmt"

	"nebolon/runtime/internal/blocks"
	"nebolon/runtime/internal/flow"
)

// Executor runs a single node within a FlowMap and returns an updated copy.
// It never mutates its input.
type Executor struct {
	libraryRoot string
	invoker     Invoker
}

// New creates an Executor that loads block definitions from libraryRoot and
// invokes blocks through the provided Invoker.
func New(libraryRoot string, invoker Invoker) *Executor {
	return &Executor{
		libraryRoot: libraryRoot,
		invoker:     invoker,
	}
}

// Execute locates nodeID within fm, loads its block definition, invokes it
// through the Invoker, and returns a new FlowMap with the node's status and
// result updated. The original fm is never modified.
func (e *Executor) Execute(fm flow.FlowMap, nodeID string) (flow.FlowMap, error) {
	// 1. Locate the requested node.
	node, found := fm.FindNode(nodeID)
	if !found {
		return flow.FlowMap{}, fmt.Errorf("node %q not found in flow map %q", nodeID, fm.ID)
	}

	// 2. Load the corresponding block definition.
	def, err := blocks.Load(e.libraryRoot, node.Type)
	if err != nil {
		return flow.FlowMap{}, err
	}

	// 3. Invoke the block through the invoker interface.
	result := e.invoker.Invoke(def, node)

	// 4. Deep-copy the input FlowMap so the original is never mutated.
	updated, err := deepCopy(fm)
	if err != nil {
		return flow.FlowMap{}, fmt.Errorf("failed to copy flow map: %w", err)
	}

	// 5. Write the execution result back onto the copied node.
	for i, n := range updated.Nodes {
		if n.ID == nodeID {
			updated.Nodes[i].Status = result.Status
			updated.Nodes[i].Outputs = result.Outputs
			break
		}
	}

	return updated, nil
}

// deepCopy produces a fully independent copy of a FlowMap via JSON round-trip.
func deepCopy(fm flow.FlowMap) (flow.FlowMap, error) {
	data, err := json.Marshal(fm)
	if err != nil {
		return flow.FlowMap{}, err
	}
	var cp flow.FlowMap
	if err := json.Unmarshal(data, &cp); err != nil {
		return flow.FlowMap{}, err
	}
	return cp, nil
}

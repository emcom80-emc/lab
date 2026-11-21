// Package flow provides types and utilities for working with NEBOLON Flow Maps.
// A Flow Map is the JSON-based source of truth that describes a program's
// canvas: its nodes, their positions, connections, configuration, and execution state.
package flow

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Position describes a node's location on the canvas.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Node represents a single block instance placed on the canvas.
type Node struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Position Position       `json:"position"`
	Config   map[string]any `json:"config"`
	Status   string         `json:"status"`
	Outputs  map[string]any `json:"outputs"`
}

// Connection represents a directed link between two ports on the canvas.
// The structure will be expanded in future milestones when flow traversal
// is introduced.
type Connection struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// FlowMap is the top-level structure of a .flow file. It describes an entire
// program canvas including all nodes and their connections.
type FlowMap struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Nodes       []Node       `json:"nodes"`
	Connections []Connection `json:"connections"`
}

// Parse deserialises a .flow JSON payload into a FlowMap.
func Parse(data []byte) (FlowMap, error) {
	var fm FlowMap
	if err := json.Unmarshal(data, &fm); err != nil {
		return FlowMap{}, fmt.Errorf("invalid flow map: %w", err)
	}
	return fm, nil
}

// Marshal serialises a FlowMap to indented JSON.
func Marshal(fm FlowMap) ([]byte, error) {
	data, err := json.MarshalIndent(fm, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialise flow map: %w", err)
	}
	return data, nil
}

// FindNode returns the Node with the given id, and whether it was found.
func (fm FlowMap) FindNode(id string) (Node, bool) {
	for _, n := range fm.Nodes {
		if n.ID == id {
			return n, true
		}
	}
	return Node{}, false
}

// New creates a new FlowMap with the given human-readable name.
func New(name string) FlowMap {
	id := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))
	if id == "" {
		id = "flow"
	}
	return FlowMap{
		ID:          id,
		Name:        name,
		Version:     "1",
		Nodes:       []Node{},
		Connections: []Connection{},
	}
}

// NextNodeID returns the next unique node ID for the given block type.
// It iterates <blockType>-1, <blockType>-2, ... until an unused ID is found.
func (fm FlowMap) NextNodeID(blockType string) string {
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d", blockType, i)
		if _, exists := fm.FindNode(candidate); !exists {
			return candidate
		}
	}
}

// AddNode adds a new block node to the FlowMap with a unique ID.
// If pos is nil, a default canvas position is assigned based on the current node count.
func (fm *FlowMap) AddNode(blockType string, pos *Position) Node {
	var position Position
	if pos != nil {
		position = *pos
	} else {
		n := len(fm.Nodes)
		position = Position{
			X: float64(100 + n*220),
			Y: 100,
		}
	}

	node := Node{
		ID:       fm.NextNodeID(blockType),
		Type:     blockType,
		Position: position,
		Config:   map[string]any{},
		Status:   "idle",
	}

	fm.Nodes = append(fm.Nodes, node)
	return node
}

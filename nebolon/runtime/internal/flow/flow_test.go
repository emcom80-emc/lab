package flow_test

import (
	"testing"

	"nebolon/runtime/internal/flow"
)

func TestNew(t *testing.T) {
	fm := flow.New("My Test Flow")
	if fm.ID != "my-test-flow" {
		t.Errorf("expected ID 'my-test-flow', got %q", fm.ID)
	}
	if fm.Name != "My Test Flow" {
		t.Errorf("expected Name 'My Test Flow', got %q", fm.Name)
	}
	if fm.Version != "1" {
		t.Errorf("expected Version '1', got %q", fm.Version)
	}
	if len(fm.Nodes) != 0 {
		t.Errorf("expected empty nodes, got %d", len(fm.Nodes))
	}
	if len(fm.Connections) != 0 {
		t.Errorf("expected empty connections, got %d", len(fm.Connections))
	}
}

func TestFlowMap_AddNode_DefaultPosition(t *testing.T) {
	fm := flow.New("Position Test")
	n1 := fm.AddNode("hello-world", nil)
	if n1.ID != "hello-world-1" {
		t.Errorf("expected ID 'hello-world-1', got %q", n1.ID)
	}
	if n1.Position.X != 100 || n1.Position.Y != 100 {
		t.Errorf("expected (100, 100), got (%v, %v)", n1.Position.X, n1.Position.Y)
	}

	n2 := fm.AddNode("hello-world", nil)
	if n2.ID != "hello-world-2" {
		t.Errorf("expected ID 'hello-world-2', got %q", n2.ID)
	}
	if n2.Position.X != 320 || n2.Position.Y != 100 {
		t.Errorf("expected (320, 100), got (%v, %v)", n2.Position.X, n2.Position.Y)
	}
}

func TestFlowMap_AddNode_CustomPosition(t *testing.T) {
	fm := flow.New("Custom Pos")
	customPos := &flow.Position{X: 450, Y: 250}
	n := fm.AddNode("my-block", customPos)
	if n.Position.X != 450 || n.Position.Y != 250 {
		t.Errorf("expected (450, 250), got (%v, %v)", n.Position.X, n.Position.Y)
	}
}

func TestFlowMap_NextNodeID_AvoidsCollisions(t *testing.T) {
	fm := flow.New("Collision Test")
	// Manually inject hello-world-1
	fm.Nodes = append(fm.Nodes, flow.Node{
		ID:   "hello-world-1",
		Type: "hello-world",
	})

	// NextNodeID should skip 1 and return 2
	id := fm.NextNodeID("hello-world")
	if id != "hello-world-2" {
		t.Errorf("expected 'hello-world-2', got %q", id)
	}

	// Add hello-world-2
	fm.Nodes = append(fm.Nodes, flow.Node{
		ID:   "hello-world-2",
		Type: "hello-world",
	})

	// NextNodeID should return 3
	id3 := fm.NextNodeID("hello-world")
	if id3 != "hello-world-3" {
		t.Errorf("expected 'hello-world-3', got %q", id3)
	}
}

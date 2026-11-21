package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type ExecutionInput struct {
	Config map[string]any `json:"config"`
	Inputs map[string]any `json:"inputs"`
}

type ExecutionOutput struct {
	Outputs map[string]any `json:"outputs"`
}

func main() {
	var input ExecutionInput
	data, err := io.ReadAll(os.Stdin)
	if err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &input)
	}

	message := "Hello, World!"
	if input.Config != nil {
		if msg, ok := input.Config["message"].(string); ok && msg != "" {
			message = msg
		} else if greeting, ok := input.Config["greeting"].(string); ok && greeting != "" {
			name := "World"
			if n, ok := input.Config["name"].(string); ok && n != "" {
				name = n
			}
			message = fmt.Sprintf("%s, %s!", greeting, name)
		}
	}

	out := ExecutionOutput{
		Outputs: map[string]any{
			"message": message,
		},
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode output: %v\n", err)
		os.Exit(1)
	}
}

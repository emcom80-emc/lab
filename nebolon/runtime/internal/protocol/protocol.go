// Package protocol defines the versioned, language-independent JSON Lines
// communication protocol used between nebolon-runtime and nebolon-lib-runtime.
//
// Every message is a single JSON object terminated by a newline (\n).
// The envelope fields "protocol", "version", and "id" are present on every
// message in both directions. Requests carry a "method" and "params" field;
// responses carry a "result" or "error" field.
//
// This package contains only message types. Transport (reading and writing
// JSON Lines) is handled by the caller.
package protocol

import "encoding/json"

// ProtocolName is the fixed identifier embedded in every message envelope.
// It allows receivers to reject unrelated traffic on the same stream.
const ProtocolName = "nebolon-lib"

// Version is the current protocol version. Bumped on incompatible changes.
const Version = "1"

// --- Envelope ----------------------------------------------------------------

// Request is the envelope for every message sent from nebolon-runtime to
// nebolon-lib-runtime.
type Request struct {
	Protocol string          `json:"protocol"`
	Version  string          `json:"version"`
	ID       string          `json:"id"`
	Method   string          `json:"method"`
	Params   json.RawMessage `json:"params"`
}

// Response is the envelope for every message sent from nebolon-lib-runtime
// back to nebolon-runtime.
// Exactly one of Result or Error is non-nil on a well-formed response.
type Response struct {
	Protocol string          `json:"protocol"`
	Version  string          `json:"version"`
	ID       string          `json:"id"`
	Result   json.RawMessage `json:"result,omitempty"`
	Error    *ResponseError  `json:"error,omitempty"`
}

// ResponseError carries a structured error returned by the lib-runtime.
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// --- Method: execute ---------------------------------------------------------

// ExecuteParams are the parameters for the "execute" method.
// BlockType identifies the block to invoke.
// Config carries the user-defined configuration values for this block instance.
type ExecuteParams struct {
	BlockType string         `json:"blockType"`
	Config    map[string]any `json:"config"`
}

// ExecuteResult is the result returned for a successful "execute" call.
// Outputs contains all block outputs (success/fail signals, data, error).
type ExecuteResult struct {
	Status  string         `json:"status"`
	Outputs map[string]any `json:"outputs"`
}

// --- Method: get-definition --------------------------------------------------

// GetDefinitionParams are the parameters for the "get-definition" method.
type GetDefinitionParams struct {
	BlockType string `json:"blockType"`
}

// GetDefinitionResult carries the raw block definition JSON as returned by
// nebolon-lib-runtime. The caller is responsible for further unmarshalling.
type GetDefinitionResult struct {
	Definition json.RawMessage `json:"definition"`
}

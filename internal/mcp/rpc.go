package mcp

import (
	"encoding/json"
	"fmt"
	"os"
)

// RPCMessage is the standard MCP JSON-RPC structure.
type RPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

// SendResponse marshals and writes a JSON-RPC response to Stdout.
// MCP strictly requires Stdout for JSON.
func SendResponse(id json.RawMessage, result interface{}, err interface{}) {
	resp := RPCMessage{JSONRPC: "2.0", ID: id, Result: result, Error: err}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}

package hook

import (
	"encoding/json"
	"fmt"
	"io"
)

// Input is the PostToolUse hook JSON payload.
type Input struct {
	SessionID   string          `json:"session_id"`
	AgentID     string          `json:"agent_id"`   // absent = main context
	ToolUseID   string          `json:"tool_use_id"`
	ToolName    string          `json:"tool_name"`
	ToolResponse json.RawMessage `json:"tool_response"`
}

// ParsedCall is the extracted data we care about.
type ParsedCall struct {
	SessionID     string
	AgentID       string // empty = main context
	ToolUseID     string
	ToolName      string
	ResponseBytes int64
	IsMainContext bool
}

// Parse reads the PostToolUse hook JSON from r and returns a ParsedCall.
func Parse(r io.Reader) (*ParsedCall, []byte, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, raw, fmt.Errorf("read stdin: %w", err)
	}

	var input Input
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, raw, fmt.Errorf("parse json: %w", err)
	}

	if input.SessionID == "" {
		return nil, raw, fmt.Errorf("missing session_id")
	}
	if input.ToolUseID == "" {
		return nil, raw, fmt.Errorf("missing tool_use_id")
	}
	if input.ToolName == "" {
		return nil, raw, fmt.Errorf("missing tool_name")
	}

	responseBytes := int64(len(input.ToolResponse))

	return &ParsedCall{
		SessionID:     input.SessionID,
		AgentID:       input.AgentID,
		ToolUseID:     input.ToolUseID,
		ToolName:      input.ToolName,
		ResponseBytes: responseBytes,
		IsMainContext: input.AgentID == "",
	}, raw, nil
}

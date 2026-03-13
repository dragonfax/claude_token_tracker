package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Input is the PostToolUse hook JSON payload.
type Input struct {
	SessionID    string          `json:"session_id"`
	AgentID      string          `json:"agent_id"` // absent = main context
	ToolUseID    string          `json:"tool_use_id"`
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
}

// ParsedCall is the extracted data we care about.
type ParsedCall struct {
	SessionID     string
	AgentID       string // empty = main context
	ToolUseID     string
	ToolName      string
	InputSummary  string
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
		InputSummary:  summarizeInput(input.ToolName, input.ToolInput),
		ResponseBytes: responseBytes,
		IsMainContext: input.AgentID == "",
	}, raw, nil
}

const maxSummaryLen = 120

// summarizeInput extracts a human-readable summary from the tool_input JSON.
// It pulls well-known fields for common tools and falls back to a truncated
// raw JSON excerpt for unknown tools.
func summarizeInput(toolName string, raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return truncateSummary(string(raw))
	}

	var s string
	switch toolName {
	case "Agent":
		s = stringField(m, "description")
	case "Bash":
		s = stringField(m, "command")
	case "Read":
		s = stringField(m, "file_path")
	case "Write":
		s = stringField(m, "file_path")
	case "Edit":
		s = stringField(m, "file_path")
	case "Glob":
		s = stringField(m, "pattern")
	case "Grep":
		pattern := stringField(m, "pattern")
		path := stringField(m, "path")
		if path != "" {
			s = pattern + " in " + path
		} else {
			s = pattern
		}
	case "WebFetch", "WebSearch":
		s = stringField(m, "url")
		if s == "" {
			s = stringField(m, "query")
		}
	case "NotebookEdit":
		s = stringField(m, "notebook_path")
	default:
		// Fall back to compact raw JSON excerpt
		s = compactJSON(raw)
	}

	return truncateSummary(s)
}

func stringField(m map[string]json.RawMessage, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return ""
	}
	return s
}

func compactJSON(raw json.RawMessage) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return string(raw)
	}
	return buf.String()
}

func truncateSummary(s string) string {
	// Normalize whitespace
	s = strings.Join(strings.Fields(s), " ")
	runes := []rune(s)
	if len(runes) <= maxSummaryLen {
		return s
	}
	return string(runes[:maxSummaryLen-1]) + "…"
}

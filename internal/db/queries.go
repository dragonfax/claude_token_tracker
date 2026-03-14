package db

import (
	"database/sql"
	"fmt"
	"time"
)

// ToolCall represents a recorded tool response.
type ToolCall struct {
	ID            int64
	RecordedAt    time.Time
	SessionID     string
	AgentID       string // empty string = main context
	ToolUseID     string
	ToolName      string
	InputSummary  string
	ResponseBytes int64
	IsMainContext bool
}

// AppError represents a recorded internal error.
type AppError struct {
	ID         int64
	RecordedAt time.Time
	SessionID  string
	Source     string
	Message    string
	RawInput   string
}

// AggregateRow is one row of the stats view.
type AggregateRow struct {
	ToolName   string
	Calls      int64
	TotalBytes int64
	AvgBytes   float64
}

func InsertToolCall(db *sql.DB, tc ToolCall) error {
	isMain := 0
	if tc.IsMainContext {
		isMain = 1
	}
	var agentID interface{}
	if tc.AgentID != "" {
		agentID = tc.AgentID
	}
	var inputSummary interface{}
	if tc.InputSummary != "" {
		inputSummary = tc.InputSummary
	}
	_, err := db.Exec(
		`INSERT INTO tool_calls (recorded_at, session_id, agent_id, tool_use_id, tool_name, response_bytes, is_main_context, input_summary)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		tc.RecordedAt.UTC().Format(time.RFC3339Nano),
		tc.SessionID,
		agentID,
		tc.ToolUseID,
		tc.ToolName,
		tc.ResponseBytes,
		isMain,
		inputSummary,
	)
	return err
}

func InsertError(db *sql.DB, e AppError) error {
	var sessionID interface{}
	if e.SessionID != "" {
		sessionID = e.SessionID
	}
	var rawInput interface{}
	if e.RawInput != "" {
		rawInput = e.RawInput
	}
	_, err := db.Exec(
		`INSERT INTO errors (recorded_at, session_id, source, message, raw_input)
		 VALUES (?, ?, ?, ?, ?)`,
		e.RecordedAt.UTC().Format(time.RFC3339Nano),
		sessionID,
		e.Source,
		e.Message,
		rawInput,
	)
	return err
}

// TailEntry is a unified row for the watch/log views.
type TailEntry struct {
	ID         int64
	RecordedAt time.Time
	IsError    bool
	// tool call fields
	SessionID     string
	AgentID       string
	ToolName      string
	InputSummary  string
	ResponseBytes int64
	IsMainContext bool
	// error fields
	Source  string
	Message string
}

// TailSince returns tool_calls and errors after the given id/time, merged and sorted.
// afterID is the last tool_calls.id seen (0 for initial load).
// afterErrID is the last errors.id seen.
// showSub controls whether subagent tool calls are included.
func TailSince(db *sql.DB, afterID, afterErrID int64, showSub bool, since time.Time) ([]TailEntry, error) {
	var entries []TailEntry

	subFilter := ""
	if !showSub {
		subFilter = "AND is_main_context = 1"
	}

	sinceStr := since.UTC().Format(time.RFC3339Nano)

	rows, err := db.Query(
		`SELECT id, recorded_at, session_id, COALESCE(agent_id,''), tool_name, COALESCE(input_summary,''), response_bytes, is_main_context
		 FROM tool_calls
		 WHERE id > ? AND recorded_at >= ? `+subFilter+`
		 ORDER BY recorded_at ASC, id ASC`,
		afterID, sinceStr,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var e TailEntry
		var recStr string
		var isMain int
		if err := rows.Scan(&e.ID, &recStr, &e.SessionID, &e.AgentID, &e.ToolName, &e.InputSummary, &e.ResponseBytes, &isMain); err != nil {
			return nil, err
		}
		var err2 error
		e.RecordedAt, err2 = time.Parse(time.RFC3339Nano, recStr)
		if err2 != nil {
			return nil, fmt.Errorf("parse recorded_at %q: %w", recStr, err2)
		}
		e.IsMainContext = isMain == 1
		entries = append(entries, e)
	}

	errRows, err := db.Query(
		`SELECT id, recorded_at, COALESCE(session_id,''), source, message
		 FROM errors
		 WHERE id > ? AND recorded_at >= ?
		 ORDER BY recorded_at ASC, id ASC`,
		afterErrID, sinceStr,
	)
	if err != nil {
		return nil, err
	}
	defer errRows.Close()
	for errRows.Next() {
		var e TailEntry
		var recStr string
		if err := errRows.Scan(&e.ID, &recStr, &e.SessionID, &e.Source, &e.Message); err != nil {
			return nil, err
		}
		var err2 error
		e.RecordedAt, err2 = time.Parse(time.RFC3339Nano, recStr)
		if err2 != nil {
			return nil, fmt.Errorf("parse error recorded_at %q: %w", recStr, err2)
		}
		e.IsError = true
		entries = append(entries, e)
	}

	// Merge-sort by RecordedAt
	sortEntries(entries)
	return entries, nil
}

func sortEntries(entries []TailEntry) {
	// Simple insertion sort — lists are typically small (polling window)
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].RecordedAt.Before(entries[j-1].RecordedAt); j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}
}

// Aggregate returns stats for main-context tool calls in the given time window.
func Aggregate(db *sql.DB, since time.Time) ([]AggregateRow, error) {
	sinceStr := since.UTC().Format(time.RFC3339Nano)
	rows, err := db.Query(
		`SELECT tool_name, COUNT(*) as calls, SUM(response_bytes) as total, AVG(response_bytes) as avg
		 FROM tool_calls
		 WHERE is_main_context = 1 AND recorded_at >= ?
		 GROUP BY tool_name
		 ORDER BY total DESC`,
		sinceStr,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AggregateRow
	for rows.Next() {
		var r AggregateRow
		if err := rows.Scan(&r.ToolName, &r.Calls, &r.TotalBytes, &r.AvgBytes); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

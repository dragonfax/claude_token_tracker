package db

import (
	"testing"
	"time"
)

// TestAggregate_FloatAvg inserts rows with byte counts that produce a
// non-integer AVG, ensuring the scan into AggregateRow succeeds.
func TestAggregate_FloatAvg(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	calls := []ToolCall{
		{RecordedAt: now, SessionID: "s1", ToolUseID: "u1", ToolName: "Read", ResponseBytes: 100, IsMainContext: true},
		{RecordedAt: now, SessionID: "s1", ToolUseID: "u2", ToolName: "Read", ResponseBytes: 201, IsMainContext: true},
	}
	for _, tc := range calls {
		if err := InsertToolCall(db, tc); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	rows, err := Aggregate(db, now.Add(-time.Minute))
	if err != nil {
		t.Fatalf("Aggregate returned error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.ToolName != "Read" {
		t.Errorf("tool name: got %q, want %q", r.ToolName, "Read")
	}
	if r.Calls != 2 {
		t.Errorf("calls: got %d, want 2", r.Calls)
	}
	if r.TotalBytes != 301 {
		t.Errorf("total bytes: got %d, want 301", r.TotalBytes)
	}
	want := 150.5
	if r.AvgBytes != want {
		t.Errorf("avg bytes: got %v, want %v", r.AvgBytes, want)
	}
}

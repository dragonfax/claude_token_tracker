package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	appdb "github.com/reshophq/token-tracker/internal/db"
)

const (
	testWidth  = 120
	testHeight = 40
)

var fixedTime = time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC)

func makeToolCall(id, offset int64, sessionID, toolName string, bytes int64, isMain bool) appdb.TailEntry {
	return appdb.TailEntry{
		ID:            id,
		RecordedAt:    fixedTime.Add(time.Duration(offset) * time.Second),
		IsError:       false,
		SessionID:     sessionID,
		ToolName:      toolName,
		ResponseBytes: bytes,
		IsMainContext: isMain,
	}
}

func makeError(id, offset int64, sessionID, source, message string) appdb.TailEntry {
	return appdb.TailEntry{
		ID:         id,
		RecordedAt: fixedTime.Add(time.Duration(offset) * time.Second),
		IsError:    true,
		SessionID:  sessionID,
		Source:     source,
		Message:    message,
	}
}

func standardEntries() []appdb.TailEntry {
	return []appdb.TailEntry{
		makeToolCall(1, 0, "aabbccdd-1111-2222-3333-444455556666", "Read", 1024, true),
		makeToolCall(2, 10, "aabbccdd-1111-2222-3333-444455556666", "Grep", 2048, true),
		makeToolCall(3, 20, "aabbccdd-1111-2222-3333-444455556666", "Write", 512, true),
		makeToolCall(4, 30, "bbccddee-1111-2222-3333-444455556666", "Bash", 768, false),
		makeError(5, 40, "aabbccdd-1111-2222-3333-444455556666", "hook", "permission denied"),
	}
}

func makeNEntries(n int) []appdb.TailEntry {
	entries := make([]appdb.TailEntry, n)
	for i := range entries {
		entries[i] = makeToolCall(int64(i+1), int64(i*5), "aabbccdd-1111-2222-3333-444455556666",
			fmt.Sprintf("Tool%02d", i+1), int64(1024*(i+1)), true)
	}
	return entries
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	dir := filepath.Join("testdata")
	rawPath := filepath.Join(dir, name+".golden")
	txtPath := filepath.Join(dir, name+".golden.txt")

	stripped := ansi.Strip(got)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(rawPath, []byte(got), 0644); err != nil {
			t.Fatalf("write %s: %v", rawPath, err)
		}
		if err := os.WriteFile(txtPath, []byte(stripped), 0644); err != nil {
			t.Fatalf("write %s: %v", txtPath, err)
		}
		t.Logf("updated golden files for %s", name)
		return
	}

	wantRaw, err := os.ReadFile(rawPath)
	if err != nil {
		t.Fatalf("golden file missing %s — run UPDATE_GOLDEN=1 go test ./internal/tui/ to generate", rawPath)
	}
	wantTxt, err := os.ReadFile(txtPath)
	if err != nil {
		t.Fatalf("golden file missing %s — run UPDATE_GOLDEN=1 go test ./internal/tui/ to generate", txtPath)
	}

	if got != string(wantRaw) {
		t.Errorf("ANSI output mismatch for %s\n--- want ---\n%s\n--- got ---\n%s\n--- diff (stripped) ---\n%s",
			name, string(wantTxt), stripped, diffLines(string(wantTxt), stripped))
	}
}

func diffLines(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	var out strings.Builder
	max := len(wantLines)
	if len(gotLines) > max {
		max = len(gotLines)
	}
	for i := 0; i < max; i++ {
		var w, g string
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if w != g {
			out.WriteString(fmt.Sprintf("line %d:\n  want: %q\n   got: %q\n", i+1, w, g))
		}
	}
	return out.String()
}

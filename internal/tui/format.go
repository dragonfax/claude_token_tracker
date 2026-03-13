package tui

import (
	"fmt"
	"time"

	appdb "github.com/reshophq/token-tracker/internal/db"
)

func formatBytes(b int64) string {
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/1024/1024)
	case b >= 1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func shortSession(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

func formatEntry(e appdb.TailEntry, showSub bool) string {
	ts := e.RecordedAt.Local().Format("15:04:05")

	if e.IsError {
		return styleError.Render(fmt.Sprintf(" %s  ──── ERROR [%s] %s", ts, e.Source, e.Message))
	}

	tag := "[main]"
	lineStyle := styleMain
	if !e.IsMainContext {
		tag = "[sub] "
		lineStyle = styleSub
	}

	sess := shortSession(e.SessionID)
	bytesStr := styleBytes.Render(fmt.Sprintf("%10s", formatBytes(e.ResponseBytes)))
	line := fmt.Sprintf(" %s  %s  %s  %-40s  %s",
		ts, sess, tag, truncate(e.ToolName, 40), bytesStr)

	return lineStyle.Render(line)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func sinceTime(window string) time.Time {
	now := time.Now()
	switch window {
	case "24h":
		return now.Add(-24 * time.Hour)
	case "48h":
		return now.Add(-48 * time.Hour)
	case "7d":
		return now.Add(-7 * 24 * time.Hour)
	default:
		return time.Time{} // all time
	}
}

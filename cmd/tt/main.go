package main

import (
	"fmt"
	"os"
	"time"

	appdb "github.com/reshophq/token-tracker/internal/db"
	"github.com/reshophq/token-tracker/internal/hook"
	"github.com/reshophq/token-tracker/internal/tui"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "record":
		runRecord()
	case "watch":
		tui.RunWatch()
	case "log":
		tui.RunLog(os.Args[2:])
	case "stats":
		tui.RunStats()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: tt <command>")
	fmt.Fprintln(os.Stderr, "  record   read PostToolUse hook JSON from stdin and record it")
	fmt.Fprintln(os.Stderr, "  watch    live tail of tool calls")
	fmt.Fprintln(os.Stderr, "  log      historical log view")
	fmt.Fprintln(os.Stderr, "  stats    aggregate stats by tool")
}

// runRecord is the hook handler. It must always exit 0.
func runRecord() {
	dbPath, err := appdb.DefaultDBPath()
	if err != nil {
		// Can't even get home dir — nothing we can do
		os.Exit(0)
	}

	db, err := appdb.Open(dbPath)
	if err != nil {
		// DB unavailable — log to stderr (goes to hook log, not Claude) and exit 0
		fmt.Fprintf(os.Stderr, "tt record: open db: %v\n", err)
		os.Exit(0)
	}
	defer db.Close()

	call, raw, parseErr := hook.Parse(os.Stdin)
	if parseErr != nil {
		_ = appdb.InsertError(db, appdb.AppError{
			RecordedAt: time.Now(),
			Source:     "parse",
			Message:    parseErr.Error(),
			RawInput:   string(raw),
		})
		os.Exit(0)
	}

	insertErr := appdb.InsertToolCall(db, appdb.ToolCall{
		RecordedAt:    time.Now(),
		SessionID:     call.SessionID,
		AgentID:       call.AgentID,
		ToolUseID:     call.ToolUseID,
		ToolName:      call.ToolName,
		InputSummary:  call.InputSummary,
		ResponseBytes: call.ResponseBytes,
		IsMainContext: call.IsMainContext,
	})
	if insertErr != nil {
		_ = appdb.InsertError(db, appdb.AppError{
			RecordedAt: time.Now(),
			SessionID:  call.SessionID,
			Source:     "db",
			Message:    insertErr.Error(),
		})
	}

	os.Exit(0)
}

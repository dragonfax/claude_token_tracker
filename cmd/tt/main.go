package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	case "install-hook":
		if err := runInstallHook(); err != nil {
			fmt.Fprintf(os.Stderr, "tt install-hook: %v\n", err)
			os.Exit(1)
		}
	case "uninstall-hook":
		if err := runUninstallHook(); err != nil {
			fmt.Fprintf(os.Stderr, "tt uninstall-hook: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: tt <command>")
	fmt.Fprintln(os.Stderr, "  record       read PostToolUse hook JSON from stdin and record it")
	fmt.Fprintln(os.Stderr, "  watch        live tail of tool calls")
	fmt.Fprintln(os.Stderr, "  log          historical log view")
	fmt.Fprintln(os.Stderr, "  stats        aggregate stats by tool")
	fmt.Fprintln(os.Stderr, "  install-hook   add PostToolUse hook to ~/.claude/settings.json")
	fmt.Fprintln(os.Stderr, "  uninstall-hook remove PostToolUse hook from ~/.claude/settings.json")
}

func runInstallHook() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	execPath, err = filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("resolve absolute path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")

	// Read existing settings or start fresh
	var settings map[string]any
	data, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read settings: %w", err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parse settings.json: %w", err)
		}
	}
	if settings == nil {
		settings = map[string]any{}
	}

	hookEntry := map[string]any{
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": execPath + " record",
			},
		},
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}

	existing, _ := hooks["PostToolUse"].([]any)
	// Check if our hook is already present
	for _, entry := range existing {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		innerHooks, _ := m["hooks"].([]any)
		for _, ih := range innerHooks {
			ihm, ok := ih.(map[string]any)
			if !ok {
				continue
			}
			if ihm["command"] == execPath+" record" {
				fmt.Println("tt: hook already configured in ~/.claude/settings.json")
				return nil
			}
		}
	}

	hooks["PostToolUse"] = append(existing, hookEntry)
	settings["hooks"] = hooks

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("create .claude dir: %w", err)
	}
	if err := os.WriteFile(settingsPath, append(out, '\n'), 0o644); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}

	fmt.Printf("tt: hook installed — %s record\n", execPath)
	fmt.Println("tt: PostToolUse hook added to ~/.claude/settings.json")
	return nil
}

func runUninstallHook() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if os.IsNotExist(err) {
		fmt.Println("tt: no ~/.claude/settings.json found — nothing to remove")
		return nil
	}
	if err != nil {
		return fmt.Errorf("read settings: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parse settings.json: %w", err)
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		fmt.Println("tt: no hooks found in ~/.claude/settings.json — nothing to remove")
		return nil
	}

	existing, _ := hooks["PostToolUse"].([]any)
	if len(existing) == 0 {
		fmt.Println("tt: no PostToolUse hooks found — nothing to remove")
		return nil
	}

	// Filter out any hook entries that contain a tt record command
	var kept []any
	removed := 0
	for _, entry := range existing {
		m, ok := entry.(map[string]any)
		if !ok {
			kept = append(kept, entry)
			continue
		}
		innerHooks, _ := m["hooks"].([]any)
		isTT := false
		for _, ih := range innerHooks {
			ihm, ok := ih.(map[string]any)
			if !ok {
				continue
			}
			cmd, _ := ihm["command"].(string)
			// Match any variant: "tt record", "/path/to/tt record"
			if len(cmd) >= len("tt record") && cmd[len(cmd)-len("tt record"):] == "tt record" {
				isTT = true
				break
			}
		}
		if isTT {
			removed++
		} else {
			kept = append(kept, entry)
		}
	}

	if removed == 0 {
		fmt.Println("tt: no tt hook entries found — nothing to remove")
		return nil
	}

	if len(kept) == 0 {
		delete(hooks, "PostToolUse")
	} else {
		hooks["PostToolUse"] = kept
	}
	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	if err := os.WriteFile(settingsPath, append(out, '\n'), 0o644); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}

	fmt.Printf("tt: removed %d hook entry(s) from ~/.claude/settings.json\n", removed)
	return nil
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

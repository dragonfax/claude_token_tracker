package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS tool_calls (
	id               INTEGER PRIMARY KEY AUTOINCREMENT,
	recorded_at      DATETIME NOT NULL,
	session_id       TEXT NOT NULL,
	agent_id         TEXT,
	tool_use_id      TEXT NOT NULL,
	tool_name        TEXT NOT NULL,
	response_bytes   INTEGER NOT NULL,
	is_main_context  INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tool_calls_recorded_at ON tool_calls(recorded_at);
CREATE INDEX IF NOT EXISTS idx_tool_calls_session ON tool_calls(session_id, recorded_at);

CREATE TABLE IF NOT EXISTS errors (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	recorded_at  DATETIME NOT NULL,
	session_id   TEXT,
	source       TEXT NOT NULL,
	message      TEXT NOT NULL,
	raw_input    TEXT
);
`

func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "token_tracker", "token_tracker.db"), nil
}

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	// modernc/sqlite uses _pragma= for connection-time PRAGMAs
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return db, nil
}

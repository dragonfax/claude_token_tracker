package tui

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	appdb "github.com/reshophq/token-tracker/internal/db"
)

type watchModel struct {
	db        *sql.DB
	entries   []appdb.TailEntry
	lastID    int64
	lastErrID int64
	showSub   bool
	err       string
	width     int
	height    int
}

type watchTickMsg struct{}
type watchDataMsg struct {
	entries   []appdb.TailEntry
	lastID    int64
	lastErrID int64
}

func watchTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return watchTickMsg{}
	})
}

func (m watchModel) fetchNew() tea.Cmd {
	return func() tea.Msg {
		// Watch shows everything live — no time filter
		since := time.Time{}
		entries, err := appdb.TailSince(m.db, m.lastID, m.lastErrID, true, since)
		if err != nil {
			return watchDataMsg{}
		}

		newLastID := m.lastID
		newLastErrID := m.lastErrID
		var filtered []appdb.TailEntry
		for _, e := range entries {
			if e.IsError {
				if e.ID > newLastErrID {
					newLastErrID = e.ID
				}
				filtered = append(filtered, e)
			} else {
				if e.ID > newLastID {
					newLastID = e.ID
				}
				if m.showSub || e.IsMainContext {
					filtered = append(filtered, e)
				}
			}
		}
		return watchDataMsg{entries: filtered, lastID: newLastID, lastErrID: newLastErrID}
	}
}

func (m watchModel) Init() tea.Cmd {
	return tea.Batch(m.fetchNew(), watchTick())
}

func (m watchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "s":
			m.showSub = !m.showSub
		}

	case watchTickMsg:
		return m, tea.Batch(m.fetchNew(), watchTick())

	case watchDataMsg:
		if len(msg.entries) > 0 {
			m.entries = append(m.entries, msg.entries...)
			// Keep last 1000 lines in memory
			if len(m.entries) > 1000 {
				m.entries = m.entries[len(m.entries)-1000:]
			}
		}
		m.lastID = msg.lastID
		m.lastErrID = msg.lastErrID
	}
	return m, nil
}

func (m watchModel) View() string {
	var sb strings.Builder

	subToggle := styleDim.Render("[s: show sub]")
	if m.showSub {
		subToggle = styleSelected.Render("[s: hide sub]")
	}
	header := styleHeader.Render("tt watch — live tool call feed")
	sb.WriteString(fmt.Sprintf("%s  %s\n", header, subToggle))
	sb.WriteString(styleSep.Render(strings.Repeat("─", max(m.width, 80))) + "\n")

	// Show last N lines that fit in terminal
	maxLines := m.height - 4
	if maxLines < 1 {
		maxLines = 10
	}
	start := 0
	if len(m.entries) > maxLines {
		start = len(m.entries) - maxLines
	}
	for _, e := range m.entries[start:] {
		sb.WriteString(formatEntry(e, m.showSub) + "\n")
	}

	// Pad to fill screen
	lines := len(m.entries[start:])
	for i := lines; i < maxLines; i++ {
		sb.WriteString("\n")
	}

	sb.WriteString(styleSep.Render(strings.Repeat("─", max(m.width, 80))) + "\n")
	sb.WriteString(styleDim.Render(" q quit    s toggle subagent calls"))
	return sb.String()
}

func RunWatch() {
	dbPath, err := appdb.DefaultDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	db, err := appdb.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening db: %v\n", err)
		return
	}
	defer db.Close()

	m := watchModel{db: db}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

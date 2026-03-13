package tui

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	appdb "github.com/reshophq/token-tracker/internal/db"
)

var logWindows = []string{"24h", "48h", "7d", "all"}

type logModel struct {
	db          *sql.DB
	entries     []appdb.TailEntry
	lastID      int64
	lastErrID   int64
	showSub     bool
	errorsOnly  bool
	windowIdx   int
	offset      int // scroll offset (top visible line)
	width       int
	height      int
	loaded      bool
}

type logDataMsg struct {
	entries   []appdb.TailEntry
	lastID    int64
	lastErrID int64
}

type logTickMsg struct{}

func logTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return logTickMsg{}
	})
}

func (m logModel) fetchAll() tea.Cmd {
	return func() tea.Msg {
		window := logWindows[m.windowIdx]
		since := sinceTime(window)
		entries, err := appdb.TailSince(m.db, 0, 0, true, since)
		if err != nil {
			return logDataMsg{}
		}
		newLastID := int64(0)
		newLastErrID := int64(0)
		for _, e := range entries {
			if e.IsError {
				if e.ID > newLastErrID {
					newLastErrID = e.ID
				}
			} else {
				if e.ID > newLastID {
					newLastID = e.ID
				}
			}
		}
		return logDataMsg{entries: entries, lastID: newLastID, lastErrID: newLastErrID}
	}
}

func (m logModel) fetchNew() tea.Cmd {
	return func() tea.Msg {
		window := logWindows[m.windowIdx]
		since := sinceTime(window)
		entries, err := appdb.TailSince(m.db, m.lastID, m.lastErrID, true, since)
		if err != nil {
			return logDataMsg{}
		}
		newLastID := m.lastID
		newLastErrID := m.lastErrID
		for _, e := range entries {
			if e.IsError {
				if e.ID > newLastErrID {
					newLastErrID = e.ID
				}
			} else {
				if e.ID > newLastID {
					newLastID = e.ID
				}
			}
		}
		return logDataMsg{entries: entries, lastID: newLastID, lastErrID: newLastErrID}
	}
}

func (m logModel) Init() tea.Cmd {
	return tea.Batch(m.fetchAll(), logTick())
}

func (m logModel) filteredEntries() []appdb.TailEntry {
	var out []appdb.TailEntry
	for _, e := range m.entries {
		if m.errorsOnly && !e.IsError {
			continue
		}
		if !m.showSub && !e.IsMainContext && !e.IsError {
			continue
		}
		out = append(out, e)
	}
	return out
}

func (m logModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "e":
			m.errorsOnly = !m.errorsOnly
			m.offset = 0
		case "tab":
			m.windowIdx = (m.windowIdx + 1) % len(logWindows)
			m.entries = nil
			m.lastID = 0
			m.lastErrID = 0
			m.offset = 0
			m.loaded = false
			return m, m.fetchAll()
		case "up", "k":
			if m.offset > 0 {
				m.offset--
			}
		case "down", "j":
			filtered := m.filteredEntries()
			maxOffset := len(filtered) - (m.height - 4)
			if maxOffset < 0 {
				maxOffset = 0
			}
			if m.offset < maxOffset {
				m.offset++
			}
		case "g":
			m.offset = 0
		case "G":
			filtered := m.filteredEntries()
			m.offset = len(filtered) - (m.height - 4)
			if m.offset < 0 {
				m.offset = 0
			}
		}

	case logTickMsg:
		return m, tea.Batch(m.fetchNew(), logTick())

	case logDataMsg:
		if !m.loaded {
			m.entries = msg.entries
			m.loaded = true
			// Start at bottom
			filtered := m.filteredEntries()
			m.offset = len(filtered) - (m.height - 4)
			if m.offset < 0 {
				m.offset = 0
			}
		} else if len(msg.entries) > 0 {
			m.entries = append(m.entries, msg.entries...)
		}
		m.lastID = msg.lastID
		m.lastErrID = msg.lastErrID
	}
	return m, nil
}

func (m logModel) View() string {
	var sb strings.Builder

	window := logWindows[m.windowIdx]
	var tabStr strings.Builder
	for i, w := range logWindows {
		tabStr.WriteString(tabLabel(w, i == m.windowIdx))
		if i < len(logWindows)-1 {
			tabStr.WriteString(" ")
		}
	}

	subToggle := styleDim.Render("s:sub")
	if m.showSub {
		subToggle = styleSelected.Render("s:sub")
	}
	errToggle := styleDim.Render("e:errors")
	if m.errorsOnly {
		errToggle = styleSelected.Render("e:errors-only")
	}
	_ = window

	header := styleHeader.Render("tt log — historical tool calls")
	sb.WriteString(fmt.Sprintf("%s  %s  %s %s\n", header, tabStr.String(), subToggle, errToggle))
	sb.WriteString(styleSep.Render(strings.Repeat("─", max(m.width, 80))) + "\n")

	filtered := m.filteredEntries()
	maxLines := m.height - 4
	if maxLines < 1 {
		maxLines = 10
	}

	end := m.offset + maxLines
	if end > len(filtered) {
		end = len(filtered)
	}

	visible := filtered
	if m.offset < len(filtered) {
		visible = filtered[m.offset:end]
	} else {
		visible = nil
	}

	for _, e := range visible {
		sb.WriteString(formatEntry(e, m.showSub) + "\n")
	}
	for i := len(visible); i < maxLines; i++ {
		sb.WriteString("\n")
	}

	sb.WriteString(styleSep.Render(strings.Repeat("─", max(m.width, 80))) + "\n")
	pos := fmt.Sprintf("%d/%d", m.offset+len(visible), len(filtered))
	sb.WriteString(styleDim.Render(fmt.Sprintf(" q quit    tab window    s sub    e errors    ↑↓/jk scroll    g/G top/bot    %s", pos)))
	return sb.String()
}

func RunLog(args []string) {
	dbPath, err := appdb.DefaultDBPath()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	db, err := appdb.Open(dbPath)
	if err != nil {
		fmt.Printf("error opening db: %v\n", err)
		return
	}
	defer db.Close()

	m := logModel{db: db}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

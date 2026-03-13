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

var statsWindows = []string{"24h", "48h", "7d"}

type statsModel struct {
	db        *sql.DB
	rows      []appdb.AggregateRow
	windowIdx int
	width     int
	height    int
	err       string
}

type statsDataMsg struct {
	rows []appdb.AggregateRow
	err  string
}

type statsTickMsg struct{}

func statsTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return statsTickMsg{}
	})
}

func (m statsModel) fetch() tea.Cmd {
	return func() tea.Msg {
		window := statsWindows[m.windowIdx]
		since := sinceTime(window)
		rows, err := appdb.Aggregate(m.db, since)
		if err != nil {
			return statsDataMsg{err: err.Error()}
		}
		return statsDataMsg{rows: rows}
	}
}

func (m statsModel) Init() tea.Cmd {
	return tea.Batch(m.fetch(), statsTick())
}

func (m statsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.windowIdx = (m.windowIdx + 1) % len(statsWindows)
			return m, m.fetch()
		}

	case statsTickMsg:
		return m, tea.Batch(m.fetch(), statsTick())

	case statsDataMsg:
		m.rows = msg.rows
		m.err = msg.err
	}
	return m, nil
}

func (m statsModel) View() string {
	var sb strings.Builder

	var tabStr strings.Builder
	for i, w := range statsWindows {
		tabStr.WriteString(tabLabel(w, i == m.windowIdx))
		if i < len(statsWindows)-1 {
			tabStr.WriteString(" ")
		}
	}

	header := styleHeader.Render("tt stats — top tools by bytes into main context")
	sb.WriteString(fmt.Sprintf("%s   %s\n", header, tabStr.String()))
	sb.WriteString(styleSep.Render(strings.Repeat("─", max(m.width, 80))) + "\n")

	if m.err != "" {
		sb.WriteString(styleError.Render(" error: "+m.err) + "\n")
		return sb.String()
	}

	// Column header
	sb.WriteString(styleDim.Render(fmt.Sprintf(" %-42s  %8s  %10s  %10s\n",
		"Tool", "Calls", "Total", "Avg/call")))
	sb.WriteString(styleDim.Render(" "+strings.Repeat("─", max(m.width-2, 76))) + "\n")

	if len(m.rows) == 0 {
		sb.WriteString(styleDim.Render(" no data for this window\n"))
	}

	for _, r := range m.rows {
		line := fmt.Sprintf(" %-42s  %8d  %10s  %10s",
			truncate(r.ToolName, 42),
			r.Calls,
			formatBytes(r.TotalBytes),
			formatBytes(r.AvgBytes),
		)
		sb.WriteString(styleMain.Render(line) + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(styleSep.Render(strings.Repeat("─", max(m.width, 80))) + "\n")
	sb.WriteString(styleDim.Render(" q quit    tab change window    (main context only, auto-refreshes)"))
	return sb.String()
}

func RunStats() {
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

	m := statsModel{db: db}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

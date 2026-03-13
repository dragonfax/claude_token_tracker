package tui

import "github.com/charmbracelet/lipgloss"

var (
	styleHeader   = lipgloss.NewStyle().Bold(true)
	styleDim      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleError    = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	styleMain     = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	styleSub      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleSelected = lipgloss.NewStyle().Bold(true).Underline(true)
	styleNormal   = lipgloss.NewStyle()
	styleSep      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleBytes    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
)

func tabLabel(label string, active bool) string {
	if active {
		return styleSelected.Render("[" + label + "]")
	}
	return styleNormal.Render(" " + label + " ")
}

package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/eskylake/network-tracker/internal/check"
)

var (
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	tabStyle         = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("244"))
	activeTabStyle   = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24"))
	selectedStyle    = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	mutedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	sectionStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	tableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	detailStyle      = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("238")).Padding(1, 2)
	ipPanelStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("31")).Padding(1, 2)
	inputStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24")).Padding(0, 1)
	inputLockedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Background(lipgloss.Color("238")).Padding(0, 1)
	statusBarStyle   = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("235"))
	statusBarLabel   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	wifiSignalStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
)

func statusStyle(status check.Status) lipgloss.Style {
	switch status {
	case check.StatusOK:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	case check.StatusWarning:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	case check.StatusError:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Bold(true)
	}
}

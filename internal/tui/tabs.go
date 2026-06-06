package tui

import "github.com/eskylake/network-tracker/internal/check"

// Tab identifies a dashboard view.
type Tab int

const (
	TabOverview Tab = iota
	TabVPN
	TabConnectivity
	TabScan
	TabRoutes
	TabDocker
	TabPing
	TabLogs
)

func (t Tab) String() string {
	switch t {
	case TabOverview:
		return "Overview"
	case TabVPN:
		return "VPN"
	case TabConnectivity:
		return "Connectivity"
	case TabScan:
		return "Scan"
	case TabRoutes:
		return "Routes"
	case TabDocker:
		return "Docker"
	case TabPing:
		return "Ping"
	case TabLogs:
		return "Logs"
	default:
		return "Overview"
	}
}

func (t Tab) Category() string {
	switch t {
	case TabVPN:
		return check.CategoryVPN
	case TabConnectivity:
		return check.CategoryConnectivity
	case TabRoutes:
		return check.CategoryRoutes
	case TabDocker:
		return check.CategoryDocker
	default:
		return ""
	}
}

var allTabs = []Tab{
	TabOverview,
	TabVPN,
	TabConnectivity,
	TabScan,
	TabRoutes,
	TabDocker,
	TabPing,
	TabLogs,
}

func tabAt(index int) Tab {
	n := len(allTabs)
	index = ((index % n) + n) % n
	return allTabs[index]
}

func tabIndex(t Tab) int {
	for i, tab := range allTabs {
		if tab == t {
			return i
		}
	}
	return 0
}

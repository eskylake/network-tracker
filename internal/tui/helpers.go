package tui

import (
	"strings"
	"time"

	"github.com/eskylake/network-tracker/internal/check"
)

func prependLog(logs []string, message string) []string {
	return prependBounded(logs, timestamped(message), 200)
}

func mergeResults(current, updates []check.Result) []check.Result {
	return check.MergeResults(current, updates)
}

func prependBounded(values []string, value string, limit int) []string {
	values = append([]string{value}, values...)
	if len(values) > limit {
		return values[:limit]
	}
	return values
}

func timestamped(message string) string {
	return time.Now().Format("15:04:05") + " " + message
}

func snapshotFromResults(results []check.Result) networkSnapshot {
	var snap networkSnapshot
	for _, result := range results {
		switch result.Name {
		case check.NameRoutes:
			snap.DefaultRoute = result.Summary
		case check.NameDNSConfig:
			snap.DNSServers = result.Summary
		}
	}
	return snap
}

func (s networkSnapshot) String() string {
	return strings.TrimSpace(s.DefaultRoute + " dns=" + s.DNSServers)
}

func (m model) filteredResults() []check.Result {
	category := m.tab.Category()
	if category == "" {
		return m.results
	}
	out := make([]check.Result, 0, len(m.results))
	for _, result := range m.results {
		if result.Name == check.NameWiFi {
			continue
		}
		if result.Category == category {
			out = append(out, result)
		}
	}
	return out
}

func findResult(results []check.Result, name string) (check.Result, bool) {
	for _, result := range results {
		if result.Name == name {
			return result, true
		}
	}
	return check.Result{}, false
}

func hasResult(results []check.Result, name string) bool {
	_, ok := findResult(results, name)
	return ok
}

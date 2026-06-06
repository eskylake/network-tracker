package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/eskylake/network-tracker/internal/check"
	"github.com/eskylake/network-tracker/internal/parse"
)

const (
	wifiColInUse    = 8
	wifiColSSID     = 26
	wifiColSignal   = 10
	wifiColSecurity = 14
	wifiColFreq     = 14
)

func (m model) View() string {
	var header strings.Builder
	header.WriteString(titleStyle.Render("network-tracker"))
	header.WriteString("  ")
	if m.loading {
		header.WriteString(m.spinner.View() + " checking")
	} else if !m.lastRefresh.IsZero() {
		header.WriteString(mutedStyle.Render("last refresh " + m.lastRefresh.Format("15:04:05")))
	}
	header.WriteString("\n")
	header.WriteString(mutedStyle.Render("Read-only diagnostics for VPN, DNS, routes, public IP, and Docker."))
	header.WriteString("\n")
	header.WriteString(m.renderStatusBar())
	header.WriteString("\n\n")
	header.WriteString(m.renderTabs())

	footerText := "tab/shift+tab switch - r refresh - enter details - up/down select - q quit"
	switch m.tab {
	case TabScan:
		footerText = "r rescan - up/down select - enter details - tab switch - q quit"
	case TabPing:
		footerText = "type host/IP - enter start/stop - backspace edit - ctrl+u clear - tab switch - q quit"
	case TabLogs:
		footerText = "up/down scroll logs - tab switch - r refresh - q quit"
	}
	footer := helpStyle.Render(footerText)
	availableBodyLines := m.height - lineCount(header.String()) - lineCount(footer) - 4
	if availableBodyLines < 6 {
		availableBodyLines = 6
	}

	body := ""
	switch m.tab {
	case TabOverview:
		body = m.renderOverview()
	case TabVPN, TabConnectivity, TabRoutes, TabDocker:
		body = m.renderResults()
	case TabScan:
		body = m.renderWiFiScan(availableBodyLines)
	case TabPing:
		body = m.renderPing()
	case TabLogs:
		body = m.renderLogs()
	}
	if m.tab != TabScan {
		body = clampLines(body, availableBodyLines)
	}

	return header.String() + "\n\n" + body + "\n\n" + footer
}

func (m model) renderStatusBar() string {
	label := statusBarLabel.Render("Wi-Fi")
	if m.loading && !hasResult(m.results, check.NameWiFi) {
		return statusBarStyle.Render(label + "  " + mutedStyle.Render("checking…"))
	}
	result, ok := findResult(m.results, check.NameWiFi)
	if !ok {
		return statusBarStyle.Render(label + "  " + mutedStyle.Render("—"))
	}
	return statusBarStyle.Render(label + "  " + m.renderWiFiStatusValue(result))
}

func (m model) renderWiFiStatusValue(result check.Result) string {
	ssid := statusStyle(result.Status).Render(result.Summary)
	signal := wifiSignalLabel(result, m.wifiNetworks)
	if signal == "" {
		return ssid
	}
	return ssid + "  " + wifiSignalStyle.Render(signal)
}

func wifiSignalLabel(result check.Result, networks []parse.WiFiNetwork) string {
	if signal := strings.TrimSpace(parseDetails(result.Details)["Signal"]); signal != "" {
		return signal
	}
	for _, network := range networks {
		if network.InUse {
			return network.SignalLabel()
		}
	}
	return ""
}

func (m model) renderTabs() string {
	parts := make([]string, 0, len(allTabs))
	for _, tab := range allTabs {
		label := tab.String()
		if tab == m.tab {
			parts = append(parts, activeTabStyle.Render(label))
		} else {
			parts = append(parts, tabStyle.Render(label))
		}
	}
	return strings.Join(parts, " ")
}

func (m model) renderOverview() string {
	if len(m.results) == 0 {
		return sectionStyle.Render("Status") + "\n" + mutedStyle.Render("No results yet. Press r to refresh.")
	}

	counts := map[check.Status]int{}
	for _, result := range m.results {
		counts[result.Status]++
	}

	var lines []string
	lines = append(lines, sectionStyle.Render("Status"))
	lines = append(lines, fmt.Sprintf("%s ok   %s warning   %s error   %s unknown",
		statusStyle(check.StatusOK).Render(fmt.Sprint(counts[check.StatusOK])),
		statusStyle(check.StatusWarning).Render(fmt.Sprint(counts[check.StatusWarning])),
		statusStyle(check.StatusError).Render(fmt.Sprint(counts[check.StatusError])),
		statusStyle(check.StatusUnknown).Render(fmt.Sprint(counts[check.StatusUnknown])),
	))

	if result, ok := findResult(m.results, check.NamePublicIP); ok {
		lines = append(lines, "", renderIPPanel(result))
	}

	var summary []check.Result
	for _, name := range []string{check.NameXVPN, check.NameV2RayAService, check.NameRoutes, check.NameDNSConfig, check.NameDockerNetwork} {
		if result, ok := findResult(m.results, name); ok {
			summary = append(summary, result)
		}
	}
	lines = append(lines, "", sectionStyle.Render("Quick Checks"), renderResultTable(summary, -1))
	return strings.Join(lines, "\n")
}

func (m model) renderResults() string {
	results := m.filteredResults()
	if len(results) == 0 {
		return mutedStyle.Render("No checks for this tab yet.")
	}
	lines := []string{sectionStyle.Render(m.tab.String()), renderResultTable(results, m.selected)}
	if m.detail && m.selected >= 0 && m.selected < len(results) {
		details := results[m.selected].Details
		if details == "" {
			details = "No details."
		}
		if results[m.selected].Name == check.NamePublicIP {
			lines = append(lines, "", renderIPPanel(results[m.selected]))
		} else {
			lines = append(lines, "", detailStyle.Render(details))
		}
	}
	return strings.Join(lines, "\n")
}

func (m model) renderWiFiScan(availableLines int) string {
	meta := []string{"nearby networks"}
	if m.wifiScanning {
		meta = append(meta, m.spinner.View()+" scanning")
	} else if !m.wifiScanAt.IsZero() {
		meta = append(meta, "last scan "+m.wifiScanAt.Format("15:04:05"))
		if m.wifiScanSource != "" {
			meta = append(meta, "via "+m.wifiScanSource)
		}
	}
	lines := []string{sectionStyle.Render("Scan") + "  " + mutedStyle.Render(strings.Join(meta, "  "))}

	if m.wifiScanning && len(m.wifiNetworks) == 0 {
		lines = append(lines, mutedStyle.Render("Scanning for nearby networks…"))
		return strings.Join(lines, "\n")
	}
	if m.wifiScanErr != nil && len(m.wifiNetworks) == 0 {
		lines = append(lines, statusStyle(check.StatusError).Render("scan failed"))
		lines = append(lines, detailStyle.Render(m.wifiScanErr.Error()))
		lines = append(lines, mutedStyle.Render("Press r to retry. Requires nmcli or iw and a wireless interface."))
		return strings.Join(lines, "\n")
	}
	if len(m.wifiNetworks) == 0 {
		lines = append(lines, mutedStyle.Render("No networks found. Press r to scan again."))
		return strings.Join(lines, "\n")
	}

	showDetail := m.detail && m.selected >= 0 && m.selected < len(m.wifiNetworks)
	maxRows := availableLines - lineCount(strings.Join(lines, "\n")) - 2
	if showDetail {
		maxRows -= 9
	}
	if maxRows < 3 {
		maxRows = 3
	}

	start, end := wifiVisibleWindow(m.selected, len(m.wifiNetworks), maxRows)
	visible := m.wifiNetworks[start:end]
	position := mutedStyle.Render(fmt.Sprintf("showing %d-%d of %d", start+1, end, len(m.wifiNetworks)))
	lines = append(lines, position)
	lines = append(lines, renderWiFiNetworkTable(visible, m.selected, start, m.wifiTableWidth()))

	if showDetail {
		network := m.wifiNetworks[m.selected]
		details := strings.Join([]string{
			"SSID: " + network.SSID,
			"BSSID: " + network.BSSID,
			"Signal: " + network.SignalLabel(),
			"Security: " + fallback(network.Security, "-"),
			"Frequency: " + fallback(network.Freq, "-"),
			"In use: " + boolLabel(network.InUse),
		}, "\n")
		lines = append(lines, "", detailStyle.Render(details))
	}
	return strings.Join(lines, "\n")
}

func (m model) wifiTableWidth() int {
	width := 2 + wifiColInUse + 1 + wifiColSSID + 1 + wifiColSignal + 1 + wifiColSecurity + 1 + wifiColFreq
	if m.width > 0 && width > m.width-2 {
		return max(48, m.width-2)
	}
	return width
}

func wifiVisibleWindow(selected, total, maxVisible int) (int, int) {
	if total <= maxVisible {
		return 0, total
	}
	start := selected - maxVisible/2
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
	}
	return start, end
}

func renderWiFiNetworkTable(networks []parse.WiFiNetwork, selected, offset, tableWidth int) string {
	var lines []string
	header := fmt.Sprintf("%s %s %s %s %s",
		pad("IN USE", wifiColInUse),
		pad("SSID", wifiColSSID),
		pad("SIGNAL", wifiColSignal),
		pad("SECURITY", wifiColSecurity),
		pad("FREQ", wifiColFreq),
	)
	lines = append(lines, tableHeaderStyle.Render(" "+header))
	lines = append(lines, mutedStyle.Render(strings.Repeat("-", tableWidth)))
	for i, network := range networks {
		rowIndex := offset + i
		cursor := " "
		if rowIndex == selected {
			cursor = ">"
		}
		inUse := pad(" ", wifiColInUse)
		if network.InUse {
			inUse = statusStyle(check.StatusOK).Render(pad("*", wifiColInUse))
		}
		line := fmt.Sprintf("%s %s %s %s %s %s",
			cursor,
			inUse,
			pad(truncate(network.SSID, wifiColSSID), wifiColSSID),
			pad(network.SignalLabel(), wifiColSignal),
			pad(truncate(fallback(network.Security, "-"), wifiColSecurity), wifiColSecurity),
			pad(truncate(fallback(network.Freq, "-"), wifiColFreq), wifiColFreq),
		)
		if rowIndex == selected {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func boolLabel(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func (m model) renderPing() string {
	state := statusStyle(check.StatusWarning).Render("STOPPED")
	if m.pingRunning {
		state = statusStyle(check.StatusOK).Render("RUNNING")
	}
	input := inputStyle.Render(m.pingTarget)
	if m.pingRunning {
		input = inputLockedStyle.Render(m.pingTarget)
	}
	lines := []string{
		sectionStyle.Render("Ping"),
		"Target  " + input + "  " + state,
		mutedStyle.Render("Press enter to start. Press enter again to stop. Edit target while stopped."),
		"",
		sectionStyle.Render("Results"),
	}
	if len(m.pingResults) == 0 {
		lines = append(lines, mutedStyle.Render("No ping results yet."))
	} else {
		limit := min(len(m.pingResults), 40)
		lines = append(lines, m.pingResults[:limit]...)
	}
	return strings.Join(lines, "\n")
}

func (m model) renderLogs() string {
	if len(m.logs) == 0 {
		return mutedStyle.Render("No log entries.")
	}
	visible := 30
	if m.height > 8 {
		visible = m.height - 8
	}
	if visible < 5 {
		visible = 5
	}
	start := min(m.logOffset, max(0, len(m.logs)-1))
	end := min(len(m.logs), start+visible)
	position := fmt.Sprintf("showing %d-%d of %d", start+1, end, len(m.logs))
	return sectionStyle.Render("Logs") + "  " + mutedStyle.Render(position) + "\n" + strings.Join(m.logs[start:end], "\n")
}

func renderResultTable(results []check.Result, selected int) string {
	if len(results) == 0 {
		return mutedStyle.Render("No checks to show.")
	}
	var lines []string
	lines = append(lines, tableHeaderStyle.Render("  "+pad("STATUS", 10)+" "+pad("CHECK", 20)+" "+pad("SUMMARY", 52)+" "+pad("TIME", 8)))
	lines = append(lines, mutedStyle.Render(strings.Repeat("-", 96)))
	for i, result := range results {
		cursor := " "
		if i == selected {
			cursor = ">"
		}
		status := statusStyle(result.Status).Render(pad(strings.ToUpper(string(result.Status)), 10))
		line := fmt.Sprintf("%s %s %s %s %s",
			cursor,
			status,
			pad(result.Name, 20),
			pad(truncate(result.Summary, 52), 52),
			pad(result.Duration.Round(1_000_000).String(), 8),
		)
		if i == selected {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func renderIPPanel(result check.Result) string {
	if result.Status != check.StatusOK {
		details := strings.TrimSpace(result.Details)
		if details == "" && result.Err != nil {
			details = result.Err.Error()
		}
		if details == "" {
			details = result.Summary
		}
		return sectionStyle.Render("Public IP Details") + "\n" + ipPanelStyle.Render(statusStyle(result.Status).Render(strings.ToUpper(string(result.Status)))+" "+result.Summary+"\n"+details)
	}
	values := parseDetails(result.Details)
	rows := []string{
		kv("IP", values["IP"]),
		kv("Hostname", values["Hostname"]),
		kv("City", values["City"]),
		kv("Region", values["Region"]),
		kv("Country", values["Country"]),
		kv("Latitude", values["Latitude"]),
		kv("Longitude", values["Longitude"]),
		kv("Location", values["Location"]),
		kv("Postal", values["Postal"]),
		kv("Org", values["Org"]),
		kv("Timezone", values["Timezone"]),
		kv("Provider", values["Provider"]),
	}
	return sectionStyle.Render("Public IP Details") + "\n" + ipPanelStyle.Render(strings.Join(rows, "\n"))
}

func kv(key, value string) string {
	if strings.TrimSpace(value) == "" {
		value = "-"
	}
	return mutedStyle.Render(pad(key, 10)) + " " + value
}

func parseDetails(details string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(details, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return values
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if runewidth.StringWidth(value) <= limit {
		return value
	}
	if limit <= 3 {
		return runewidth.Truncate(value, limit, "")
	}
	return runewidth.Truncate(value, limit-3, "") + "..."
}

func pad(value string, width int) string {
	value = truncate(value, width)
	if lipgloss.Width(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-lipgloss.Width(value))
}

func lineCount(value string) int {
	if value == "" {
		return 0
	}
	return len(strings.Split(value, "\n"))
}

func clampLines(value string, limit int) string {
	lines := strings.Split(value, "\n")
	if len(lines) <= limit {
		return value
	}
	if limit <= 1 {
		return mutedStyle.Render("content truncated")
	}
	visible := append([]string{}, lines[:limit-1]...)
	visible = append(visible, mutedStyle.Render(fmt.Sprintf("... %d more line(s). Open this tab in a taller terminal or use details view.", len(lines)-limit+1)))
	return strings.Join(visible, "\n")
}

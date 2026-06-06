package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.tab = tabAt(tabIndex(m.tab) + 1)
			m.selected = 0
			m.detail = false
			return m.startWiFiScanIfNeeded()
		case "shift+tab":
			m.tab = tabAt(tabIndex(m.tab) - 1)
			m.selected = 0
			m.detail = false
			return m.startWiFiScanIfNeeded()
		case "r":
			if m.tab == TabScan {
				m.wifiScanning = true
				m.logs = prependLog(m.logs, "wifi scan")
				return m, tea.Batch(m.spinner.Tick, m.scanWiFi())
			}
			m.loading = true
			m.logs = prependLog(m.logs, "manual refresh")
			return m, tea.Batch(m.spinner.Tick, m.refresh())
		case "up", "k":
			if m.tab == TabLogs {
				if m.logOffset < max(0, len(m.logs)-1) {
					m.logOffset++
				}
			} else if m.tab == TabScan {
				if m.selected > 0 {
					m.selected--
				}
			} else if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down", "j":
			if m.tab == TabLogs {
				if m.logOffset > 0 {
					m.logOffset--
				}
			} else if m.tab == TabScan {
				if m.selected < len(m.wifiNetworks)-1 {
					m.selected++
				}
			} else if m.selected < len(m.filteredResults())-1 {
				m.selected++
			}
			return m, nil
		case "enter":
			if m.tab == TabPing {
				if m.pingRunning {
					m.pingRunning = false
					m.pingResults = prependBounded(m.pingResults, timestamped("stopped ping"), 200)
					return m, nil
				}
				target := strings.TrimSpace(m.pingTarget)
				if target == "" {
					m.pingResults = prependBounded(m.pingResults, timestamped("enter a host or IP before starting"), 200)
					return m, nil
				}
				m.pingTarget = target
				m.pingRunning = true
				m.pingResults = prependBounded(m.pingResults, timestamped("started ping "+target), 200)
				return m, pingOnce(target)
			}
			m.detail = !m.detail
			return m, nil
		case "backspace", "ctrl+h":
			if m.tab == TabPing && !m.pingRunning && len(m.pingTarget) > 0 {
				m.pingTarget = m.pingTarget[:len(m.pingTarget)-1]
			}
			return m, nil
		case "ctrl+u":
			if m.tab == TabPing && !m.pingRunning {
				m.pingTarget = ""
			}
			return m, nil
		}
		if m.tab == TabPing && !m.pingRunning && len(msg.Runes) > 0 {
			for _, r := range msg.Runes {
				if r > 32 && r < 127 {
					m.pingTarget += string(r)
				}
			}
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tickMsg:
		m.loading = true
		return m, tea.Batch(m.refresh(), tick(m.cfg.RefreshInterval.Duration))
	case publicIPTickMsg:
		return m, tea.Batch(m.refreshPublicIP(), publicIPTick(m.cfg.PublicIPRefreshInterval.Duration))
	case pingMsg:
		if msg.target != strings.TrimSpace(m.pingTarget) {
			return m, nil
		}
		m.pingResults = prependBounded(m.pingResults, msg.line, 200)
		if m.pingRunning {
			return m, tea.Tick(time.Second, func(time.Time) tea.Msg { return runPing(msg.target) })
		}
		return m, nil
	case refreshMsg:
		m.loading = false
		m.results = mergeResults(m.results, msg.results)
		m.lastRefresh = time.Now()
		m.err = msg.err
		if m.lastSnap != (networkSnapshot{}) && m.lastSnap != msg.snap {
			m.logs = prependLog(m.logs, fmt.Sprintf("route/dns changed: %s", msg.snap.String()))
		}
		m.lastSnap = msg.snap
		m.logs = prependLog(m.logs, fmt.Sprintf("refreshed %d checks", len(msg.results)))
		if m.selected >= len(m.filteredResults()) {
			m.selected = max(0, len(m.filteredResults())-1)
		}
		return m, nil
	case publicIPRefreshMsg:
		m.results = mergeResults(m.results, msg.results)
		m.logs = prependLog(m.logs, "refreshed public ip")
		if msg.err != nil {
			m.err = msg.err
		}
		if m.selected >= len(m.filteredResults()) {
			m.selected = max(0, len(m.filteredResults())-1)
		}
		return m, nil
	case wifiScanMsg:
		m.wifiScanning = false
		m.wifiNetworks = msg.networks
		m.wifiScanSource = msg.source
		m.wifiScanErr = msg.err
		m.wifiScanAt = time.Now()
		if msg.err != nil {
			m.logs = prependLog(m.logs, "wifi scan failed: "+msg.err.Error())
		} else {
			m.logs = prependLog(m.logs, fmt.Sprintf("wifi scan found %d network(s) via %s", len(msg.networks), msg.source))
		}
		if m.selected >= len(m.wifiNetworks) {
			m.selected = max(0, len(m.wifiNetworks)-1)
		}
		return m, nil
	}
	return m, nil
}

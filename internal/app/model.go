package app

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eskylake/network-tracker/internal/checks"
	"github.com/eskylake/network-tracker/internal/config"
)

type model struct {
	cfg      config.Config
	spinner  spinner.Model
	tabs     []string
	tab      int
	selected int
	detail   bool
	loading  bool
	width    int
	height   int

	results     []checks.Result
	logs        []string
	logOffset   int
	pingTarget  string
	pingRunning bool
	pingResults []string
	lastRefresh time.Time
	lastSnap    networkSnapshot
	err         error

	wifiScanning   bool
	wifiNetworks   []checks.WiFiNetwork
	wifiScanErr    error
	wifiScanSource string
	wifiScanAt     time.Time
}

type refreshMsg struct {
	results []checks.Result
	snap    networkSnapshot
	err     error
}

type tickMsg time.Time
type publicIPTickMsg time.Time

type publicIPRefreshMsg struct {
	results []checks.Result
	err     error
}

type pingMsg struct {
	target string
	line   string
}

type wifiScanMsg struct {
	networks []checks.WiFiNetwork
	source   string
	err      error
}

type networkSnapshot struct {
	DefaultRoute string
	DNSServers   string
}

func New(cfg config.Config) tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = mutedStyle
	return model{
		cfg:        cfg,
		spinner:    s,
		tabs:       []string{"Overview", "VPN", "Connectivity", "Scan", "Routes", "Docker", "Ping", "Logs"},
		logs:       []string{"network-tracker started"},
		pingTarget: "8.8.8.8",
		loading:    true,
		width:      100,
		height:     30,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.refresh(),
		m.refreshPublicIP(),
		tick(m.cfg.RefreshInterval.Duration),
		publicIPTick(m.cfg.PublicIPRefreshInterval.Duration),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		currentTab := m.tabs[m.tab]
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.tab = (m.tab + 1) % len(m.tabs)
			m.selected = 0
			m.detail = false
			return m.startWiFiScanIfNeeded()
		case "shift+tab":
			m.tab--
			if m.tab < 0 {
				m.tab = len(m.tabs) - 1
			}
			m.selected = 0
			m.detail = false
			return m.startWiFiScanIfNeeded()
		case "r":
			if currentTab == "Scan" {
				m.wifiScanning = true
				m.logs = prependLog(m.logs, "wifi scan")
				return m, tea.Batch(m.spinner.Tick, m.scanWiFi())
			}
			m.loading = true
			m.logs = prependLog(m.logs, "manual refresh")
			return m, tea.Batch(m.spinner.Tick, m.refresh())
		case "up", "k":
			if currentTab == "Logs" {
				if m.logOffset < max(0, len(m.logs)-1) {
					m.logOffset++
				}
			} else if currentTab == "Scan" {
				if m.selected > 0 {
					m.selected--
				}
			} else if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down", "j":
			if currentTab == "Logs" {
				if m.logOffset > 0 {
					m.logOffset--
				}
			} else if currentTab == "Scan" {
				if m.selected < len(m.wifiNetworks)-1 {
					m.selected++
				}
			} else if m.selected < len(m.filteredResults())-1 {
				m.selected++
			}
			return m, nil
		case "enter":
			if currentTab == "Ping" {
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
			if currentTab == "Ping" && !m.pingRunning && len(m.pingTarget) > 0 {
				m.pingTarget = m.pingTarget[:len(m.pingTarget)-1]
			}
			return m, nil
		case "ctrl+u":
			if currentTab == "Ping" && !m.pingRunning {
				m.pingTarget = ""
			}
			return m, nil
		}
		if currentTab == "Ping" && !m.pingRunning && len(msg.Runes) > 0 {
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

func (m model) startWiFiScanIfNeeded() (tea.Model, tea.Cmd) {
	if m.tabs[m.tab] == "Scan" && !m.wifiScanning && len(m.wifiNetworks) == 0 {
		m.wifiScanning = true
		return m, m.scanWiFi()
	}
	return m, nil
}

func (m model) refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		runner := checks.Runner{Timeout: m.cfg.CheckTimeout.Duration, Checks: checks.BuildWithoutPublicIP(m.cfg)}
		results := runner.Run(ctx)
		return refreshMsg{results: results, snap: snapshotFromResults(results)}
	}
}

func (m model) refreshPublicIP() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		runner := checks.Runner{Timeout: m.cfg.CheckTimeout.Duration, Checks: checks.BuildPublicIP(m.cfg)}
		return publicIPRefreshMsg{results: runner.Run(ctx)}
	}
}

func (m model) scanWiFi() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		networks, source, err := checks.ScanWiFiNetworks(ctx)
		return wifiScanMsg{networks: networks, source: source, err: err}
	}
}

func tick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func publicIPTick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg { return publicIPTickMsg(t) })
}

func pingOnce(target string) tea.Cmd {
	return func() tea.Msg { return runPing(target) }
}

func runPing(target string) tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	start := time.Now()
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "2", target)
	raw, err := cmd.CombinedOutput()
	duration := time.Since(start).Round(time.Millisecond)
	line := compactPingOutput(string(raw), err, duration)
	return pingMsg{target: target, line: timestamped(line)}
}

func compactPingOutput(output string, err error, duration time.Duration) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "bytes from") || strings.Contains(line, "time=") {
			return line
		}
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "PING ") {
			if err != nil {
				return "failed in " + duration.String() + ": " + line
			}
			return line
		}
	}
	if err != nil {
		return "failed in " + duration.String() + ": " + err.Error()
	}
	return "ok in " + duration.String()
}

func snapshotFromResults(results []checks.Result) networkSnapshot {
	var snap networkSnapshot
	for _, result := range results {
		if result.Name == "routes" {
			snap.DefaultRoute = result.Summary
		}
		if result.Name == "dns config" {
			snap.DNSServers = result.Summary
		}
	}
	return snap
}

func (s networkSnapshot) String() string {
	return strings.TrimSpace(s.DefaultRoute + " dns=" + s.DNSServers)
}

func prependLog(logs []string, message string) []string {
	return prependBounded(logs, timestamped(message), 200)
}

func mergeResults(current, updates []checks.Result) []checks.Result {
	merged := append([]checks.Result{}, current...)
	for _, update := range updates {
		replaced := false
		for i, result := range merged {
			if result.Name == update.Name {
				merged[i] = update
				replaced = true
				break
			}
		}
		if !replaced {
			merged = append(merged, update)
		}
	}
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Category == merged[j].Category {
			return merged[i].Name < merged[j].Name
		}
		return merged[i].Category < merged[j].Category
	})
	return merged
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

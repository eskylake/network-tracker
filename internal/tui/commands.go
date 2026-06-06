package tui

import (
	"context"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/eskylake/network-tracker/internal/check"
	"github.com/eskylake/network-tracker/internal/parse"
)

type refreshMsg struct {
	results []check.Result
	snap    networkSnapshot
	err     error
}

type tickMsg time.Time
type publicIPTickMsg time.Time

type publicIPRefreshMsg struct {
	results []check.Result
	err     error
}

type pingMsg struct {
	target string
	line   string
}

type wifiScanMsg struct {
	networks []parse.WiFiNetwork
	source   string
	err      error
}

const pingTimeout = 4 * time.Second
const wifiScanTimeout = 15 * time.Second

func (m model) refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		runner := check.Runner{Timeout: m.cfg.CheckTimeout.Duration, Checks: check.BuildWithoutPublicIP(m.cfg)}
		results := runner.Run(ctx)
		return refreshMsg{results: results, snap: snapshotFromResults(results)}
	}
}

func (m model) refreshPublicIP() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		runner := check.Runner{Timeout: m.cfg.CheckTimeout.Duration, Checks: check.BuildPublicIP(m.cfg)}
		return publicIPRefreshMsg{results: runner.Run(ctx)}
	}
}

func (m model) scanWiFi() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), wifiScanTimeout)
		defer cancel()
		networks, source, err := check.ScanWiFiNetworks(ctx)
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
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
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

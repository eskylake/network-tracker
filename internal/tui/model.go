package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eskylake/network-tracker/internal/check"
	"github.com/eskylake/network-tracker/internal/config"
	"github.com/eskylake/network-tracker/internal/parse"
)

type model struct {
	cfg      config.Config
	spinner  spinner.Model
	tab      Tab
	selected int
	detail   bool
	loading  bool
	width    int
	height   int

	results     []check.Result
	logs        []string
	logOffset   int
	pingTarget  string
	pingRunning bool
	pingResults []string
	lastRefresh time.Time
	lastSnap    networkSnapshot
	err         error

	wifiScanning   bool
	wifiNetworks   []parse.WiFiNetwork
	wifiScanErr    error
	wifiScanSource string
	wifiScanAt     time.Time
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

func (m model) startWiFiScanIfNeeded() (tea.Model, tea.Cmd) {
	if m.tab == TabScan && !m.wifiScanning && len(m.wifiNetworks) == 0 {
		m.wifiScanning = true
		return m, m.scanWiFi()
	}
	return m, nil
}

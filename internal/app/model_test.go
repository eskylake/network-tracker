package app

import (
	"strings"
	"testing"

	"github.com/eskylake/network-tracker/internal/checks"
)

func TestMergeResultsReplacesExistingByName(t *testing.T) {
	current := []checks.Result{
		{Name: "public ip", Summary: "old"},
		{Name: "routes", Summary: "route"},
	}
	updates := []checks.Result{{Name: "public ip", Summary: "new"}}

	got := mergeResults(current, updates)
	if len(got) != 2 {
		t.Fatalf("unexpected result count: %d", len(got))
	}
	if got[0].Name != "public ip" || got[0].Summary != "new" {
		t.Fatalf("public ip was not replaced: %#v", got[0])
	}
	if got[1].Name != "routes" || got[1].Summary != "route" {
		t.Fatalf("unrelated result changed: %#v", got[1])
	}
}

func TestMergeResultsAppendsNewResults(t *testing.T) {
	current := []checks.Result{{Name: "routes", Summary: "route"}}
	updates := []checks.Result{{Name: "wifi", Summary: "Home Wifi"}}

	got := mergeResults(current, updates)
	if len(got) != 2 || got[1].Name != "wifi" || got[1].Summary != "Home Wifi" {
		t.Fatalf("new result was not appended: %#v", got)
	}
}

func TestFilteredResultsExcludesWiFiFromConnectivityTab(t *testing.T) {
	m := model{
		tabs: []string{"Overview", "VPN", "Connectivity", "Routes", "Docker", "Ping", "Logs"},
		tab:  2,
		results: []checks.Result{
			{Name: "wifi", Category: "connectivity", Summary: "Home Wifi"},
			{Name: "tcp 1.1.1.1:443", Category: "connectivity", Summary: "reachable"},
		},
	}

	got := m.filteredResults()
	if len(got) != 1 || got[0].Name != "tcp 1.1.1.1:443" {
		t.Fatalf("wifi should be excluded from connectivity tab: %#v", got)
	}
}

func TestRenderStatusBarShowsWiFiSummary(t *testing.T) {
	m := model{
		results: []checks.Result{
			{
				Name:     "wifi",
				Category: "connectivity",
				Status:   checks.StatusOK,
				Summary:  "Home Wifi",
				Details:  "SSID: Home Wifi\nSignal: 72%\nSource: nmcli",
			},
		},
	}

	got := m.renderStatusBar()
	if !strings.Contains(got, "Home Wifi") {
		t.Fatalf("status bar should include wifi summary: %q", got)
	}
	if !strings.Contains(got, "72%") {
		t.Fatalf("status bar should include wifi signal: %q", got)
	}
}

func TestStartWiFiScanIfNeededTriggersOnEmptyScanTab(t *testing.T) {
	m := model{
		tabs: []string{"Overview", "Scan"},
		tab:  1,
	}

	updated, cmd := m.startWiFiScanIfNeeded()
	next := updated.(model)
	if !next.wifiScanning {
		t.Fatal("expected wifiScanning to be true")
	}
	if cmd == nil {
		t.Fatal("expected scan command")
	}
}

func TestStartWiFiScanIfNeededSkipsWhenResultsExist(t *testing.T) {
	m := model{
		tabs:         []string{"Overview", "Scan"},
		tab:          1,
		wifiNetworks: []checks.WiFiNetwork{{SSID: "Home Wifi"}},
	}

	_, cmd := m.startWiFiScanIfNeeded()
	if cmd != nil {
		t.Fatal("expected no scan command when results already exist")
	}
}

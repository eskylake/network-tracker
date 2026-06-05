package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/eskylake/network-tracker/internal/checks"
)

func TestPadUsesVisualWidthWithANSI(t *testing.T) {
	styled := statusStyle(checks.StatusOK).Render(pad("*", wifiColInUse))
	if lipgloss.Width(styled) != wifiColInUse {
		t.Fatalf("styled cell width = %d, want %d", lipgloss.Width(styled), wifiColInUse)
	}
}

func TestRenderWiFiNetworkTableAlignsColumns(t *testing.T) {
	networks := []checks.WiFiNetwork{
		{SSID: "salam", SignalPercent: 58, Security: "WPA2", Freq: "2432 MHz", InUse: true},
		{SSID: "Very Long Network Name", SignalPercent: 97, Security: "WPA2 WPA3", Freq: "5240 MHz"},
	}
	tableWidth := model{width: 120}.wifiTableWidth()
	got := renderWiFiNetworkTable(networks, 0, 0, tableWidth)
	lines := strings.Split(got, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected table lines, got %q", got)
	}
	if !strings.Contains(lines[2], "salam") {
		t.Fatalf("first row missing ssid: %q", lines[2])
	}
	if lipgloss.Width(lines[2]) > tableWidth+2 {
		t.Fatalf("row wider than table: %d > %d", lipgloss.Width(lines[2]), tableWidth+2)
	}
}

func TestWiFiVisibleWindowKeepsSelectionCentered(t *testing.T) {
	start, end := wifiVisibleWindow(10, 20, 5)
	if start != 8 || end != 13 {
		t.Fatalf("unexpected window: %d-%d", start, end)
	}
}

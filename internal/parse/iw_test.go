package parse

import "testing"

func TestWirelessInterfaces(t *testing.T) {
	raw := "phy#0\n\tInterface wlan0\n\t\ttype managed\n\tInterface p2p-dev-wlan0\n"
	got := WirelessInterfaces(raw)
	if len(got) != 2 || got[0] != "wlan0" || got[1] != "p2p-dev-wlan0" {
		t.Fatalf("unexpected interfaces: %#v", got)
	}
}

func TestIWLinkSSID(t *testing.T) {
	raw := "Connected to aa:bb:cc:dd:ee:ff (on wlan0)\n\tSSID: Home Wifi\n"
	got := IWLinkSSID(raw)
	if got != "Home Wifi" {
		t.Fatalf("unexpected ssid: %q", got)
	}
}

func TestIWLinkSignal(t *testing.T) {
	raw := "Connected to aa:bb:cc:dd:ee:ff (on wlan0)\n\tsignal: -58.00 dBm\n\tSSID: Home Wifi\n"
	got := IWLinkSignal(raw)
	if got != -58 {
		t.Fatalf("unexpected signal: %d", got)
	}
}

func TestWiFiListFromIWScan(t *testing.T) {
	raw := "" +
		"BSS aa:bb:cc:dd:ee:01(on wlan0)\n" +
		"\tsignal: -58.00 dBm\n" +
		"\tSSID: Home Wifi\n" +
		"\tfreq: 2432\n" +
		"BSS aa:bb:cc:dd:ee:02(on wlan0)\n" +
		"\tsignal: -72 dBm\n" +
		"\tfreq: 5240\n"

	got := WiFiListFromIWScan(raw)
	if len(got) != 2 {
		t.Fatalf("unexpected network count: %d", len(got))
	}
	if got[0].SSID != "Home Wifi" || got[0].SignalDBM != -58 || got[0].Freq != "2432 MHz" {
		t.Fatalf("unexpected first network: %#v", got[0])
	}
	if got[1].SSID != "(hidden)" || got[1].SignalDBM != -72 {
		t.Fatalf("hidden network not parsed: %#v", got[1])
	}
}

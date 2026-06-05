package checks

import "testing"

func TestParseNMCLIWiFiList(t *testing.T) {
	raw := "" +
		" :DC\\:62\\:79\\:B6\\:EE\\:A8:Tilo:97:WPA2:2422 MHz\n" +
		"*:48\\:EE\\:0C\\:DC\\:D3\\:D5:salam:58:WPA2:2432 MHz\n" +
		" :AA\\:BB\\:CC\\:DD\\:EE\\:FF::45:WPA2:2437 MHz\n"

	got := ParseNMCLIWiFiList(raw)
	if len(got) != 3 {
		t.Fatalf("unexpected network count: %d", len(got))
	}
	if !got[0].InUse || got[0].SSID != "salam" || got[0].SignalPercent != 58 {
		t.Fatalf("connected network should sort first: %#v", got[0])
	}
	if got[1].SSID != "Tilo" || got[1].SignalPercent != 97 {
		t.Fatalf("unexpected strongest network: %#v", got[1])
	}
	if got[2].SSID != "(hidden)" {
		t.Fatalf("empty ssid should be hidden: %#v", got[2])
	}
}

func TestParseIWScan(t *testing.T) {
	raw := "" +
		"BSS aa:bb:cc:dd:ee:01(on wlan0)\n" +
		"\tsignal: -58.00 dBm\n" +
		"\tSSID: Home Wifi\n" +
		"\tfreq: 2432\n" +
		"BSS aa:bb:cc:dd:ee:02(on wlan0)\n" +
		"\tsignal: -72 dBm\n" +
		"\tfreq: 5240\n"

	got := ParseIWScan(raw)
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

func TestWiFiNetworkSignalLabel(t *testing.T) {
	if got := (WiFiNetwork{SignalPercent: 58}).SignalLabel(); got != "58%" {
		t.Fatalf("unexpected percent label: %q", got)
	}
	if got := (WiFiNetwork{SignalDBM: -58}).SignalLabel(); got != "-58 dBm" {
		t.Fatalf("unexpected dbm label: %q", got)
	}
}

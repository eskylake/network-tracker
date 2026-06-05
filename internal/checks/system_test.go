package checks

import "testing"

func TestParseDefaultRoute(t *testing.T) {
	raw := "default via 192.168.1.1 dev wlan0 proto dhcp metric 600\n172.17.0.0/16 dev docker0 proto kernel"
	got := ParseDefaultRoute(raw)
	if got.Interface != "wlan0" || got.Gateway != "192.168.1.1" || got.Metric != "600" {
		t.Fatalf("unexpected route: %+v", got)
	}
}

func TestParseResolvConf(t *testing.T) {
	got := ParseResolvConf("nameserver 1.1.1.1\nsearch lan\nnameserver 8.8.8.8\n")
	if len(got) != 2 || got[0] != "1.1.1.1" || got[1] != "8.8.8.8" {
		t.Fatalf("unexpected nameservers: %#v", got)
	}
}

func TestParseNMCLIActiveSSID(t *testing.T) {
	raw := "no:Other\nyes:Cafe\\:Guest\\\\5G\n"
	got := ParseNMCLIActiveSSID(raw)
	if got != "Cafe:Guest\\5G" {
		t.Fatalf("unexpected ssid: %q", got)
	}
}

func TestParseNMCLIActiveWiFi(t *testing.T) {
	raw := "no:Tilo:97\nyes:salam:55\n"
	got, ok := ParseNMCLIActiveWiFi(raw)
	if !ok || got.SSID != "salam" || got.Signal != 55 {
		t.Fatalf("unexpected active wifi: %#v ok=%v", got, ok)
	}
}

func TestParseWirelessInterfaces(t *testing.T) {
	raw := "phy#0\n\tInterface wlan0\n\t\ttype managed\n\tInterface p2p-dev-wlan0\n"
	got := ParseWirelessInterfaces(raw)
	if len(got) != 2 || got[0] != "wlan0" || got[1] != "p2p-dev-wlan0" {
		t.Fatalf("unexpected interfaces: %#v", got)
	}
}

func TestParseIWLinkSSID(t *testing.T) {
	raw := "Connected to aa:bb:cc:dd:ee:ff (on wlan0)\n\tSSID: Home Wifi\n"
	got := ParseIWLinkSSID(raw)
	if got != "Home Wifi" {
		t.Fatalf("unexpected ssid: %q", got)
	}
}

func TestParseIWLinkSignal(t *testing.T) {
	raw := "Connected to aa:bb:cc:dd:ee:ff (on wlan0)\n\tsignal: -58.00 dBm\n\tSSID: Home Wifi\n"
	got := ParseIWLinkSignal(raw)
	if got != -58 {
		t.Fatalf("unexpected signal: %d", got)
	}
}

func TestParseDockerNetworkList(t *testing.T) {
	raw := `{"Name":"bridge","ID":"abc"}` + "\n" + `{"Name":"vpn","ID":"def"}`
	got := ParseDockerNetworkList(raw)
	if len(got) != 2 || got[0].Name != "bridge" || got[1].Name != "vpn" {
		t.Fatalf("unexpected docker list: %#v", got)
	}
}

func TestParseDockerInspect(t *testing.T) {
	raw := `[{"Name":"bridge","Driver":"bridge","IPAM":{"Config":[{"Subnet":"172.17.0.0/16","Gateway":"172.17.0.1"}]}}]`
	got, warnings := ParseDockerInspect(raw)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %#v", warnings)
	}
	if len(got) != 1 || got[0].Subnet != "172.17.0.0/16" || got[0].Gateway != "172.17.0.1" {
		t.Fatalf("unexpected inspect: %#v", got)
	}
}

func TestCIDROverlap(t *testing.T) {
	if !CIDROverlap("172.17.0.0/16", "172.17.1.0/24") {
		t.Fatal("expected overlap")
	}
	if CIDROverlap("172.17.0.0/16", "192.168.1.0/24") {
		t.Fatal("did not expect overlap")
	}
}

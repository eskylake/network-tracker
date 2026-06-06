package parse

import "testing"

func TestDefaultRouteFromIP(t *testing.T) {
	raw := "default via 192.168.1.1 dev wlan0 proto dhcp metric 600\n172.17.0.0/16 dev docker0 proto kernel"
	got := DefaultRouteFromIP(raw)
	if got.Interface != "wlan0" || got.Gateway != "192.168.1.1" || got.Metric != "600" {
		t.Fatalf("unexpected route: %+v", got)
	}
}

func TestNameservers(t *testing.T) {
	got := Nameservers("nameserver 1.1.1.1\nsearch lan\nnameserver 8.8.8.8\n")
	if len(got) != 2 || got[0] != "1.1.1.1" || got[1] != "8.8.8.8" {
		t.Fatalf("unexpected nameservers: %#v", got)
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

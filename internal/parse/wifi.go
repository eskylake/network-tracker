package parse

import (
	"fmt"
	"sort"
)

// WiFiNetwork is a nearby or connected wireless network.
type WiFiNetwork struct {
	SSID          string
	BSSID         string
	SignalPercent int
	SignalDBM     int
	Security      string
	Freq          string
	InUse         bool
}

func (n WiFiNetwork) SignalLabel() string {
	if n.SignalPercent > 0 {
		return fmt.Sprintf("%d%%", n.SignalPercent)
	}
	if n.SignalDBM != 0 {
		return fmt.Sprintf("%d dBm", n.SignalDBM)
	}
	return "-"
}

func sortWiFiNetworks(networks []WiFiNetwork) {
	sort.Slice(networks, func(i, j int) bool {
		left := wifiSortKey(networks[i])
		right := wifiSortKey(networks[j])
		if left != right {
			return left > right
		}
		if networks[i].SSID != networks[j].SSID {
			return networks[i].SSID < networks[j].SSID
		}
		return networks[i].BSSID < networks[j].BSSID
	})
}

func wifiSortKey(network WiFiNetwork) int {
	if network.InUse {
		return 1_000_000
	}
	if network.SignalPercent > 0 {
		return network.SignalPercent
	}
	if network.SignalDBM != 0 {
		return network.SignalDBM + 120
	}
	return 0
}

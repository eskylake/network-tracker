package parse

import (
	"strconv"
	"strings"
)

func WirelessInterfaces(raw string) []string {
	var interfaces []string
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "Interface" {
			interfaces = append(interfaces, fields[1])
		}
	}
	return interfaces
}

func IWLinkSSID(raw string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SSID:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		}
	}
	return ""
}

func IWLinkSignal(raw string) int {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "signal:") {
			continue
		}
		fields := strings.Fields(strings.TrimPrefix(line, "signal:"))
		if len(fields) == 0 {
			continue
		}
		value := strings.TrimSuffix(strings.TrimSuffix(fields[0], ".00"), ".0")
		signal, err := strconv.Atoi(value)
		if err != nil {
			continue
		}
		return signal
	}
	return 0
}

func WiFiListFromIWScan(raw string) []WiFiNetwork {
	var networks []WiFiNetwork
	var current WiFiNetwork
	flush := func() {
		if current.BSSID == "" && current.SSID == "" {
			return
		}
		if current.SSID == "" {
			current.SSID = "(hidden)"
		}
		networks = append(networks, current)
		current = WiFiNetwork{}
	}

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "BSS "):
			flush()
			bssid := strings.TrimPrefix(line, "BSS ")
			bssid, _, _ = strings.Cut(bssid, "(")
			current.BSSID = strings.TrimSpace(bssid)
		case strings.HasPrefix(line, "SSID:"):
			current.SSID = strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		case strings.HasPrefix(line, "signal:"):
			fields := strings.Fields(strings.TrimPrefix(line, "signal:"))
			if len(fields) > 0 {
				value := strings.TrimSuffix(fields[0], ".00")
				value = strings.TrimSuffix(value, ".0")
				if dbm, err := strconv.Atoi(value); err == nil {
					current.SignalDBM = dbm
				}
			}
		case strings.HasPrefix(line, "freq:"):
			fields := strings.Fields(strings.TrimPrefix(line, "freq:"))
			if len(fields) > 0 {
				current.Freq = fields[0] + " MHz"
			}
		}
	}
	flush()
	sortWiFiNetworks(networks)
	return networks
}

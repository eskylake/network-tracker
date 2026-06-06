package parse

import (
	"strconv"
	"strings"
)

// ActiveWiFi is the currently connected network from nmcli output.
type ActiveWiFi struct {
	SSID   string
	Signal int
}

func ActiveSSID(raw string) string {
	active, ok := ActiveWiFiFromNMCLI(raw)
	if !ok {
		return ""
	}
	return active.SSID
}

func ActiveWiFiFromNMCLI(raw string) (ActiveWiFi, bool) {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		active, rest, ok := strings.Cut(line, ":")
		if !ok || active != "yes" || rest == "" {
			continue
		}
		wifi := ActiveWiFi{SSID: unescapeNMCLITerse(rest)}
		if cut := lastUnescapedColon(rest); cut >= 0 {
			if signal, err := strconv.Atoi(strings.TrimSpace(rest[cut+1:])); err == nil {
				wifi.Signal = signal
				wifi.SSID = unescapeNMCLITerse(rest[:cut])
			}
		}
		return wifi, true
	}
	return ActiveWiFi{}, false
}

func WiFiListFromNMCLI(raw string) []WiFiNetwork {
	var networks []WiFiNetwork
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := splitNMCLITerse(line)
		if len(fields) < 6 {
			continue
		}
		signal, _ := strconv.Atoi(strings.TrimSpace(fields[3]))
		ssid := unescapeNMCLITerse(fields[2])
		if ssid == "" || ssid == "--" {
			ssid = "(hidden)"
		}
		networks = append(networks, WiFiNetwork{
			InUse:         strings.TrimSpace(fields[0]) == "*",
			BSSID:         unescapeNMCLITerse(fields[1]),
			SSID:          ssid,
			SignalPercent: signal,
			Security:      strings.TrimSpace(fields[4]),
			Freq:          strings.TrimSpace(fields[5]),
		})
	}
	sortWiFiNetworks(networks)
	return networks
}

func lastUnescapedColon(value string) int {
	for i := len(value) - 1; i >= 0; i-- {
		if value[i] != ':' {
			continue
		}
		backslashes := 0
		for j := i - 1; j >= 0 && value[j] == '\\'; j-- {
			backslashes++
		}
		if backslashes%2 == 0 {
			return i
		}
	}
	return -1
}

func unescapeNMCLITerse(value string) string {
	var b strings.Builder
	escaped := false
	for _, r := range value {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		b.WriteRune(r)
	}
	if escaped {
		b.WriteRune('\\')
	}
	return b.String()
}

func splitNMCLITerse(line string) []string {
	var fields []string
	var current strings.Builder
	escaped := false
	for _, r := range line {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == ':' {
			fields = append(fields, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	if escaped {
		current.WriteRune('\\')
	}
	fields = append(fields, current.String())
	return fields
}

package checks

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type WiFiNetwork struct {
	SSID           string
	BSSID          string
	SignalPercent  int
	SignalDBM      int
	Security       string
	Freq           string
	InUse          bool
}

func ScanWiFiNetworks(ctx context.Context) ([]WiFiNetwork, string, error) {
	runner := ShellCommandRunner{}

	_, _ = runner.Run(ctx, "nmcli device wifi rescan")
	list, listErr := runner.Run(ctx, "nmcli -t -f IN-USE,BSSID,SSID,SIGNAL,SECURITY,FREQ dev wifi list")
	if networks := ParseNMCLIWiFiList(list); len(networks) > 0 {
		return networks, "nmcli", listErr
	}

	iwdev, iwdevErr := runner.Run(ctx, "iw dev")
	var scanErrs []error
	if listErr != nil {
		scanErrs = append(scanErrs, listErr)
	}
	if iwdevErr != nil {
		scanErrs = append(scanErrs, iwdevErr)
	}

	for _, iface := range ParseWirelessInterfaces(iwdev) {
		if strings.HasPrefix(iface, "p2p-dev-") {
			continue
		}
		scan, err := runner.Run(ctx, "iw dev "+shellQuote(iface)+" scan")
		if err != nil {
			scanErrs = append(scanErrs, err)
			continue
		}
		networks := ParseIWScan(scan)
		if len(networks) > 0 {
			return networks, "iw (" + iface + ")", errors.Join(scanErrs...)
		}
	}

	return nil, "", errors.Join(scanErrs...)
}

func ParseNMCLIWiFiList(raw string) []WiFiNetwork {
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

func ParseIWScan(raw string) []WiFiNetwork {
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

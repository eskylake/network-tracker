package check

import (
	"context"
	"errors"
	"strings"

	"github.com/eskylake/network-tracker/internal/parse"
	"github.com/eskylake/network-tracker/internal/shell"
)

// ScanWiFiNetworks scans nearby networks using nmcli or iw.
func ScanWiFiNetworks(ctx context.Context) ([]parse.WiFiNetwork, string, error) {
	runner := shell.Shell{}

	_, _ = runner.Run(ctx, "nmcli device wifi rescan")
	list, listErr := runner.Run(ctx, "nmcli -t -f IN-USE,BSSID,SSID,SIGNAL,SECURITY,FREQ dev wifi list")
	if networks := parse.WiFiListFromNMCLI(list); len(networks) > 0 {
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

	for _, iface := range parse.WirelessInterfaces(iwdev) {
		if strings.HasPrefix(iface, "p2p-dev-") {
			continue
		}
		scan, err := runner.Run(ctx, "iw dev "+shell.Quote(iface)+" scan")
		if err != nil {
			scanErrs = append(scanErrs, err)
			continue
		}
		networks := parse.WiFiListFromIWScan(scan)
		if len(networks) > 0 {
			return networks, "iw (" + iface + ")", errors.Join(scanErrs...)
		}
	}

	return nil, "", errors.Join(scanErrs...)
}

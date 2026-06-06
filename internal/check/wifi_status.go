package check

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eskylake/network-tracker/internal/parse"
	"github.com/eskylake/network-tracker/internal/shell"
)

func WiFiChecker() Checker {
	return CheckFunc{
		CheckName:     NameWiFi,
		CheckCategory: CategoryConnectivity,
		Fn: func(ctx context.Context) Result {
			start := time.Now()
			runner := shell.Shell{}
			nmcli, nmcliErr := runner.Run(ctx, "nmcli -t -f ACTIVE,SSID,SIGNAL dev wifi")
			if active, ok := parse.ActiveWiFiFromNMCLI(nmcli); ok {
				details := fmt.Sprintf("SSID: %s\nSignal: %d%%\nSource: nmcli", active.SSID, active.Signal)
				return finished(NameWiFi, CategoryConnectivity, StatusOK, active.SSID, details, start, nil)
			}

			iwdev, iwdevErr := runner.Run(ctx, "iw dev")
			for _, iface := range parse.WirelessInterfaces(iwdev) {
				link, linkErr := runner.Run(ctx, "iw dev "+shell.Quote(iface)+" link")
				if ssid := parse.IWLinkSSID(link); ssid != "" {
					details := fmt.Sprintf("SSID: %s\nInterface: %s\nSource: iw", ssid, iface)
					if signal := parse.IWLinkSignal(link); signal != 0 {
						details = fmt.Sprintf("SSID: %s\nSignal: %d dBm\nInterface: %s\nSource: iw", ssid, signal, iface)
					}
					return finished(NameWiFi, CategoryConnectivity, StatusOK, ssid, details, start, nil)
				}
				if linkErr != nil && iwdevErr == nil {
					iwdevErr = linkErr
				}
			}

			details := strings.TrimSpace(outputOrError(nmcli, nmcliErr) + "\n" + outputOrError(iwdev, iwdevErr))
			return finished(NameWiFi, CategoryConnectivity, StatusUnknown, "not connected or unavailable", details, start, errors.Join(nmcliErr, iwdevErr))
		},
	}
}

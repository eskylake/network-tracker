package check

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type PublicIP struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

func TCPChecker(target string) Checker {
	name := "tcp " + target
	return CheckFunc{
		CheckName:     name,
		CheckCategory: CategoryConnectivity,
		Fn: func(ctx context.Context) Result {
			var dialer net.Dialer
			start := time.Now()
			conn, err := dialer.DialContext(ctx, "tcp", target)
			if err != nil {
				return finished(name, CategoryConnectivity, StatusError, "connection failed", err.Error(), start, err)
			}
			_ = conn.Close()
			return finished(name, CategoryConnectivity, StatusOK, "reachable", "", start, nil)
		},
	}
}

func DNSChecker(domain string) Checker {
	name := "dns " + domain
	return CheckFunc{
		CheckName:     name,
		CheckCategory: CategoryConnectivity,
		Fn: func(ctx context.Context) Result {
			start := time.Now()
			addrs, err := net.DefaultResolver.LookupHost(ctx, domain)
			if err != nil {
				return finished(name, CategoryConnectivity, StatusError, "resolution failed", err.Error(), start, err)
			}
			return finished(name, CategoryConnectivity, StatusOK, fmt.Sprintf("%d address(es)", len(addrs)), strings.Join(addrs, "\n"), start, nil)
		},
	}
}

func PublicIPChecker(provider string) Checker {
	return CheckFunc{
		CheckName:     NamePublicIP,
		CheckCategory: CategoryConnectivity,
		Fn: func(ctx context.Context) Result {
			start := time.Now()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, provider, nil)
			if err != nil {
				return finished(NamePublicIP, CategoryConnectivity, StatusError, "invalid provider", err.Error(), start, err)
			}
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return finished(NamePublicIP, CategoryConnectivity, StatusError, "request failed", err.Error(), start, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				err := fmt.Errorf("unexpected status %s", resp.Status)
				return finished(NamePublicIP, CategoryConnectivity, StatusError, "provider error", err.Error(), start, err)
			}
			var info PublicIP
			if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
				return finished(NamePublicIP, CategoryConnectivity, StatusError, "parse failed", err.Error(), start, err)
			}
			lat, lon := splitLocation(info.Loc)
			summary := strings.Join(nonEmpty(info.IP, info.City, info.Country, info.Org), " | ")
			details := fmt.Sprintf("IP: %s\nHostname: %s\nCity: %s\nRegion: %s\nCountry: %s\nLatitude: %s\nLongitude: %s\nLocation: %s\nPostal: %s\nOrg: %s\nTimezone: %s\nProvider: %s",
				info.IP, info.Hostname, info.City, info.Region, info.Country, lat, lon, info.Loc, info.Postal, info.Org, info.Timezone, provider)
			return finished(NamePublicIP, CategoryConnectivity, StatusOK, summary, details, start, nil)
		},
	}
}

func splitLocation(loc string) (string, string) {
	parts := strings.SplitN(loc, ",", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

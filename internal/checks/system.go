package checks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func CommandChecker(name, category, command string, ok func(string) bool, classify func(string, error) (Status, string)) Checker {
	return commandChecker{name: name, category: category, command: command, ok: ok, classify: classify, runner: ShellCommandRunner{}}
}

type commandChecker struct {
	name     string
	category string
	command  string
	ok       func(string) bool
	classify func(string, error) (Status, string)
	runner   CommandRunner
}

func (c commandChecker) Name() string     { return c.name }
func (c commandChecker) Category() string { return c.category }
func (c commandChecker) Run(ctx context.Context) Result {
	start := time.Now()
	output, err := c.runner.Run(ctx, c.command)
	if c.classify != nil {
		status, summary := c.classify(output, err)
		return finished(c.name, c.category, status, summary, output, start, err)
	}
	if err != nil {
		return finished(c.name, c.category, StatusWarning, "command failed", outputOrError(output, err), start, err)
	}
	if c.ok == nil || c.ok(output) {
		return finished(c.name, c.category, StatusOK, firstLine(output, "ok"), output, start, nil)
	}
	return finished(c.name, c.category, StatusWarning, firstLine(output, "unexpected output"), output, start, nil)
}

func XVPNChecker(command string) Checker {
	return CommandChecker("xvpn", "vpn", command, nil, func(output string, err error) (Status, string) {
		lower := strings.ToLower(output)
		if err != nil {
			if strings.Contains(lower, "password") || strings.Contains(lower, "sudo") {
				return StatusWarning, "xvpn command should not require sudo"
			}
			return StatusWarning, "xvpn status unavailable"
		}
		if strings.Contains(lower, "connected") {
			return StatusOK, "connected"
		}
		if strings.Contains(lower, "disconnect") || strings.Contains(lower, "not connected") {
			return StatusWarning, "not connected"
		}
		return StatusUnknown, firstLine(output, "status unknown")
	})
}

func V2RayAServiceChecker(command string) Checker {
	return CommandChecker("v2raya service", "vpn", command, nil, func(output string, err error) (Status, string) {
		trimmed := strings.TrimSpace(output)
		if err != nil {
			if trimmed == "inactive" || trimmed == "failed" {
				return StatusWarning, trimmed
			}
			return StatusWarning, "service status unavailable"
		}
		if trimmed == "active" {
			return StatusOK, "active"
		}
		return StatusWarning, firstLine(trimmed, "not active")
	})
}

func ProcessChecker(name, pattern string) Checker {
	return CommandChecker(name, "vpn", "pgrep -a "+pattern, nil, func(output string, err error) (Status, string) {
		if err != nil || strings.TrimSpace(output) == "" {
			return StatusWarning, "process not found"
		}
		return StatusOK, "process found"
	})
}

func WiFiChecker() Checker {
	return CheckFunc{
		CheckName:     "wifi",
		CheckCategory: "connectivity",
		Fn: func(ctx context.Context) Result {
			start := time.Now()
			runner := ShellCommandRunner{}
			nmcli, nmcliErr := runner.Run(ctx, "nmcli -t -f ACTIVE,SSID,SIGNAL dev wifi")
			if active, ok := ParseNMCLIActiveWiFi(nmcli); ok {
				details := fmt.Sprintf("SSID: %s\nSignal: %d%%\nSource: nmcli", active.SSID, active.Signal)
				return finished("wifi", "connectivity", StatusOK, active.SSID, details, start, nil)
			}

			iwdev, iwdevErr := runner.Run(ctx, "iw dev")
			for _, iface := range ParseWirelessInterfaces(iwdev) {
				link, linkErr := runner.Run(ctx, "iw dev "+shellQuote(iface)+" link")
				if ssid := ParseIWLinkSSID(link); ssid != "" {
					details := fmt.Sprintf("SSID: %s\nInterface: %s\nSource: iw", ssid, iface)
					if signal := ParseIWLinkSignal(link); signal != 0 {
						details = fmt.Sprintf("SSID: %s\nSignal: %d dBm\nInterface: %s\nSource: iw", ssid, signal, iface)
					}
					return finished("wifi", "connectivity", StatusOK, ssid, details, start, nil)
				}
				if linkErr != nil && iwdevErr == nil {
					iwdevErr = linkErr
				}
			}

			details := strings.TrimSpace(outputOrError(nmcli, nmcliErr) + "\n" + outputOrError(iwdev, iwdevErr))
			return finished("wifi", "connectivity", StatusUnknown, "not connected or unavailable", details, start, errors.Join(nmcliErr, iwdevErr))
		},
	}
}

func RouteChecker() Checker {
	return CheckFunc{
		CheckName:     "routes",
		CheckCategory: "routes",
		Fn: func(ctx context.Context) Result {
			start := time.Now()
			runner := ShellCommandRunner{}
			route, routeErr := runner.Run(ctx, "ip route")
			rule, ruleErr := runner.Run(ctx, "ip rule")
			dns := readResolvConf()
			defaultRoute := ParseDefaultRoute(route)

			status := StatusOK
			summary := defaultRoute.Summary()
			var errs []string
			if routeErr != nil {
				status = StatusWarning
				errs = append(errs, routeErr.Error())
			}
			if ruleErr != nil {
				status = StatusWarning
				errs = append(errs, ruleErr.Error())
			}
			if summary == "" {
				status = StatusWarning
				summary = "default route not detected"
			}

			details := "ip route:\n" + route + "\n\nip rule:\n" + rule + "\n\n/etc/resolv.conf:\n" + dns
			return finished("routes", "routes", status, summary, details, start, errors.Join(stringErrors(errs)...))
		},
	}
}

func DNSConfigChecker() Checker {
	return CheckFunc{
		CheckName:     "dns config",
		CheckCategory: "routes",
		Fn: func(ctx context.Context) Result {
			_ = ctx
			start := time.Now()
			raw := readResolvConf()
			servers := ParseResolvConf(raw)
			if len(servers) == 0 {
				return finished("dns config", "routes", StatusWarning, "no nameservers found", raw, start, nil)
			}
			return finished("dns config", "routes", StatusOK, strings.Join(servers, ", "), raw, start, nil)
		},
	}
}

func DockerChecker(enabled bool) Checker {
	return CheckFunc{
		CheckName:     "docker networks",
		CheckCategory: "docker",
		Fn: func(ctx context.Context) Result {
			start := time.Now()
			if !enabled {
				return finished("docker networks", "docker", StatusUnknown, "disabled", "", start, nil)
			}

			runner := ShellCommandRunner{}
			rawList, err := runner.Run(ctx, "docker network ls --format '{{json .}}'")
			if err != nil {
				return finished("docker networks", "docker", StatusWarning, "docker unavailable", outputOrError(rawList, err), start, err)
			}
			networks := ParseDockerNetworkList(rawList)
			if len(networks) == 0 {
				return finished("docker networks", "docker", StatusWarning, "no networks returned", rawList, start, nil)
			}

			routeRaw, _ := runner.Run(ctx, "ip route")
			routeCIDRs := RouteCIDRs(routeRaw)
			var rows []DockerNetwork
			var warnings []string
			for _, network := range networks {
				inspect, err := runner.Run(ctx, "docker network inspect "+shellQuote(network.Name))
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("%s inspect failed: %s", network.Name, firstLine(outputOrError(inspect, err), "unknown error")))
					rows = append(rows, DockerNetwork{Name: network.Name, Driver: "?", Subnet: "-", Gateway: "-"})
					continue
				}
				dockerNets, parseWarnings := ParseDockerInspect(inspect)
				warnings = append(warnings, parseWarnings...)
				if len(dockerNets) == 0 {
					rows = append(rows, DockerNetwork{Name: network.Name, Driver: "-", Subnet: "-", Gateway: "-"})
				}
				for _, dn := range dockerNets {
					rows = append(rows, dn)
					for _, routeCIDR := range routeCIDRs {
						if CIDROverlap(dn.Subnet, routeCIDR) {
							warnings = append(warnings, fmt.Sprintf("%s (%s) overlaps route %s", dn.Name, dn.Subnet, routeCIDR))
						}
					}
				}
			}
			warnings = unique(warnings)
			status := StatusOK
			summary := fmt.Sprintf("%d network(s)", len(networks))
			if len(warnings) > 0 {
				status = StatusWarning
				summary = fmt.Sprintf("%d network(s), %d warning(s)", len(networks), len(warnings))
			}
			return finished("docker networks", "docker", status, summary, FormatDockerDetails(rows, warnings), start, nil)
		},
	}
}

type DefaultRoute struct {
	Interface string
	Gateway   string
	Metric    string
}

func (r DefaultRoute) Summary() string {
	if r.Interface == "" {
		return ""
	}
	parts := []string{"dev " + r.Interface}
	if r.Gateway != "" {
		parts = append(parts, "via "+r.Gateway)
	}
	if r.Metric != "" {
		parts = append(parts, "metric "+r.Metric)
	}
	return strings.Join(parts, " ")
}

func ParseDefaultRoute(raw string) DefaultRoute {
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "default" {
			continue
		}
		var route DefaultRoute
		for i := 1; i < len(fields)-1; i++ {
			switch fields[i] {
			case "via":
				route.Gateway = fields[i+1]
			case "dev":
				route.Interface = fields[i+1]
			case "metric":
				route.Metric = fields[i+1]
			}
		}
		return route
	}
	return DefaultRoute{}
}

func ParseResolvConf(raw string) []string {
	var servers []string
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "nameserver" {
			servers = append(servers, fields[1])
		}
	}
	return servers
}

type ActiveWiFi struct {
	SSID   string
	Signal int
}

func ParseNMCLIActiveSSID(raw string) string {
	active, ok := ParseNMCLIActiveWiFi(raw)
	if !ok {
		return ""
	}
	return active.SSID
}

func ParseNMCLIActiveWiFi(raw string) (ActiveWiFi, bool) {
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

func ParseWirelessInterfaces(raw string) []string {
	var interfaces []string
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "Interface" {
			interfaces = append(interfaces, fields[1])
		}
	}
	return interfaces
}

func ParseIWLinkSSID(raw string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SSID:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		}
	}
	return ""
}

func ParseIWLinkSignal(raw string) int {
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

type dockerListItem struct {
	Name string `json:"Name"`
}

func ParseDockerNetworkList(raw string) []dockerListItem {
	var networks []dockerListItem
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var item dockerListItem
		if json.Unmarshal([]byte(line), &item) == nil && item.Name != "" {
			networks = append(networks, item)
		}
	}
	return networks
}

type dockerInspectItem struct {
	Name   string `json:"Name"`
	Driver string `json:"Driver"`
	IPAM   struct {
		Config []struct {
			Subnet  string `json:"Subnet"`
			Gateway string `json:"Gateway"`
		} `json:"Config"`
	} `json:"IPAM"`
}

type DockerNetwork struct {
	Name    string
	Driver  string
	Subnet  string
	Gateway string
}

func (n DockerNetwork) String() string {
	return fmt.Sprintf("%s driver=%s subnet=%s gateway=%s", n.Name, n.Driver, n.Subnet, n.Gateway)
}

func FormatDockerDetails(networks []DockerNetwork, warnings []string) string {
	var b strings.Builder
	b.WriteString("Docker Networks\n")
	b.WriteString(fmt.Sprintf("%-22s %-12s %-20s %-20s\n", "NETWORK", "DRIVER", "SUBNET", "GATEWAY"))
	b.WriteString(strings.Repeat("-", 78) + "\n")
	if len(networks) == 0 {
		b.WriteString("No Docker networks found.\n")
	} else {
		for _, network := range networks {
			b.WriteString(fmt.Sprintf("%-22s %-12s %-20s %-20s\n", cleanCell(network.Name, 22), cleanCell(network.Driver, 12), cleanCell(network.Subnet, 20), cleanCell(network.Gateway, 20)))
		}
	}

	if len(warnings) > 0 {
		b.WriteString("\nWarnings\n")
		for _, warning := range warnings {
			b.WriteString("- " + warning + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func cleanCell(value string, width int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "-"
	}
	if len(value) <= width {
		return value
	}
	if width <= 3 {
		return value[:width]
	}
	return value[:width-3] + "..."
}

func ParseDockerInspect(raw string) ([]DockerNetwork, []string) {
	var items []dockerInspectItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, []string{"docker inspect parse failed"}
	}
	var networks []DockerNetwork
	for _, item := range items {
		for _, cfg := range item.IPAM.Config {
			if cfg.Subnet == "" {
				continue
			}
			networks = append(networks, DockerNetwork{
				Name:    item.Name,
				Driver:  item.Driver,
				Subnet:  cfg.Subnet,
				Gateway: cfg.Gateway,
			})
		}
	}
	return networks, nil
}

func RouteCIDRs(raw string) []string {
	var cidrs []string
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "default" {
			continue
		}
		if strings.Contains(fields[0], "/") {
			cidrs = append(cidrs, fields[0])
		}
	}
	return cidrs
}

func CIDROverlap(a, b string) bool {
	_, an, errA := net.ParseCIDR(a)
	_, bn, errB := net.ParseCIDR(b)
	if errA != nil || errB != nil {
		return false
	}
	return an.Contains(bn.IP) || bn.Contains(an.IP)
}

func readResolvConf() string {
	raw, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return err.Error()
	}
	return string(raw)
}

func firstLine(output, fallback string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return fallback
	}
	return strings.Split(output, "\n")[0]
}

func outputOrError(output string, err error) string {
	output = strings.TrimSpace(output)
	if output != "" {
		return output
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func stringErrors(values []string) []error {
	errs := make([]error, 0, len(values))
	for _, value := range values {
		if value != "" {
			errs = append(errs, errors.New(value))
		}
	}
	return errs
}

func unique(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

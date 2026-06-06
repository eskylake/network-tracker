package parse

import (
	"net"
	"strings"
)

// DefaultRoute is parsed from `ip route` output.
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

func DefaultRouteFromIP(raw string) DefaultRoute {
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

func Nameservers(raw string) []string {
	var servers []string
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "nameserver" {
			servers = append(servers, fields[1])
		}
	}
	return servers
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

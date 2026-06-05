package checks

import "github.com/eskylake/network-tracker/internal/config"

func Build(cfg config.Config) []Checker {
	out := BuildWithoutPublicIP(cfg)
	out = append(out, PublicIPChecker(cfg.PublicIPProvider))
	return out
}

func BuildWithoutPublicIP(cfg config.Config) []Checker {
	var out []Checker
	out = append(out, XVPNChecker(cfg.XVPNStatusCommand))
	out = append(out, V2RayAServiceChecker(cfg.V2RayAStatusCommand))
	out = append(out, ProcessChecker("v2ray process", "v2ray"))
	out = append(out, ProcessChecker("v2raya process", "v2raya"))
	out = append(out, WiFiChecker())
	for _, target := range cfg.TCPTargets {
		out = append(out, TCPChecker(target))
	}
	for _, domain := range cfg.DomainsToResolve {
		out = append(out, DNSChecker(domain))
	}
	out = append(out, RouteChecker(), DNSConfigChecker())
	out = append(out, DockerChecker(cfg.DockerEnabled))
	return out
}

func BuildPublicIP(cfg config.Config) []Checker {
	return []Checker{PublicIPChecker(cfg.PublicIPProvider)}
}

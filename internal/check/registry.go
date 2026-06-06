package check

import "github.com/eskylake/network-tracker/internal/config"

func Build(cfg config.Config) []Checker {
	out := BuildWithoutPublicIP(cfg)
	out = append(out, PublicIPChecker(cfg.PublicIPProvider))
	return out
}

func BuildWithoutPublicIP(cfg config.Config) []Checker {
	out := make([]Checker, 0, 8+len(cfg.TCPTargets)+len(cfg.DomainsToResolve))
	out = append(out,
		XVPNChecker(cfg.XVPNStatusCommand),
		V2RayAServiceChecker(cfg.V2RayAStatusCommand),
		ProcessChecker("v2ray process", "v2ray"),
		ProcessChecker("v2raya process", "v2raya"),
		WiFiChecker(),
	)
	for _, target := range cfg.TCPTargets {
		out = append(out, TCPChecker(target))
	}
	for _, domain := range cfg.DomainsToResolve {
		out = append(out, DNSChecker(domain))
	}
	out = append(out, RouteChecker(), DNSConfigChecker(), DockerChecker(cfg.DockerEnabled))
	return out
}

func BuildPublicIP(cfg config.Config) []Checker {
	return []Checker{PublicIPChecker(cfg.PublicIPProvider)}
}

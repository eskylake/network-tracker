package check

import (
	"context"
	"fmt"
	"time"

	"github.com/eskylake/network-tracker/internal/parse"
	"github.com/eskylake/network-tracker/internal/shell"
)

func DockerChecker(enabled bool) Checker {
	return CheckFunc{
		CheckName:     NameDockerNetwork,
		CheckCategory: CategoryDocker,
		Fn: func(ctx context.Context) Result {
			start := time.Now()
			if !enabled {
				return finished(NameDockerNetwork, CategoryDocker, StatusUnknown, "disabled", "", start, nil)
			}

			runner := shell.Shell{}
			rawList, err := runner.Run(ctx, "docker network ls --format '{{json .}}'")
			if err != nil {
				return finished(NameDockerNetwork, CategoryDocker, StatusWarning, "docker unavailable", outputOrError(rawList, err), start, err)
			}
			networks := parse.DockerNetworkList(rawList)
			if len(networks) == 0 {
				return finished(NameDockerNetwork, CategoryDocker, StatusWarning, "no networks returned", rawList, start, nil)
			}

			routeRaw, _ := runner.Run(ctx, "ip route")
			routeCIDRs := parse.RouteCIDRs(routeRaw)
			var rows []parse.DockerNetwork
			var warnings []string
			for _, network := range networks {
				inspect, err := runner.Run(ctx, "docker network inspect "+shell.Quote(network.Name))
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("%s inspect failed: %s", network.Name, firstLine(outputOrError(inspect, err), "unknown error")))
					rows = append(rows, parse.DockerNetwork{Name: network.Name, Driver: "?", Subnet: "-", Gateway: "-"})
					continue
				}
				dockerNets, parseWarnings := parse.DockerNetworksFromInspect(inspect)
				warnings = append(warnings, parseWarnings...)
				if len(dockerNets) == 0 {
					rows = append(rows, parse.DockerNetwork{Name: network.Name, Driver: "-", Subnet: "-", Gateway: "-"})
				}
				for _, dn := range dockerNets {
					rows = append(rows, dn)
					for _, routeCIDR := range routeCIDRs {
						if parse.CIDROverlap(dn.Subnet, routeCIDR) {
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
			return finished(NameDockerNetwork, CategoryDocker, status, summary, parse.FormatDockerDetails(rows, warnings), start, nil)
		},
	}
}

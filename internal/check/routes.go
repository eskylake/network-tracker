package check

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/eskylake/network-tracker/internal/parse"
	"github.com/eskylake/network-tracker/internal/shell"
)

func RouteChecker() Checker {
	return CheckFunc{
		CheckName:     NameRoutes,
		CheckCategory: CategoryRoutes,
		Fn: func(ctx context.Context) Result {
			start := time.Now()
			runner := shell.Shell{}
			route, routeErr := runner.Run(ctx, "ip route")
			rule, ruleErr := runner.Run(ctx, "ip rule")
			dns := readResolvConf()
			defaultRoute := parse.DefaultRouteFromIP(route)

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
			return finished(NameRoutes, CategoryRoutes, status, summary, details, start, errors.Join(stringErrors(errs)...))
		},
	}
}

func DNSConfigChecker() Checker {
	return CheckFunc{
		CheckName:     NameDNSConfig,
		CheckCategory: CategoryRoutes,
		Fn: func(ctx context.Context) Result {
			_ = ctx
			start := time.Now()
			raw := readResolvConf()
			servers := parse.Nameservers(raw)
			if len(servers) == 0 {
				return finished(NameDNSConfig, CategoryRoutes, StatusWarning, "no nameservers found", raw, start, nil)
			}
			return finished(NameDNSConfig, CategoryRoutes, StatusOK, strings.Join(servers, ", "), raw, start, nil)
		},
	}
}

func readResolvConf() string {
	raw, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return err.Error()
	}
	return string(raw)
}

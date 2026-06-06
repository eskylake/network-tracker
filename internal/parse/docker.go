package parse

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DockerNetwork holds inspect details for one Docker network subnet.
type DockerNetwork struct {
	Name    string
	Driver  string
	Subnet  string
	Gateway string
}

// DockerListEntry is a row from `docker network ls --format json`.
type DockerListEntry struct {
	Name string
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

func DockerNetworkList(raw string) []DockerListEntry {
	var networks []DockerListEntry
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var item struct {
			Name string `json:"Name"`
		}
		if json.Unmarshal([]byte(line), &item) == nil && item.Name != "" {
			networks = append(networks, DockerListEntry{Name: item.Name})
		}
	}
	return networks
}

func DockerNetworksFromInspect(raw string) ([]DockerNetwork, []string) {
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

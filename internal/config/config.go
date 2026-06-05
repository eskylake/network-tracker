package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return err
	}
	d.Duration = parsed
	return nil
}

type Config struct {
	RefreshInterval         Duration `yaml:"refresh_interval"`
	PublicIPRefreshInterval Duration `yaml:"public_ip_refresh_interval"`
	CheckTimeout            Duration `yaml:"check_timeout"`

	DomainsToResolve []string `yaml:"domains_to_resolve"`
	TCPTargets       []string `yaml:"tcp_targets"`
	PublicIPProvider string   `yaml:"public_ip_provider"`

	DockerEnabled bool `yaml:"docker_enabled"`

	XVPNStatusCommand   string `yaml:"xvpn_status_command"`
	V2RayAStatusCommand string `yaml:"v2raya_status_command"`
}

func Default() Config {
	return Config{
		RefreshInterval:         Duration{Duration: 5 * time.Second},
		PublicIPRefreshInterval: Duration{Duration: 60 * time.Second},
		CheckTimeout:            Duration{Duration: 4 * time.Second},
		DomainsToResolve: []string{
			"google.com",
			"github.com",
			"cloudflare.com",
		},
		TCPTargets: []string{
			"1.1.1.1:53",
			"8.8.8.8:53",
		},
		PublicIPProvider:    "https://ipinfo.io/json",
		DockerEnabled:       true,
		XVPNStatusCommand:   "xvpn status",
		V2RayAStatusCommand: "systemctl is-active v2raya",
	}
}

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Config{}, err
		}
		raw, err := yaml.Marshal(cfg)
		if err != nil {
			return Config{}, err
		}
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			return Config{}, err
		}
		return cfg, nil
	} else if err != nil {
		return Config{}, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, err
	}
	normalize(&cfg)
	return cfg, nil
}

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "network-tracker", "config.yaml"), nil
}

func normalize(cfg *Config) {
	defaults := Default()
	if cfg.RefreshInterval.Duration == 0 {
		cfg.RefreshInterval = defaults.RefreshInterval
	}
	if cfg.PublicIPRefreshInterval.Duration == 0 {
		cfg.PublicIPRefreshInterval = defaults.PublicIPRefreshInterval
	}
	if cfg.CheckTimeout.Duration == 0 {
		cfg.CheckTimeout = defaults.CheckTimeout
	}
	if len(cfg.DomainsToResolve) == 0 {
		cfg.DomainsToResolve = defaults.DomainsToResolve
	}
	if len(cfg.TCPTargets) == 0 {
		cfg.TCPTargets = defaults.TCPTargets
	}
	if cfg.PublicIPProvider == "" {
		cfg.PublicIPProvider = defaults.PublicIPProvider
	}
	if cfg.XVPNStatusCommand == "" || cfg.XVPNStatusCommand == "sudo -n xvpn status" {
		cfg.XVPNStatusCommand = defaults.XVPNStatusCommand
	}
	if cfg.V2RayAStatusCommand == "" {
		cfg.V2RayAStatusCommand = defaults.V2RayAStatusCommand
	}
}

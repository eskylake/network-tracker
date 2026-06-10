# network-tracker

A read-only Linux terminal dashboard for Wi-Fi scanning, VPN/DNS/route checks, Docker networks, public IP info, and ping tests.

---

## Quick start

**Install and run:**

```bash
go install github.com/eskylake/network-tracker/cmd/network-tracker@latest
network-tracker
```

**Or build from source:**

```bash
git clone https://github.com/eskylake/network-tracker.git
cd network-tracker
go build -o network-tracker ./cmd/network-tracker
./network-tracker
```

**Run without installing:**

```bash
go run github.com/eskylake/network-tracker/cmd/network-tracker@latest
```

Make sure `$HOME/go/bin` is in your `PATH` when using `go install`.

---

## What it does

`network-tracker` is a TUI that keeps network diagnostics in one place. It is **read-only** — it does not restart services, change routes, modify Docker networks, or toggle VPNs.

### Highlights

- **Always-visible Wi-Fi bar** — current SSID and signal strength at the top on every tab
- **Wi-Fi scan tab** — list nearby networks with signal, security, and frequency
- **Overview dashboard** — quick status for VPN, DNS, routes, Docker, and public IP
- **Public IP panel** — city, region, country, coordinates, org, timezone (via ipinfo.io)
- **VPN checks** — `xvpn`, `v2raya` service, and related processes
- **Connectivity checks** — TCP targets and DNS resolution
- **Route/DNS checks** — `ip route`, `ip rule`, `/etc/resolv.conf`
- **Docker diagnostics** — network list and route overlap warnings
- **Ping tab** — continuous ping to any host or IP
- **Logs tab** — refresh history and route/DNS change events

---

## Keyboard shortcuts

### Global

| Key | Action |
| --- | --- |
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `r` | Refresh (or rescan Wi-Fi on Scan tab) |
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Toggle details for selected row |
| `q` / `Ctrl+C` | Quit |

### Scan tab

| Key | Action |
| --- | --- |
| `r` | Rescan nearby Wi-Fi networks |
| `↑` / `↓` | Select a network |
| `Enter` | Show BSSID, signal, security, frequency |

### Ping tab

| Key | Action |
| --- | --- |
| Type | Edit target while stopped |
| `Enter` | Start ping; press again to stop |
| `Backspace` | Delete one character |
| `Ctrl+U` | Clear target |

### Logs tab

| Key | Action |
| --- | --- |
| `↑` / `↓` | Scroll through log history |

---

## Tabs

| Tab | What you see |
| --- | --- |
| **Overview** | Status summary, public IP panel, quick checks |
| **VPN** | xvpn, v2raya service, v2ray/v2raya processes |
| **Connectivity** | TCP reachability, DNS resolution, public IP check |
| **Scan** | Nearby Wi-Fi networks (nmcli / iw) |
| **Routes** | Default route, routing rules, DNS config |
| **Docker** | Docker networks and overlap warnings |
| **Ping** | Live ping results for a custom target |
| **Logs** | App events and refresh history |

---

## Requirements

- Linux
- Go 1.22+

### Optional tools

Missing tools show warnings in the TUI instead of crashing the app.

| Tool | Used for |
| --- | --- |
| `nmcli` / `iw` | Wi-Fi SSID, signal, and scanning |
| `xvpn` | VPN status |
| `systemctl` | v2raya service status |
| `pgrep` | Process checks |
| `ip` | Routes and rules |
| `docker` | Docker network diagnostics |
| `ping` | Ping tab |

---

## Configuration

Config file:

```text
~/.config/network-tracker/config.yaml
```

Created automatically on first run with sensible defaults.

Edit manually:

```bash
mkdir -p ~/.config/network-tracker
$EDITOR ~/.config/network-tracker/config.yaml
```

Restart the app after changes.

### Full config reference

```yaml
# Auto-refresh interval for most checks (not public IP).
refresh_interval: 5s

# Public IP refresh interval (slower to avoid rate limits).
public_ip_refresh_interval: 60s

# Per-check timeout.
check_timeout: 4s

# Domains resolved in Connectivity tab.
domains_to_resolve:
  - google.com
  - github.com
  - cloudflare.com

# TCP targets in host:port format.
tcp_targets:
  - 1.1.1.1:53
  - 8.8.8.8:53

# Public IP provider (ipinfo.io-compatible JSON).
public_ip_provider: https://ipinfo.io/json

# Enable Docker diagnostics.
docker_enabled: true

# xvpn status command (runs without sudo).
xvpn_status_command: xvpn status

# v2raya service status command.
v2raya_status_command: systemctl is-active v2raya
```

### Common tweaks

Disable Docker:

```yaml
docker_enabled: false
```

Faster refresh:

```yaml
refresh_interval: 2s
check_timeout: 2s
```

Custom xvpn path:

```yaml
xvpn_status_command: /usr/local/bin/xvpn status
```

---

## Development

```bash
git clone https://github.com/eskylake/network-tracker.git
cd network-tracker
go run ./cmd/network-tracker
```

---

## Notes

- `xvpn status` runs without sudo.
- Docker checks need permission to access the Docker daemon.
- Ping behavior depends on your system `ping` and firewall rules.
- Public IP details require outbound access to your configured provider.
- Wi-Fi scanning uses `nmcli` first, then falls back to `iw`.

---

## License

MIT — see [LICENSE](LICENSE).

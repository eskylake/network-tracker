# Graph Report - network-tracker  (2026-06-05)

## Corpus Check
- 15 files · ~8,529 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 211 nodes · 384 edges · 9 communities
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 56 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 12|Community 12]]

## God Nodes (most connected - your core abstractions)
1. `network-tracker` - 14 edges
2. `DockerChecker()` - 13 edges
3. `BuildWithoutPublicIP()` - 12 edges
4. `model` - 11 edges
5. `WiFiChecker()` - 11 edges
6. `statusStyle()` - 10 edges
7. `finished()` - 9 edges
8. `Configuration` - 9 edges
9. `RouteChecker()` - 8 edges
10. `model` - 7 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Load()`  [INFERRED]
  cmd/network-tracker/main.go → internal/config/config.go
- `main()` --calls--> `New()`  [INFERRED]
  cmd/network-tracker/main.go → internal/app/model.go
- `ScanWiFiNetworks()` --calls--> `ParseWirelessInterfaces()`  [INFERRED]
  internal/checks/wifi_scan.go → internal/checks/system.go
- `ScanWiFiNetworks()` --calls--> `shellQuote()`  [INFERRED]
  internal/checks/wifi_scan.go → internal/checks/system.go
- `ParseNMCLIWiFiList()` --calls--> `unescapeNMCLITerse()`  [INFERRED]
  internal/checks/wifi_scan.go → internal/checks/system.go

## Communities (9 total, 0 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.17
Nodes (17): compactPingOutput(), pingOnce(), prependBounded(), prependLog(), publicIPTick(), runPing(), snapshotFromResults(), tick() (+9 more)

### Community 1 - "Community 1"
Cohesion: 0.13
Nodes (22): max(), statusStyle(), boolLabel(), clampLines(), fallback(), findResult(), hasResult(), kv() (+14 more)

### Community 2 - "Community 2"
Cohesion: 0.07
Nodes (38): code:bash (go install github.com/eskylake/network-tracker/cmd/network-t), code:bash (git clone https://github.com/eskylake/network-tracker.git), code:bash (go test ./internal/app/... -v), code:yaml (refresh_interval: 2s), code:yaml (xvpn_status_command: /usr/local/bin/xvpn status), code:yaml (v2raya_status_command: systemctl is-active v2raya.service), code:yaml (docker_enabled: false), code:bash (git clone https://github.com/eskylake/network-tracker.git) (+30 more)

### Community 3 - "Community 3"
Cohesion: 0.15
Nodes (7): Checker, CheckFunc, CommandRunner, Result, Runner, ShellCommandRunner, Status

### Community 4 - "Community 4"
Cohesion: 0.21
Nodes (9): New(), stringErrors(), Config, Default(), Load(), normalize(), Path(), Duration (+1 more)

### Community 6 - "Community 6"
Cohesion: 0.08
Nodes (40): ActiveWiFi, commandChecker, finished(), DefaultRoute, dockerInspectItem, dockerListItem, DockerNetwork, CIDROverlap() (+32 more)

### Community 7 - "Community 7"
Cohesion: 0.24
Nodes (9): ParseIWScan(), ParseNMCLIWiFiList(), ScanWiFiNetworks(), sortWiFiNetworks(), splitNMCLITerse(), TestParseIWScan(), TestParseNMCLIWiFiList(), wifiSortKey() (+1 more)

### Community 10 - "Community 10"
Cohesion: 0.20
Nodes (14): DNSChecker(), nonEmpty(), PublicIPChecker(), splitLocation(), TCPChecker(), Build(), BuildPublicIP(), BuildWithoutPublicIP() (+6 more)

### Community 12 - "Community 12"
Cohesion: 0.29
Nodes (3): mergeResults(), TestMergeResultsAppendsNewResults(), TestMergeResultsReplacesExistingByName()

## Knowledge Gaps
- **28 isolated node(s):** `Config`, `refreshMsg`, `publicIPRefreshMsg`, `pingMsg`, `wifiScanMsg` (+23 more)
  These have ≤1 connection - possible missing edges or undocumented components.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `max()` connect `Community 1` to `Community 0`?**
  _High betweenness centrality (0.195) - this node is a cross-community bridge._
- **Why does `New()` connect `Community 4` to `Community 0`?**
  _High betweenness centrality (0.155) - this node is a cross-community bridge._
- **Why does `stringErrors()` connect `Community 4` to `Community 6`?**
  _High betweenness centrality (0.118) - this node is a cross-community bridge._
- **Are the 2 inferred relationships involving `DockerChecker()` (e.g. with `finished()` and `BuildWithoutPublicIP()`) actually correct?**
  _`DockerChecker()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `BuildWithoutPublicIP()` (e.g. with `.refresh()` and `XVPNChecker()`) actually correct?**
  _`BuildWithoutPublicIP()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `WiFiChecker()` (e.g. with `finished()` and `BuildWithoutPublicIP()`) actually correct?**
  _`WiFiChecker()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **What connects `Config`, `refreshMsg`, `publicIPRefreshMsg` to the rest of the system?**
  _28 weakly-connected nodes found - possible documentation gaps or missing edges._
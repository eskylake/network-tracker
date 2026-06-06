# Graph Report - network-tracker  (2026-06-06)

## Corpus Check
- 34 files · ~8,908 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 472 nodes · 843 edges · 22 communities (20 shown, 2 thin omitted)
- Extraction: 75% EXTRACTED · 25% INFERRED · 0% AMBIGUOUS · INFERRED: 212 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d2552c75`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]

## God Nodes (most connected - your core abstractions)
1. `BuildWithoutPublicIP()` - 19 edges
2. `network-tracker` - 14 edges
3. `DockerChecker()` - 13 edges
4. `BuildWithoutPublicIP()` - 12 edges
5. `DockerChecker()` - 12 edges
6. `DockerChecker()` - 12 edges
7. `model` - 11 edges
8. `WiFiChecker()` - 11 edges
9. `model` - 10 edges
10. `statusStyle()` - 10 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Load()`  [INFERRED]
  cmd/network-tracker/main.go → internal/config/config.go
- `main()` --calls--> `New()`  [INFERRED]
  cmd/network-tracker/main.go → internal/tui/model.go
- `main()` --calls--> `New()`  [INFERRED]
  cmd/network-tracker/main.go → internal/app/model.go
- `ParseNMCLIWiFiList()` --calls--> `splitNMCLITerse()`  [INFERRED]
  internal/checks/wifi_scan.go → internal/checks/nmcli.go
- `BuildWithoutPublicIP()` --calls--> `ProcessChecker()`  [INFERRED]
  internal/check/registry.go → internal/check/vpn.go

## Communities (22 total, 2 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.09
Nodes (34): compactPingOutput(), pingOnce(), publicIPTick(), runPing(), tick(), prependBounded(), prependLog(), snapshotFromResults() (+26 more)

### Community 1 - "Community 1"
Cohesion: 0.12
Nodes (24): findResult(), hasResult(), statusStyle(), boolLabel(), clampLines(), fallback(), findResult(), hasResult() (+16 more)

### Community 2 - "Community 2"
Cohesion: 0.07
Nodes (38): code:bash (go install github.com/eskylake/network-tracker/cmd/network-t), code:bash (git clone https://github.com/eskylake/network-tracker.git), code:bash (go test ./internal/tui/... -v), code:yaml (refresh_interval: 2s), code:yaml (xvpn_status_command: /usr/local/bin/xvpn status), code:yaml (v2raya_status_command: systemctl is-active v2raya.service), code:yaml (docker_enabled: false), code:bash (git clone https://github.com/eskylake/network-tracker.git) (+30 more)

### Community 3 - "Community 3"
Cohesion: 0.12
Nodes (11): Checker, CheckFunc, CommandRunner, Result, Runner, MergeResults(), sortResults(), ShellCommandRunner (+3 more)

### Community 4 - "Community 4"
Cohesion: 0.16
Nodes (12): New(), stringErrors(), stringErrors(), Config, Default(), Load(), normalize(), Path() (+4 more)

### Community 5 - "Community 5"
Cohesion: 0.11
Nodes (23): ActiveWiFi, DefaultRoute, lastUnescapedColon(), ParseNMCLIActiveSSID(), ParseNMCLIActiveWiFi(), splitNMCLITerse(), unescapeNMCLITerse(), DNSConfigChecker() (+15 more)

### Community 6 - "Community 6"
Cohesion: 0.06
Nodes (62): commandChecker, DNSChecker(), finished(), nonEmpty(), PublicIPChecker(), splitLocation(), TCPChecker(), CIDROverlap() (+54 more)

### Community 7 - "Community 7"
Cohesion: 0.24
Nodes (9): ParseIWScan(), ParseNMCLIWiFiList(), ScanWiFiNetworks(), sortWiFiNetworks(), splitNMCLITerse(), TestParseIWScan(), TestParseNMCLIWiFiList(), wifiSortKey() (+1 more)

### Community 10 - "Community 10"
Cohesion: 0.05
Nodes (40): commandChecker, DNSChecker(), PublicIPChecker(), splitLocation(), TCPChecker(), DockerChecker(), finished(), firstLine() (+32 more)

### Community 11 - "Community 11"
Cohesion: 0.08
Nodes (26): ScanWiFiNetworks(), WiFiChecker(), ActiveWiFi, IWLinkSignal(), IWLinkSSID(), TestIWLinkSignal(), TestIWLinkSSID(), TestWiFiListFromIWScan() (+18 more)

### Community 12 - "Community 12"
Cohesion: 0.29
Nodes (3): mergeResults(), TestMergeResultsAppendsNewResults(), TestMergeResultsReplacesExistingByName()

### Community 13 - "Community 13"
Cohesion: 0.14
Nodes (20): model, findResult(), hasResult(), statusStyle(), boolLabel(), clampLines(), fallback(), kv() (+12 more)

### Community 14 - "Community 14"
Cohesion: 0.11
Nodes (19): model, model, compactPingOutput(), pingOnce(), publicIPTick(), runPing(), tick(), prependBounded() (+11 more)

### Community 15 - "Community 15"
Cohesion: 0.20
Nodes (5): Checker, CheckFunc, Result, Status, Runner

### Community 17 - "Community 17"
Cohesion: 0.50
Nodes (3): MergeResults(), sortResults(), Runner

## Knowledge Gaps
- **45 isolated node(s):** `PublicIP`, `Status`, `Result`, `Checker`, `Runner` (+40 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `New()` connect `Community 4` to `Community 0`, `Community 6`?**
  _High betweenness centrality (0.349) - this node is a cross-community bridge._
- **Why does `stringErrors()` connect `Community 4` to `Community 10`?**
  _High betweenness centrality (0.271) - this node is a cross-community bridge._
- **Are the 17 inferred relationships involving `BuildWithoutPublicIP()` (e.g. with `.refresh()` and `WiFiChecker()`) actually correct?**
  _`BuildWithoutPublicIP()` has 17 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `DockerChecker()` (e.g. with `BuildWithoutPublicIP()` and `finished()`) actually correct?**
  _`DockerChecker()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `BuildWithoutPublicIP()` (e.g. with `XVPNChecker()` and `V2RayAServiceChecker()`) actually correct?**
  _`BuildWithoutPublicIP()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 11 inferred relationships involving `DockerChecker()` (e.g. with `BuildWithoutPublicIP()` and `finished()`) actually correct?**
  _`DockerChecker()` has 11 INFERRED edges - model-reasoned connections that need verification._
- **What connects `PublicIP`, `Status`, `Result` to the rest of the system?**
  _45 weakly-connected nodes found - possible documentation gaps or missing edges._
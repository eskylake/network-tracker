package parse

import "testing"

func TestDockerNetworkList(t *testing.T) {
	raw := `{"Name":"bridge","ID":"abc"}` + "\n" + `{"Name":"vpn","ID":"def"}`
	got := DockerNetworkList(raw)
	if len(got) != 2 || got[0].Name != "bridge" || got[1].Name != "vpn" {
		t.Fatalf("unexpected docker list: %#v", got)
	}
}

func TestDockerNetworksFromInspect(t *testing.T) {
	raw := `[{"Name":"bridge","Driver":"bridge","IPAM":{"Config":[{"Subnet":"172.17.0.0/16","Gateway":"172.17.0.1"}]}}]`
	got, warnings := DockerNetworksFromInspect(raw)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %#v", warnings)
	}
	if len(got) != 1 || got[0].Subnet != "172.17.0.0/16" || got[0].Gateway != "172.17.0.1" {
		t.Fatalf("unexpected inspect: %#v", got)
	}
}

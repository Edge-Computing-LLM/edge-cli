package kubernetes

import "testing"

func TestValidateLocalNodeAddresses(t *testing.T) {
	nodes := []byte(`{"items":[{"status":{"addresses":[{"type":"InternalIP","address":"192.0.2.10"}]}}]}`)
	current := []byte(`[{"addr_info":[{"family":"inet","local":"192.0.2.10"}]}]`)
	stale := []byte(`[{"addr_info":[{"family":"inet","local":"192.0.2.20"}]}]`)
	if err := ValidateLocalNodeAddresses(nodes, current); err != nil {
		t.Fatalf("current address rejected: %v", err)
	}
	if err := ValidateLocalNodeAddresses(nodes, stale); err == nil {
		t.Fatal("stale node address accepted")
	}
}

func TestValidatePodReadiness(t *testing.T) {
	healthy := []byte(`{"items":[
		{"metadata":{"name":"operator"},"status":{"phase":"Running","containerStatuses":[{"ready":true}]}},
		{"metadata":{"name":"validator"},"status":{"phase":"Succeeded","containerStatuses":[{"ready":false}]}}
	]}`)
	if _, err := ValidatePodReadiness(healthy); err != nil {
		t.Fatalf("healthy pod set rejected: %v", err)
	}

	unhealthy := []byte(`{"items":[
		{"metadata":{"name":"operator"},"status":{"phase":"Running","containerStatuses":[{"ready":false}]}},
		{"metadata":{"name":"worker"},"status":{"phase":"Failed","containerStatuses":[{"ready":false}]}}
	]}`)
	if _, err := ValidatePodReadiness(unhealthy); err == nil {
		t.Fatal("unhealthy pod set accepted")
	}
}

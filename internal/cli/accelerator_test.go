package cli

import "testing"

func TestProfileForAccelerator(t *testing.T) {
	if got := profileForAccelerator("cpu"); got != "cpu-k3s" {
		t.Fatalf("CPU profile = %q", got)
	}
	if got := profileForAccelerator("nvidia"); got != "geforce-940m-k3s" {
		t.Fatalf("NVIDIA profile = %q", got)
	}
}

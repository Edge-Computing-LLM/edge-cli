package infra

import "testing"

func TestReplaceSignedBy(t *testing.T) {
	got := replaceSignedBy("deb https://nvidia.github.io/libnvidia-container/stable/deb/amd64 /\n")
	want := "deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://nvidia.github.io/libnvidia-container/stable/deb/amd64 /\n"
	if got != want {
		t.Fatalf("replaceSignedBy() = %q, want %q", got, want)
	}
}

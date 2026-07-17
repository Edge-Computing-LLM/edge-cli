package nvidia

import "testing"

func TestHasAllocatableGPU(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{name: "gpu", output: "1", want: true},
		{name: "multiple nodes", output: "\n0\n2\n", want: true},
		{name: "zero", output: "0", want: false},
		{name: "missing", output: "<none>", want: false},
		{name: "empty", output: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAllocatableGPU(tt.output); got != tt.want {
				t.Fatalf("HasAllocatableGPU(%q) = %t, want %t", tt.output, got, tt.want)
			}
		})
	}
}

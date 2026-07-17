package nvidia

import (
	"context"
	"strconv"
	"strings"

	"github.com/Edge-Computing-LLM/edge-cli/internal/execx"
)

const GPUResource = "nvidia.com/gpu"

func HostAvailable(ctx context.Context, r execx.Runner) bool {
	_, err := r.Output(ctx, execx.Command{Name: "nvidia-smi", Args: []string{"--query-gpu=name", "--format=csv,noheader"}})
	return err == nil
}

func ClusterAvailable(ctx context.Context, r execx.Runner) bool {
	out, err := r.Output(ctx, execx.Command{Name: "kubectl", Args: []string{
		"get", "nodes", "-o", "jsonpath={range .items[*]}{.status.allocatable.nvidia\\.com/gpu}{\"\\n\"}{end}",
	}})
	return err == nil && HasAllocatableGPU(out)
}

func HasAllocatableGPU(output string) bool {
	for _, field := range strings.Fields(output) {
		count, err := strconv.ParseInt(field, 10, 64)
		if err == nil && count > 0 {
			return true
		}
	}
	return false
}

func Checks(ctx context.Context, r execx.Runner) []execx.Result {
	return []execx.Result{
		r.Check(ctx, "nvidia-smi", execx.Command{Name: "nvidia-smi"}),
		r.Check(ctx, "NVIDIA container toolkit package", execx.Command{Name: "dpkg-query", Args: []string{"-W", "nvidia-container-toolkit"}}),
	}
}

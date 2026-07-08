package nvidia

import (
	"context"

	"github.com/Edge-Computing-LLM/edge-cli/internal/execx"
)

func Checks(ctx context.Context, r execx.Runner) []execx.Result {
	return []execx.Result{
		r.Check(ctx, "nvidia-smi", execx.Command{Name: "nvidia-smi"}),
		r.Check(ctx, "NVIDIA container toolkit package", execx.Command{Name: "dpkg-query", Args: []string{"-W", "nvidia-container-toolkit"}}),
	}
}

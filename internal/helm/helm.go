package helm

import (
	"context"

	"github.com/Edge-Computing-LLM/edge-cli/internal/execx"
)

type Client struct {
	Runner execx.Runner
}

func (c Client) Status(ctx context.Context, release, namespace string) execx.Result {
	return c.Runner.Check(ctx, "helm release "+release, execx.Command{Name: "helm", Args: []string{"status", release, "-n", namespace}})
}

func (c Client) ListAll(ctx context.Context) execx.Result {
	return c.Runner.Check(ctx, "helm releases", execx.Command{Name: "helm", Args: []string{"list", "-A"}})
}

func (c Client) DependencyBuild(ctx context.Context, dir string) error {
	return c.Runner.Run(ctx, execx.Command{Name: "helm", Args: []string{"dependency", "build", "."}, Dir: dir, Mutates: true})
}

func (c Client) UpgradeInstall(ctx context.Context, dir string, args []string) error {
	return c.Runner.Run(ctx, execx.Command{Name: "helm", Args: args, Dir: dir, Mutates: true})
}

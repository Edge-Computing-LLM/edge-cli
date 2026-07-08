package kubernetes

import (
	"context"

	"github.com/Edge-Computing-LLM/edge-cli/internal/execx"
)

type Client struct {
	Runner execx.Runner
}

func (c Client) ClusterInfo(ctx context.Context) execx.Result {
	return c.Runner.Check(ctx, "kubectl cluster-info", execx.Command{Name: "kubectl", Args: []string{"cluster-info"}})
}

func (c Client) Nodes(ctx context.Context) execx.Result {
	return c.Runner.Check(ctx, "cluster nodes", execx.Command{Name: "kubectl", Args: []string{"get", "nodes", "-o", "wide"}})
}

func (c Client) RuntimeClassNvidia(ctx context.Context) execx.Result {
	return c.Runner.Check(ctx, "NVIDIA RuntimeClass", execx.Command{Name: "kubectl", Args: []string{"get", "runtimeclass", "nvidia"}})
}

func (c Client) Namespace(ctx context.Context, ns string) execx.Result {
	return c.Runner.Check(ctx, "namespace "+ns, execx.Command{Name: "kubectl", Args: []string{"get", "namespace", ns}})
}

func (c Client) Workloads(ctx context.Context, ns string) execx.Result {
	return c.Runner.Check(ctx, "workloads "+ns, execx.Command{Name: "kubectl", Args: []string{"get", "pods,deploy,statefulset,svc,pvc", "-n", ns, "-o", "wide"}})
}

func (c Client) Service(ctx context.Context, ns, name string) execx.Result {
	return c.Runner.Check(ctx, "service "+name, execx.Command{Name: "kubectl", Args: []string{"get", "svc", "-n", ns, name}})
}

func (c Client) Rollout(ctx context.Context, kindName, ns, timeout string) error {
	return c.Runner.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"rollout", "status", kindName, "-n", ns, "--timeout", timeout}, Mutates: true})
}

func (c Client) WaitAllPods(ctx context.Context, ns, timeout string) error {
	return c.Runner.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"wait", "--for=condition=Ready", "pod", "--all", "-n", ns, "--timeout", timeout}, Mutates: true})
}

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

func (c Client) LocalNodeAddress(ctx context.Context) execx.Result {
	nodes, err := c.Runner.Output(ctx, execx.Command{Name: "kubectl", Args: []string{"get", "nodes", "-o", "json"}})
	if err != nil {
		return execx.Result{Label: "k3s node address", Output: nodes, Err: err}
	}
	addresses, err := c.Runner.Output(ctx, execx.Command{Name: "ip", Args: []string{"-j", "-4", "address", "show"}})
	if err != nil {
		return execx.Result{Label: "k3s node address", Output: addresses, Err: err}
	}
	if err := ValidateLocalNodeAddresses([]byte(nodes), []byte(addresses)); err != nil {
		return execx.Result{Label: "k3s node address", Err: err}
	}
	return execx.Result{Label: "k3s node address", Output: "advertised InternalIP matches a current host interface"}
}

func (c Client) RuntimeClassNvidia(ctx context.Context) execx.Result {
	return c.Runner.Check(ctx, "NVIDIA RuntimeClass", execx.Command{Name: "kubectl", Args: []string{"get", "runtimeclass", "nvidia"}})
}

func (c Client) PodsReady(ctx context.Context, namespace string) execx.Result {
	output, err := c.Runner.Output(ctx, execx.Command{Name: "kubectl", Args: []string{"get", "pods", "-n", namespace, "-o", "json"}})
	if err != nil {
		return execx.Result{Label: namespace + " pod health", Output: output, Err: err}
	}
	summary, err := ValidatePodReadiness([]byte(output))
	return execx.Result{Label: namespace + " pod health", Output: summary, Err: err}
}

type nodeList struct {
	Items []struct {
		Status struct {
			Addresses []struct {
				Type    string `json:"type"`
				Address string `json:"address"`
			} `json:"addresses"`
		} `json:"status"`
	} `json:"items"`
}

type hostInterface struct {
	AddressInfo []struct {
		Family string `json:"family"`
		Local  string `json:"local"`
	} `json:"addr_info"`
}

func ValidateLocalNodeAddresses(nodesJSON, addressesJSON []byte) error {
	var nodes nodeList
	if err := json.Unmarshal(nodesJSON, &nodes); err != nil {
		return fmt.Errorf("decode Kubernetes nodes: %w", err)
	}
	var interfaces []hostInterface
	if err := json.Unmarshal(addressesJSON, &interfaces); err != nil {
		return fmt.Errorf("decode host interfaces: %w", err)
	}
	local := map[string]bool{}
	for _, iface := range interfaces {
		for _, address := range iface.AddressInfo {
			if address.Family == "inet" {
				local[address.Local] = true
			}
		}
	}
	if len(nodes.Items) == 0 {
		return fmt.Errorf("no Kubernetes nodes found")
	}
	for _, node := range nodes.Items {
		internalIP := ""
		for _, address := range node.Status.Addresses {
			if address.Type == "InternalIP" {
				internalIP = address.Address
				break
			}
		}
		if internalIP == "" {
			return fmt.Errorf("k3s node has no InternalIP")
		}
		if !local[internalIP] {
			return fmt.Errorf("k3s node InternalIP is not assigned to a current host interface; check node-ip and flannel-iface in /etc/rancher/k3s/config.yaml")
		}
	}
	return nil
}

type podList struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Status struct {
			Phase             string `json:"phase"`
			ContainerStatuses []struct {
				Ready bool `json:"ready"`
			} `json:"containerStatuses"`
		} `json:"status"`
	} `json:"items"`
}

func ValidatePodReadiness(data []byte) (string, error) {
	var pods podList
	if err := json.Unmarshal(data, &pods); err != nil {
		return "", fmt.Errorf("decode Kubernetes pods: %w", err)
	}
	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found")
	}
	active, completed := 0, 0
	var unhealthy []string
	for _, pod := range pods.Items {
		if pod.Status.Phase == "Succeeded" {
			completed++
			continue
		}
		active++
		ready := pod.Status.Phase == "Running" && len(pod.Status.ContainerStatuses) > 0
		for _, status := range pod.Status.ContainerStatuses {
			ready = ready && status.Ready
		}
		if !ready {
			unhealthy = append(unhealthy, pod.Metadata.Name+" ("+pod.Status.Phase+")")
		}
	}
	if len(unhealthy) > 0 {
		return fmt.Sprintf("%d active, %d completed", active, completed), fmt.Errorf("unhealthy pods: %s", strings.Join(unhealthy, ", "))
	}
	return fmt.Sprintf("%d active pods ready, %d completed", active, completed), nil
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

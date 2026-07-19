# Configuration

Default config path:

```text
~/.edge-cli/config.yaml
```

Create a config file:

```bash
edge config init
```

Show the resolved config:

```bash
edge config show
```

## Schema

```yaml
repos:
  k3sNvidiaEdge: /media/waqasm86/External1/Waqas-Projects/Project-Linux-Kubernetes-Nvidia/Project-Edge-Computing-LLM/k3s-nvidia-edge
  llmObservabilityStack: /media/waqasm86/External1/Waqas-Projects/Project-Linux-Kubernetes-Nvidia/Project-Edge-Computing-LLM/llm-observability-stack

cluster:
  kubeconfig: ""
  defaultNamespace: llm-observability

gpu:
  vendor: nvidia
```

## Repo Path Overrides

Use config for persistent paths, or use flags for one-off commands:

```bash
edge install infra --repo-path /path/to/k3s-nvidia-edge --yes
edge install observability --repo-path /path/to/llm-observability-stack --yes
```

## Kubeconfig

When `cluster.kubeconfig` is set, `edge-cli` exports it to `KUBECONFIG` for the
current command process before running Kubernetes or Helm operations.

## GPU Vendor

Only `nvidia` is supported. Any other value is rejected during config loading.

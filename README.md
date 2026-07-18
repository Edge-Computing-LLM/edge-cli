# edge-cli

`edge-cli` is the unified Go command-line entry point for the `Edge-Computing-LLM`
GitHub organization.

It coordinates two current layers:

- `k3s-nvidia-edge`: infrastructure for a local Linux + k3s + NVIDIA GPU host.
- `llm-observability-stack`: LLMOps workloads, Helm charts, Open WebUI, Ollama,
  OpenTelemetry, Prometheus, Grafana, and related services.

The organization also publishes
[`qwen-gguf-observability`](https://github.com/Edge-Computing-LLM/qwen-gguf-observability),
a read-only evidence companion for the deployed Qwen runtime. It consumes the
status produced by these layers; `edge-cli` does not install it as another
cluster layer.

Dashboard presentation is owned by the Helm-provisioned Grafana dashboards in
`llm-observability-stack`; there is no separate browser repository or
browser-side Kubernetes credential path.

The CLI supports automatic NVIDIA/CPU selection. It does not currently install
AMD, Intel, or Apple Silicon accelerator runtimes.

## Why this exists

The organization has separate repositories for infrastructure and LLMOps
workloads. `edge-cli` provides one installable control plane so operators can run
preflight checks, installs, validation, status, logs, and safe uninstall flows
without treating each repository as a disconnected CLI application.

Future repositories can be added as new layers in `internal/platform` and new Go
modules under `internal/modules`.

## Install

Build from source:

```bash
go mod tidy
go build -o edge ./cmd/edge
sudo install -m 0755 edge /usr/local/bin/edge
edge version
```

## Configuration

Default config path:

```text
~/.edge-cli/config.yaml
```

Create it:

```bash
edge config init
edge config show
```

Default configuration:

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

Repository paths can also be overridden per command:

```bash
edge install infra --repo-path /path/to/k3s-nvidia-edge --yes
edge install observability --repo-path /path/to/llm-observability-stack --yes
```

## Basic Commands

```bash
edge doctor
edge status
edge install infra --yes
edge validate infra
edge install observability --yes
edge validate observability
edge install all --yes
edge logs
edge repo list
edge repo doctor
```

## Full Install

```bash
edge install all --accelerator auto --yes
edge status
```

`auto` deploys the NVIDIA infrastructure layer when `nvidia-smi` detects a
working GPU. Without NVIDIA hardware it skips the toolkit and GPU Operator,
installs or validates basic k3s, and selects `values.cpu-k3s.yaml`.

On the validated GeForce 940M profile, the application smoke test targets
`qwen-1-8b-chat-q4-k-m-local`; CPU and general profiles retain their configured
Gemma model defaults.

Explicit modes are useful in automation:

```bash
edge install all --accelerator nvidia --yes
edge install all --accelerator cpu --yes
```

`edge install all` runs:

1. Host accelerator detection.
2. Linux/k3s installation and accelerator-specific infrastructure setup.
3. Basic k3s validation, plus NVIDIA and CUDA validation in NVIDIA mode.
4. `llm-observability-stack` Helm install.
5. Observability validation.
6. Next command and URL hints.

An empty local k3s cluster with only the default k3s system components is a
valid starting point. `edge install infra` is responsible for preparing the
NVIDIA runtime and GPU Operator layer before any GPU-backed LLM workload is
installed.

## Infra Layer

`edge install infra` uses the local `k3s-nvidia-edge` project as the
infrastructure source. It validates the repo path, checks Linux and NVIDIA
preconditions, installs required host packages and NVIDIA Container Toolkit when
not skipped, installs k3s when missing, installs the GPU Operator through the
bundled Helm chart by default, and validates GPU readiness with a CUDA pod.

Useful commands:

```bash
edge install infra --yes
edge install infra --skip-k3s-install --yes
edge install infra --skip-toolkit-install --yes
edge validate infra
edge uninstall infra --yes
```

## Observability Layer

`edge install observability` uses the local `llm-observability-stack` chart. It
validates the infra layer first, builds Helm dependencies, installs the chart,
waits for Ollama, Open WebUI, and OpenTelemetry Collector, then runs validation.
For GPU profiles, `edge-cli` does not allow bypassing infra validation and
forcibly keeps `gpu-operator`, `nvidia-device-plugin`, and `dcgm-exporter`
disabled in the observability chart. Those components belong to
`k3s-nvidia-edge`.

On single-GPU laptops, run `edge validate infra` before installing
observability. After Ollama is running and reserving `nvidia.com/gpu: 1`,
observability validation checks base readiness without launching another CUDA
pod that would contend for the only GPU.

Useful commands:

```bash
edge install observability --profile geforce-940m-k3s --yes
edge install observability --values custom.yaml --set ollama.enabled=true --yes
edge validate observability
edge logs --tail 200
edge uninstall observability --yes --keep-namespace
```

## Safe Uninstall

Uninstall commands require `--yes`. Observability uninstall does not delete user
data outside Kubernetes resources. Use `--keep-namespace` to retain the
namespace. Infra uninstall keeps k3s unless `--k3s` is explicitly set.

```bash
edge uninstall observability --yes --keep-namespace
edge uninstall infra --yes
edge uninstall all --yes
```

## Troubleshooting

Run the broad checks first:

```bash
edge doctor
edge repo doctor
edge status
```

Common issues:

- Missing `nvidia-smi`: install or repair the NVIDIA driver on the host.
- Missing `runtimeclass/nvidia`: validate the GPU Operator and NVIDIA Container
  Toolkit setup with `edge validate infra`.
- Helm release missing: run `edge install observability --yes`.
- Pods not ready: run `edge logs` and inspect the namespace with
  `kubectl get pods -n llm-observability -o wide`.

## Development

```bash
go test ./...
go build -o edge ./cmd/edge
```

The project is 100% Go. It does not create Bash scripts. External platform
commands such as `kubectl`, `helm`, `apt-get`, `systemctl`, `k3s`, and
`nvidia-smi` are executed through Go's `os/exec`.

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Programming language and script boundaries](docs/LANGUAGE-BOUNDARIES.md)
- [Accelerator selection](docs/ACCELERATOR-MODES.md)
- [Local dependency and repository inventory](docs/LOCAL-DEPENDENCY-INVENTORY.md)
- [Live validation - 2026-07-17](docs/LIVE-VALIDATION-2026-07-17.md)
- [Live validation - 2026-07-08](docs/LIVE-VALIDATION-2026-07-08.md)
- [Command reference](docs/COMMANDS.md)
- [Configuration](docs/CONFIGURATION.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)
- [Contributing](CONTRIBUTING.md)
- [Security](SECURITY.md)
- [Qwen GGUF runtime evidence companion](https://github.com/Edge-Computing-LLM/qwen-gguf-observability)

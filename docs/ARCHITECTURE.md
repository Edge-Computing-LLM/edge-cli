# Architecture

`edge-cli` is the unified Go CLI control plane for the
`Edge-Computing-LLM` organization.

## Layers

The platform is split into ordered layers:

- Layer 0, `edge-cli`: command-line control plane, configuration, checks, workflows, and
  module orchestration.
- Layer 1, `k3s-nvidia-edge`: infrastructure layer for Linux, k3s, NVIDIA Container
  Toolkit, GPU Operator, RuntimeClass, and CUDA validation.
- Layer 2, `llm-observability-stack`: LLMOps layer for Helm workloads such as Ollama,
  Open WebUI, OpenTelemetry Collector, Prometheus, Grafana, and related tools.
- Evidence companion, `gguf-observability`: read-only, model-selectable GGUF runtime contract
  checks and sanitized point-in-time evidence. It owns no cluster resources and
  is intentionally outside the ordered install/uninstall layer graph.
- Future Layer 3 repositories that deploy resources, such as a data/storage
  stack, should declare their dependency on the previous layer before they can
  be installed.

## Module Model

Modules live under `internal/modules`.

Current modules:

- `infra`: owns local NVIDIA+k3s infrastructure workflows.
- `observability`: owns LLMOps Helm workflows and validates infra before
  installing GPU-backed workloads.

Future organization repositories should be added as modules with a narrow
interface:

- repo path validation
- doctor checks
- install workflow
- validation workflow
- status/logs where relevant
- safe uninstall behavior

Layer metadata lives in `internal/platform`. Install workflows run in catalog
order. Uninstall workflows run in reverse order where layers own resources that
depend on earlier layers.

## Dependency Gates

`llm-observability-stack` GPU profiles require a valid infra layer first:

- `RuntimeClass/nvidia` exists
- nodes advertise allocatable `nvidia.com/gpu`
- GPU Operator/device plugin/DCGM are owned by `k3s-nvidia-edge`
- CUDA validation succeeds from the infra layer

`edge-cli` rejects GPU observability installs that try to skip infra validation.
It also appends Helm overrides that keep `gpu-operator.enabled`,
`nvidia-device-plugin.enabled`, and `dcgm-exporter.enabled` false in the
observability chart.

Full infra validation owns CUDA pod execution. Observability dependency checks
verify the ready base layer without launching a CUDA pod, so validation remains
usable after Ollama has reserved the only GPU on a low-VRAM laptop.

After Layer 2 is healthy, `gguf-observability` may independently read the
Kubernetes, Helm, Ollama, and `nvidia-smi` status surfaces. This avoids copying
deployment logic into an evidence repository or coupling evidence capture to an
`edge install` operation.

## Execution Model

The CLI is 100% Go. It does not create Bash scripts. Host and cluster operations
that must be delegated to platform tools are executed through Go's `os/exec`.

Examples of external commands:

- `kubectl`
- `helm`
- `systemctl`
- `apt-get`
- `k3s`
- `nvidia-smi`

Argument handling, checks, workflow ordering, config loading, path validation,
and safety gates are implemented in Go.

## NVIDIA-Only Scope

This project targets NVIDIA GPU based LLMOps. The config schema intentionally
accepts only:

```yaml
gpu:
  vendor: nvidia
```

AMD, Intel, Apple Silicon, and CPU-first accelerator abstractions are out of
scope.

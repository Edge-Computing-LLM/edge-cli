# Architecture

`edge-cli` is the unified Go CLI control plane for the
`Edge-Computing-LLM` organization.

## Layers

The platform is split into three layers:

- `edge-cli`: command-line control plane, configuration, checks, workflows, and
  module orchestration.
- `k3s-nvidia-edge`: infrastructure layer for Linux, k3s, NVIDIA Container
  Toolkit, GPU Operator, RuntimeClass, and CUDA validation.
- `llm-observability-stack`: LLMOps layer for Helm workloads such as Ollama,
  Open WebUI, OpenTelemetry Collector, Prometheus, Grafana, and related tools.

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

AMD, Intel, Apple Silicon, and CPU-only accelerator abstractions are out of
scope.

# Command Reference

## Global Flags

```bash
edge --config /path/to/config.yaml --timeout 10m --verbose <command>
edge --dry-run <command>
```

## Version

```bash
edge version
```

Prints the CLI version.

## Doctor

```bash
edge doctor
```

Runs broad diagnostics for Linux, required commands, sudo, k3s, kubectl, Helm,
NVIDIA driver/runtime, GPU Operator, CUDA validation support, and observability
readiness.

## Status

```bash
edge status
```

Prints live infrastructure and observability state.

## Install

```bash
edge install infra --yes
edge install observability --yes
edge install all --yes
```

`edge install all` executes infra preflight, infra install, infra validation,
observability install, and observability validation.

## Validate

```bash
edge validate infra
edge validate observability
```

`infra` validation includes k3s, node readiness, NVIDIA RuntimeClass, GPU
Operator pods, GPU allocatable resources, and a CUDA `nvidia-smi` validation pod.

`observability` validation checks infra first, then verifies the Helm release,
namespace, Ollama, Open WebUI, OpenTelemetry Collector, optional
Prometheus/Grafana services, pod readiness, and optional Ollama smoke behavior.

## Uninstall

```bash
edge uninstall observability --yes
edge uninstall observability --yes --keep-namespace
edge uninstall infra --yes
edge uninstall infra --yes --k3s
edge uninstall all --yes
```

Uninstall commands require `--yes`. k3s is kept unless `--k3s` is explicitly
provided for infra uninstall.

## Logs

```bash
edge logs
edge logs --tail 200
```

Prints recent logs from core observability workloads.

## Config

```bash
edge config init
edge config show
edge config path
```

## Repos

```bash
edge repo list
edge repo doctor
```

Lists configured repositories and checks local path, Git remote, and branch
state.

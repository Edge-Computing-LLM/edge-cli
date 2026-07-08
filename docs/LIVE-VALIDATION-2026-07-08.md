# Live Validation - 2026-07-08

Validated on local Xubuntu 24, single-node k3s, NVIDIA GeForce 940M.

Layer order tested:

1. Empty/mostly-empty k3s baseline: CoreDNS and local-path-provisioner only.
2. `edge install observability --skip-infra-check --profile geforce-940m-k3s --yes` refused the GPU install.
3. `edge validate infra` failed before Layer 1 because `nvidia.com/gpu` was not allocatable.
4. `edge install infra --skip-base-package-install --skip-toolkit-install --skip-k3s-install --yes` installed/upgraded the `k3s-nvidia-edge` GPU Operator wrapper.
5. `edge validate infra` passed and ran a CUDA `nvidia-smi` pod.
6. `edge install observability --profile geforce-940m-k3s --yes` deployed Ollama, Open WebUI, Redis, and OpenTelemetry Collector.
7. Repeated install/validate upgraded the observability Helm release safely.
8. `edge uninstall observability --yes` removed Layer 2 while Layer 1 stayed healthy.
9. Reinstall and validation passed again.

Observed ownership:

- `k3s-nvidia-edge` Helm release lives in `gpu-operator`.
- GPU Operator, NVIDIA device plugin, DCGM exporter, NFD, RuntimeClass, and `nvidia.com/gpu` are provided by Layer 1.
- `llm-observability-stack` rendered no GPU Operator, NVIDIA device plugin, or DCGM exporter resources.
- `llm-observability-stack` package did not include the old substrate chart dependencies.

Single-GPU note:

`edge validate infra` runs CUDA validation before Layer 2 is installed. Once Ollama is running and reserving `nvidia.com/gpu: 1`, observability validation checks base readiness without launching another CUDA pod.

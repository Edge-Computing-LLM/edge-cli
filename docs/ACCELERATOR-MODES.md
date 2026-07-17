# Accelerator selection

`edge install all` and `edge install observability` accept
`--accelerator auto|nvidia|cpu`.

## Automatic mode

For the full layered installation, `auto` probes the host with `nvidia-smi` before
making changes:

- NVIDIA detected: install Ubuntu prerequisites, NVIDIA Container Toolkit, k3s,
  `k3s-nvidia-edge`, and the NVIDIA observability profile.
- NVIDIA not detected: skip NVIDIA toolkit and GPU Operator components, install or
  validate basic k3s, and use `values.cpu-k3s.yaml`.

For an observability-only installation, `auto` inspects Kubernetes allocatable
resources. It selects NVIDIA only when at least one node advertises a positive
`nvidia.com/gpu` value; otherwise it selects CPU.

## Explicit modes

Use `--accelerator nvidia` where the GPU layer must be present and a missing GPU
should fail validation. Use `--accelerator cpu` to guarantee that no NVIDIA runtime
class, GPU request, GPU Operator, device plugin, or DCGM dependency is required by
the application layer.

An explicit `--profile` on `edge install observability` takes precedence over
accelerator-based profile selection.

## Detection boundaries

Host detection answers whether NVIDIA infrastructure should be installed. Cluster
detection answers whether a workload can request a Kubernetes GPU. A host may pass
`nvidia-smi` while Kubernetes still reports no GPU; that means the GPU layer needs
installation or repair rather than that a GPU workload is ready.

## Interrupted first installs

If a first Helm client times out while large images continue pulling, Kubernetes
may later become healthy while Helm remains `pending-install`. With `--yes`, the
observability installer detects that exact state, reconciles the recorded revision,
and then proceeds with the normal idempotent upgrade. Other pending Helm states are
not rewritten automatically.

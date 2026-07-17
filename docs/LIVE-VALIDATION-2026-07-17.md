# Live validation — 2026-07-17

## Host

- Ubuntu 24.04.3 LTS
- k3s v1.36.2+k3s1 with containerd 2.3.2-k3s2
- NVIDIA GeForce 940M, 1 GiB VRAM
- NVIDIA driver 580.95.05, reported CUDA capability 13.0
- Helm v4.2.3
- Go 1.26.5
- Python 3.11.15 at `/usr/local/bin/python3.11`

## Accelerator orchestration

The new `--accelerator auto` cluster probe selected CPU before the NVIDIA device
plugin advertised `nvidia.com/gpu`, then selected NVIDIA after Layer 1 completed.
Explicit CPU dry-run rendered `values.cpu-k3s.yaml` and skipped NVIDIA toolkit,
GPU Operator, RuntimeClass, DCGM, and GPU scheduling checks.

## NVIDIA layer

`edge install infra --yes` installed NVIDIA Container Toolkit 1.19.1 and the local
`k3s-nvidia-edge` GPU Operator chart v26.3.3. GPU Operator, Node Feature Discovery,
the toolkit, device plugin, validator, and DCGM Exporter became healthy. Kubernetes
advertised one `nvidia.com/gpu` resource.

The CUDA validation pod used `nvidia/cuda:12.8.1-base-ubuntu24.04`, completed, and
printed the host GeForce 940M through `nvidia-smi`.

## Repository tests

- `go test ./...`: pass
- `go vet ./...`: pass
- Go builds for all three CLIs: pass
- Helm lint and default/CPU/NVIDIA renders: pass
- Python 3.11 smoke suite: 17 passed
- Frontend typecheck and production build: pass
- CPU full-install dry-run: pass
- Live NVIDIA infra validation with `--skip-cuda` while the GPU was reserved: pass

The frontend build reports a non-failing Vite warning for a JavaScript chunk over
500 kB. It is a performance optimization opportunity, not a correctness failure.

## Operational observations

First-time Ollama and Open WebUI pulls are multi-gigabyte operations on this node.
The initial application Helm wait reached 15 minutes while those pulls continued.
Ollama took approximately 25 minutes and Open WebUI approximately 33 minutes to
pull. Redis, OpenTelemetry Collector, PVC provisioning, GPU scheduling, and
services were already healthy.

After the pulls completed, Helm revision 1 still recorded `pending-install`
because the original client had timed out. `helm rollback ... 1` reconciled the
already healthy resources as deployed, and unchanged idempotent installs then
completed successfully in under 20 seconds. Final Helm revision 4 is `deployed`.

Final live application checks passed:

- all four application pods Ready with zero restarts;
- Open WebUI `/health` returned `{"status":true}`;
- Ollama loaded `gemma3-1b-it-gguf-local` and returned exactly `validation ok`;
- Ollama selected CUDA on the GeForce 940M and offloaded 23 of 27 layers;
- `ollama ps` reported a 57% CPU / 43% GPU split for the 1.2 GB loaded model.

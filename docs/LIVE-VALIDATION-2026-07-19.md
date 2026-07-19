# Live validation: 2026-07-19

All local Go gates passed on Ubuntu 24.04: module verification, formatting,
unit tests, race tests, vet, build, and `govulncheck` (no reachable
vulnerabilities).

The live `edge validate infra --skip-cuda` audit exposed a false-positive in
the earlier validator. A successful `kubectl get pods` command was reported as
healthy even when the returned GPU Operator pods included CrashLoopBackOff and
unready containers. The host had also rebooted onto a different network
interface while k3s retained a stale InternalIP and Flannel interface.

Validation now parses pod phase and container readiness, permits successfully
Completed validator pods, and rejects every other unready pod. It also compares
the advertised k3s InternalIP with the host's current IPv4 interfaces. These
checks run before CUDA validation, so a stale network cannot be mistaken for a
healthy GPU substrate merely because `nvidia.com/gpu` remains allocatable.

No kubeconfig content, Secrets, host addresses, pod logs, prompts, responses,
or model data are recorded here.

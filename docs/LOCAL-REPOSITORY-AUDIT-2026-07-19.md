# Local repository audit: 2026-07-19

## Scope and method

The audit recursively inspected all 64 Git worktrees under
`Project-Linux-Kubernetes-Nvidia`, including remotes, tracking branches,
working-tree state, latest commits, repository instructions, modules, charts,
CI workflows, and direct Edge platform relationships. Every `origin` fetch
completed successfully. All worktrees were clean at the baseline.

The four `Edge-Computing-LLM` organization repositories were compared with the
GitHub organization API, pull requests, branches, and Actions runs. Their prior
dated branches had already been merged and had byte-identical trees to `main`,
so the local checkouts were safely fast-forwarded to `origin/main` before this
audit's new branches were created.

## Collection status

- 64 Git worktrees total
- 4 organization-owned deployable/evidence repositories
- 7 third-party dashboard template candidates
- 53 other upstream/reference worktrees
- 42 clean reference branches behind their freshly fetched upstream tip
- 22 clean branches synchronized at the baseline
- 0 locally divergent or ahead-only reference branches
- 0 failed remote fetches

Reference branches were not fast-forwarded because they are vendor research
inputs, not organization-owned deployment sources. Their remote-tracking refs
are current, which is sufficient for dependency and release comparison without
rewriting the user's chosen checkout commits.

## Reference branches behind upstream

The values below are commit counts (`ahead/behind`) after fetching.

| Area | Repository | Ahead/behind |
|---|---|---:|
| Apache | echarts | 0/172 |
| Artifact Hub | hub | 0/2 |
| Cloudflare | cloudflared | 0/13 |
| CoreDNS | coredns | 0/82 |
| Dashboard candidate | next-shadcn-admin-dashboard | 0/18 |
| Dashboard candidate | next-shadcn-dashboard-starter | 0/86 |
| Dashboard candidate | nextadmin-dashboard | 0/1 |
| Go | go | 0/5 |
| Helm | helm | 0/6 |
| Kubernetes | kubernetes-client/python | 0/22 |
| Kubernetes | dra-driver-nvidia-gpu (two worktrees) | 0/36 each |
| Kubernetes | node-feature-discovery | 0/19 |
| Kubernetes | client-go | 0/19 |
| Kubernetes | kubectl | 0/20 |
| NVIDIA | DCGM | 0/1 |
| NVIDIA | dcgm-exporter | 0/1 |
| NVIDIA | go-dcgm | 0/2 |
| NVIDIA | gpu-operator | 0/85 |
| NVIDIA | k8s-device-plugin | 0/43 |
| NVIDIA | libnvidia-container | 0/4 |
| NVIDIA | nvidia-container-toolkit | 0/71 |
| Ollama | ollama | 0/35 |
| OpenAI | codex | 0/11 |
| OpenAI | openai-agents-python | 0/4 |
| OpenTelemetry | collector | 0/25 |
| OpenTelemetry | collector-contrib | 0/183 |
| OpenTelemetry | Helm charts (primary) | 0/19 |
| OpenTelemetry | operator | 0/22 |
| OpenTelemetry | Python | 0/28 |
| OpenTelemetry | Python contrib | 0/20 |
| OpenTelemetry | Python GenAI | 0/71 |
| OpenTelemetry | semantic-conventions-genai | 0/10 |
| Prometheus | community Helm charts (primary) | 0/40 |
| Prometheus | kube-prometheus | 0/2 |
| Rancher/k3s | k3s | 0/32 |
| Rancher | local-path-provisioner | 0/4 |
| Traefik | traefik | 0/8 |
| Zensical | zensical | 0/2 |
| Duplicate worktree | Prometheus Community charts | 0/109 |
| Duplicate worktree | Grafana charts | 0/4 |
| Duplicate worktree | OpenTelemetry charts | 0/41 |

## Edge repository gates

All four repositories passed formatting, `go mod verify`, unit tests, race
tests, vet, builds, and `govulncheck`; no reachable Go vulnerabilities were
found. The Helm repositories additionally passed dependency resolution, lint,
and default/CPU/local/GeForce/full-NVIDIA rendering. The current GitHub `main`
workflow is green in all four repositories and Dependabot reports no open alert
in the three repositories where alert access is enabled.

Coverage remains the largest engineering-quality gap. Core package coverage
ranges from roughly 17% to 56%, while orchestration and CLI entry packages are
mostly untested. Code scanning has no analysis workflow, secret scanning is
disabled, and Dependabot alerts are disabled for `llm-observability-stack`.

## Live Ubuntu/k3s/NVIDIA result

The host runs Ubuntu 24.04, k3s v1.36.2+k3s1, and GPU Operator v26.3.3 with a
GeForce 940M advertising one allocatable GPU. Qwen is registered, resident,
using GPU layers, and within the 850 MiB evidence ceiling.

The cluster is not fully healthy after a network-adapter change. k3s advertises
a stale InternalIP and Flannel interface that no longer exist on the host.
Pod-to-API routing fails, which cascades into GPU Operator, Node Feature
Discovery, metrics-server, kube-state-metrics, and node-exporter failures. The
new validators reject this state instead of accepting GPU capacity as a proxy
for substrate health.

Repairing `/etc/rancher/k3s/config.yaml` and restarting k3s requires local sudo
authentication and was therefore not performed automatically. No addresses,
kubeconfig content, Secrets, prompts, responses, or model weights are recorded
in this audit.

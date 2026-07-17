# Local dependency and repository inventory

Validated on 2026-07-17 against
`/media/waqasm86/External1/Waqas-Projects/Project-Linux-Kubernetes-Nvidia`.
The scan found 58 local Git worktrees. This document distinguishes the platform's
direct source/runtime dependencies from incidental links in vendored charts,
lockfiles, generated package metadata, and documentation.

## Layer ownership

| Layer | Organization repository | Responsibility |
|---|---|---|
| Control | `Edge-Computing-LLM/edge-cli` | Detect the accelerator, orchestrate installation, and expose status and validation commands. |
| Infrastructure | `Edge-Computing-LLM/k3s-nvidia-edge` | Install or validate k3s and, conditionally, NVIDIA Container Toolkit, GPU Operator, RuntimeClass, device plugin, and DCGM. |
| Application | `Edge-Computing-LLM/llm-observability-stack` | Deploy CPU- or GPU-profiled Ollama, Open WebUI, OpenTelemetry, Prometheus, and Grafana workloads. |
| Evidence | `Edge-Computing-LLM/qwen-gguf-observability` | Validate the live Qwen runtime contract and capture sanitized evidence without owning cluster resources. |
| Dashboard | `Edge-Computing-LLM/Frontend-Edge-LLM-Observability` | Present LLM, Kubernetes, and accelerator telemetry. |

The NVIDIA infrastructure layer is conditional. CPU hosts skip its GPU-specific
components and deploy the application layer with `values.cpu-k3s.yaml`.

## Direct upstream repositories available locally

The following major upstreams used by installation, charts, runtime images, or
platform operations are present under the scanned parent directory:

- Kubernetes/k3s: `k3s-io/k3s`, `kubernetes/kubectl`,
  `kubernetes/client-go`, `rancher/local-path-provisioner`,
  `kubernetes-sigs/metrics-server`, `kubernetes-sigs/node-feature-discovery`,
  `coredns/coredns`, and `traefik/traefik`.
- NVIDIA: `NVIDIA/gpu-operator`, `NVIDIA/nvidia-container-toolkit`,
  `NVIDIA/libnvidia-container`, `NVIDIA/k8s-device-plugin`, `NVIDIA/DCGM`,
  `NVIDIA/dcgm-exporter`, `NVIDIA/go-dcgm`, and `NVIDIA/cuda-samples`.
- LLM applications: `ollama/ollama`, `otwld/ollama-helm`,
  `open-webui/open-webui`, `open-webui/helm-charts`, and
  `open-webui/pipelines`.
- Observability: `prometheus-community/helm-charts`,
  `prometheus-operator/prometheus-operator`,
  `prometheus-operator/kube-prometheus`, `grafana/helm-charts`, and the
  OpenTelemetry collector, collector-contrib, collector-releases, operator,
  Helm charts, Python, Python contrib, and GenAI repositories.
- Supporting tools: `helm/helm`, `cloudflare/cloudflared`,
  `apache/tika-helm`, and `apache/echarts`.

The application chart vendors its exact Helm dependencies under `charts/`, so a
normal local install does not require sibling source clones to build those chart
dependencies.

## Referenced repositories not available locally

These are source repositories for direct dependencies or operational tools. Their
absence does not block the validated deployment because released Go modules,
Python packages, npm packages, Helm charts, or container images are consumed
instead.

| Area | Missing source repository | How it is consumed |
|---|---|---|
| edge-cli | `spf13/cobra`, `spf13/pflag`, `go-yaml/yaml`, `inconshreveable/mousetrap` | Go modules |
| Kubernetes | `kubernetes/kubernetes` | Released k3s/Kubernetes binaries and APIs |
| NVIDIA | `NVIDIA/k8s-dra-driver-gpu` | Reference/documentation; the current platform uses the device plugin and GPU Operator path |
| Metrics | `prometheus/prometheus`, `prometheus/alertmanager`, `prometheus/node_exporter`, `prometheus/blackbox_exporter`, `kubernetes/kube-state-metrics`, `grafana/grafana` | Vendored Helm charts and released images |
| Python service | `fastapi/fastapi`, `encode/uvicorn`, `pydantic/pydantic`, `langchain-ai/langchain`, `encode/httpx`, `prometheus/client_python` | Pinned Python 3.11 packages in `langchain-demo/requirements.txt` |
| Frontend | `ag-grid/ag-grid`, `primefaces/primevue`, `primefaces/primeicons`, `vuejs/core`, `ecomfe/vue-echarts`, `tailwindlabs/tailwindcss`, `vitejs/vite` | Locked npm packages; `apache/echarts` is already local |
| Tika | `apache/tika`, `apache/tika-docker` | Released chart/image; `apache/tika-helm` is local |
| Documentation tooling | `norwoodj/helm-docs`, `helm-unittest/helm-unittest` | Optional development tooling |

Transitive repositories appearing only inside vendored charts or npm/Go package
metadata are intentionally not treated as required local clones. Cloning every
transitive source would not make the build more reproducible; the relevant lock
files, vendored charts, image tags, and package versions are the reproducibility
boundary.

## Local duplicates and naming notes

- `prometheus-community/helm-charts` and
  `open-telemetry/opentelemetry-helm-charts` each have duplicate worktrees. One
  clean, current clone is sufficient.
- The NVIDIA DRA repository referenced as `NVIDIA/k8s-dra-driver-gpu` is distinct
  from the locally available `kubernetes-sigs/dra-driver-nvidia-gpu`.
- `Frontend-Edge-LLM-Observability` is the actual organization repository name;
  use that case in documentation and remotes.

## Audit method

The inventory combined Git remote enumeration for every `.git` worktree with
repository URL extraction, Go module files, Helm `Chart.yaml`/`Chart.lock`, Python
requirements, npm lock metadata, CI workflows, and runtime manifests from all five
organization projects. Credentials and generated build output were excluded.

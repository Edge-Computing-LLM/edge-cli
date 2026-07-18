# Programming language and script boundaries

The organization uses languages by operational responsibility, not by forcing
every repository into one implementation language.

| Language or format | Use it for | Do not use it for |
|---|---|---|
| Go | Durable CLIs, install/uninstall workflows, validation gates, typed configuration, command execution, and cross-repository orchestration | Browser UI or one-off data analysis |
| Python | Optional Jupyter learning material only | Deployed services, host validation, benchmarks, evidence collection, or a second control plane |
| Bash | Short transparent wrappers around Helm, kubectl, container image import, and workshop command sequences | Growing business logic, complex parsing, persistent state, or duplicated Go workflows |
| Grafana JSON | Reproducible dashboard presentation provisioned by Helm | Direct privileged Kubernetes access or host orchestration |
| Helm/YAML | Declarative Kubernetes resources and deployment profiles | Imperative host installation logic |

## Repository mapping

- `edge-cli`: Go only for product behavior; Make and YAML remain build/CI glue.
- `k3s-nvidia-edge`: Go for infrastructure workflows and Helm/YAML for the GPU
  Operator profile.
- `llm-observability-stack`: Helm/YAML for resources and Go for its helper CLI,
  gateway, toolbox, benchmark, Kubernetes inspection, telemetry, and tests.
- `qwen-gguf-observability`: dependency-free Go for structured, read-only
  evidence collection.
- Grafana JSON in `llm-observability-stack`: browser presentation.

New deployable repositories require a clear ownership boundary. A new
repository should not be created merely to move an existing script to another
language.

# Programming language and script boundaries

The organization uses languages by operational responsibility, not by forcing
every repository into one implementation language.

| Language or format | Use it for | Do not use it for |
|---|---|---|
| Go | Durable CLIs, install/uninstall workflows, validation gates, typed configuration, command execution, and cross-repository orchestration | Browser UI or one-off data analysis |
| Python 3.11 | Structured evidence, benchmarks, API services, notebooks, Kubernetes reporting, and test tooling | Host package installation or a second deployment control plane |
| Bash | Short transparent wrappers around Helm, kubectl, container image import, and workshop command sequences | Growing business logic, complex parsing, persistent state, or duplicated Go workflows |
| TypeScript/Vue | Browser dashboards, typed frontend models, chart interaction, and client-side Prometheus parsing | Direct privileged Kubernetes access or host orchestration |
| Helm/YAML | Declarative Kubernetes resources and deployment profiles | Imperative host installation logic |

## Repository mapping

- `edge-cli`: Go only for product behavior; Make and YAML remain build/CI glue.
- `k3s-nvidia-edge`: Go for infrastructure workflows and Helm/YAML for the GPU
  Operator profile.
- `llm-observability-stack`: Helm/YAML for resources, Go for its legacy/helper
  CLI, Python 3.11 for application/benchmark/diagnostic code, and Bash only for
  thin local operator wrappers.
- `qwen-gguf-observability`: dependency-free Python 3.11 for structured,
  read-only evidence collection.
- `Frontend-Edge-LLM-Observability`: Vue 3 and TypeScript for the browser.

New deployable repositories require a clear ownership boundary. A new
repository should not be created merely to move an existing script to another
language.

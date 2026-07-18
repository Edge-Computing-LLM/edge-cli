# Repository instructions

This repository is the Go control plane for the Edge-Computing-LLM organization.
Keep Layer 1 (`k3s-nvidia-edge`) and Layer 2 (`llm-observability-stack`)
ownership explicit. Do not embed application services, Helm templates, model
weights, credentials, or arbitrary model-generated shell execution here.

Before completing a change run `gofmt`, `go test ./...`, `go vet ./...`, and
`go build -o /tmp/edge ./cmd/edge`. Mutating cluster operations must retain
dry-run behavior and require `--yes`. Never log Secrets or kubeconfig content.

# Contributing

This project is the unified Go CLI for the `Edge-Computing-LLM` organization.

## Development Rules

- Keep the CLI 100% Go.
- Do not add Python orchestration.
- Do not add Bash scripts.
- Use Go for workflow ordering, validation, config handling, and safety checks.
- External platform tools may be executed through Go's `os/exec`.
- Keep modules narrow and tied to organization repositories.
- Keep NVIDIA as the only accelerator target.

## Local Checks

```bash
go mod tidy
gofmt -w ./cmd ./internal ./pkg
go test ./...
go build -o edge ./cmd/edge
```

## Module Additions

New repository integrations should be added under `internal/modules/<name>` and
should include:

- repo path validation
- doctor/status checks
- install workflow
- validation workflow
- safe uninstall workflow
- focused tests for path/config/profile logic

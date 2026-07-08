# Security

`edge-cli` can run host-level and cluster-level commands such as `apt-get`,
`systemctl`, `helm`, and `kubectl`.

## Reporting

Open a private security advisory or contact the repository maintainers through
the GitHub organization.

## Operational Safety

- Review commands with `--dry-run` before destructive workflows.
- Uninstall commands require `--yes`.
- `edge uninstall infra` does not remove k3s unless `--k3s` is explicitly set.
- `edge uninstall observability` can keep the namespace with `--keep-namespace`.
- Do not commit kubeconfigs, secrets, tokens, local `.env` files, or generated
  runtime artifacts.

# Troubleshooting

Start with:

```bash
edge repo doctor
edge doctor
edge status
```

## `nvidia-smi` Missing or Failing

The host NVIDIA driver is not installed or not healthy. Repair the host driver
first, then rerun:

```bash
edge doctor
```

## k3s Not Running

Check systemd:

```bash
systemctl status k3s --no-pager
```

Then rerun:

```bash
edge validate infra
```

## NVIDIA RuntimeClass Missing

The GPU Operator or NVIDIA Container Toolkit setup is incomplete.

```bash
edge install infra --yes
edge validate infra
```

## GPU Not Allocatable

Check node capacity and GPU Operator pods:

```bash
kubectl get nodes -o wide
kubectl get runtimeclass
kubectl get pods -n gpu-operator -o wide
```

Then run:

```bash
edge validate infra
```

## Observability Pods Not Ready

Inspect status and logs:

```bash
edge status
edge logs --tail 200
kubectl get pods -n llm-observability -o wide
```

## Helm Release Missing

Install the observability stack:

```bash
edge install observability --yes
edge validate observability
```

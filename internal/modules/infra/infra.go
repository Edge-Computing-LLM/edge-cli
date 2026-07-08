package infra

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Edge-Computing-LLM/edge-cli/internal/execx"
	"github.com/Edge-Computing-LLM/edge-cli/internal/kubernetes"
	"github.com/Edge-Computing-LLM/edge-cli/internal/linux"
	"github.com/Edge-Computing-LLM/edge-cli/internal/nvidia"
)

type Options struct {
	RepoPath           string
	GPUOperatorVersion string
	CUDATestImage      string
	DriverEnabled      bool
	UseLocalChart      bool
	SkipBasePackages   bool
	SkipToolkitInstall bool
	SkipK3sInstall     bool
	SkipGPUOperator    bool
	UninstallK3s       bool
	Yes                bool
	Timeout            string
	Verbose            bool
	DryRun             bool
	RequireHostCUDA    bool
	InstallK3sChannel  string
	InstallK3sExec     string
}

func DefaultOptions(repoPath string) Options {
	return Options{
		RepoPath:           repoPath,
		GPUOperatorVersion: "v26.3.3",
		CUDATestImage:      "nvidia/cuda:12.8.1-base-ubuntu24.04",
		UseLocalChart:      true,
		Timeout:            "5m",
		RequireHostCUDA:    false,
		InstallK3sChannel:  "stable",
		InstallK3sExec:     "server --write-kubeconfig-mode 0644 --disable traefik --disable servicelb --disable metrics-server --node-label gpu=nvidia --node-label workload=edge-ai",
	}
}

func Doctor(ctx context.Context, opts Options) error {
	r := runner(opts)
	results := []execx.Result{}
	if out, err := linux.CheckUbuntu22Plus(); err != nil {
		results = append(results, execx.Result{Label: "Linux OS", Err: err})
	} else {
		results = append(results, execx.Result{Label: "Linux OS", Output: out})
	}
	results = append(results, execx.RequiredCommands("sudo", "curl", "apt-get", "systemctl", "kubectl", "helm", "nvidia-smi")...)
	results = append(results, nvidia.Checks(ctx, r)...)
	results = append(results,
		r.Check(ctx, "k3s service", execx.Command{Name: "systemctl", Args: []string{"is-active", "k3s"}}),
		r.Check(ctx, "kubectl status", execx.Command{Name: "kubectl", Args: []string{"cluster-info"}}),
		r.Check(ctx, "helm status", execx.Command{Name: "helm", Args: []string{"version", "--short"}}),
		r.Check(ctx, "containerd runtime", execx.Command{Name: "systemctl", Args: []string{"is-active", "k3s"}}),
	)
	k := kubernetes.Client{Runner: r}
	results = append(results,
		k.Nodes(ctx),
		k.RuntimeClassNvidia(ctx),
		r.Check(ctx, "GPU Operator pods", execx.Command{Name: "kubectl", Args: []string{"get", "pods", "-n", "gpu-operator", "-o", "wide"}}),
		r.Check(ctx, "GPU allocatable", execx.Command{Name: "kubectl", Args: []string{"get", "nodes", "-o", "custom-columns=NAME:.metadata.name,GPU:.status.allocatable.nvidia\\.com/gpu"}}),
		r.Check(ctx, "CUDA validation support", execx.Command{Name: "kubectl", Args: []string{"get", "runtimeclass", "nvidia"}}),
	)
	return execx.PrintResults(results)
}

func Status(ctx context.Context, opts Options) error {
	r := runner(opts)
	results := []execx.Result{
		r.Check(ctx, "k3s resources", execx.Command{Name: "kubectl", Args: []string{"get", "all", "-A"}}),
		r.Check(ctx, "nodes", execx.Command{Name: "kubectl", Args: []string{"get", "nodes", "-o", "wide"}}),
		r.Check(ctx, "runtime classes", execx.Command{Name: "kubectl", Args: []string{"get", "runtimeclass"}}),
		r.Check(ctx, "helm releases", execx.Command{Name: "helm", Args: []string{"list", "-A"}}),
		r.Check(ctx, "GPU Operator values", execx.Command{Name: "helm", Args: []string{"get", "values", "gpu-operator", "-n", "gpu-operator", "-o", "yaml"}}),
		r.Check(ctx, "GPU allocatable", execx.Command{Name: "kubectl", Args: []string{"get", "nodes", "-o", "custom-columns=NAME:.metadata.name,GPU:.status.allocatable.nvidia\\.com/gpu"}}),
	}
	return execx.PrintResults(results)
}

func Install(ctx context.Context, opts Options) error {
	if err := ValidateRepo(opts.RepoPath); err != nil {
		return err
	}
	if err := hostPreflight(ctx, opts); err != nil {
		return err
	}
	r := runner(opts)
	if !opts.SkipBasePackages {
		if err := r.Run(ctx, execx.Command{Name: "apt-get", Args: []string{"update"}, Sudo: true, Mutates: true}); err != nil {
			return err
		}
		if err := r.Run(ctx, execx.Command{Name: "apt-get", Args: []string{"install", "-y", "ca-certificates", "curl", "gnupg", "lsb-release", "jq", "apt-transport-https", "software-properties-common"}, Sudo: true, Mutates: true}); err != nil {
			return err
		}
	}
	if !opts.SkipToolkitInstall {
		if err := installNvidiaToolkit(ctx, r); err != nil {
			return err
		}
	}
	if !opts.SkipK3sInstall && !execx.Exists("k3s") {
		if err := installK3s(ctx, r, opts); err != nil {
			return err
		}
	}
	if err := prepareKubeconfig(ctx, r); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"wait", "--for=condition=Ready", "node", "--all", "--timeout=180s"}, Mutates: true}); err != nil {
		return err
	}
	if !opts.SkipGPUOperator {
		if err := installGPUOperator(ctx, r, opts); err != nil {
			return err
		}
	}
	return Validate(ctx, opts)
}

func Validate(ctx context.Context, opts Options) error {
	r := runner(opts)
	k := kubernetes.Client{Runner: r}
	results := []execx.Result{
		k.ClusterInfo(ctx),
		k.Nodes(ctx),
		k.RuntimeClassNvidia(ctx),
		r.Check(ctx, "GPU Operator pods", execx.Command{Name: "kubectl", Args: []string{"get", "pods", "-n", "gpu-operator", "-o", "wide"}}),
		r.Check(ctx, "GPU allocatable", execx.Command{Name: "kubectl", Args: []string{"get", "nodes", "-o", "custom-columns=NAME:.metadata.name,GPU:.status.allocatable.nvidia\\.com/gpu"}}),
	}
	if err := execx.PrintResults(results); err != nil {
		return err
	}
	return cudaValidation(ctx, r, opts)
}

func Uninstall(ctx context.Context, opts Options) error {
	if !opts.Yes && !opts.DryRun {
		return errors.New("uninstall requires --yes; user data is not deleted by default")
	}
	r := runner(opts)
	if err := r.Run(ctx, execx.Command{Name: "helm", Args: []string{"uninstall", "gpu-operator", "-n", "gpu-operator", "--wait"}, Mutates: true}); err != nil {
		fmt.Printf("gpu-operator uninstall returned: %v\n", err)
	}
	if err := r.Run(ctx, execx.Command{Name: "helm", Args: []string{"uninstall", "k3s-nvidia-edge", "-n", "gpu-operator", "--wait"}, Mutates: true}); err != nil {
		fmt.Printf("k3s-nvidia-edge uninstall returned: %v\n", err)
	}
	if err := r.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"delete", "namespace", "gpu-operator", "--ignore-not-found"}, Mutates: true}); err != nil {
		return err
	}
	if opts.UninstallK3s {
		if _, err := os.Stat("/usr/local/bin/k3s-uninstall.sh"); err == nil {
			if err := r.Run(ctx, execx.Command{Name: "/usr/local/bin/k3s-uninstall.sh", Sudo: true, Mutates: true}); err != nil {
				return err
			}
		}
	}
	return nil
}

func ValidateRepo(path string) error {
	if path == "" {
		return errors.New("infra repo path is empty")
	}
	for _, rel := range []string{"go.mod", "charts/k3s-nvidia-edge/Chart.yaml"} {
		if _, err := os.Stat(filepath.Join(path, rel)); err != nil {
			return fmt.Errorf("invalid k3s-nvidia-edge repo path %q: missing %s", path, rel)
		}
	}
	return nil
}

func runner(opts Options) execx.Runner {
	return execx.Runner{Verbose: opts.Verbose, DryRun: opts.DryRun}
}

func hostPreflight(ctx context.Context, opts Options) error {
	r := runner(opts)
	results := []execx.Result{}
	if out, err := linux.CheckUbuntu22Plus(); err != nil {
		results = append(results, execx.Result{Label: "Linux OS", Err: err})
	} else {
		results = append(results, execx.Result{Label: "Linux OS", Output: out})
	}
	results = append(results, execx.RequiredCommands("sudo", "curl", "apt-get", "systemctl", "kubectl", "helm", "nvidia-smi")...)
	results = append(results, r.Check(ctx, "NVIDIA driver", execx.Command{Name: "nvidia-smi"}))
	if err := execx.PrintResults(results); err != nil {
		return err
	}
	return ValidateRepo(opts.RepoPath)
}

func installNvidiaToolkit(ctx context.Context, r execx.Runner) error {
	if err := r.Run(ctx, execx.Command{Name: "install", Args: []string{"-d", "-m", "0755", "/usr/share/keyrings", "/etc/apt/sources.list.d"}, Sudo: true, Mutates: true}); err != nil {
		return err
	}
	keyURL := "https://nvidia.github.io/libnvidia-container/gpgkey"
	listURL := "https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list"
	keyOut := "/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg"
	tmpKey := filepath.Join(os.TempDir(), "nvidia-container-toolkit.gpgkey")
	tmpList := filepath.Join(os.TempDir(), "nvidia-container-toolkit.list")
	if err := downloadFile(keyURL, tmpKey); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "gpg", Args: []string{"--dearmor", "-o", keyOut, tmpKey}, Sudo: true, Mutates: true}); err != nil {
		return err
	}
	if err := downloadFile(listURL, tmpList); err != nil {
		return err
	}
	listData, err := os.ReadFile(tmpList)
	if err != nil {
		return err
	}
	replaced := []byte(replaceSignedBy(string(listData)))
	if err := os.WriteFile(tmpList, replaced, 0o644); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "cp", Args: []string{tmpList, "/etc/apt/sources.list.d/nvidia-container-toolkit.list"}, Sudo: true, Mutates: true}); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "apt-get", Args: []string{"update"}, Sudo: true, Mutates: true}); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "apt-get", Args: []string{"install", "-y", "nvidia-container-toolkit", "libnvidia-container-tools", "libnvidia-container1"}, Sudo: true, Mutates: true}); err != nil {
		return err
	}
	return r.Run(ctx, execx.Command{Name: "apt-get", Args: []string{"remove", "-y", "nvidia-container-runtime", "nvidia-docker2", "nvidia-docker"}, Sudo: true, Mutates: true})
}

func installK3s(ctx context.Context, r execx.Runner, opts Options) error {
	installer := filepath.Join(os.TempDir(), "get-k3s-install.sh")
	if err := downloadFile("https://get.k3s.io", installer); err != nil {
		return err
	}
	if err := os.Chmod(installer, 0o700); err != nil {
		return err
	}
	env := []string{"INSTALL_K3S_CHANNEL=" + opts.InstallK3sChannel, "INSTALL_K3S_EXEC=" + opts.InstallK3sExec}
	r.Env = append(r.Env, env...)
	return r.Run(ctx, execx.Command{Name: "sh", Args: []string{installer}, Sudo: true, Mutates: true})
}

func prepareKubeconfig(ctx context.Context, r execx.Runner) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	kubeDir := filepath.Join(home, ".kube")
	if err := os.MkdirAll(kubeDir, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(kubeDir, "config")
	if err := r.Run(ctx, execx.Command{Name: "cp", Args: []string{"/etc/rancher/k3s/k3s.yaml", dst}, Sudo: true, Mutates: true}); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "chown", Args: []string{fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), dst}, Sudo: true, Mutates: true}); err != nil {
		return err
	}
	return os.Chmod(dst, 0o600)
}

func installGPUOperator(ctx context.Context, r execx.Runner, opts Options) error {
	if opts.UseLocalChart {
		chart := filepath.Join(opts.RepoPath, "charts", "k3s-nvidia-edge")
		if err := r.Run(ctx, execx.Command{Name: "helm", Args: []string{"dependency", "update", chart}, Mutates: true}); err != nil {
			return err
		}
		args := []string{"upgrade", "--install", "k3s-nvidia-edge", chart, "-n", "gpu-operator", "--create-namespace",
			"--set", fmt.Sprintf("gpu-operator.driver.enabled=%t", opts.DriverEnabled),
			"--set", "gpu-operator.toolkit.enabled=true",
			"--set", "gpu-operator.gfd.enabled=false",
			"--set", "gpu-operator.toolkit.env[0].name=CONTAINERD_CONFIG",
			"--set", "gpu-operator.toolkit.env[0].value=/var/lib/rancher/k3s/agent/etc/containerd/config.toml",
			"--set", "gpu-operator.toolkit.env[1].name=CONTAINERD_SOCKET",
			"--set", "gpu-operator.toolkit.env[1].value=/run/k3s/containerd/containerd.sock",
			"--set", "gpu-operator.toolkit.env[2].name=RUNTIME_CONFIG_SOURCE",
			"--set-string", "gpu-operator.toolkit.env[2].value=file=/var/lib/rancher/k3s/agent/etc/containerd/config.toml",
			"--wait"}
		return r.Run(ctx, execx.Command{Name: "helm", Args: args, Mutates: true})
	}
	if err := r.Run(ctx, execx.Command{Name: "helm", Args: []string{"repo", "add", "nvidia", "https://helm.ngc.nvidia.com/nvidia"}, Mutates: true}); err != nil {
		fmt.Printf("helm repo add nvidia returned: %v\n", err)
	}
	if err := r.Run(ctx, execx.Command{Name: "helm", Args: []string{"repo", "update"}, Mutates: true}); err != nil {
		return err
	}
	args := []string{"upgrade", "--install", "gpu-operator", "nvidia/gpu-operator", "-n", "gpu-operator", "--create-namespace",
		"--version", opts.GPUOperatorVersion,
		"--set", fmt.Sprintf("driver.enabled=%t", opts.DriverEnabled),
		"--set", "toolkit.enabled=true",
		"--set", "gfd.enabled=false",
		"--set", "toolkit.env[0].name=CONTAINERD_CONFIG",
		"--set", "toolkit.env[0].value=/var/lib/rancher/k3s/agent/etc/containerd/config.toml",
		"--set", "toolkit.env[1].name=CONTAINERD_SOCKET",
		"--set", "toolkit.env[1].value=/run/k3s/containerd/containerd.sock",
		"--set", "toolkit.env[2].name=RUNTIME_CONFIG_SOURCE",
		"--set-string", "toolkit.env[2].value=file=/var/lib/rancher/k3s/agent/etc/containerd/config.toml",
		"--wait"}
	return r.Run(ctx, execx.Command{Name: "helm", Args: args, Mutates: true})
}

func cudaValidation(ctx context.Context, r execx.Runner, opts Options) error {
	manifest := cudaManifest(opts.CUDATestImage)
	tmp := filepath.Join(os.TempDir(), "edge-cli-cuda-validation.yaml")
	if err := os.WriteFile(tmp, []byte(manifest), 0o600); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"delete", "pod", "cuda-test", "--ignore-not-found"}, Mutates: true}); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"apply", "-f", tmp}, Mutates: true}); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"wait", "--for=jsonpath={.status.phase}=Succeeded", "pod/cuda-test", "--timeout=180s"}, Mutates: true}); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"logs", "cuda-test"}}); err != nil {
		return err
	}
	return r.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"delete", "pod", "cuda-test"}, Mutates: true})
}

func cudaManifest(image string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: cuda-test
  labels:
    app.kubernetes.io/name: edge-cli
    app.kubernetes.io/component: cuda-validation
spec:
  restartPolicy: Never
  runtimeClassName: nvidia
  nodeSelector:
    nvidia.com/gpu.present: "true"
  containers:
  - name: cuda-test
    image: %s
    command: ["nvidia-smi"]
    securityContext:
      allowPrivilegeEscalation: false
    resources:
      limits:
        nvidia.com/gpu: 1
`, image)
}

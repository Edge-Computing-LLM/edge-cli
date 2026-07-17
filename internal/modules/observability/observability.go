package observability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Edge-Computing-LLM/edge-cli/internal/execx"
	"github.com/Edge-Computing-LLM/edge-cli/internal/kubernetes"
	"github.com/Edge-Computing-LLM/edge-cli/internal/modules/infra"
)

type Options struct {
	RepoPath       string
	InfraRepoPath  string
	Namespace      string
	Release        string
	Profile        string
	ValuesFiles    []string
	SetValues      []string
	Timeout        string
	Model          string
	OllamaSmoke    bool
	KeepNamespace  bool
	Yes            bool
	DryRun         bool
	Verbose        bool
	SkipInfraCheck bool
}

var forcedBaseLayerDisables = []string{
	"gpu-operator.enabled=false",
	"nvidia-device-plugin.enabled=false",
	"dcgm-exporter.enabled=false",
}

func DefaultOptions(repoPath, infraRepoPath, namespace string) Options {
	return Options{
		RepoPath:      repoPath,
		InfraRepoPath: infraRepoPath,
		Namespace:     namespace,
		Release:       "llm-observability-stack",
		Profile:       "geforce-940m-k3s",
		Timeout:       "5m",
		Model:         "gemma3-1b-it-gguf-local",
		OllamaSmoke:   true,
	}
}

func Doctor(ctx context.Context, opts Options) error {
	if err := ValidateRepo(opts.RepoPath); err != nil {
		return err
	}
	r := runner(opts)
	if !opts.SkipInfraCheck {
		infraOpts := infraDependencyOptions(opts)
		if err := infra.Validate(ctx, infraOpts); err != nil {
			return fmt.Errorf("infra is not ready: %w", err)
		}
	}
	k := kubernetes.Client{Runner: r}
	results := []execx.Result{
		k.Namespace(ctx, opts.Namespace),
		r.Check(ctx, "Helm release", execx.Command{Name: "helm", Args: []string{"status", opts.Release, "-n", opts.Namespace}}),
		k.Workloads(ctx, opts.Namespace),
		k.Service(ctx, opts.Namespace, "ollama"),
		k.Service(ctx, opts.Namespace, "open-webui"),
		k.Service(ctx, opts.Namespace, "opentelemetry-collector"),
		optional(ctx, r, "Prometheus service", execx.Command{Name: "kubectl", Args: []string{"get", "svc", "-n", opts.Namespace, "kube-prometheus-stack-prometheus"}}),
		optional(ctx, r, "Grafana service", execx.Command{Name: "kubectl", Args: []string{"get", "svc", "-n", opts.Namespace, "llm-observability-stack-grafana"}}),
	}
	return execx.PrintResults(results)
}

func Status(ctx context.Context, opts Options) error {
	r := runner(opts)
	results := []execx.Result{
		r.Check(ctx, "Helm release", execx.Command{Name: "helm", Args: []string{"status", opts.Release, "-n", opts.Namespace}}),
		r.Check(ctx, "LLM workloads", execx.Command{Name: "kubectl", Args: []string{"get", "pods,deploy,statefulset,svc,pvc", "-n", opts.Namespace, "-o", "wide"}}),
		r.Check(ctx, "Ollama models", execx.Command{Name: "kubectl", Args: []string{"exec", "-n", opts.Namespace, "deploy/ollama", "--", "ollama", "list"}}),
	}
	if GPUProfile(opts.Profile) {
		results = append([]execx.Result{
			r.Check(ctx, "base GPU layer", execx.Command{Name: "kubectl", Args: []string{"get", "pods", "-n", "gpu-operator", "-o", "wide"}}),
			r.Check(ctx, "NVIDIA RuntimeClass", execx.Command{Name: "kubectl", Args: []string{"get", "runtimeclass", "nvidia"}}),
		}, results...)
	}
	return execx.PrintResults(results)
}

func Install(ctx context.Context, opts Options) error {
	if !opts.Yes && !opts.DryRun {
		return errors.New("install requires --yes; use --dry-run to preview changes")
	}
	if err := ValidateRepo(opts.RepoPath); err != nil {
		return err
	}
	if err := ValidateInstallOptions(opts); err != nil {
		return err
	}
	if !opts.SkipInfraCheck {
		infraOpts := infraDependencyOptions(opts)
		if err := infra.Validate(ctx, infraOpts); err != nil {
			return fmt.Errorf("infra must validate before observability install: %w", err)
		}
	}
	r := runner(opts)
	if err := applyOptionalCRDs(ctx, r, opts.RepoPath); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "helm", Args: []string{"dependency", "build", "."}, Dir: opts.RepoPath, Mutates: true}); err != nil {
		return err
	}
	if err := recoverPendingInstall(ctx, r, opts); err != nil {
		return err
	}
	if err := r.Run(ctx, execx.Command{Name: "helm", Args: helmInstallArgs(opts), Dir: opts.RepoPath, Mutates: true}); err != nil {
		return err
	}
	if opts.DryRun {
		return nil
	}
	k := kubernetes.Client{Runner: r}
	for _, rollout := range []string{"deploy/ollama", "statefulset/open-webui", "deploy/opentelemetry-collector"} {
		if err := k.Rollout(ctx, rollout, opts.Namespace, opts.Timeout); err != nil {
			return err
		}
	}
	return Validate(ctx, opts)
}

func Validate(ctx context.Context, opts Options) error {
	if !opts.SkipInfraCheck {
		infraOpts := infraDependencyOptions(opts)
		if err := infra.Validate(ctx, infraOpts); err != nil {
			return fmt.Errorf("infra validation failed: %w", err)
		}
	}
	r := runner(opts)
	k := kubernetes.Client{Runner: r}
	results := []execx.Result{
		r.Check(ctx, "Helm release deployed", execx.Command{Name: "helm", Args: []string{"status", opts.Release, "-n", opts.Namespace}}),
		k.Namespace(ctx, opts.Namespace),
		k.Service(ctx, opts.Namespace, "ollama"),
		k.Service(ctx, opts.Namespace, "open-webui"),
		k.Service(ctx, opts.Namespace, "opentelemetry-collector"),
		optional(ctx, r, "Prometheus service", execx.Command{Name: "kubectl", Args: []string{"get", "svc", "-n", opts.Namespace, "kube-prometheus-stack-prometheus"}}),
		optional(ctx, r, "Grafana service", execx.Command{Name: "kubectl", Args: []string{"get", "svc", "-n", opts.Namespace, "llm-observability-stack-grafana"}}),
	}
	if GPUProfile(opts.Profile) {
		results = append(results, r.Check(ctx, "Ollama GPU limits", execx.Command{Name: "kubectl", Args: []string{"get", "deploy", "-n", opts.Namespace, "ollama", "-o", "jsonpath={.spec.template.spec.runtimeClassName}{\"\\n\"}{.spec.template.spec.containers[0].resources.limits.nvidia\\.com/gpu}{\"\\n\"}"}}))
	}
	if err := execx.PrintResults(results); err != nil {
		return err
	}
	for _, rollout := range []string{"deploy/ollama", "statefulset/open-webui", "deploy/opentelemetry-collector"} {
		if err := k.Rollout(ctx, rollout, opts.Namespace, opts.Timeout); err != nil {
			return err
		}
	}
	if err := k.WaitAllPods(ctx, opts.Namespace, opts.Timeout); err != nil {
		return err
	}
	if opts.OllamaSmoke {
		return r.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"exec", "-n", opts.Namespace, "deploy/ollama", "--", "ollama", "run", opts.Model, "Reply with exactly: validation ok"}})
	}
	return nil
}

func Uninstall(ctx context.Context, opts Options) error {
	if !opts.Yes && !opts.DryRun {
		return errors.New("uninstall requires --yes; namespace and user data are kept unless deletion is requested by flags")
	}
	r := runner(opts)
	if err := r.Run(ctx, execx.Command{Name: "helm", Args: []string{"uninstall", opts.Release, "-n", opts.Namespace, "--wait", "--ignore-not-found"}, Mutates: true}); err != nil {
		fmt.Printf("observability uninstall returned: %v\n", err)
	}
	if !opts.KeepNamespace {
		return r.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"delete", "namespace", opts.Namespace, "--ignore-not-found"}, Mutates: true})
	}
	return nil
}

func Logs(ctx context.Context, opts Options, tail string) error {
	r := runner(opts)
	selectors := [][]string{
		{"deploy/ollama"},
		{"statefulset/open-webui"},
		{"deploy/opentelemetry-collector"},
		{"deploy/kube-prometheus-stack-operator"},
	}
	for _, target := range selectors {
		args := append([]string{"logs", "-n", opts.Namespace}, target...)
		args = append(args, "--tail", tail)
		if err := r.Run(ctx, execx.Command{Name: "kubectl", Args: args}); err != nil {
			fmt.Printf("logs for %s returned: %v\n", target[0], err)
		}
	}
	return nil
}

func ValidateRepo(path string) error {
	if path == "" {
		return errors.New("observability repo path is empty")
	}
	for _, rel := range []string{"Chart.yaml", "values.yaml", "templates"} {
		if _, err := os.Stat(filepath.Join(path, rel)); err != nil {
			return fmt.Errorf("invalid llm-observability-stack repo path %q: missing %s", path, rel)
		}
	}
	return nil
}

func ProfileValuesFile(profile string) string {
	switch profile {
	case "", "geforce-940m-k3s":
		return "values.geforce-940m-k3s.yaml"
	case "default":
		return "values.yaml"
	case "local-k3s":
		return "values.local-k3s.yaml"
	case "local-k3s-example":
		return "values.local-k3s.example.yaml"
	case "enterprise-pilot-k3s":
		return "values.enterprise-pilot-k3s.yaml"
	case "validation-k3s":
		return "values.validation-k3s.yaml"
	case "cpu-k3s":
		return "values.cpu-k3s.yaml"
	case "full-stack-nvidia":
		return "values.full-stack-nvidia.example.yaml"
	default:
		return profile
	}
}

func GPUProfile(profile string) bool {
	file := ProfileValuesFile(profile)
	name := profile + " " + file
	return !contains(name, "cpu")
}

func ValidateInstallOptions(opts Options) error {
	if opts.SkipInfraCheck && GPUProfile(opts.Profile) && !opts.DryRun {
		return errors.New("cannot skip infra validation for GPU observability profiles; run edge install infra or edge validate infra first")
	}
	return nil
}

func runner(opts Options) execx.Runner {
	return execx.Runner{Verbose: opts.Verbose, DryRun: opts.DryRun}
}

func infraDependencyOptions(opts Options) infra.Options {
	infraOpts := infra.DefaultOptions(opts.InfraRepoPath)
	infraOpts.Verbose = opts.Verbose
	infraOpts.DryRun = opts.DryRun
	infraOpts.SkipCUDAValidation = true
	infraOpts.NVIDIAEnabled = GPUProfile(opts.Profile)
	return infraOpts
}

func optional(ctx context.Context, r execx.Runner, label string, cmd execx.Command) execx.Result {
	res := r.Check(ctx, label+" (optional)", cmd)
	if res.Err != nil {
		res.Output = "not present or disabled"
		res.Err = nil
	}
	return res
}

func helmInstallArgs(opts Options) []string {
	args := []string{"upgrade", "--install", opts.Release, ".", "-n", opts.Namespace, "--create-namespace", "-f", ProfileValuesFile(opts.Profile), "--wait", "--timeout", opts.Timeout}
	for _, vf := range opts.ValuesFiles {
		args = append(args, "-f", vf)
	}
	for _, set := range opts.SetValues {
		args = append(args, "--set", set)
	}
	if GPUProfile(opts.Profile) {
		for _, set := range forcedBaseLayerDisables {
			args = append(args, "--set", set)
		}
	}
	return args
}

type releaseStatus struct {
	Info struct {
		Status string `json:"status"`
	} `json:"info"`
	Version int `json:"version"`
}

func recoverPendingInstall(ctx context.Context, r execx.Runner, opts Options) error {
	out, err := r.Output(ctx, execx.Command{Name: "helm", Args: []string{"status", opts.Release, "-n", opts.Namespace, "-o", "json"}})
	if err != nil {
		if strings.Contains(strings.ToLower(out), "release: not found") || strings.Contains(strings.ToLower(out), "release not found") {
			return nil
		}
		return fmt.Errorf("inspect existing Helm release: %s: %w", out, err)
	}
	var status releaseStatus
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		return fmt.Errorf("decode Helm release status: %w", err)
	}
	if status.Info.Status != "pending-install" {
		return nil
	}
	if status.Version < 1 {
		return errors.New("Helm reports pending-install without a valid revision")
	}
	fmt.Printf("Recovering interrupted Helm install at revision %d\n", status.Version)
	return r.Run(ctx, execx.Command{Name: "helm", Args: []string{
		"rollback", opts.Release, strconv.Itoa(status.Version), "-n", opts.Namespace,
		"--wait", "--timeout", opts.Timeout,
	}, Mutates: true})
}

func contains(value, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return needle == ""
}

func applyOptionalCRDs(ctx context.Context, r execx.Runner, repoPath string) error {
	crdDir := filepath.Join(repoPath, "charts", "kube-prometheus-stack", "charts", "crds", "crds")
	matches, err := filepath.Glob(filepath.Join(crdDir, "*.yaml"))
	if err != nil {
		return err
	}
	for _, file := range matches {
		if err := r.Run(ctx, execx.Command{Name: "kubectl", Args: []string{"apply", "--server-side", "-f", file}, Mutates: true}); err != nil {
			return err
		}
	}
	return nil
}

package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Edge-Computing-LLM/edge-cli/internal/app"
	"github.com/Edge-Computing-LLM/edge-cli/internal/config"
	"github.com/Edge-Computing-LLM/edge-cli/internal/execx"
	"github.com/Edge-Computing-LLM/edge-cli/internal/modules/infra"
	"github.com/Edge-Computing-LLM/edge-cli/internal/modules/observability"
	"github.com/spf13/cobra"
)

type globals struct {
	ConfigPath string
	Verbose    bool
	DryRun     bool
	Timeout    string
}

type repeated []string

func (r *repeated) String() string { return fmt.Sprint([]string(*r)) }
func (r *repeated) Type() string   { return "stringArray" }
func (r *repeated) Set(v string) error {
	*r = append(*r, v)
	return nil
}

func NewRootCommand() *cobra.Command {
	g := &globals{Timeout: "10m"}
	root := &cobra.Command{
		Use:   "edge",
		Short: "Unified NVIDIA edge LLMOps CLI for Edge-Computing-LLM",
		Long:  "edge is the single Go CLI entry point for k3s-nvidia-edge infrastructure and llm-observability-stack workloads.",
	}
	root.PersistentFlags().StringVar(&g.ConfigPath, "config", "", "config file path")
	root.PersistentFlags().BoolVar(&g.Verbose, "verbose", false, "print command details")
	root.PersistentFlags().BoolVar(&g.DryRun, "dry-run", false, "print mutating operations without executing them")
	root.PersistentFlags().StringVar(&g.Timeout, "timeout", g.Timeout, "overall command timeout")

	root.AddCommand(versionCmd())
	root.AddCommand(doctorCmd(g))
	root.AddCommand(statusCmd(g))
	root.AddCommand(installCmd(g))
	root.AddCommand(uninstallCmd(g))
	root.AddCommand(validateCmd(g))
	root.AddCommand(logsCmd(g))
	root.AddCommand(configCmd(g))
	root.AddCommand(repoCmd(g))
	return root
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print edge-cli version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s %s\n", app.Name, app.Version)
		},
	}
}

func doctorCmd(g *globals) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run full NVIDIA edge platform diagnostics",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			if err := infra.Doctor(ctx, infraOpts(cfg, g, "")); err != nil {
				return err
			}
			return observability.Doctor(ctx, obsOpts(cfg, g, "", nil, nil))
		},
	}
}

func statusCmd(g *globals) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Print live infra and observability status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			if err := infra.Status(ctx, infraOpts(cfg, g, "")); err != nil {
				return err
			}
			return observability.Status(ctx, obsOpts(cfg, g, "", nil, nil))
		},
	}
}

func installCmd(g *globals) *cobra.Command {
	cmd := &cobra.Command{Use: "install", Short: "Install platform modules"}
	cmd.AddCommand(installInfraCmd(g), installObsCmd(g), installAllCmd(g))
	return cmd
}

func installInfraCmd(g *globals) *cobra.Command {
	var repoPath string
	var yes bool
	var useUpstream bool
	var skipBase, skipToolkit, skipK3s, skipGPUOperator bool
	c := &cobra.Command{
		Use:   "infra",
		Short: "Install Linux + k3s + NVIDIA GPU infrastructure",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			opts := infraOpts(cfg, g, repoPath)
			opts.Yes = yes
			opts.UseLocalChart = !useUpstream
			opts.SkipBasePackages = skipBase
			opts.SkipToolkitInstall = skipToolkit
			opts.SkipK3sInstall = skipK3s
			opts.SkipGPUOperator = skipGPUOperator
			return infra.Install(ctx, opts)
		},
	}
	c.Flags().StringVar(&repoPath, "repo-path", "", "path to k3s-nvidia-edge repo")
	c.Flags().BoolVar(&yes, "yes", false, "execute mutating operations")
	c.Flags().BoolVar(&useUpstream, "use-upstream-gpu-operator", false, "install upstream NVIDIA GPU Operator chart directly")
	c.Flags().BoolVar(&skipBase, "skip-base-package-install", false, "skip apt base package setup")
	c.Flags().BoolVar(&skipToolkit, "skip-toolkit-install", false, "skip NVIDIA Container Toolkit setup")
	c.Flags().BoolVar(&skipK3s, "skip-k3s-install", false, "skip k3s install")
	c.Flags().BoolVar(&skipGPUOperator, "skip-gpu-operator-install", false, "skip GPU Operator install")
	return c
}

func installObsCmd(g *globals) *cobra.Command {
	var repoPath string
	var yes bool
	var profile string
	var values repeated
	var sets repeated
	var skipInfra bool
	c := &cobra.Command{
		Use:   "observability",
		Short: "Install llm-observability-stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			opts := obsOpts(cfg, g, repoPath, values, sets)
			opts.Yes = yes
			opts.Profile = first(profile, opts.Profile)
			opts.SkipInfraCheck = skipInfra
			return observability.Install(ctx, opts)
		},
	}
	c.Flags().StringVar(&repoPath, "repo-path", "", "path to llm-observability-stack repo")
	c.Flags().BoolVar(&yes, "yes", false, "execute mutating operations")
	c.Flags().StringVar(&profile, "profile", "", "values profile or values file")
	c.Flags().Var(&values, "values", "additional Helm values file; may be repeated")
	c.Flags().Var(&sets, "set", "additional Helm --set override; may be repeated")
	c.Flags().BoolVar(&skipInfra, "skip-infra-check", false, "skip infra validation before install")
	return c
}

func installAllCmd(g *globals) *cobra.Command {
	var yes bool
	var infraRepoPath, obsRepoPath string
	c := &cobra.Command{
		Use:   "all",
		Short: "Install infra, validate it, install observability, validate everything",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			iopts := infraOpts(cfg, g, infraRepoPath)
			iopts.Yes = yes
			if err := infra.Doctor(ctx, iopts); err != nil {
				return err
			}
			if err := infra.Install(ctx, iopts); err != nil {
				return err
			}
			if err := infra.Validate(ctx, iopts); err != nil {
				return err
			}
			oopts := obsOpts(cfg, g, obsRepoPath, nil, nil)
			oopts.Yes = yes
			if err := observability.Install(ctx, oopts); err != nil {
				return err
			}
			if err := observability.Validate(ctx, oopts); err != nil {
				return err
			}
			fmt.Println("Next commands:")
			fmt.Println("  edge status")
			fmt.Println("  kubectl -n llm-observability port-forward svc/open-webui 8080:8080")
			fmt.Println("  kubectl -n llm-observability port-forward svc/llm-observability-stack-grafana 3000:80")
			return nil
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "execute mutating operations")
	c.Flags().StringVar(&infraRepoPath, "infra-repo-path", "", "path to k3s-nvidia-edge repo")
	c.Flags().StringVar(&obsRepoPath, "observability-repo-path", "", "path to llm-observability-stack repo")
	return c
}

func uninstallCmd(g *globals) *cobra.Command {
	cmd := &cobra.Command{Use: "uninstall", Short: "Safely uninstall platform modules"}
	cmd.AddCommand(uninstallInfraCmd(g), uninstallObsCmd(g), uninstallAllCmd(g))
	return cmd
}

func uninstallInfraCmd(g *globals) *cobra.Command {
	var yes, k3s bool
	c := &cobra.Command{
		Use:   "infra",
		Short: "Uninstall NVIDIA infra components; keeps k3s unless --k3s is set",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			opts := infraOpts(cfg, g, "")
			opts.Yes = yes
			opts.UninstallK3s = k3s
			return infra.Uninstall(ctx, opts)
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "confirm uninstall")
	c.Flags().BoolVar(&k3s, "k3s", false, "also uninstall k3s")
	return c
}

func uninstallObsCmd(g *globals) *cobra.Command {
	var yes, keepNamespace bool
	c := &cobra.Command{
		Use:   "observability",
		Short: "Uninstall llm-observability-stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			opts := obsOpts(cfg, g, "", nil, nil)
			opts.Yes = yes
			opts.KeepNamespace = keepNamespace
			return observability.Uninstall(ctx, opts)
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "confirm uninstall")
	c.Flags().BoolVar(&keepNamespace, "keep-namespace", false, "keep namespace and data")
	return c
}

func uninstallAllCmd(g *globals) *cobra.Command {
	var yes bool
	c := &cobra.Command{
		Use:   "all",
		Short: "Uninstall observability then infra GPU layer",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			oopts := obsOpts(cfg, g, "", nil, nil)
			oopts.Yes = yes
			if err := observability.Uninstall(ctx, oopts); err != nil {
				return err
			}
			iopts := infraOpts(cfg, g, "")
			iopts.Yes = yes
			return infra.Uninstall(ctx, iopts)
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "confirm uninstall")
	return c
}

func validateCmd(g *globals) *cobra.Command {
	cmd := &cobra.Command{Use: "validate", Short: "Validate platform modules"}
	cmd.AddCommand(validateInfraCmd(g), validateObsCmd(g))
	return cmd
}

func validateInfraCmd(g *globals) *cobra.Command {
	return &cobra.Command{
		Use:   "infra",
		Short: "Validate k3s + NVIDIA GPU readiness",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			return infra.Validate(ctx, infraOpts(cfg, g, ""))
		},
	}
}

func validateObsCmd(g *globals) *cobra.Command {
	return &cobra.Command{
		Use:   "observability",
		Short: "Validate LLMOps/observability stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			return observability.Validate(ctx, obsOpts(cfg, g, "", nil, nil))
		},
	}
}

func logsCmd(g *globals) *cobra.Command {
	var tail string
	c := &cobra.Command{
		Use:   "logs",
		Short: "Print recent observability workload logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			return observability.Logs(ctx, obsOpts(cfg, g, "", nil, nil), tail)
		},
	}
	c.Flags().StringVar(&tail, "tail", "100", "lines per workload")
	return c
}

func configCmd(g *globals) *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Manage edge-cli config"}
	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Write default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := g.ConfigPath
			if path == "" {
				path = config.DefaultPath()
			}
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("config already exists: %s", path)
			}
			return config.Save(path, config.Default())
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print active config path",
		Run: func(cmd *cobra.Command, args []string) {
			path := g.ConfigPath
			if path == "" {
				path = config.DefaultPath()
			}
			fmt.Println(path)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print resolved config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			fmt.Printf("repos:\n  k3sNvidiaEdge: %s\n  llmObservabilityStack: %s\ncluster:\n  kubeconfig: %s\n  defaultNamespace: %s\ngpu:\n  vendor: %s\n",
				cfg.Repos.K3sNvidiaEdge, cfg.Repos.LLMObservabilityStack, cfg.Cluster.Kubeconfig, cfg.Cluster.DefaultNamespace, cfg.GPU.Vendor)
			return nil
		},
	})
	return cmd
}

func repoCmd(g *globals) *cobra.Command {
	cmd := &cobra.Command{Use: "repo", Short: "Inspect organization repos"}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List configured repos",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			fmt.Printf("%s\t%s\n", app.InfraRepoName, cfg.Repos.K3sNvidiaEdge)
			fmt.Printf("%s\t%s\n", app.ObservabilityRepoName, cfg.Repos.LLMObservabilityStack)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "doctor",
		Short: "Validate configured repo paths and git state",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := load(g)
			if err != nil {
				return err
			}
			r := execx.Runner{Verbose: g.Verbose}
			ctx, cancel, err := contextWithTimeout(g.Timeout)
			if err != nil {
				return err
			}
			defer cancel()
			results := []execx.Result{}
			for name, path := range map[string]string{app.InfraRepoName: cfg.Repos.K3sNvidiaEdge, app.ObservabilityRepoName: cfg.Repos.LLMObservabilityStack} {
				if _, err := os.Stat(path); err != nil {
					results = append(results, execx.Result{Label: name + " path", Err: err})
					continue
				}
				results = append(results, execx.Result{Label: name + " path", Output: path})
				results = append(results, r.Check(ctx, name+" remote", execx.Command{Name: "git", Args: []string{"remote", "-v"}, Dir: path}))
				results = append(results, r.Check(ctx, name+" branch", execx.Command{Name: "git", Args: []string{"status", "--short", "--branch"}, Dir: path}))
			}
			return execx.PrintResults(results)
		},
	})
	return cmd
}

func load(g *globals) (config.Config, error) {
	cfg, err := config.Load(g.ConfigPath)
	if err != nil {
		return config.Config{}, err
	}
	if cfg.Cluster.Kubeconfig != "" {
		os.Setenv("KUBECONFIG", cfg.Cluster.Kubeconfig)
	}
	return cfg, nil
}

func infraOpts(cfg config.Config, g *globals, repoPath string) infra.Options {
	opts := infra.DefaultOptions(first(repoPath, cfg.Repos.K3sNvidiaEdge))
	opts.Verbose = g.Verbose
	opts.DryRun = g.DryRun
	opts.Timeout = g.Timeout
	return opts
}

func obsOpts(cfg config.Config, g *globals, repoPath string, values, sets []string) observability.Options {
	opts := observability.DefaultOptions(first(repoPath, cfg.Repos.LLMObservabilityStack), cfg.Repos.K3sNvidiaEdge, cfg.Cluster.DefaultNamespace)
	opts.Verbose = g.Verbose
	opts.DryRun = g.DryRun
	opts.Timeout = "5m"
	opts.ValuesFiles = values
	opts.SetValues = sets
	return opts
}

func contextWithTimeout(value string) (context.Context, context.CancelFunc, error) {
	d, err := time.ParseDuration(value)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid --timeout: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), d)
	return ctx, cancel, nil
}

func first(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Edge-Computing-LLM/edge-cli/internal/app"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Repos   ReposConfig   `yaml:"repos"`
	Cluster ClusterConfig `yaml:"cluster"`
	GPU     GPUConfig     `yaml:"gpu"`
}

type ReposConfig struct {
	K3sNvidiaEdge         string `yaml:"k3sNvidiaEdge"`
	LLMObservabilityStack string `yaml:"llmObservabilityStack"`
}

type ClusterConfig struct {
	Kubeconfig       string `yaml:"kubeconfig"`
	DefaultNamespace string `yaml:"defaultNamespace"`
}

type GPUConfig struct {
	Vendor string `yaml:"vendor"`
}

func Default() Config {
	return Config{
		Repos: ReposConfig{
			K3sNvidiaEdge:         app.DefaultInfraRepoPath,
			LLMObservabilityStack: app.DefaultObsRepoPath,
		},
		Cluster: ClusterConfig{
			DefaultNamespace: "llm-observability",
		},
		GPU: GPUConfig{Vendor: "nvidia"},
	}
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".edge-cli/config.yaml"
	}
	return filepath.Join(home, ".edge-cli", "config.yaml")
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		path = DefaultPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.GPU.Vendor != "" && cfg.GPU.Vendor != "nvidia" {
		return Config{}, fmt.Errorf("unsupported gpu vendor %q: edge-cli targets NVIDIA GPUs only", cfg.GPU.Vendor)
	}
	fillDefaults(&cfg)
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if path == "" {
		path = DefaultPath()
	}
	fillDefaults(&cfg)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func fillDefaults(cfg *Config) {
	def := Default()
	if cfg.Repos.K3sNvidiaEdge == "" {
		cfg.Repos.K3sNvidiaEdge = def.Repos.K3sNvidiaEdge
	}
	if cfg.Repos.LLMObservabilityStack == "" {
		cfg.Repos.LLMObservabilityStack = def.Repos.LLMObservabilityStack
	}
	if cfg.Cluster.DefaultNamespace == "" {
		cfg.Cluster.DefaultNamespace = def.Cluster.DefaultNamespace
	}
	if cfg.GPU.Vendor == "" {
		cfg.GPU.Vendor = def.GPU.Vendor
	}
}

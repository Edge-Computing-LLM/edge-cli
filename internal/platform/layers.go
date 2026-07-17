package platform

type Operation string

const (
	Doctor    Operation = "doctor"
	Install   Operation = "install"
	Validate  Operation = "validate"
	Status    Operation = "status"
	Logs      Operation = "logs"
	Uninstall Operation = "uninstall"
)

type Layer struct {
	Name         string
	RepoName     string
	DependsOn    []string
	Operations   []Operation
	UninstallRev bool
}

func Catalog() []Layer {
	return []Layer{
		{
			Name:       "infra",
			RepoName:   "k3s-nvidia-edge",
			Operations: []Operation{Doctor, Install, Validate, Status, Uninstall},
		},
		{
			Name:         "observability",
			RepoName:     "llm-observability-stack",
			DependsOn:    []string{"infra"},
			Operations:   []Operation{Doctor, Install, Validate, Status, Logs, Uninstall},
			UninstallRev: true,
		},
	}
}

func InstallOrder() []string {
	layers := Catalog()
	order := make([]string, 0, len(layers))
	for _, layer := range layers {
		order = append(order, layer.Name)
	}
	return order
}

func UninstallOrder() []string {
	install := InstallOrder()
	for left, right := 0, len(install)-1; left < right; left, right = left+1, right-1 {
		install[left], install[right] = install[right], install[left]
	}
	return install
}

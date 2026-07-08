package observability

import "testing"

func TestProfileValuesFile(t *testing.T) {
	cases := map[string]string{
		"":                  "values.geforce-940m-k3s.yaml",
		"geforce-940m-k3s":  "values.geforce-940m-k3s.yaml",
		"local-k3s":         "values.local-k3s.yaml",
		"custom-values.yml": "custom-values.yml",
	}
	for input, want := range cases {
		if got := ProfileValuesFile(input); got != want {
			t.Fatalf("ProfileValuesFile(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestGPUInstallCannotSkipInfraCheck(t *testing.T) {
	opts := DefaultOptions("/tmp/obs", "/tmp/infra", "llm-observability")
	opts.SkipInfraCheck = true
	if err := ValidateInstallOptions(opts); err == nil {
		t.Fatal("ValidateInstallOptions accepted --skip-infra-check for GPU profile")
	}
}

func TestCPUInstallCanSkipInfraCheck(t *testing.T) {
	opts := DefaultOptions("/tmp/obs", "/tmp/infra", "llm-observability")
	opts.Profile = "cpu-k3s"
	opts.SkipInfraCheck = true
	if err := ValidateInstallOptions(opts); err != nil {
		t.Fatalf("ValidateInstallOptions rejected CPU skip infra check: %v", err)
	}
}

func TestGPUHelmInstallForcesBaseLayerChartsDisabled(t *testing.T) {
	opts := DefaultOptions("/tmp/obs", "/tmp/infra", "llm-observability")
	opts.SetValues = []string{"nvidia-device-plugin.enabled=true"}
	args := helmInstallArgs(opts)
	for _, want := range forcedBaseLayerDisables {
		if lastSetValue(args, want[:len(want)-len("=false")]) != want {
			t.Fatalf("last override for %q was not forced false in %#v", want, args)
		}
	}
}

func TestRepoValidationRequiresChartAndTemplates(t *testing.T) {
	dir := t.TempDir()
	if err := ValidateRepo(dir); err == nil {
		t.Fatal("ValidateRepo accepted incomplete repo")
	}
}

func TestObservabilityDependencyCheckDoesNotConsumeGPU(t *testing.T) {
	opts := DefaultOptions("/tmp/obs", "/tmp/infra", "llm-observability")
	infraOpts := infraDependencyOptions(opts)
	if !infraOpts.SkipCUDAValidation {
		t.Fatal("observability dependency check should skip CUDA pod validation")
	}
}

func lastSetValue(args []string, key string) string {
	last := ""
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--set" && len(args[i+1]) > len(key) && args[i+1][:len(key)+1] == key+"=" {
			last = args[i+1]
		}
	}
	return last
}

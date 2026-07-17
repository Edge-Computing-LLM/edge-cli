package platform

import "testing"

func TestLayerDependencyOrder(t *testing.T) {
	got := InstallOrder()
	want := []string{"infra", "observability"}
	if len(got) != len(want) {
		t.Fatalf("InstallOrder length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("InstallOrder[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	uninstall := UninstallOrder()
	if uninstall[0] != "observability" || uninstall[1] != "infra" {
		t.Fatalf("UninstallOrder = %#v, want reverse layer order", uninstall)
	}
}

func TestCatalogKeepsFutureLayersAdditive(t *testing.T) {
	for _, layer := range Catalog() {
		if layer.Name == "" || layer.RepoName == "" {
			t.Fatalf("layer must declare name and repo: %#v", layer)
		}
		if len(layer.Operations) == 0 {
			t.Fatalf("layer %q must declare supported operations", layer.Name)
		}
	}
}

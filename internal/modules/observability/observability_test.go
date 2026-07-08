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

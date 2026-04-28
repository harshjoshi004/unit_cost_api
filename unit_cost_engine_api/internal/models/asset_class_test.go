package models

import "testing"

func TestAssetClass(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		want         string
	}{
		{name: "gpu", resourceType: "nvidia-tesla-t4", want: AssetClassGPU},
		{name: "pv", resourceType: "pd-ssd-disk", want: AssetClassPV},
		{name: "node", resourceType: "n2-standard-4", want: AssetClassNode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AssetClass(tt.resourceType); got != tt.want {
				t.Fatalf("AssetClass(%q) = %q, want %q", tt.resourceType, got, tt.want)
			}
		})
	}
}

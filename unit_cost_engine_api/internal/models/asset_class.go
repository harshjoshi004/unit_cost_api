package models

import "strings"

const (
	AssetClassNode = "node"
	AssetClassGPU  = "gpu"
	AssetClassPV   = "pv"
)

var gpuWords = []string{
	"gpu",
	"nvidia",
	"tesla",
	"a100",
	"h100",
	"l4",
	"t4",
	"v100",
}

var pvWords = []string{
	"disk",
	"rwo",
	"hyperdisk",
	"storage",
	"volume",
	"pvc",
	"persistent",
	"pd-",
	"ssd",
	"hdd",
	"filestore",
}

func AssetClass(resourceType string) string {
	value := strings.ToLower(resourceType)

	for _, word := range gpuWords {
		if strings.Contains(value, word) {
			return AssetClassGPU
		}
	}

	for _, word := range pvWords {
		if strings.Contains(value, word) {
			return AssetClassPV
		}
	}

	return AssetClassNode
}

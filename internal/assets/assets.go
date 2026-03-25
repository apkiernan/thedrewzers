package assets

import (
	_ "embed"
	"encoding/json"
	"sync"
)

//go:embed js-manifest.json
var embeddedManifest []byte

// AssetManifest maps original asset paths to fingerprinted paths
type AssetManifest map[string]string

var (
	manifest AssetManifest
	once     sync.Once
)

// Asset returns the fingerprinted asset path for the given original path.
// If the asset is not found in the manifest, returns the original path.
func Asset(originalPath string) string {
	once.Do(func() {
		_ = json.Unmarshal(embeddedManifest, &manifest)
	})

	if manifest == nil {
		return originalPath
	}

	if fingerprintedPath, ok := manifest[originalPath]; ok {
		return fingerprintedPath
	}

	return originalPath
}

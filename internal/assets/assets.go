package assets

import (
	"encoding/json"
	"os"
	"sync"
)

// AssetManifest maps original asset paths to fingerprinted paths
type AssetManifest map[string]string

var (
	manifest AssetManifest
	once     sync.Once
)

// LoadManifest loads the asset manifest file (called once on startup)
func LoadManifest(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &manifest)
}

// Asset returns the fingerprinted asset path for the given original path.
// If the asset is not found in the manifest, returns the original path.
// This allows graceful degradation if manifest is not loaded.
func Asset(originalPath string) string {
	once.Do(func() {
		// Try to load manifest if not already loaded
		if manifest == nil {
			_ = LoadManifest("dist/js-manifest.json")
		}
	})

	if manifest == nil {
		return originalPath
	}

	if fingerprintedPath, ok := manifest[originalPath]; ok {
		return fingerprintedPath
	}

	return originalPath
}

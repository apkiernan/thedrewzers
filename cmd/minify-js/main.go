package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AssetManifest maps original filenames to fingerprinted filenames
type AssetManifest map[string]string

func main() {
	sourceDir := "static/js"
	outputDir := "dist/js"
	manifestPath := "dist/js-manifest.json"

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	manifest := make(AssetManifest)

	// Find all JS files in source directory
	files, err := filepath.Glob(filepath.Join(sourceDir, "*.js"))
	if err != nil {
		fmt.Printf("Error finding JS files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d JavaScript files to process\n", len(files))

	for _, inputPath := range files {
		filename := filepath.Base(inputPath)
		fmt.Printf("Processing %s...\n", filename)

		// Read original file for hashing
		content, err := os.ReadFile(inputPath)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", filename, err)
			continue
		}

		// Generate content hash (first 8 chars of MD5)
		hash := md5.Sum(content)
		hashStr := hex.EncodeToString(hash[:])[:8]

		// Generate fingerprinted filename
		ext := filepath.Ext(filename)
		baseName := strings.TrimSuffix(filename, ext)
		fingerprintedName := fmt.Sprintf("%s.%s.min.js", baseName, hashStr)
		outputPath := filepath.Join(outputDir, fingerprintedName)

		// Minify using terser
		cmd := exec.Command("npx", "terser", inputPath,
			"--compress",
			"--mangle",
			"--output", outputPath)

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error minifying %s: %v\n%s\n", filename, err, string(output))
			continue
		}

		// Get minified file size
		stat, _ := os.Stat(outputPath)
		originalSize := len(content)
		minifiedSize := stat.Size()
		savings := 100 - (float64(minifiedSize) / float64(originalSize) * 100)

		fmt.Printf("  ✓ %s → %s (%.1f%% smaller)\n", filename, fingerprintedName, savings)

		// Add to manifest (using /static/js/ prefix for template references)
		manifest["/static/js/"+filename] = "/static/js/" + fingerprintedName
	}

	// Write manifest file
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fmt.Printf("Error creating manifest: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		fmt.Printf("Error writing manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Minified %d files\n", len(manifest))
	fmt.Printf("✓ Manifest written to %s\n", manifestPath)
}

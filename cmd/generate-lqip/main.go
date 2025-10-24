package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

func main() {
	sourceDir := "static/images"
	distDir := "dist/images"

	// Ensure output directories exist
	if err := os.MkdirAll(distDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(distDir, "slideshow"), 0755); err != nil {
		log.Fatalf("Failed to create slideshow directory: %v", err)
	}

	// Process images from multiple directories
	imageDirs := []struct {
		source string
		output string
	}{
		{sourceDir, distDir},
		{filepath.Join(sourceDir, "slideshow"), filepath.Join(distDir, "slideshow")},
	}

	totalProcessed := 0

	for _, dir := range imageDirs {
		// Process all JPG images in this directory
		images, err := filepath.Glob(filepath.Join(dir.source, "*.jpg"))
		if err != nil {
			log.Fatal(err)
		}

		if len(images) == 0 {
			log.Printf("No JPG images found in %s\n", dir.source)
			continue
		}

		log.Printf("Generating LQIP placeholders for %d images in %s\n", len(images), dir.source)

		for _, imagePath := range images {
			baseFilename := filepath.Base(imagePath)

			// Skip already processed images
			if strings.Contains(baseFilename, "-lqip.jpg") ||
				strings.Contains(baseFilename, "w.jpg") {
				continue
			}

			log.Printf("Processing: %s\n", imagePath)
			if err := generateLQIP(imagePath, dir.output); err != nil {
				log.Printf("Error generating LQIP for %s: %v\n", imagePath, err)
			} else {
				totalProcessed++
			}
		}
	}

	log.Printf("LQIP generation complete! Processed %d images\n", totalProcessed)
}

func generateLQIP(inputPath, distDir string) error {
	// Open and decode image
	src, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	// Resize to 20px wide (maintains aspect ratio)
	// Using Box filter for speed since we're making it tiny
	tiny := imaging.Resize(src, 20, 0, imaging.Box)

	// Apply blur for smooth placeholder effect
	blurred := imaging.Blur(tiny, 2.0)

	// Generate output filename
	baseFilename := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputPath := filepath.Join(distDir, baseFilename+"-lqip.jpg")

	// Save with very low quality (we want 2-5KB files)
	if err := imaging.Save(blurred, outputPath, imaging.JPEGQuality(20)); err != nil {
		return fmt.Errorf("failed to save LQIP: %w", err)
	}

	// Get file size for logging
	info, _ := os.Stat(outputPath)
	sizeKB := info.Size() / 1024

	log.Printf("  Generated LQIP: %s (%dKB)\n", filepath.Base(outputPath), sizeKB)

	return nil
}

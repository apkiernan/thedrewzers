package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

var widths = []int{640, 768, 1024, 1280, 1920, 2560}

func main() {
	sourceDir := "static/images"
	distDir := "dist/images"

	// Create output directories
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

		log.Printf("Found %d images to process in %s\n", len(images), dir.source)

		for _, imagePath := range images {
			baseFilename := filepath.Base(imagePath)

			// Skip already processed images (those with width suffixes or lqip)
			if strings.Contains(baseFilename, "-lqip.jpg") ||
				strings.Contains(baseFilename, "w.jpg") {
				log.Printf("Skipping already processed image: %s\n", baseFilename)
				continue
			}

			log.Printf("Processing: %s\n", imagePath)
			if err := generateResponsiveSizes(imagePath, dir.output); err != nil {
				log.Printf("Error processing %s: %v\n", imagePath, err)
			} else {
				totalProcessed++
			}
		}
	}

	log.Printf("Image optimization complete! Processed %d images\n", totalProcessed)
}

func generateResponsiveSizes(inputPath, distDir string) error {
	// Open and decode image
	src, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	baseFilename := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	originalWidth := src.Bounds().Dx()
	originalHeight := src.Bounds().Dy()

	log.Printf("  Original size: %dx%d\n", originalWidth, originalHeight)

	for _, width := range widths {
		// Skip if original is smaller than target width
		if originalWidth < width {
			log.Printf("  Skipping %dw (original is smaller)\n", width)
			continue
		}

		// Resize using Lanczos filter for best quality
		resized := imaging.Resize(src, width, 0, imaging.Lanczos)

		outputPath := filepath.Join(distDir, fmt.Sprintf("%s-%dw.jpg", baseFilename, width))

		// Save with quality 85 (good balance between quality and file size)
		if err := imaging.Save(resized, outputPath, imaging.JPEGQuality(85)); err != nil {
			return fmt.Errorf("failed to save resized image: %w", err)
		}

		log.Printf("  Generated: %s (resized to %dw)\n", filepath.Base(outputPath), width)
	}

	return nil
}

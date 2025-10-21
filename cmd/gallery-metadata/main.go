package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ImageMetadata represents metadata for a gallery image
type ImageMetadata struct {
	Filename    string  `json:"filename"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	AspectRatio float64 `json:"aspectRatio"`
	GridRowSpan int     `json:"gridRowSpan"`
}

func main() {
	staticDir := "static"
	var images []ImageMetadata

	// Scan directory for all molly_andrewENG*.jpg files
	pattern := filepath.Join(staticDir, "molly_andrewENG*.jpg")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("Error scanning for images: %v\n", err)
		os.Exit(1)
	}

	if len(matches) == 0 {
		fmt.Printf("Warning: No images found matching pattern: %s\n", pattern)
		os.Exit(1)
	}

	fmt.Printf("Found %d images to process...\n", len(matches))

	// Sort matches numerically by the number in the filename
	sort.Slice(matches, func(i, j int) bool {
		numI := extractNumber(matches[i])
		numJ := extractNumber(matches[j])
		return numI < numJ
	})

	// Process each image found
	for i, path := range matches {
		filename := filepath.Base(path)

		if (i+1)%10 == 0 {
			fmt.Printf("Processing image %d/%d...\n", i+1, len(matches))
		}

		if metadata, err := processImage(path); err == nil {
			images = append(images, metadata)
		} else {
			fmt.Printf("Warning: Failed to process %s: %v\n", filename, err)
		}
	}

	// Write metadata JSON
	outputPath := filepath.Join(staticDir, "gallery-metadata.json")
	data, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		fmt.Printf("Error writing metadata file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Successfully generated metadata for %d images\n", len(images))
	fmt.Printf("✓ Output written to: %s\n", outputPath)
}

func processImage(path string) (ImageMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return ImageMetadata{}, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return ImageMetadata{}, err
	}

	aspectRatio := float64(config.Width) / float64(config.Height)

	// Calculate grid row span (assuming 30px row height - changed from 20px)
	// Normalize to a base width of 300px
	normalizedHeight := 300.0 / aspectRatio
	gridRowSpan := int(normalizedHeight / 30)

	// Ensure minimum row span of 7
	if gridRowSpan < 7 {
		gridRowSpan = 7
	}

	// Add variety to create more visual interest in masonry layout
	// Extract number from filename to create consistent pseudo-random pattern
	filename := filepath.Base(path)
	var imageNum int
	fmt.Sscanf(filename, "molly_andrewENG-%d.jpg", &imageNum)

	pattern := (imageNum*7 + 3) % 11

	// Apply size variations based on pattern (reduced multipliers)
	switch {
	case pattern == 0:
		// Make some images taller (20% taller instead of 30%)
		gridRowSpan = int(float64(gridRowSpan) * 1.2)
	case pattern%3 == 0:
		// Some images slightly taller (10% instead of 15%)
		gridRowSpan = int(float64(gridRowSpan) * 1.1)
	case pattern%5 == 0:
		// Some images shorter for balance
		gridRowSpan = int(float64(gridRowSpan) * 0.85)
	}

	return ImageMetadata{
		Filename:    filename,
		Width:       config.Width,
		Height:      config.Height,
		AspectRatio: aspectRatio,
		GridRowSpan: gridRowSpan,
	}, nil
}

// extractNumber extracts the numeric part from a filename like "molly_andrewENG-123.jpg"
func extractNumber(path string) int {
	filename := filepath.Base(path)
	// Remove extension
	filename = strings.TrimSuffix(filename, ".jpg")
	// Split by dash to get the number part
	parts := strings.Split(filename, "-")
	if len(parts) < 2 {
		return 0
	}
	num, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0
	}
	return num
}

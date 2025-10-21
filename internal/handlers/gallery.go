package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/apkiernan/thedrewzers/internal/views"
)

func HandleGalleryPage(w http.ResponseWriter, r *http.Request) {
	// Read gallery metadata
	metadataPath := "static/gallery-metadata.json"
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		log.Printf("Warning: Could not read gallery metadata: %v", err)
		// Fallback to empty gallery
		views.App(views.GalleryPage([]views.ImageMetadata{})).Render(r.Context(), w)
		return
	}

	var images []views.ImageMetadata
	if err := json.Unmarshal(data, &images); err != nil {
		log.Printf("Warning: Could not parse gallery metadata: %v", err)
		views.App(views.GalleryPage([]views.ImageMetadata{})).Render(r.Context(), w)
		return
	}

	// Render the gallery page with images
	views.App(views.GalleryPage(images)).Render(r.Context(), w)
}

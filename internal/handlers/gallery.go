package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/apkiernan/thedrewzers/internal/logger"
	"github.com/apkiernan/thedrewzers/internal/views"
)

func HandleGalleryPage(w http.ResponseWriter, r *http.Request) {
	// Read gallery metadata
	metadataPath := "static/gallery-metadata.json"
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		logger.Warn("could not read gallery metadata", "error", err)
		// Fallback to empty gallery
		views.App(views.GalleryPage([]views.ImageMetadata{})).Render(r.Context(), w)
		return
	}

	var images []views.ImageMetadata
	if err := json.Unmarshal(data, &images); err != nil {
		logger.Warn("could not parse gallery metadata", "error", err)
		views.App(views.GalleryPage([]views.ImageMetadata{})).Render(r.Context(), w)
		return
	}

	// Render the gallery page with images
	views.App(views.GalleryPage(images)).Render(r.Context(), w)
}

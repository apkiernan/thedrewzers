package handlers

import (
	"net/http"

	"github.com/apkiernan/thedrewzers/internal/views"
)

func HandleGalleryPage(w http.ResponseWriter, r *http.Request) {
	// Render the gallery page
	views.App(views.GalleryPage()).Render(r.Context(), w)
}

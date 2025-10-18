package handlers

import (
	"net/http"

	"github.com/apkiernan/thedrewzers/internal/views"
)

func HandleHomePage(w http.ResponseWriter, r *http.Request) {
	// Render the index view directly (no first view overlay)
	views.App(views.Index()).Render(r.Context(), w)
}

package handlers

import (
	"net/http"

	"github.com/apkiernan/thedrewzers/internal/views"
)

func HandleVenue(w http.ResponseWriter, r *http.Request) {
	views.App(views.Venue()).Render(r.Context(), w)
}

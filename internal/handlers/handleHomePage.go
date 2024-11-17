package handlers

import (
	"net/http"

	"github.com/apkiernan/thedrewzers/internal/views"
)

func HandleHomePage(w http.ResponseWriter, r *http.Request) {
	page := views.Index()
	app := views.App(page)
	app.Render(r.Context(), w)
}

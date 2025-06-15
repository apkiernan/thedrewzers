package handlers

import (
	"net/http"

	"github.com/apkiernan/thedrewzers/internal/views"
)

func HandleHomePage(w http.ResponseWriter, r *http.Request) {
	// Check for the first_view cookie
	cookie, err := r.Cookie("first_view")
	
	// Determine whether to show the first view overlay
	showFirstView := err != nil || cookie == nil
	
	// Render the combined view template
	views.App(views.CombinedView(showFirstView)).Render(r.Context(), w)
}

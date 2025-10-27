package handlers

import (
	"net/http"

	"github.com/apkiernan/thedrewzers/internal/views"
)

func HandleWeddingPartyPage(w http.ResponseWriter, r *http.Request) {
	views.App(views.WeddingPartySection()).Render(r.Context(), w)
}

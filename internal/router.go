package router

import (
	"net/http"

	"github.com/apkiernan/thedrewzers/internal/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HandleHomePage)
	return mux
}

package main

import (
	"fmt"
	"net/http"

	"github.com/apkiernan/thedrewzers/internal/handlers"
)

func main() {
	server := http.NewServeMux()

	fs := http.FileServer(http.Dir("./static"))
	server.Handle("GET /static/", http.StripPrefix("/static/", fs))
	server.HandleFunc("GET /", handlers.HandleHomePage)
	server.HandleFunc("GET /gallery", handlers.HandleGalleryPage)

	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", server)
}

package main

import (
	"fmt"
	"net/http"

	"github.com/apkiernan/thedrewzers/internal/handlers"
)

func main() {
	server := http.NewServeMux()

	// Serve static files from dist/ directory (includes optimized images)
	fs := http.FileServer(http.Dir("./dist"))
	server.Handle("GET /static/", http.StripPrefix("/static/", fs))
	server.HandleFunc("GET /", handlers.HandleHomePage)
	server.HandleFunc("GET /gallery", handlers.HandleGalleryPage)

	fmt.Println("Server started on port 8080")
	fmt.Println("Serving static files from ./dist")
	if err := http.ListenAndServe(":8080", server); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
		panic(err)
	}
}

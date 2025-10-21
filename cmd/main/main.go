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
	if err := http.ListenAndServe(":8080", server); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
		panic(err)
	}
}

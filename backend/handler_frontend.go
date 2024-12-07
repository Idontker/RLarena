package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
)

// Serves the HTML page
func ServeIndex(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

// Serve static files with correct MIME types
func ServeStatic(w http.ResponseWriter, r *http.Request) {
	LogRequest(r)

	filePath := "." + r.URL.Path // Static files are located in the ./static folder
	ext := filepath.Ext(filePath)
	slog.Debug("Serving static file", "path", filePath, "ext", ext)

	// Set correct Content-Type based on file extension
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}

	http.ServeFile(w, r, filePath)
}

func InitHttpHandler_Frontend_Handler() {
	// Custom handler for static files
	http.HandleFunc("GET /static/", ServeStatic)
	// Custom handler for static files
	http.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/favicon.ico")
	})

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		ServeIndex(w, r)
	})

	http.HandleFunc("GET /leaderboard", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		http.ServeFile(w, r, "./templates/leaderboard.html")
	})

	http.HandleFunc("GET /player", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		http.ServeFile(w, r, "./templates/player.html")
	})
}

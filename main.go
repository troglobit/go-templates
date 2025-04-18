package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type contextKey string

const keyIsHTMX contextKey = "isHTMX"

var templates map[string]*template.Template

func main() {
	loadTemplates()

	// Create a new router using Gorilla Mux
	r := mux.NewRouter()

	// Apply middlewares
	r.Use(loggingMiddleware)
	r.Use(htmxMiddleware)

	// Define routes
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/page/{name}", pageHandler)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/page/dashboard", http.StatusSeeOther)
	})

	// Create server with timeouts for better security
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Listening on :8080")
	log.Fatal(srv.ListenAndServe())
}

func loadTemplates() {
	templates = make(map[string]*template.Template)

	layout := "templates/layout.html"
	pages, err := filepath.Glob("templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	for _, page := range pages {
		if strings.HasSuffix(page, "layout.html") {
			continue
		}

		name := strings.TrimSuffix(filepath.Base(page), ".html")
		tmpl := template.Must(template.ParseFiles(layout, page))
		templates[name] = tmpl
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf(
			"%s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}

func htmxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isHTMX := r.Header.Get("HX-Request") == "true"
		ctx := context.WithValue(r.Context(), keyIsHTMX, isHTMX)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page := vars["name"]

	tmpl, ok := templates[page]
	if !ok {
		http.NotFound(w, r)
		return
	}

	data := map[string]any{
		"Title": strings.Title(page),
	}

	isHTMX := r.Context().Value(keyIsHTMX).(bool)
	if isHTMX {
		tmpl.ExecuteTemplate(w, "content", data)
	} else {
		tmpl.ExecuteTemplate(w, "layout.html", data)
	}
}

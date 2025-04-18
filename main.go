package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

type contextKey string

const keyIsHTMX contextKey = "isHTMX"

var templates map[string]*template.Template

func main() {
	loadTemplates()

	mux := http.NewServeMux()
	mux.HandleFunc("/page/", pageHandler)
	// mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/page/dashboard", http.StatusSeeOther)
	})

	log.Println("Listening on :8080")
	err := http.ListenAndServe(":8080", htmxMiddleware(mux))
	if err != nil {
		log.Fatal(err)
	}
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

func htmxMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isHTMX := r.Header.Get("HX-Request") == "true"
		ctx := context.WithValue(r.Context(), keyIsHTMX, isHTMX)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	page := filepath.Base(r.URL.Path[len("/page/"):])
	tmpl, ok := templates[page]
	if !ok {
		http.NotFound(w, r)
		return
	}

	data := map[string]any{
		"Title": page,
	}

	isHTMX := r.Context().Value(keyIsHTMX).(bool)
	if isHTMX {
		tmpl.ExecuteTemplate(w, "content", data)
	} else {
		tmpl.ExecuteTemplate(w, "layout.html", data)
	}
}

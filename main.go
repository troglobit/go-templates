package main

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func renderTemplate(w http.ResponseWriter, name string, data any, full bool) {
	files := []string{filepath.Join("templates", name+".html")}
	if full {
		files = append([]string{"templates/layout.html"}, files...)
	}

	tmpl := template.Must(template.ParseFiles(files...))

	if full {
		tmpl.ExecuteTemplate(w, "layout.html", data)
	} else {
		tmpl.ExecuteTemplate(w, "content", data)
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	page := filepath.Base(r.URL.Path[len("/page/"):])
	if page == "" {
		http.NotFound(w, r)
		return
	}

	full := r.Header.Get("HX-Request") != "true"

	renderTemplate(w, page, map[string]any{
		"Title": page,
	}, full)
}

func main() {
	http.HandleFunc("/page/", pageHandler)
	http.Handle("/", http.FileServer(http.Dir("static")))

	http.ListenAndServe(":8080", nil)
}

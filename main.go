package main

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func renderTemplate(w http.ResponseWriter, name string, data any) {
	tmplFiles := []string{
		"templates/layout.html",
		filepath.Join("templates", name+".html"),
	}

	tmpl := template.Must(template.ParseFiles(tmplFiles...))
	tmpl.ExecuteTemplate(w, "layout.html", data)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	page := filepath.Base(r.URL.Path[len("/page/"):])
	if page == "" {
		http.NotFound(w, r)
		return
	}

	renderTemplate(w, page, map[string]any{
		"Title": page,
	})
}

func main() {
	http.HandleFunc("/page/", pageHandler)
	http.Handle("/", http.FileServer(http.Dir("static")))

	http.ListenAndServe(":8080", nil)
}

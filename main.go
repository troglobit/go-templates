package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed templates
var templatesFS embed.FS

var templates map[string]*template.Template

func main() {
	if err := loadTemplates(); err != nil {
		log.Fatal("Failed to load templates:", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/page/dashboard", http.StatusSeeOther)
	})
	r.Get("/page/{name}", pageHandler)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func loadTemplates() error {
	templates = make(map[string]*template.Template)

	tmplFiles, err := templatesFS.ReadDir("templates")
	if err != nil {
		return err
	}

	layoutContent, err := templatesFS.ReadFile("templates/layout.html")
	if err != nil {
		return err
	}

	for _, file := range tmplFiles {
		if file.Name() == "layout.html" || file.IsDir() {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".html")

		pageContent, err := templatesFS.ReadFile("templates/" + file.Name())
		if err != nil {
			return err
		}

		// Create a new template and first parse the layout
		t := template.New("layout.html")
		t, err = t.Parse(string(layoutContent))
		if err != nil {
			return err
		}

		// Then parse the page template with its content definition
		t, err = t.Parse(string(pageContent))
		if err != nil {
			return err
		}

		templates[name] = t
	}

	return nil
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	page := chi.URLParam(r, "name")

	tmpl, ok := templates[page]
	if !ok {
		http.NotFound(w, r)
		return
	}

	data := map[string]any{"Title": strings.Title(page)}

	// Choose template based on whether this is an HTMX request
	templateName := "layout.html"
	if r.Header.Get("HX-Request") == "true" {
		templateName = "content"
	}

	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

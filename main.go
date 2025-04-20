package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/msteinert/pam"
	"github.com/spf13/pflag"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

var templates map[string]*template.Template
var debugMode bool
var sessionSecret string
var sessionStoragePath string

type SessionConfig struct {
	Secret string `json:"secret"`
}

func main() {
	pflag.BoolVarP(&debugMode, "debug", "d", false, "Enable debug mode (allows admin/admin login)")
	pflag.StringVarP(&sessionStoragePath, "secret", "s", "/var/lib/misc/", "Directory for session secret")
	pflag.Parse()

	if debugMode {
		log.Println("WARNING: Debug mode enabled - insecure authentication is active")
	}

	if err := ensureSessionSecret(); err != nil {
		log.Fatal("Failed to secure session secret:", err)
	}

	if err := loadTemplates(); err != nil {
		log.Fatal("Failed to load templates:", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Serve static files (for favicon.ico and other assets)
	fileServer := http.FileServer(http.FS(staticFS))
	r.Handle("/static/*", http.StripPrefix("/", fileServer))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		})
		r.Get("/login", loginPageHandler)
		r.Post("/login", loginHandler)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/logout", logoutHandler)
		r.Get("/page/{name}", pageHandler)
	})

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

	data := map[string]any{
		"Title":    strings.Title(page),
		"Username": getUsername(r),
	}

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

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	// Check if already logged in
	if _, err := r.Cookie("session"); err == nil {
		http.Redirect(w, r, "/page/dashboard", http.StatusSeeOther)
		return
	}

	tmpl, ok := templates["login"]
	if !ok {
		http.Error(w, "Login template not found", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title": "Login",
		"Error": r.URL.Query().Get("error"),
	}

	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")

	if username == "" || password == "" {
		http.Redirect(w, r, "/login?error=empty_fields", http.StatusSeeOther)
		return
	}

	// Authenticate user
	authenticated := false

	// In debug mode, allow admin/admin
	if debugMode && username == "admin" && password == "admin" {
		authenticated = true
	} else {
		// Use PAM for authentication
		authenticated = authenticateWithPAM(username, password)
	}

	if !authenticated {
		http.Redirect(w, r, "/login?error=invalid_credentials", http.StatusSeeOther)
		return
	}

	// Set session cookie
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{
		Name:     "session",
		Value:    createSessionToken(username),
		Expires:  expiration,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)

	// Redirect to dashboard
	http.Redirect(w, r, "/page/dashboard", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	cookie := http.Cookie{
		Name:     "session",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")

		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validate session token (in a real app, you'd check a session store)
		username := validateSessionToken(cookie.Value)
		if username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Store username in request context for use in handlers
		ctx := r.Context()
		ctx = context.WithValue(ctx, "username", username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUsername(r *http.Request) string {
	if username, ok := r.Context().Value("username").(string); ok {
		return username
	}
	return ""
}

// For a real application, use a proper session management library
func createSessionToken(username string) string {
	// This is a simplified example - use a proper session library in production
	return username + "-" + sessionSecret
}

func validateSessionToken(token string) string {
	// Simple validation for demonstration purposes
	if parts := strings.Split(token, "-"); len(parts) >= 2 && parts[1] == sessionSecret {
		return parts[0]
	}
	return ""
}

func authenticateWithPAM(username, password string) bool {
	// Initialize PAM transaction
	t, err := pam.StartFunc("system-auth", username, func(style pam.Style, msg string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			return password, nil
		case pam.PromptEchoOn:
			return "", nil
		case pam.ErrorMsg, pam.TextInfo:
			log.Println(msg)
			return "", nil
		default:
			return "", errors.New("unrecognized message style")
		}
	})

	if err != nil {
		log.Printf("PAM start error: %v", err)
		return false
	}

	// Authenticate the user
	err = t.Authenticate(0)
	if err != nil {
		log.Printf("PAM authentication error: %v", err)
		return false
	}

	// Check account validity
	err = t.AcctMgmt(0)
	if err != nil {
		log.Printf("PAM account error: %v", err)
		return false
	}

	return true
}

// ensureSessionSecret makes sure we have a valid session secret
// It will try to load one from disk, or generate a new one if needed
func ensureSessionSecret() error {
	// If provided on command line, use that
	if sessionSecret != "" {
		return nil
	}

	configPath := filepath.Join(sessionStoragePath, "session.json")

	// Try to load existing config
	data, err := os.ReadFile(configPath)
	if err == nil {
		var config SessionConfig
		if err := json.Unmarshal(data, &config); err == nil && config.Secret != "" {
			sessionSecret = config.Secret
			log.Println("Using existing session secret from", configPath)
			return nil
		}
	}

	// Generate a new secret
	secret := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, secret); err != nil {
		return err
	}
	sessionSecret = base64.StdEncoding.EncodeToString(secret)

	// Save to disk
	config := SessionConfig{Secret: sessionSecret}
	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return err
	}

	log.Println("Generated and saved new session secret to", configPath)
	return nil
}

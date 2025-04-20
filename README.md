# Go Templates Example

A lightweight single-page application for embedded systems, built with
Go, HTMX, and Bootstrap. Features PAM authentication and a simple,
responsive interface.


## Features

- Minimal dependencies - only Chi router, PAM authentication, and pflag
- Templates and static assets embedded directly in the binary
- Standard library logging
- PAM-based authentication with debug mode for development
- HTMX for SPA-like navigation without JavaScript
- Bootstrap 5 for responsive UI
- Simple structure optimized for embedded systems


## Project Structure

```
├── main.go                 # Main application with auth and embedded resources
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── favicon.ico             # Application favicon
├── templates/              # HTML templates (embedded in binary)
│   ├── layout.html         # Base layout template
│   ├── login.html          # Login page
│   ├── dashboard.html      # Dashboard page content
│   ├── settings.html       # Settings page content
│   └── users.html          # User management page content
└── static/                 # Static assets (embedded in binary)
    └── favicon.ico         # Application favicon
```


## Requirements

- Go 1.19+ (for embed package)
- Chi router (`github.com/go-chi/chi/v5`)
- PAM authentication (`github.com/msteinert/pam`)
- Command-line flags (`github.com/spf13/pflag`)


## Building and Running

### Development

```bash
# Get dependencies
go mod tidy

# Run the application in debug mode (allows admin/admin login)
go run main.go --debug
```

## Authentication

The application uses Linux PAM for authentication. In debug mode, it
also allows using "admin" as both username and password.

The application uses PAM "system-auth" for authentication.  If you need
to customize PAM authentication, you can modify the PAM service name in
the code.


## Customization

### Adding a New Page

1. Create a new HTML template in the `templates` directory with your
   content inside the "content" block:

   ```html
   {{ define "content" }}
   <!-- Your content here -->
   {{ end }}
   ```

2. Add a link to your new page in the sidebar menu in

   `templates/layout.html`


## License

MIT

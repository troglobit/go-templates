# Go Templates Example

This project demonstrates a modern, responsive admin dashboard built
with Go, HTMX, and Bootstrap 5. It uses Gorilla Mux for routing and
middleware capabilities.

## Features

- Single-page application behavior with HTMX
- Responsive design using Bootstrap 5
- Clean routing with Gorilla Mux
- Middleware for logging and HTMX detection
- Template inheritance for layout consistency
- Dashboard with mock data and visualizations
- Settings page with form elements
- User management interface

## Project Structure

```
├── main.go                   # Main application file
├── static/                   # Static assets directory
├── templates/
│   ├── layout.html           # Base layout template
│   ├── dashboard.html        # Dashboard page content
│   ├── settings.html         # Settings page content
│   └── users.html            # User management page content
└── README.md                 # This file
```

## Requirements

- Go 1.16 or higher
- Gorilla Mux package

## Installation

1. Clone this repository:
   ```
   git clone https://github.com/troglobit/go-templates.git
   cd go-templates
   ```

2. Run the application:
   ```
   go run main.go
   ```

3. Open your browser and navigate to `http://localhost:8080`

## Customization

### Adding a New Page

1. Create a new HTML template in the `templates` directory
2. Add the content inside a content block:
   ```html
   {{ define "content" }}
   <!-- Your content here -->
   {{ end }}
   ```
3. Add a new link in the sidebar within `layout.html`

### Modifying Styles

The main layout uses Bootstrap 5. You can:
- Modify the existing classes
- Add custom CSS in the `<style>` section of `layout.html`
- Link external stylesheets

## Technologies Used

- **Go** - Backend server and templating
- **HTMX** - Frontend interactivity without writing JavaScript
- **Bootstrap 5** - Responsive UI framework
- **Gorilla Mux** - Enhanced HTTP routing

## License

MIT

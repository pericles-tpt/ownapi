package views

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func RegisterViews(m *http.ServeMux) {
	// Serve `js` and `css` at "/static/"
	m.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./dist/"))))

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			return
		}

		start := time.Now()

		// Handle "/data" paths, with file directory suffixes
		switch r.URL.Path {
		case "/":
			Home(w, r)
		default:
			http.NotFound(w, r)
		}

		fmt.Printf("[fe]: %s request took: %v\n", r.URL.Path, time.Since(start))
	})
}

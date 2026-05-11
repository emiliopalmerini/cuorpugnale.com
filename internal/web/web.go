package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"time"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

var trailerLaunch = time.Date(2026, 5, 26, 10, 0, 0, 0, time.FixedZone("CEST", 2*60*60))

type indexData struct {
	Launched     bool
	LaunchUnixMs int64
	SpotifyEmbed string
}

type Server struct {
	tmpl *template.Template
}

func New() (*Server, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{tmpl: tmpl}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	staticSub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticSub)))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := indexData{
			Launched:     time.Now().After(trailerLaunch),
			LaunchUnixMs: trailerLaunch.UnixMilli(),
			SpotifyEmbed: "",
		}
		if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return mux
}

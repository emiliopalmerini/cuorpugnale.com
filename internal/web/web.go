package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"time"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

var defaultTrailerLaunch = time.Date(2026, 5, 26, 10, 0, 0, 0, time.FixedZone("CEST", 2*60*60))

const siteURL = "https://cuorpugnale.com"
const defaultSpotifyURL = "https://open.spotify.com/search/Cuorpugnale"
const instagramURL = "https://www.instagram.com/cuorpugnale"

type indexData struct {
	Launched     bool
	LaunchUnixMs int64
	SpotifyEmbed string
	SpotifyURL   string
	InstagramURL string
	SiteURL      string
	OGImageURL   string
}

type Server struct {
	tmpl          *template.Template
	trailerLaunch time.Time
}

func New() (*Server, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{
		tmpl:          tmpl,
		trailerLaunch: trailerLaunchTime(time.Now()),
	}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	mux.Handle("GET /static/", securityHeaders(cacheStatic(http.StripPrefix("/static/", http.FileServerFS(staticSub)))))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		data := indexData{
			Launched:     now.After(s.trailerLaunch),
			LaunchUnixMs: s.trailerLaunch.UnixMilli(),
			SpotifyEmbed: "",
			SpotifyURL:   spotifyURL(),
			InstagramURL: instagramURL,
			SiteURL:      siteURL,
			OGImageURL:   siteURL + "/static/img/cuorpugnale_logotipo.jpg",
		}
		if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self'; img-src 'self' data:; frame-src https://open.spotify.com; base-uri 'self'; form-action 'self'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func trailerLaunchTime(now time.Time) time.Time {
	if delay := os.Getenv("TRAILER_LAUNCH_DELAY"); delay != "" {
		duration, err := time.ParseDuration(delay)
		if err == nil && duration > 0 {
			return now.Add(duration)
		}
	}

	if value := os.Getenv("TRAILER_LAUNCH_AT"); value != "" {
		launch, err := time.Parse(time.RFC3339, value)
		if err == nil {
			return launch
		}
	}

	return defaultTrailerLaunch
}

func spotifyURL() string {
	if url := os.Getenv("SPOTIFY_URL"); url != "" {
		return url
	}
	return defaultSpotifyURL
}

func cacheStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

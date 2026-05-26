package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHomePageRendersSocialLinks(t *testing.T) {
	rec := renderHome(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	for _, link := range []string{
		"https://youtube.com/@cuorpugnale?si=GMnp_eG1ujakmclG",
		"https://www.instagram.com/cuorpugnale",
	} {
		if !strings.Contains(body, link) {
			t.Errorf("home page does not contain %q", link)
		}
	}
}

func TestHomePageUsesOptimizedHeroImage(t *testing.T) {
	rec := renderHome(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	for _, snippet := range []string{
		`rel="preload"`,
		`as="image"`,
		`href="/static/img/cuorpugnale_logotipo-960.avif"`,
		`imagesizes="(max-width: 956px) 92vw, 880px"`,
		`<picture class="hero__brand-picture">`,
		`type="image/avif"`,
		`/static/img/cuorpugnale_logotipo-640.avif 640w`,
		`/static/img/cuorpugnale_logotipo-960.webp 960w`,
		`src="/static/img/cuorpugnale_logotipo-960.jpg"`,
		`width="960"`,
		`height="414"`,
	} {
		if !strings.Contains(body, snippet) {
			t.Errorf("home page does not contain %q", snippet)
		}
	}
}

func TestHomePageEmbedsLatestYouTubeVideoAfterLaunch(t *testing.T) {
	oldLaunch := trailerLaunch
	trailerLaunch = time.Now().Add(-time.Hour)
	t.Cleanup(func() {
		trailerLaunch = oldLaunch
	})

	rec := renderHome(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	for _, snippet := range []string{
		`class="hero__video"`,
		`title="Ultimo video di Cuorpugnale su YouTube"`,
		`src="https://www.youtube-nocookie.com/embed/videoseries?list=UU0hhZyFibLeVk9KDIatuIag&amp;rel=0"`,
		`allowfullscreen`,
	} {
		if !strings.Contains(body, snippet) {
			t.Errorf("home page does not contain %q", snippet)
		}
	}
}

func TestOptimizedHeroAssetsAreServed(t *testing.T) {
	server, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	handler := server.Handler()
	for _, path := range []string{
		"/static/img/cuorpugnale_logotipo-640.avif",
		"/static/img/cuorpugnale_logotipo-960.avif",
		"/static/img/cuorpugnale_logotipo-640.webp",
		"/static/img/cuorpugnale_logotipo-960.webp",
		"/static/img/cuorpugnale_logotipo-960.jpg",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("%s status = %d, want %d", path, rec.Code, http.StatusOK)
		}
		if cacheControl := rec.Header().Get("Cache-Control"); !strings.Contains(cacheControl, "immutable") {
			t.Errorf("%s Cache-Control = %q, want immutable static cache", path, cacheControl)
		}
	}
}

func TestHomePageDoesNotLoadExternalFonts(t *testing.T) {
	rec := renderHome(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	for _, blocked := range []string{
		"fonts.googleapis.com",
		"fonts.gstatic.com",
		"Playfair Display",
		"Lora",
	} {
		if strings.Contains(body, blocked) {
			t.Errorf("home page contains external font dependency %q", blocked)
		}
	}
}

func TestSecurityPolicyDoesNotAllowExternalFonts(t *testing.T) {
	server, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	for _, blocked := range []string{
		"fonts.googleapis.com",
		"fonts.gstatic.com",
	} {
		if strings.Contains(csp, blocked) {
			t.Errorf("Content-Security-Policy contains external font source %q", blocked)
		}
	}
}

func TestSecurityPolicyAllowsYouTubeEmbed(t *testing.T) {
	server, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "frame-src https://www.youtube-nocookie.com") {
		t.Errorf("Content-Security-Policy = %q, want YouTube no-cookie frame source", csp)
	}
	if strings.Contains(csp, "https://open.spotify.com") {
		t.Errorf("Content-Security-Policy contains old Spotify frame source")
	}
}

func renderHome(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()

	server, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	return rec
}

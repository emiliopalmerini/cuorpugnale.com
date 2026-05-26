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
		"https://open.spotify.com/search/Cuorpugnale",
		"https://www.instagram.com/cuorpugnale",
	} {
		if !strings.Contains(body, link) {
			t.Errorf("home page does not contain %q", link)
		}
	}
	if strings.Contains(body, `>YouTube</a>`) {
		t.Errorf("home page contains YouTube social link, want Spotify")
	}
}

func TestHomePageCanOverrideSpotifyLink(t *testing.T) {
	t.Setenv("SPOTIFY_URL", "https://open.spotify.com/show/example")

	rec := renderHome(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !strings.Contains(rec.Body.String(), "https://open.spotify.com/show/example") {
		t.Errorf("home page does not contain overridden Spotify URL")
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

func TestHomePageShowsCountdownBeforeLaunch(t *testing.T) {
	t.Setenv("TRAILER_LAUNCH_DELAY", "10s")

	rec := renderHome(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	for _, snippet := range []string{
		`class="countdown"`,
		`Il teaser trailer arriva tra`,
		`data-target="`,
	} {
		if !strings.Contains(body, snippet) {
			t.Errorf("home page does not contain %q", snippet)
		}
	}
	if strings.Contains(body, `class="hero__video"`) {
		t.Errorf("home page contains launched YouTube embed during countdown")
	}
}

func TestHomePageEmbedsSpotifyShowAfterLaunch(t *testing.T) {
	t.Setenv("TRAILER_LAUNCH_AT", time.Now().Add(-time.Hour).Format(time.RFC3339))

	rec := renderHome(t)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	for _, snippet := range []string{
		`data-testid="embed-iframe"`,
		`class="hero__spotify"`,
		`src="https://open.spotify.com/embed/show/033nh6SpB5aFTDuQ6FMkSI?utm_source=generator"`,
		`height="352"`,
		`allowfullscreen`,
		`allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"`,
	} {
		if !strings.Contains(body, snippet) {
			t.Errorf("home page does not contain %q", snippet)
		}
	}
}

func TestTrailerLaunchTimeCanUseAbsoluteOverride(t *testing.T) {
	now := time.Date(2026, 5, 26, 8, 0, 0, 0, time.UTC)
	want := time.Date(2026, 5, 26, 10, 15, 0, 0, time.UTC)
	t.Setenv("TRAILER_LAUNCH_AT", want.Format(time.RFC3339))

	got := trailerLaunchTime(now)

	if !got.Equal(want) {
		t.Fatalf("trailerLaunchTime() = %v, want %v", got, want)
	}
}

func TestTrailerLaunchTimeCanUseDelayOverride(t *testing.T) {
	now := time.Date(2026, 5, 26, 8, 0, 0, 0, time.UTC)
	t.Setenv("TRAILER_LAUNCH_DELAY", "10s")

	got := trailerLaunchTime(now)
	want := now.Add(10 * time.Second)

	if !got.Equal(want) {
		t.Fatalf("trailerLaunchTime() = %v, want %v", got, want)
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

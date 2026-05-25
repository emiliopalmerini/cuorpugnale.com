package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomePageRendersSocialLinks(t *testing.T) {
	server, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

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

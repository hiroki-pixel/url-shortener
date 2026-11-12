package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hiroki-pixel/url-shortener/service"
)

type FakeRedirectService struct {
	OriginalURL string
}

func (f *FakeRedirectService) ResolveRedirect(
	ctx context.Context, input service.RedirectInput,
) (service.RedirectOutput, error) {
	return service.RedirectOutput{
		// f.OriginalURLを設定する
		OriginalURL: f.OriginalURL,
	}, nil
}

func TestHandleRedirect(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()

	fakeService := &FakeRedirectService{
		OriginalURL: "http://example.com",
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{short_code}", func(w http.ResponseWriter, r *http.Request) {
		HandleRedirect(w, r, fakeService)
	})

	// Act
	mux.ServeHTTP(rec, req)
	got := rec.Code
	want := http.StatusFound

	gotLocation := rec.Header().Get("Location")
	wantLocation := "http://example.com"

	// Assert
	if got != want {
		t.Errorf("got %v,want %v", got, want)
	}

	if gotLocation != wantLocation {
		t.Errorf("gotLocation %q, wantLocation %q", gotLocation, wantLocation)
	}
}

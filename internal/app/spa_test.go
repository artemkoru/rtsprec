package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testUI() fstest.MapFS {
	return fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte("<html>UI-SHELL</html>")},
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('app')")},
	}
}

func serve(t *testing.T, h http.Handler, path string) (int, string, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	body, _ := io.ReadAll(rec.Result().Body)
	return rec.Code, rec.Header().Get("Content-Type"), string(body)
}

func TestSPAServesExistingAsset(t *testing.T) {
	code, _, body := serve(t, newSPAHandler(testUI()), "/assets/app.js")
	if code != http.StatusOK || !strings.Contains(body, "console.log") {
		t.Fatalf("asset: status=%d body=%q", code, body)
	}
}

func TestSPAFallsBackToIndexForClientRoutes(t *testing.T) {
	h := newSPAHandler(testUI())
	for _, p := range []string{"/", "/live", "/archive", "/settings", "/login", "/some/deep/route"} {
		code, ct, body := serve(t, h, p)
		if code != http.StatusOK {
			t.Fatalf("%s: status=%d, want 200", p, code)
		}
		if !strings.Contains(body, "UI-SHELL") {
			t.Fatalf("%s: body=%q, want index.html", p, body)
		}
		if !strings.HasPrefix(ct, "text/html") {
			t.Fatalf("%s: content-type=%q, want text/html", p, ct)
		}
	}
}

func TestSPAMissingAssetIsNotFound(t *testing.T) {
	code, _, body := serve(t, newSPAHandler(testUI()), "/assets/missing.js")
	if code != http.StatusNotFound {
		t.Fatalf("missing asset: status=%d, want 404", code)
	}
	if strings.Contains(body, "UI-SHELL") {
		t.Fatal("index.html served for a missing asset")
	}
}

func TestSPANeverServesHTMLForAPIOrMediaPaths(t *testing.T) {
	h := newSPAHandler(testUI())
	for _, p := range []string{"/api/unknown", "/api/cameras", "/hls/cam1/playlist.m3u8"} {
		code, _, body := serve(t, h, p)
		if code != http.StatusNotFound {
			t.Fatalf("%s: status=%d, want 404", p, code)
		}
		if strings.Contains(body, "UI-SHELL") {
			t.Fatalf("%s: index.html served for an API path", p)
		}
	}
}

func TestSPAPlaceholderOnlyReportsNotBuilt(t *testing.T) {
	placeholder := fstest.MapFS{".gitkeep": &fstest.MapFile{}}
	code, _, body := serve(t, newSPAHandler(placeholder), "/live")
	if code != http.StatusNotFound || !strings.Contains(body, "not built") {
		t.Fatalf("placeholder: status=%d body=%q", code, body)
	}
}

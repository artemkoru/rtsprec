package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artemkoru/rtsprec/internal/storage"
)

// setupTestEnv points configuration at temporary files and returns the paths.
func setupTestEnv(t *testing.T) {
	t.Helper()
	clearConfigEnv(t)

	dir := t.TempDir()
	t.Setenv("DB_PATH", filepath.Join(dir, "test.db"))
	t.Setenv("ENCRYPTION_KEY_PATH", filepath.Join(dir, "encryption.key"))
	t.Setenv("HLS_OUTPUT_PATH", filepath.Join(dir, "hls"))
	t.Setenv("HTTP_PORT", "18080")
	t.Setenv("AUTH_BASIC_USERNAME", "admin")
	t.Setenv("AUTH_BASIC_PASSWORD", "test-password")
}

// newInitializedApp creates and initializes an App with test configuration.
func newInitializedApp(t *testing.T) *App {
	t.Helper()
	setupTestEnv(t)

	a, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	// Tests must not contact AWS: use the uploader without the connection check
	a.newS3Uploader = storage.NewS3UploaderWithoutCheck
	if err := a.Init(context.Background()); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	t.Cleanup(func() {
		_ = a.Shutdown(context.Background())
	})
	return a
}

func TestNewDoesNotOpenDatabase(t *testing.T) {
	setupTestEnv(t)
	// A path that cannot be opened: New must not touch it
	t.Setenv("DB_PATH", "/nonexistent-dir-for-test/rtsprec.db")

	a, err := New()
	if err != nil {
		t.Fatalf("New() error = %v, want nil (New must not open the database)", err)
	}
	if a.db != nil {
		t.Fatal("New() opened the database")
	}
}

func TestInitFailsForUnusableDatabasePath(t *testing.T) {
	setupTestEnv(t)
	t.Setenv("DB_PATH", "/nonexistent-dir-for-test/rtsprec.db")

	a, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := a.Init(context.Background()); err == nil {
		t.Fatal("Init() error = nil, want error for unusable DB_PATH")
	}
}

func TestInitFailsForUnusableEncryptionKeyPath(t *testing.T) {
	setupTestEnv(t)
	t.Setenv("ENCRYPTION_KEY_PATH", "/nonexistent-dir-for-test/encryption.key")

	a, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := a.Init(context.Background()); err == nil {
		_ = a.Shutdown(context.Background())
		t.Fatal("Init() error = nil, want error for unusable ENCRYPTION_KEY_PATH")
	}
}

func TestServerAddressUsesHTTPPort(t *testing.T) {
	a := newInitializedApp(t)

	if want := ":18080"; a.server.Addr != want {
		t.Fatalf("server.Addr = %q, want %q", a.server.Addr, want)
	}
}

func TestRoutesAreServedNotStubbed(t *testing.T) {
	a := newInitializedApp(t)

	token, err := a.authManager.AuthenticateBasic("admin", "test-password")
	if err != nil {
		t.Fatalf("AuthenticateBasic() error = %v", err)
	}

	cases := []struct {
		method, path string
	}{
		{http.MethodGet, "/api/cameras"},
		{http.MethodPost, "/api/cameras"},
		{http.MethodGet, "/api/archive"},
		{http.MethodGet, "/api/events"},
		{http.MethodGet, "/api/diagnostics"},
		{http.MethodPost, "/api/auth/logout"},
		{http.MethodGet, "/hls/cam1/playlist.m3u8"},
		{http.MethodGet, "/"},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader("{}"))
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			a.server.Handler.ServeHTTP(rec, req)

			if rec.Code == http.StatusNotImplemented {
				t.Fatalf("status = 501 Not Implemented, route is still a stub")
			}
		})
	}
}

func TestProtectedRoutesRequireToken(t *testing.T) {
	a := newInitializedApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/cameras", nil)
	rec := httptest.NewRecorder()
	a.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d without token, want 401", rec.Code)
	}
}

func TestInitCreatesStaticAndHLSHandlers(t *testing.T) {
	a := newInitializedApp(t)

	if a.apiServer == nil || a.hlsHandler == nil || a.archiveManager == nil || a.s3Uploader == nil {
		t.Fatal("Init() did not build all dependencies")
	}
	if _, err := os.Stat(a.config.EncryptionKeyPath); err != nil {
		t.Fatalf("encryption key was not created: %v", err)
	}
}

func TestLoginIsReachableWithoutToken(t *testing.T) {
	a := newInitializedApp(t)

	body := strings.NewReader(`{"username":"admin","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	rec := httptest.NewRecorder()
	a.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"token"`) {
		t.Fatalf("login body %q has no token", rec.Body.String())
	}

	// The browser must send the session cookie to every API path, so its Path must be "/"
	var session *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session_token" {
			session = c
		}
	}
	if session == nil {
		t.Fatal("login did not set session_token cookie")
	}
	if session.Path != "/" {
		t.Fatalf("session_token Path = %q, want \"/\"", session.Path)
	}
}

func TestLoginWithWrongPasswordIsRejectedByHandler(t *testing.T) {
	a := newInitializedApp(t)

	body := strings.NewReader(`{"username":"admin","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	rec := httptest.NewRecorder()
	a.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "missing session token") {
		t.Fatal("login was blocked by the auth middleware")
	}
}

func TestLogoutIsReachableWithoutToken(t *testing.T) {
	a := newInitializedApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rec := httptest.NewRecorder()
	a.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want 200", rec.Code)
	}
}

func TestCameraStatusRouteExtractsCameraID(t *testing.T) {
	a := newInitializedApp(t)

	token, err := a.authManager.AuthenticateBasic("admin", "test-password")
	if err != nil {
		t.Fatalf("AuthenticateBasic() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/cameras/cam1/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	a.server.Handler.ServeHTTP(rec, req)

	// The handler answers "camera not found" for an unknown ID, which proves the
	// route matched and the ID was extracted; a plain 404 would mean no route matched.
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "camera not found") {
		t.Fatalf("status = %d body = %q, want handler's camera-not-found response", rec.Code, rec.Body.String())
	}
}

func TestStartCamerasWithoutEnabledCameras(t *testing.T) {
	a := newInitializedApp(t)

	// Must return without starting goroutines when there are no enabled cameras
	a.StartCameras()
}

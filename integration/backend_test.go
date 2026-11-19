package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// do sends a request to the shared backend and returns the status code and body.
func do(t *testing.T, method, path string, body any, token string) (int, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, shared.baseURL+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, data
}

// login returns a session token for the configured basic-auth user.
func login(t *testing.T) string {
	t.Helper()

	status, body := do(t, http.MethodPost, "/api/auth/login",
		map[string]string{"username": adminUser, "password": adminPassword}, "")
	if status != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", status, body)
	}

	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &out); err != nil || out.Token == "" {
		t.Fatalf("login response has no token: %s", body)
	}
	return out.Token
}

func TestLoginReturnsToken(t *testing.T) {
	login(t)
}

func TestLoginWithWrongPasswordIsRejected(t *testing.T) {
	status, body := do(t, http.MethodPost, "/api/auth/login",
		map[string]string{"username": adminUser, "password": "wrong"}, "")

	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", status)
	}
	if strings.Contains(string(body), "missing session token") {
		t.Fatal("login was blocked by the auth middleware instead of the login handler")
	}
}

func TestProtectedRouteRequiresToken(t *testing.T) {
	status, _ := do(t, http.MethodGet, "/api/cameras", nil, "")
	if status != http.StatusUnauthorized {
		t.Fatalf("status without token = %d, want 401", status)
	}
}

func TestProtectedRouteAcceptsToken(t *testing.T) {
	status, body := do(t, http.MethodGet, "/api/cameras", nil, login(t))
	if status != http.StatusOK {
		t.Fatalf("status with token = %d, body = %s, want 200", status, body)
	}
}

func TestConfigS3ReturnsJSON(t *testing.T) {
	status, body := do(t, http.MethodGet, "/api/config/s3", nil, login(t))
	if status != http.StatusOK {
		t.Fatalf("status = %d, body = %s, want 200", status, body)
	}
	if !json.Valid(body) {
		t.Fatalf("body is not valid JSON: %s", body)
	}
}

func TestCameraStatusUnknownCameraIsNotFound(t *testing.T) {
	status, body := do(t, http.MethodGet, "/api/cameras/does-not-exist/status", nil, login(t))

	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
	if !strings.Contains(string(body), "camera not found") {
		t.Fatalf("body = %s, want the handler's 'camera not found' response", body)
	}
}

func TestLogoutWithoutTokenReturnsOK(t *testing.T) {
	status, _ := do(t, http.MethodPost, "/api/auth/logout", nil, "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
}

func TestContainerStopsCleanlyOnSIGTERM(t *testing.T) {
	ctx := context.Background()

	// A separate stack, so stopping the backend does not affect the shared one.
	s, err := startStack(ctx, nil)
	if err != nil {
		t.Fatalf("start stack: %v", err)
	}
	t.Cleanup(func() { s.terminate(ctx) })

	timeout := 15 * time.Second
	if err := s.backend.Stop(ctx, &timeout); err != nil {
		t.Fatalf("stop backend: %v", err)
	}

	state, err := s.backend.State(ctx)
	if err != nil {
		t.Fatalf("container state: %v", err)
	}
	if state.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0 after SIGTERM", state.ExitCode)
	}

	logs, err := s.backend.Logs(ctx)
	if err != nil {
		t.Fatalf("container logs: %v", err)
	}
	defer logs.Close()
	data, err := io.ReadAll(logs)
	if err != nil {
		t.Fatalf("read logs: %v", err)
	}
	if !strings.Contains(string(data), "shutdown complete") {
		t.Fatalf("logs do not contain 'shutdown complete':\n%s", data)
	}
}

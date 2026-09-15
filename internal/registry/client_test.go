package registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"HarborDownloader/internal/model"
)

func TestRegistryPing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/" || r.URL.Path == "/v2":
			w.WriteHeader(http.StatusOK)
		case strings.HasPrefix(r.URL.Path, "/api/v2.0/"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"healthy"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cfg := model.RegistryConfig{
		Registry: ts.URL,
		Insecure: true,
	}
	c := NewClient(cfg)
	res := c.Test(context.Background(), cfg)
	if !res.RegistryReachable {
		t.Fatalf("registry not reachable: %+v", res)
	}
	if !res.AuthSuccess {
		t.Fatalf("auth not success: %+v", res)
	}
	if !res.OK {
		t.Fatalf("expected ok: %+v", res)
	}
}

func TestRegistryUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("WWW-Authenticate", `Bearer realm="https://example/token"`)
	}))
	defer ts.Close()

	cfg := model.RegistryConfig{
		Registry: ts.URL,
		Username: "robot$project",
		Password: "wrong",
		Insecure: true,
	}
	c := NewClient(cfg)
	res := c.Test(context.Background(), cfg)
	if res.OK {
		t.Fatalf("expected failure: %+v", res)
	}
}

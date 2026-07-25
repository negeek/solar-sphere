package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/negeek/solar-sphere/solar-spectrum/httpapi"
)

func TestGatewayProxiesJSON(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/join/" {
			t.Errorf("upstream got path %q, want /auth/v1/join/", r.URL.Path)
		}
		httpapi.JsonResponse(w, true, http.StatusCreated, "Successfully joined", map[string]interface{}{"email": "alice@example.com"})
	}))
	defer upstream.Close()

	gw := NewGateway(map[string]string{"auth": upstream.URL})

	req := httptest.NewRequest(http.MethodPost, "/auth/v1/join/", nil)
	rec := httptest.NewRecorder()
	gw.Handle(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var resp httpapi.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success || resp.Message != "Successfully joined" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGatewayProxiesCSV(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=device-1.csv")
		w.Write([]byte("lux,device_id\n100,device-1\n"))
	}))
	defer upstream.Close()

	gw := NewGateway(map[string]string{"sentinel": upstream.URL})

	req := httptest.NewRequest(http.MethodGet, "/sentinel/v1/download/device-1", nil)
	rec := httptest.NewRecorder()
	gw.Handle(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if body := rec.Body.String(); body != "lux,device_id\n100,device-1\n" {
		t.Errorf("body = %q", body)
	}
}

func TestGatewayUnknownService(t *testing.T) {
	gw := NewGateway(map[string]string{"auth": "http://unused"})

	req := httptest.NewRequest(http.MethodGet, "/nonexistent/v1/whatever", nil)
	rec := httptest.NewRecorder()
	gw.Handle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGatewayUpstreamUnreachable(t *testing.T) {
	gw := NewGateway(map[string]string{"auth": "http://127.0.0.1:1"})

	req := httptest.NewRequest(http.MethodPost, "/auth/v1/join/", nil)
	rec := httptest.NewRecorder()
	gw.Handle(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}

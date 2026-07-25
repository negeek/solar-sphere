package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	service "github.com/negeek/solar-sphere/solar-auth/service/v1"
	"github.com/negeek/solar-sphere/solar-spectrum/accesskey"
	"github.com/negeek/solar-sphere/solar-spectrum/httpapi"
	"github.com/negeek/solar-sphere/solar-spectrum/shared"
)

// fakeRepo satisfies service.NewAuthService's repository dependency (an
// unexported interface, matched structurally) without a real database.
type fakeRepo struct {
	revoked map[string]bool
}

func newFakeRepo() *fakeRepo { return &fakeRepo{revoked: map[string]bool{}} }

func (f *fakeRepo) CreateUser(context.Context, *shared.User) error { return nil }

func (f *fakeRepo) RevokeKey(_ context.Context, key, _ string) error {
	f.revoked[key] = true
	return nil
}

func (f *fakeRepo) IsKeyRevoked(_ context.Context, key string) (bool, error) {
	return f.revoked[key], nil
}

func newTestHandler(t *testing.T) (*Handler, string) {
	t.Helper()
	verifyKey, signKey, err := accesskey.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}
	return NewHandler(service.NewAuthService(newFakeRepo(), signKey, verifyKey)), verifyKey
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder) httpapi.Response {
	t.Helper()
	var resp httpapi.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response body %q: %v", rec.Body.String(), err)
	}
	return resp
}

func TestJoinHandler(t *testing.T) {
	h, _ := newTestHandler(t)

	body, _ := json.Marshal(map[string]string{"email": "alice@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/auth/v1/join/", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Join(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	resp := decodeResponse(t, rec)
	if !resp.Success {
		t.Errorf("expected success=true, got message=%q", resp.Message)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected data shape: %#v", resp.Data)
	}
	if data["email"] != "alice@example.com" {
		t.Errorf("email = %v, want alice@example.com", data["email"])
	}
	if _, ok := data["access_key"].(string); !ok {
		t.Errorf("expected access_key string in response, got %#v", data["access_key"])
	}
}

func TestJoinHandlerInvalidBody(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/v1/join/", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()

	h.Join(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestJoinHandlerInvalidEmail(t *testing.T) {
	h, _ := newTestHandler(t)

	body, _ := json.Marshal(map[string]string{"email": "not-an-email"})
	req := httptest.NewRequest(http.MethodPost, "/auth/v1/join/", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Join(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestNewAccessKeyHandler(t *testing.T) {
	h, _ := newTestHandler(t)

	joinBody, _ := json.Marshal(map[string]string{"email": "bob@example.com"})
	joinReq := httptest.NewRequest(http.MethodPost, "/auth/v1/join/", bytes.NewReader(joinBody))
	joinRec := httptest.NewRecorder()
	h.Join(joinRec, joinReq)

	joinData := decodeResponse(t, joinRec).Data.(map[string]interface{})
	accessKey := joinData["access_key"].(string)

	rotateBody, _ := json.Marshal(map[string]string{"key": accessKey, "email": "bob@example.com"})
	rotateReq := httptest.NewRequest(http.MethodPost, "/auth/v1/new_key/", bytes.NewReader(rotateBody))
	rotateRec := httptest.NewRecorder()
	h.NewAccessKey(rotateRec, rotateReq)

	if rotateRec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rotateRec.Code, http.StatusCreated, rotateRec.Body.String())
	}

	rotateData := decodeResponse(t, rotateRec).Data.(map[string]interface{})
	if rotateData["access_key"] == accessKey {
		t.Error("expected rotation to return a different access key")
	}
}

func TestNewAccessKeyHandlerRejectsRevokedKey(t *testing.T) {
	h, _ := newTestHandler(t)

	joinBody, _ := json.Marshal(map[string]string{"email": "carol@example.com"})
	joinReq := httptest.NewRequest(http.MethodPost, "/auth/v1/join/", bytes.NewReader(joinBody))
	joinRec := httptest.NewRecorder()
	h.Join(joinRec, joinReq)
	accessKey := decodeResponse(t, joinRec).Data.(map[string]interface{})["access_key"].(string)

	rotateBody, _ := json.Marshal(map[string]string{"key": accessKey, "email": "carol@example.com"})

	// First rotation succeeds and revokes the original key.
	firstReq := httptest.NewRequest(http.MethodPost, "/auth/v1/new_key/", bytes.NewReader(rotateBody))
	firstRec := httptest.NewRecorder()
	h.NewAccessKey(firstRec, firstReq)
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("first rotation status = %d, want %d", firstRec.Code, http.StatusCreated)
	}

	// Reusing the now-revoked original key must be rejected.
	secondReq := httptest.NewRequest(http.MethodPost, "/auth/v1/new_key/", bytes.NewReader(rotateBody))
	secondRec := httptest.NewRecorder()
	h.NewAccessKey(secondRec, secondReq)
	if secondRec.Code != http.StatusBadRequest {
		t.Fatalf("reusing a revoked key: status = %d, want %d; body=%s", secondRec.Code, http.StatusBadRequest, secondRec.Body.String())
	}
}

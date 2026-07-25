package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"

	authmw "github.com/negeek/solar-sphere/solar-sentinel/middlewares/v1"
	service "github.com/negeek/solar-sphere/solar-sentinel/service/v1"
	"github.com/negeek/solar-sphere/solar-spectrum/accesskey"
	"github.com/negeek/solar-sphere/solar-spectrum/shared"
)

type fakeDeviceRepo struct {
	devices []shared.Device
}

func (f *fakeDeviceRepo) CreateDevice(_ context.Context, d *shared.Device) error {
	f.devices = append(f.devices, *d)
	return nil
}

type fakeIrradianceRepo struct {
	devices  map[string]shared.Device
	readings map[string][]shared.SolarIrradiance
}

func newFakeIrradianceRepo() *fakeIrradianceRepo {
	return &fakeIrradianceRepo{devices: map[string]shared.Device{}, readings: map[string][]shared.SolarIrradiance{}}
}

func (f *fakeIrradianceRepo) FindDeviceByID(_ context.Context, id string) (*shared.Device, error) {
	d, ok := f.devices[id]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	return &d, nil
}

func (f *fakeIrradianceRepo) CreateIrradianceReading(_ context.Context, reading *shared.SolarIrradiance) error {
	f.readings[reading.DeviceID] = append(f.readings[reading.DeviceID], *reading)
	return nil
}

func (f *fakeIrradianceRepo) FindIrradianceByDevice(_ context.Context, deviceID string) ([]shared.SolarIrradiance, error) {
	return f.readings[deviceID], nil
}

// fakeAccessChecker satisfies authmw.AccessChecker.
type fakeAccessChecker struct {
	revoked map[string]bool
	users   map[string]bool
}

func (f *fakeAccessChecker) IsKeyRevoked(_ context.Context, key string) (bool, error) {
	return f.revoked[key], nil
}

func (f *fakeAccessChecker) UserExists(_ context.Context, email string) (bool, error) {
	return f.users[email], nil
}

// testStack wires a full request pipeline (router + auth middleware +
// handler + services) against fakes, so handler tests exercise the same
// path a real request takes, including authentication and ownership checks.
type testStack struct {
	router     *mux.Router
	deviceRepo *fakeDeviceRepo
	irrRepo    *fakeIrradianceRepo
	signingKey string
}

func newTestStack(t *testing.T, knownUsers ...string) *testStack {
	t.Helper()

	verifyKey, signKey, err := accesskey.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}

	deviceRepo := &fakeDeviceRepo{}
	irrRepo := newFakeIrradianceRepo()
	handler := NewHandler(service.NewDeviceService(deviceRepo), service.NewIrradianceService(irrRepo))

	users := map[string]bool{}
	for _, email := range knownUsers {
		users[email] = true
	}
	checker := &fakeAccessChecker{revoked: map[string]bool{}, users: users}
	authMiddleware := authmw.Authentication(checker, verifyKey)

	router := mux.NewRouter()
	Routes(router, handler, authMiddleware)

	return &testStack{router: router, deviceRepo: deviceRepo, irrRepo: irrRepo, signingKey: signKey}
}

func (ts *testStack) authedRequest(t *testing.T, method, path, email string, body []byte) *http.Request {
	t.Helper()
	key, err := accesskey.Generate(email, ts.signingKey, accesskey.GenerateOptions{})
	if err != nil {
		t.Fatalf("generate access key: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+key)
	return req
}

func TestCreateDeviceHandler(t *testing.T) {
	ts := newTestStack(t, "alice@example.com")

	body, _ := json.Marshal(map[string]string{"name": "backyard sensor"})
	req := ts.authedRequest(t, http.MethodPost, "/sentinel/v1/device/", "alice@example.com", body)
	rec := httptest.NewRecorder()

	ts.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(ts.deviceRepo.devices) != 1 {
		t.Fatalf("expected 1 device created, got %d", len(ts.deviceRepo.devices))
	}
	if got := ts.deviceRepo.devices[0].Owner; got != "alice@example.com" {
		t.Errorf("Owner = %q, want alice@example.com", got)
	}
}

func TestCreateDeviceHandlerRequiresAuth(t *testing.T) {
	ts := newTestStack(t)

	body, _ := json.Marshal(map[string]string{"name": "backyard sensor"})
	req := httptest.NewRequest(http.MethodPost, "/sentinel/v1/device/", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	ts.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestDownloadHandlerRejectsNonOwner(t *testing.T) {
	ts := newTestStack(t, "alice@example.com")
	ts.irrRepo.devices["device-1"] = shared.Device{ID: "device-1", Owner: "bob@example.com"}

	req := ts.authedRequest(t, http.MethodGet, "/sentinel/v1/download/device-1", "alice@example.com", nil)
	rec := httptest.NewRecorder()

	ts.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestDownloadHandlerAllowsOwner(t *testing.T) {
	ts := newTestStack(t, "alice@example.com")
	ts.irrRepo.devices["device-1"] = shared.Device{ID: "device-1", Owner: "alice@example.com"}
	ts.irrRepo.readings["device-1"] = []shared.SolarIrradiance{
		{DeviceID: "device-1", Data: map[string]interface{}{"lux": 100}},
	}

	req := ts.authedRequest(t, http.MethodGet, "/sentinel/v1/download/device-1", "alice@example.com", nil)
	rec := httptest.NewRecorder()

	ts.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
}

func TestDownloadHandlerUnknownDevice(t *testing.T) {
	ts := newTestStack(t, "alice@example.com")

	req := ts.authedRequest(t, http.MethodGet, "/sentinel/v1/download/does-not-exist", "alice@example.com", nil)
	rec := httptest.NewRecorder()

	ts.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

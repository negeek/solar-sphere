package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/negeek/solar-sphere/solar-spectrum/shared"
)

type fakeDeviceRepo struct {
	devices   []shared.Device
	createErr error
}

func (f *fakeDeviceRepo) CreateDevice(_ context.Context, d *shared.Device) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.devices = append(f.devices, *d)
	return nil
}

func TestCreateDevice(t *testing.T) {
	repo := &fakeDeviceRepo{}
	svc := NewDeviceService(repo)

	device, err := svc.CreateDevice(context.Background(), CreateDeviceInput{Name: "backyard sensor", Owner: "alice@example.com"})
	if err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}
	if device.Owner != "alice@example.com" {
		t.Errorf("Owner = %q, want alice@example.com", device.Owner)
	}
	if device.ID == "" {
		t.Error("expected a generated device ID")
	}
	if len(repo.devices) != 1 {
		t.Fatalf("expected 1 device created, got %d", len(repo.devices))
	}
}

func TestCreateDeviceRequiresName(t *testing.T) {
	repo := &fakeDeviceRepo{}
	svc := NewDeviceService(repo)

	_, err := svc.CreateDevice(context.Background(), CreateDeviceInput{Name: "", Owner: "alice@example.com"})
	if !errors.Is(err, ErrInvalidDeviceName) {
		t.Fatalf("error = %v, want ErrInvalidDeviceName", err)
	}
	if len(repo.devices) != 0 {
		t.Error("expected no device created for an empty name")
	}
}

func TestCreateDeviceRepositoryError(t *testing.T) {
	repo := &fakeDeviceRepo{createErr: errors.New("boom")}
	svc := NewDeviceService(repo)

	if _, err := svc.CreateDevice(context.Background(), CreateDeviceInput{Name: "x", Owner: "alice@example.com"}); err == nil {
		t.Fatal("expected an error when the repository fails")
	}
}

// A user is allowed to register more than one device — there's no admin
// gate and no "one device per user" limit.
func TestCreateDeviceAllowsMultiplePerOwner(t *testing.T) {
	repo := &fakeDeviceRepo{}
	svc := NewDeviceService(repo)

	for i := 0; i < 3; i++ {
		if _, err := svc.CreateDevice(context.Background(), CreateDeviceInput{Name: "sensor", Owner: "alice@example.com"}); err != nil {
			t.Fatalf("CreateDevice #%d: %v", i, err)
		}
	}
	if len(repo.devices) != 3 {
		t.Fatalf("expected 3 devices for the same owner, got %d", len(repo.devices))
	}
}

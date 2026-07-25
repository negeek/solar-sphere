package v1

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/negeek/solar-sphere/solar-spectrum/shared"
	"go.mongodb.org/mongo-driver/mongo"
)

type fakeIrradianceRepo struct {
	devices  map[string]shared.Device
	readings map[string][]shared.SolarIrradiance
}

func newFakeIrradianceRepo() *fakeIrradianceRepo {
	return &fakeIrradianceRepo{
		devices:  map[string]shared.Device{},
		readings: map[string][]shared.SolarIrradiance{},
	}
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

func TestSaveRejectsUnknownDevice(t *testing.T) {
	repo := newFakeIrradianceRepo()
	svc := NewIrradianceService(repo)

	err := svc.Save(context.Background(), "unknown-device", map[string]interface{}{"lux": 100})
	if !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("error = %v, want ErrDeviceNotFound", err)
	}
}

func TestSaveStoresReadingForKnownDevice(t *testing.T) {
	repo := newFakeIrradianceRepo()
	repo.devices["device-1"] = shared.Device{ID: "device-1", Owner: "alice@example.com"}
	svc := NewIrradianceService(repo)

	if err := svc.Save(context.Background(), "device-1", map[string]interface{}{"lux": 100}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if len(repo.readings["device-1"]) != 1 {
		t.Fatalf("expected 1 reading stored, got %d", len(repo.readings["device-1"]))
	}
}

func TestExportRequiresOwnership(t *testing.T) {
	repo := newFakeIrradianceRepo()
	repo.devices["device-1"] = shared.Device{ID: "device-1", Owner: "alice@example.com"}
	svc := NewIrradianceService(repo)

	if _, err := svc.Export(context.Background(), "device-1", "mallory@example.com"); !errors.Is(err, ErrNotDeviceOwner) {
		t.Fatalf("error = %v, want ErrNotDeviceOwner", err)
	}
}

func TestExportUnknownDevice(t *testing.T) {
	repo := newFakeIrradianceRepo()
	svc := NewIrradianceService(repo)

	if _, err := svc.Export(context.Background(), "unknown-device", "alice@example.com"); !errors.Is(err, ErrDeviceNotFound) {
		t.Fatalf("error = %v, want ErrDeviceNotFound", err)
	}
}

// Regression test for a real bug found in the original handler: the CSV
// header included a trailing column with no corresponding value in each
// row, and the sensor-field columns were ordered by Go's randomized map
// iteration, so header and row columns could both miscount and misalign.
func TestExportReturnsDeterministicCSVRows(t *testing.T) {
	repo := newFakeIrradianceRepo()
	repo.devices["device-1"] = shared.Device{ID: "device-1", Owner: "alice@example.com"}
	repo.readings["device-1"] = []shared.SolarIrradiance{
		{
			DeviceID:    "device-1",
			Data:        map[string]interface{}{"lux": 100, "volts": 3.3},
			DateUpdated: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			DeviceID:    "device-1",
			Data:        map[string]interface{}{"lux": 120, "volts": 3.4},
			DateUpdated: time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC),
		},
	}
	svc := NewIrradianceService(repo)

	rows, err := svc.Export(context.Background(), "device-1", "alice@example.com")
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (header + 2 readings), got %d", len(rows))
	}
	for i, row := range rows {
		if len(row) != len(rows[0]) {
			t.Errorf("row %d has %d columns, header has %d", i, len(row), len(rows[0]))
		}
	}

	wantHeader := []string{"lux", "volts", "device_id", "date_updated"}
	if !reflect.DeepEqual(rows[0], wantHeader) {
		t.Errorf("header = %v, want %v", rows[0], wantHeader)
	}
	wantFirstRow := []string{"100", "3.3", "device-1", "2026-01-01T00:00:00Z"}
	if !reflect.DeepEqual(rows[1], wantFirstRow) {
		t.Errorf("row 1 = %v, want %v", rows[1], wantFirstRow)
	}
}

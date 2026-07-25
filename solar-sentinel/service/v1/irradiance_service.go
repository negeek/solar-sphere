package v1

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/negeek/solar-sphere/solar-spectrum/shared"
	"go.mongodb.org/mongo-driver/mongo"
)

// irradianceRepository is the subset of the repository IrradianceService
// needs, so it can be tested against a hand-written fake instead of a real
// database.
type irradianceRepository interface {
	FindDeviceByID(ctx context.Context, id string) (*shared.Device, error)
	CreateIrradianceReading(ctx context.Context, reading *shared.SolarIrradiance) error
	FindIrradianceByDevice(ctx context.Context, deviceID string) ([]shared.SolarIrradiance, error)
}

type IrradianceService struct {
	repo irradianceRepository
}

func NewIrradianceService(r irradianceRepository) *IrradianceService {
	return &IrradianceService{repo: r}
}

// Save records a reading published by a device over MQTT. MQTT publishes
// aren't authenticated (there's no per-device broker ACL wired up), so this
// at least checks the device was actually registered before data gets
// stored under its ID.
func (s *IrradianceService) Save(ctx context.Context, deviceID string, data map[string]interface{}) error {
	if _, err := s.repo.FindDeviceByID(ctx, deviceID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("%w: %s", ErrDeviceNotFound, deviceID)
		}
		return err
	}

	reading := &shared.SolarIrradiance{DeviceID: deviceID, Data: data}
	return s.repo.CreateIrradianceReading(ctx, reading)
}

// Export returns a device's readings as CSV rows (row 0 is the header),
// after checking requesterEmail actually owns the device.
func (s *IrradianceService) Export(ctx context.Context, deviceID, requesterEmail string) ([][]string, error) {
	device, err := s.repo.FindDeviceByID(ctx, deviceID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDeviceNotFound
		}
		return nil, err
	}
	if device.Owner != requesterEmail {
		return nil, ErrNotDeviceOwner
	}

	readings, err := s.repo.FindIrradianceByDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	return toCSVRows(readings), nil
}

// toCSVRows lays out readings as a fixed-column CSV: one column per sensor
// field seen across all readings (sorted, so column order is deterministic
// instead of depending on Go's randomized map iteration order), followed by
// device_id and date_updated.
func toCSVRows(readings []shared.SolarIrradiance) [][]string {
	dataKeys := collectDataKeys(readings)

	headers := make([]string, 0, len(dataKeys)+2)
	headers = append(headers, dataKeys...)
	headers = append(headers, "device_id", "date_updated")

	rows := make([][]string, 0, len(readings)+1)
	rows = append(rows, headers)

	for _, reading := range readings {
		row := make([]string, 0, len(headers))
		for _, key := range dataKeys {
			row = append(row, fmt.Sprintf("%v", reading.Data[key]))
		}
		row = append(row, reading.DeviceID, reading.DateUpdated.Format(time.RFC3339))
		rows = append(rows, row)
	}
	return rows
}

func collectDataKeys(readings []shared.SolarIrradiance) []string {
	keySet := make(map[string]struct{})
	for _, reading := range readings {
		for key := range reading.Data {
			keySet[key] = struct{}{}
		}
	}

	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

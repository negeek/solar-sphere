// Package v1 is solar-sentinel's service layer: it owns validation and
// business rules for device registration and irradiance data. Handlers only
// decode requests and call into this package; this package is the only
// caller of the repository.
package v1

import (
	"context"

	repo "github.com/negeek/solar-sphere/solar-sentinel/repository/v1"
	"github.com/negeek/solar-sphere/solar-spectrum/idgen"
	"github.com/negeek/solar-sphere/solar-spectrum/shared"
)

type DeviceService struct {
	repo *repo.Repository
}

func NewDeviceService(r *repo.Repository) *DeviceService {
	return &DeviceService{repo: r}
}

type CreateDeviceInput struct {
	Name string
	// Owner is the authenticated caller's email, taken from their verified
	// access key — never from the request body, so a device is always
	// created under whoever's key created it. A user may register any
	// number of devices this way; there is no separate admin role.
	Owner string
}

func (s *DeviceService) CreateDevice(ctx context.Context, in CreateDeviceInput) (*shared.Device, error) {
	if in.Name == "" {
		return nil, ErrInvalidDeviceName
	}

	device := &shared.Device{
		ID:    idgen.New(""),
		Name:  in.Name,
		Owner: in.Owner,
	}
	if err := s.repo.CreateDevice(ctx, device); err != nil {
		return nil, err
	}
	return device, nil
}

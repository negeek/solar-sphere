package v1

import "errors"

var (
	ErrInvalidDeviceName = errors.New("device name is required")
	ErrDeviceNotFound    = errors.New("device not found")
	ErrNotDeviceOwner    = errors.New("you do not own this device")
)

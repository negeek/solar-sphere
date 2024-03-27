package v1

import (
	"errors"
	model "github.com/negeek/solar-sphere/solar-sentinel/repository/v1"
)

func SaveSolarIrrdianceData(deviceID string, data map[string]interface{}) error {
	irr := &model.SolarIrradiance{
		DeviceID: deviceID,
		Data:     data,
	}

	// Create data
	if err := irr.Create(); err != nil {
		return errors.New("Error saving data")
	}

	return nil
}

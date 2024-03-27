package v1

import(
	"errors"
	model"github.com/negeek/solar-sphere/solar-sentinel/repository/v1"
)


func SaveSolarIrrdianceData(device_id string, data map[string]interface) error {
	var (
		irr = &model.SolarIrradiance{}
		err error
	)
	irr.DeviceID = device_id
	irr.Data = data

	// Create data
	err=irr.Create()
	if err != nil{
		return	errors.New("Error saving data")
	}

	return nil
}
package v1

import (
	"context"
	"github.com/negeek/solar-sphere/solar-sentinel/utils"
	"github.com/negeek/solar-sphere/solar-sentinel/db"
)

var (
	DEVICE_COLLECTION string = "devices"
	IRR_COLLECTION string = "solar-irradiance"
)

func (d *Device) Create() error {
	collection := db.MongoDB.Collection(DEVICE_COLLECTION)
	utils.Time(d, true)
	_, err := collection.InsertOne(context.Background(), d)
	if err != nil {
		return err
	}
	return nil
}

func (s *SolarIrradiance) Create() error {
	collection := db.MongoDB.Collection(IRR_COLLECTION)
	utils.Time(s, true)
	_, err := collection.InsertOne(context.Background(), s)
	if err != nil {
		return err
	}
	return nil

}

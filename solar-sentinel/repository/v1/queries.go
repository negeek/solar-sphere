package v1

import (
	"context"
	"github.com/negeek/solar-sphere/solar-sentinel/utils"
	"github.com/negeek/solar-sphere/solar-sentinel/db"
	"go.mongodb.org/mongo-driver/bson"
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

func (d *Device) GetAllSolarData() ([]SolarIrradiance, error){
	var data []SolarIrradiance
	collection := db.MongoDB.Collection(IRR_COLLECTION)
	cursor, err := collection.Find(context.Background(), bson.D{{"device_id", d.ID}})
	if err != nil {
		return data, err
	}

	if err = cursor.All(context.Background(), &data); err != nil {
		return data, err
	}
	return data, nil
} 
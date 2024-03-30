package v1

import (
	"context"
	"github.com/negeek/solar-sphere/solar-sentinel/utils"
	"github.com/negeek/solar-sphere/solar-sentinel/db"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/negeek/solar-sphere/solar-spectrum/consts"
	"github.com/negeek/solar-sphere/solar-spectrum/types"
)

const (
	DEVICE_COLLECTION = "devices"
	IRR_COLLECTION = "solar-irradiance"
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

func (u *types.User) FindUser()bool{
	collection := db.MongoDB.Collection(consts.USER_COLLECTION)
	var result bson.M
	err := collection.FindOne(context..Background(),bson.D{{"_id", id}}).Decode(&result)
	if err != nil {
		return false
	}
	return true
}
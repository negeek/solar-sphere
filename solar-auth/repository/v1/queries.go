package v1

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/negeek/solar-sphere/solar-auth/utils"
	"github.com/negeek/solar-sphere/solar-auth/db"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	USER_COLLECTION string = "users"
	KEY_COLLECTION string = "keys"
	DEVICE_COLLECTION string = "devices"
)

func (u *User) Create() error {
	collection := db.MongoDB.Collection(USER_COLLECTION)
	utils.Time(u, true)
	_, err := collection.InsertOne(context.Background(), u)
	if err != nil {
		return err
	}
	return nil
}

func SaveDeviceID(device_id string)bool{
	collection := db.MongoDB.Collection(DEVICE_COLLECTION)
	// device ids with name of custom is for those that have their own solar-irradiance meter
	_, err := collection.InsertOne(context.Background(), bson.D{{"_id":device_id, "name":"custom"}})
	if err != nil {
		return false
	}
	return true
}

func (u *User) Delete() error {
	collection := db.MongoDB.Collection(USER_COLLECTION)
	_, err := collection.DeleteOne(context.Background(), bson.D{{"_id",u.ID}})
	if err != nil {
		return err
	}
	return nil
}

func (k *RevokedKey) Revoke() error {
	collection := db.MongoDB.Collection(KEY_COLLECTION)
	_, err := collection.InsertOne(context.Background(), k)
	if err != nil {
		return err
	}
	return nil
}
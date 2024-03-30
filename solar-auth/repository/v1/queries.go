package v1

import (
	"time"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/negeek/solar-sphere/solar-auth/utils"
	"github.com/negeek/solar-sphere/solar-auth/db"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/negeek/solar-sphere/solar-spectrum/consts"
)


const (
	USER_COLLECTION = "users"
	KEY_COLLECTION="keys"
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
	collection := db.MongoDB.Collection(consts.DEVICE_COLLECTION)
	var device = shared.Device{device_id, "custom", time.Now().UTC(), time.Now().UTC()}
	_, err := collection.InsertOne(context.Background(), device)
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
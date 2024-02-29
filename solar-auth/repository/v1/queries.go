package v1

import (
	"context"
	"github.com/negeek/solar-sphere/solar-auth/utils"
	db "github.com/negeek/solar-sphere/solar-auth/db/v1"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	USER_COLLECTION string = "users"
	KEY_COLLECTION string = "keys"
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

func (u *User) Delete() error {
	collection := db.MongoDB.Collection(USER_COLLECTION)
	_, err := collection.DeleteOne(context.Background(), bson.D{{"_id",u.ID}})
	if err != nil {
		return err
	}
	return nil
}

func (k *AccessKey) RevokeStatus() error {
	collection := db.MongoDB.Collection(KEY_COLLECTION)
	err := collection.FindOne(context.Background(), bson.D{{"key", k.Key}}).Decode(k)
	if err != nil {
		return err
	}
	return nil
}

func (k *AccessKey) Revoke() error {
	k.Revoked = true
	collection := db.MongoDB.Collection(KEY_COLLECTION)
	_, err := collection.InsertOne(context.Background(), k)
	if err != nil {
		return err
	}
	return nil
}
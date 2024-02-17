package v1

import (
	"context"
	"github.com/negeek/solar-sphere/solar-auth/utils"
	"github.com/negeek/solar-sphere/solar-auth/db"
	"go.mongodb.org/mongo-driver/bson"
)

func (u *User) Create() error {
	collection := db.Client.Database("solar-sphere-db").Collection("users")
	utils.Time(u, true)
	_, err := collection.InsertOne(context.TODO(), u)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) Delete() error {
	collection := db.Client.Database("solar-sphere-db").Collection("users")
	_, err := collection.DeleteOne(context.TODO(), bson.D{{"email",u.Email}})
	if err != nil {
		return err
	}
	return nil
}
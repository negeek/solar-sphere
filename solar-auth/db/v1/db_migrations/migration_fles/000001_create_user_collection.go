package main

import (
	"log"
	"context"
	"github.com/negeek/solar-sphere/solar-auth/utils"
	db "github.com/negeek/solar-sphere/solar-auth/db/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
var (
	userSchema primitive.M = bson.M{
									"$jsonSchema":{
										"bsonType": "object",
										"required": []string{"_id", "email"},
										"properties": bson.M{
											"_id": bson.M{
												"bsonType": "string",
												"description": "id is required and must be a string",
											},
											"email": bson.M{
												"bsonType": "string",
												"description": "email is required and must be a string",
											},
										},
									}
								}

	userIndexModel = mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	}
	userOptions = &options.CreateCollectionOptions{}
	err error
)

func MakeMigration(){
	userOptions.SetValidator(userSchema)
	err = db.MongoDB.CreateCollection(ctx, USER_COLLECTION, userOptions)
	if err != nil {
		log.Fatal(err)
	}
	
	// Create unique index on email field for users collection
	_, err = db.MongoDB.Collection(USER_COLLECTION).Indexes().CreateOne(ctx, userIndexModel)
	if err != nil {
		log.Fatal(err)
	}
}
package migrations

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
	keySchema primitive.M = bson.M{
									"$jsonSchema":{
										"bsonType": "object",
										"required": []string{"key", "email"},
										"properties": bson.M{
											"key": bson.M{
												"bsonType": "string",
												"description": "key is required and must be a string",
											},
											"email": bson.M{
												"bsonType": "string",
												"description": "email is required and must be a string",
											},
										},
									}
								}
	

	keyOptions = &options.CreateCollectionOptions{}
	keyIndexModel = mongo.IndexModel{
		Keys:    bson.M{"key": 1}, 
		Options: options.Index().SetUnique(true),
	}
	err error
)

func MakeMigration(){
	keyOptions.SetValidator(keySchema)
	err = db.MongoDB.CreateCollection(ctx, KEY_COLLECTION, keyOptions)
	if err != nil {
		log.Fatal(err)
	}
	
	// Create unique index on email field for users collection
	_, err = db.MongoDB.Collection(KEY_COLLECTION).Indexes().CreateOne(ctx, keyIndexModel)
	if err != nil {
		log.Fatal(err)
	}
}
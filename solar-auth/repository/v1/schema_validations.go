package v1

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	UserSchema, KeySchema primitive.M
)

UserSchema = bson.M{
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
}}

KeySchema =  bson.M{
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
}}

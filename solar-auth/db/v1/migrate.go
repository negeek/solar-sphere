// Command migrate applies solar-auth's MongoDB schema migrations.
package main

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/negeek/solar-sphere/solar-spectrum/env"
	"github.com/negeek/solar-sphere/solar-spectrum/logging"
	"github.com/negeek/solar-sphere/solar-spectrum/migrate"
	"github.com/negeek/solar-sphere/solar-spectrum/mongoutil"
	"github.com/negeek/solar-sphere/solar-spectrum/shared"
)

var migrations = []migrate.Migration{
	{
		Name: "000001_create_user_collection",
		Up: func(ctx context.Context, db *mongo.Database) error {
			validator := bson.M{
				"$jsonSchema": bson.M{
					"bsonType": "object",
					"required": []string{"_id", "email"},
					"properties": bson.M{
						"_id":   bson.M{"bsonType": "string", "description": "id is required and must be a string"},
						"email": bson.M{"bsonType": "string", "description": "email is required and must be a string"},
					},
				},
			}
			if err := db.CreateCollection(ctx, shared.USER_COLLECTION, options.CreateCollection().SetValidator(validator)); err != nil {
				return err
			}
			_, err := db.Collection(shared.USER_COLLECTION).Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			})
			return err
		},
	},
	{
		Name: "000002_create_key_collection",
		Up: func(ctx context.Context, db *mongo.Database) error {
			validator := bson.M{
				"$jsonSchema": bson.M{
					"bsonType": "object",
					"required": []string{"key", "email"},
					"properties": bson.M{
						"key":   bson.M{"bsonType": "string", "description": "key is required and must be a string"},
						"email": bson.M{"bsonType": "string", "description": "email is required and must be a string"},
					},
				},
			}
			if err := db.CreateCollection(ctx, shared.REVOKED_KEY_COLLECTION, options.CreateCollection().SetValidator(validator)); err != nil {
				return err
			}
			_, err := db.Collection(shared.REVOKED_KEY_COLLECTION).Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys:    bson.D{{Key: "key", Value: 1}},
				Options: options.Index().SetUnique(true),
			})
			return err
		},
	},
}

func main() {
	if os.Getenv("APP_ENV") == "dev" {
		if err := env.Load(".env"); err != nil {
			panic("solar-auth migrate: loading .env: " + err.Error())
		}
	}

	log := logging.New("solar-auth-migrate")
	ctx := context.Background()

	client, db, err := mongoutil.Connect(ctx, os.Getenv("DATABASE_URL"), os.Getenv("DB_NAME"))
	if err != nil {
		log.Error("connect to mongo", "error", err)
		os.Exit(1)
	}
	defer mongoutil.Disconnect(context.Background(), client, 15*time.Second)

	if err := migrate.Run(ctx, db, migrations); err != nil {
		log.Error("migration failed", "error", err)
		os.Exit(1)
	}
	log.Info("migrations up to date")
}

// Command migrate applies solar-sentinel's MongoDB schema migrations.
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
		Name: "000001_create_device_collection",
		Up: func(ctx context.Context, db *mongo.Database) error {
			validator := bson.M{
				"$jsonSchema": bson.M{
					"bsonType": "object",
					"required": []string{"_id", "name", "owner"},
					"properties": bson.M{
						"_id":   bson.M{"bsonType": "string", "description": "id is required and must be a string"},
						"name":  bson.M{"bsonType": "string", "description": "name is required and must be a string"},
						"owner": bson.M{"bsonType": "string", "description": "owner is required and must be a string"},
					},
				},
			}
			if err := db.CreateCollection(ctx, shared.DEVICE_COLLECTION, options.CreateCollection().SetValidator(validator)); err != nil {
				return err
			}
			// Not unique: a user can own more than one device.
			_, err := db.Collection(shared.DEVICE_COLLECTION).Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: bson.D{{Key: "owner", Value: 1}},
			})
			return err
		},
	},
	{
		Name: "000002_create_solar_irradiance_collection",
		Up: func(ctx context.Context, db *mongo.Database) error {
			if err := db.CreateCollection(ctx, shared.IRR_COLLECTION); err != nil {
				return err
			}
			// Every read of this collection filters by device_id
			// (downloads, MQTT ingestion's device check) — without this
			// index those become full collection scans as data grows.
			_, err := db.Collection(shared.IRR_COLLECTION).Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: bson.D{{Key: "device_id", Value: 1}},
			})
			return err
		},
	},
}

func main() {
	if os.Getenv("APP_ENV") == "dev" {
		if err := env.Load(".env"); err != nil {
			panic("solar-sentinel migrate: loading .env: " + err.Error())
		}
	}

	log := logging.New("solar-sentinel-migrate")
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

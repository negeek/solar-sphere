package db

import (
	"context"
	"log"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
	MongoDB *mongo.Database
)

func Connect(connString string, dbName string)(context.Context, context.CancelFunc, error) {
	var err error
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(connString).SetServerAPIOptions(serverAPI))

	if err != nil {
		return ctx, cancel, err
	}
	
	// Send a ping to confirm a successful connection
	var result bson.M
	if err := Client.Database("admin").RunCommand(ctx, bson.D{{"ping", 1}}).Decode(&result); err != nil {
		return ctx, cancel, err
	}
	log.Println("Successfully connected to db")

	//db
	MongoDB = Client.Database(dbName)

	return ctx, cancel, nil
}

func Disconnect(ctx context.Context, cancel context.CancelFunc){
	defer cancel()
	log.Println("Disconnecting db")
	if err := Client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

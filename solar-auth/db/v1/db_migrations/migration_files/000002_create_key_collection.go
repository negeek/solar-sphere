package main

import (
	"os"
	"time"
	"context"
	"log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/joho/godotenv"
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

	//db
	MongoDB = Client.Database(dbName)

	return ctx, cancel, nil
}

func Disconnect(ctx context.Context, cancel context.CancelFunc){
	defer cancel()
	if err := Client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

func GetEnv(){
	appEnv:=os.Getenv("APP_ENV")
	if appEnv=="dev"{
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

// The above is perequisite for your migration functionality.
// Write your logic below

var (
	keySchema primitive.M = bson.M{
									"$jsonSchema":bson.M{
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
									},
								}
	

	keyOptions = &options.CreateCollectionOptions{}
	keyIndexModel = mongo.IndexModel{
		Keys:    bson.M{"key": 1}, 
		Options: options.Index().SetUnique(true),
	}
	err error
	KEY_COLLECTION string = "keys"
	Client *mongo.Client
	MongoDB *mongo.Database
)


func MakeMigration(){
	// Optional
	GetEnv()

	// Connect to DB
	dbctx, dbcancel, err:= Connect(os.Getenv("DATABASE_URL"), os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatal(err)
	}
	
	// Create Key collection
	keyOptions.SetValidator(keySchema)
	err = MongoDB.CreateCollection(context.Background(), KEY_COLLECTION, keyOptions)
	if err != nil {
		log.Fatal(err)
	}
	
	// Create unique index for key collection
	_, err = MongoDB.Collection(KEY_COLLECTION).Indexes().CreateOne(context.Background(), keyIndexModel)
	if err != nil {
		log.Fatal(err)
	}

	// Disconnect DB
	defer Disconnect(dbctx,dbcancel)
}
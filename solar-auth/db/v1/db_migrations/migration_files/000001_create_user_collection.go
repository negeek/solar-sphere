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
	userSchema primitive.M = bson.M{
									"$jsonSchema":bson.M{
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
									},
								}

	userIndexModel = mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	}
	userOptions = &options.CreateCollectionOptions{}
	err error
	USER_COLLECTION string = "users"
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
	
	// Create user collection
	userOptions.SetValidator(userSchema)
	err = MongoDB.CreateCollection(context.Background(), USER_COLLECTION, userOptions)
	if err != nil {
		log.Fatal(err)
	}
	
	// Create unique index on email field for users collection
	_, err = MongoDB.Collection(USER_COLLECTION).Indexes().CreateOne(context.Background(), userIndexModel)
	if err != nil {
		log.Fatal(err)
	}

	// End connection
	defer Disconnect(dbctx,dbcancel)
}
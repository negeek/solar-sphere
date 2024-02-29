package v1

import (
	"log"
	"context"
	"github.com/negeek/solar-sphere/solar-auth/utils"
	"github.com/negeek/solar-sphere/solar-auth/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (

	DB string = "solar-sphere-db"
	USER_COLLECTION string = "users"
	KEY_COLLECTION string = "keys"
	err error
	userOptions = &options.CreateCollectionOptions{} 
	keyOptions = &options.CreateCollectionOptions{}

)
userIndexModel := mongo.IndexModel{
	Keys:    bson.M{"email": 1},
	Options: options.Index().SetUnique(true),
}

keyIndexModel := mongo.IndexModel{
	Keys:    bson.M{"key": 1}, 
	Options: options.Index().SetUnique(true),
}


userOptions = options.CreateCollection().SetValidator(UserSchema)
err = db.Client.Database(DB).CreateCollection(context.Background(), USER_COLLECTION, userOptions)
if err != nil {
	log.Fatal(err)
}
_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		fmt.Println("Error creating index:", err)
		return
	}

keyOptions = options.CreateCollection().SetValidator(KeySchema)
err = db.Client.Database(DB).CreateCollection(context.Background(), KEY_COLLECTION, keyOptions)
if err != nil {
	log.Fatal(err)
}

_, err =db.Client.Database(DB).Collection(USER_COLLECTION).Indexes().CreateOne(ctx, userIndexModel)
if err != nil {
	fmt.Println("Error creating user index:", err)
	return
}

_, err =db.Client.Database(DB).Collection(KEY_COLLECTION).Indexes().CreateOne(ctx, keyIndexModel)
if err != nil {
	fmt.Println("Error creating key index:", err)
	return
}






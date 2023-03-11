package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetClientOptions() *options.ClientOptions {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://doadmin:g3k615i2p89A7IwD@henar-db-0d7d8f8e.mongo.ondigitalocean.com/?retryWrites=true&w=majority"). // do not use credentials like this, use process env!!!!
		SetServerAPIOptions(serverAPIOptions)

	return clientOptions
}

func GetCollection(collection string) *mongo.Collection {
	clientOptions := GetClientOptions()

	client, _ := mongo.Connect(context.TODO(), clientOptions)

	return client.Database("henar").Collection(collection)
}

package db

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetClientOptions() *options.ClientOptions {
	dburi := os.Getenv("mongodb+srv://doadmin:g3k615i2p89A7IwD@henar-db-0d7d8f8e.mongo.ondigitalocean.com/?retryWrites=true&w=majority")

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(dburi).
		SetServerAPIOptions(serverAPIOptions)

	return clientOptions
}

func GetCollection(collection string) *mongo.Collection {
	clientOptions := GetClientOptions()

	client, _ := mongo.Connect(context.TODO(), clientOptions)

	return client.Database("henar").Collection(collection)
}

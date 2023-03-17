package db

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetClientOptions() *options.ClientOptions {
	dburi := os.Getenv("DBURI")

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

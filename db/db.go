package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func GetClientOptions() *options.ClientOptions {
	dburi := "mongodb+srv://doadmin:g3k615i2p89A7IwD@henar-db-0d7d8f8e.mongo.ondigitalocean.com/?retryWrites=true&w=majority"
	// dburi := os.Getenv("DB_URI")

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(dburi).
		SetServerAPIOptions(serverAPIOptions)

	return clientOptions
}

func GetCollection(collection string) (*mongo.Collection, error) {
	client := GetMongoClient()

	return client.Database("henar").Collection(collection), nil
}

func createIndex(collection *mongo.Collection) {
	indexModel := []mongo.IndexModel{
		{
			Keys: bson.M{
				"slug": 1,
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{
				"tags": 1,
			},
		},
		{
			Keys: bson.M{
				"title": 1,
			},
		},
		{
			Keys: bson.M{
				"location": 1,
			},
		},
	}

	indexName, err := collection.Indexes().CreateMany(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created index %s on collection %s", indexName, collection.Name())
}

func InitDb() {
	clientOptions := GetClientOptions()

	newClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	} else {
		client = newClient

		// Add indexes
		researches, _ := GetCollection("researches")
		createIndex(researches)
		projects, _ := GetCollection("projects")
		createIndex(projects)
		events, _ := GetCollection("events")
		createIndex(events)
	}
}

func GetMongoClient() mongo.Client {
	return *client
}

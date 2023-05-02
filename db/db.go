package db

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func GetClientOptions() *options.ClientOptions {
	dburi := "mongodb+srv://doadmin:Y6krY4thlZAM7jeP@cluster0.fz184bf.mongodb.net/test"
	// dburi := "mongodb+srv://doadmin:g3k615i2p89A7IwD@henar-db-0d7d8f8e.mongo.ondigitalocean.com/?retryWrites=true&w=majority"

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(dburi).
		SetServerAPIOptions(serverAPIOptions)

	return clientOptions
}

func GetCollection(collection string) (*mongo.Collection, error) {
	client := GetMongoClient()

	return client.Database("test").Collection(collection), nil
}

type Index struct {
	Keys    string
	Options *options.IndexOptions
}

type Indexes []Index

func createIndex(coll *mongo.Collection, indexes Indexes) {
	for _, idx := range indexes {
		indexName, err := coll.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
			Keys:    bson.M{idx.Keys: 1},
			Options: idx.Options,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(indexName)
	}
}

func InitDb() {
	clientOptions := GetClientOptions()

	newClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	} else {
		client = newClient

		indexes := Indexes{
			{
				Keys:    "slug",
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: "tags",
			},
			{
				Keys: "title",
			},

			{
				Keys: "location",
			},
		}
		usersIndexes := Indexes{
			{
				Keys: "full_name",
			},
			{
				Keys: "job",
			},
		}

		// Add indexes
		researches, _ := GetCollection("researches")
		createIndex(researches, indexes)
		projects, _ := GetCollection("projects")
		createIndex(projects, indexes)
		events, _ := GetCollection("events")
		createIndex(events, indexes)
		users, _ := GetCollection("users")
		createIndex(users, usersIndexes)
	}
}

func GetMongoClient() mongo.Client {
	return *client
}

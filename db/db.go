package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func GetClientOptions() *options.ClientOptions {
	dburi := os.Getenv("DB_URI")

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

func initIndexes() {
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
	}
	usersIndexes := Indexes{
		{
			Keys: "full_name",
		},
		{
			Keys: "job",
		},
	}
	researchIndexes := Indexes{
		{
			Keys: "title",
		},
	}
	verificationIndexes := Indexes{
		{
			Keys:    "email",
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: "code",
		},
	}
	eventsIndexes := append(indexes, Indexes{
		{
			Keys: "location",
		},
	}...)

	projectIndexes := append(indexes, Indexes{
		{
			Keys: "project_status",
		},
		{
			Keys: "how_to_help_the_project",
		},
		{
			Keys: "location",
		},
	}...)

	// Add indexes
	researches, _ := GetCollection("researches")
	createIndex(researches, researchIndexes)
	projects, _ := GetCollection("projects")
	createIndex(projects, projectIndexes)
	events, _ := GetCollection("events")
	createIndex(events, eventsIndexes)
	users, _ := GetCollection("users")
	createIndex(users, usersIndexes)
	verificationDataCollection, _ := GetCollection("verificationdata")
	createIndex(verificationDataCollection, verificationIndexes)
}

func InitDb() {
	clientOptions := GetClientOptions()

	newClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	} else {
		client = newClient
		// initIndexes()
	}
}

func GetMongoClient() mongo.Client {
	return *client
}

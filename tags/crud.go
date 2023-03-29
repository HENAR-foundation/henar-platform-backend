package tags

import (
	"context"
	"encoding/json"
	"henar-backend/db"
	"henar-backend/types"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTags(w http.ResponseWriter, r *http.Request) {
	// Set a timeout for the context
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Get the MongoDB collection for projects
	collection, err := db.GetCollection("tags")
	if err != nil {
		http.Error(w, "Error getting collection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	filter := bson.M{}

	// Query the database and get the cursor
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Error finding projects", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Get the results from the cursor
	var results []types.Tag
	if err = cursor.All(ctx, &results); err != nil {
		panic(err)
	}

	// Marshal the project struct to JSON format
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response headers and write the response body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetTag(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Get the MongoDB collection for projects
	collection, err := db.GetCollection("tags")
	if err != nil {
		http.Error(w, "Error getting collection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the project ID from the URL path parameter
	vars := mux.Vars(r)
	id := vars["tagId"]
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	var result types.Tag

	// Find the tag by ID
	err = collection.FindOne(
		ctx,
		filter,
	).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Tag not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error getting tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the project struct to JSON format
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response headers and write the response body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

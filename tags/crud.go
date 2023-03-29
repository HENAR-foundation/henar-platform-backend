package tags

import (
	"context"
	"encoding/json"
	"henar-backend/db"
	"henar-backend/types"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/go-playground/validator.v9"
)

func GetTags(w http.ResponseWriter, r *http.Request) {
	collection, _ := db.GetCollection("tags")

	filter := bson.M{}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Error finding tags", http.StatusInternalServerError)
		return
	}

	// Get the results from the cursor
	var results []types.Tag
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	// Marshal the tag struct to JSON format
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
	collection, _ := db.GetCollection("tags")

	// Get the tag ID from the URL path parameter
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
		context.TODO(),
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

	// Marshal the tag struct to JSON format
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

func CreateTag(w http.ResponseWriter, r *http.Request) {
	collection, _ := db.GetCollection("tags")

	// Parse request body into tag struct
	var tag types.Tag
	err := json.NewDecoder(r.Body).Decode(&tag)
	if err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(tag)
	if err != nil {
		http.Error(w, "Error retrieving created tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert tag document into MongoDB
	result, err := collection.InsertOne(context.TODO(), tag)
	if err != nil {
		http.Error(w, "Error creating tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the ID of the inserted tag document
	objId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated tag from MongoDB
	filter := bson.M{"_id": objId}
	var createdTag types.Tag
	err = collection.FindOne(context.TODO(), filter).Decode(&createdTag)
	if err != nil {
		http.Error(w, "Error retrieving updated tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the tag struct to JSON format
	jsonBytes, err := json.Marshal(createdTag)
	if err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response headers and write the response body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(jsonBytes)
	if err != nil {
		http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func UpdateTag(w http.ResponseWriter, r *http.Request) {
	collection, _ := db.GetCollection("tags")

	// Get the tag ID from the URL path parameter
	params := mux.Vars(r)
	id := params["tagId"]
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse the request body into a tag struct
	var tag types.Tag
	err = json.NewDecoder(r.Body).Decode(&tag)
	if err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(tag)
	if err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Update the tag document in MongoDB
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": tag}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Error updating tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve the updated tag from MongoDB
	filter = bson.M{"_id": objId}
	var updatedTag types.Tag
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedTag)
	if err != nil {
		http.Error(w, "Error retrieving updated tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the updated tag struct to JSON format
	jsonBytes, err := json.Marshal(updatedTag)
	if err != nil {
		http.Error(w, "Error creating response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func DeleteTag(w http.ResponseWriter, r *http.Request) {
	collection, _ := db.GetCollection("tags")

	// Get the tag ID from the URL path parameter
	vars := mux.Vars(r)
	id := vars["tagId"]
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	// Delete tag document from MongoDB
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Error deleting tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if any documents were deleted
	if result.DeletedCount == 0 {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Tag deleted successfully"))
	if err != nil {
		http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

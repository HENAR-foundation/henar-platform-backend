package projects

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
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/go-playground/validator.v9"
)

func GetProject(w http.ResponseWriter, r *http.Request) {
	collection, _ := db.GetCollection("projects")

	// Get the project ID from the URL path parameter
	vars := mux.Vars(r)
	id := vars["projectId"]
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	filter := bson.D{{Key: "_id", Value: objId}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "views", Value: 1}}}}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result types.Project

	// Find the document by ID, increment its "views" and retrieve the updated document
	err = collection.FindOneAndUpdate(
		context.TODO(),
		filter,
		update,
		options,
	).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error updating project: "+err.Error(), http.StatusInternalServerError)
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

func GetProjects(w http.ResponseWriter, r *http.Request) {
	collection, _ := db.GetCollection("projects")

	filter := bson.M{}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Error finding projects", http.StatusInternalServerError)
		return
	}

	// Get the results from the cursor
	var results []types.Project
	if err := cursor.All(context.TODO(), &results); err != nil {
		http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the result to JSON
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

func CreateProject(w http.ResponseWriter, r *http.Request) {
	collection, _ := db.GetCollection("projects")

	// Parse request body into project struct
	var project types.Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(project)
	if err != nil {
		http.Error(w, "Error retrieving created project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert project document into MongoDB
	result, err := collection.InsertOne(context.TODO(), project)
	if err != nil {
		http.Error(w, "Error creating project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the ID of the inserted project document
	objId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated project from MongoDB
	filter := bson.M{"_id": objId}
	var createdProject types.Project
	err = collection.FindOne(context.TODO(), filter).Decode(&createdProject)
	if err != nil {
		http.Error(w, "Error retrieving updated project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the project struct to JSON format
	jsonBytes, err := json.Marshal(createdProject)
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

func UpdateProject(w http.ResponseWriter, r *http.Request) {
	collection, _ := db.GetCollection("projects")

	// Get the project ID from the URL path parameter
	params := mux.Vars(r)
	id := params["projectId"]
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse the request body into a project struct
	var project types.Project
	err = json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(project)
	if err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Update the project document in MongoDB
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": project}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Error updating project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve the updated project from MongoDB
	filter = bson.M{"_id": objId}
	var updatedProject types.Project
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedProject)
	if err != nil {
		http.Error(w, "Error retrieving updated project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the updated project struct to JSON format
	jsonBytes, err := json.Marshal(updatedProject)
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

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	collection, _ := db.GetCollection("projects")

	// Get the project ID from the URL path parameter
	vars := mux.Vars(r)
	id := vars["projectId"]
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	// Delete project document from MongoDB
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Error deleting project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if any documents were deleted
	if result.DeletedCount == 0 {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Project deleted successfully"))
	if err != nil {
		http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

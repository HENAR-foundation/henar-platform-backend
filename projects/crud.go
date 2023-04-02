package projects

import (
	"context"
	"encoding/json"
	"henar-backend/db"
	"henar-backend/types"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/go-playground/validator.v9"
)

func GetProject(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	// Get the project ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
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
			return c.Status(http.StatusNotFound).SendString("Project not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error updating project: " + err.Error())
	}

	// Marshal the project struct to JSON format
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set("Content-Type", "application/json")
	c.Status(http.StatusOK)
	_, err = c.Write(jsonBytes)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error writing response: " + err.Error())
	}
	return nil
}

func GetProjects(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	filter := bson.M{}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error finding projects")
	}

	// Get the results from the cursor
	var results []types.Project
	if err := cursor.All(context.TODO(), &results); err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error querying database: " + err.Error())
	}

	// Marshal the result to JSON
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set("Content-Type", "application/json")
	c.Status(http.StatusOK)
	_, err = c.Write(jsonBytes)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error writing response: " + err.Error())
	}
	return nil
}

func CreateProject(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	// Parse request body into project struct
	var project types.Project
	err := c.BodyParser(&project)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(project)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving created project: " + err.Error())
	}

	// Insert project document into MongoDB
	result, err := collection.InsertOne(context.TODO(), project)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error creating project: " + err.Error())
	}

	// Get the ID of the inserted project document
	objId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated project from MongoDB
	filter := bson.M{"_id": objId}
	var createdProject types.Project
	err = collection.FindOne(context.TODO(), filter).Decode(&createdProject)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated project: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusCreated).JSON(createdProject)
}

func UpdateProject(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	// Get the project ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Parse the request body into a project struct
	var project types.Project
	err = c.BodyParser(&project)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(project)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Validation error: " + err.Error())
	}

	// Update the project document in MongoDB
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": project}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating project: " + err.Error())
	}

	// Retrieve the updated project from MongoDB
	filter = bson.M{"_id": objId}
	var updatedProject types.Project
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedProject)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated project: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusOK).JSON(updatedProject)
}

func DeleteProject(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	// Get the project ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	// Delete project document from MongoDB
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error deleting project: " + err.Error())
	}

	// Check if any documents were deleted
	if result.DeletedCount == 0 {
		return c.Status(http.StatusNotFound).SendString("Project not found")
	}

	return c.SendString("Project deleted successfully")
}

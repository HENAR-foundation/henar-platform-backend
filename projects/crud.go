package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"henar-backend/db"
	"henar-backend/types"
	"henar-backend/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/go-playground/validator.v9"
)

// GetProject retrieves a project by its slug and increments its view count.
// @Summary Get a project by slug
// @Description Retrieves a project by its slug and increments its view count.
// @Tags projects
// @Accept json
// @Produce json
// @Param slug path string true "Project slug"
// @Success 200 {object} types.Project
// @Failure 400 {string} string "Invalid slug"
// @Failure 404 {string} string "Project not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/projects/{slug} [get]
func GetProject(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	slug := c.Params("slug")

	filter := bson.D{{Key: "slug", Value: slug}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "views", Value: 1}}}}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result types.Project

	// Find the document by slug, increment its "views" and retrieve the updated document
	err := collection.FindOneAndUpdate(
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

// GetProjects retrieves a list of all projects in the database.
// @Summary Get all projects
// @Description Retrieves a list of all projects in the database.
// @Tags projects
// @Accept json
// @Produce json
// @Success 200 {array} types.Project
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/projects [get]
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param sort query string false "Comma-separated list of sort fields and directions, e.g. views,-applicants,tags"
// @Param language query string false "Language code for the title (default 'en')"
// @Param title query string false "Substring to match in the title"
// @Param tags query string false "Comma-separated list of tag IDs to filter by"
// @Param location query string false "Location ID to filter by"
func GetProjects(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	// Get the filter and options for the query
	findOptions, err := utils.GetPaginationOptions(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid pagination parameters")
	}

	filter, err := utils.GetFilter(c)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting projects filter: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString(errMsg)
	}

	sort := utils.GetSort(c)
	if len(sort) != 0 {
		findOptions.SetSort(sort)
	}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter, findOptions)
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

// CreateProject creates a new project in the database.
// @Summary Create a project
// @Description Creates a new project in the database.
// @Tags projects
// @Accept json
// @Produce json
// @Param project body types.Project true "Project"
// @Success 201 {object} types.Project
// @Failure 400 {string} string "Error parsing request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/projects [post]
func CreateProject(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	// Parse request body into project struct
	var project types.Project
	err := c.BodyParser(&project)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	slugText := utils.CreateSlug(project.Title)
	project.Slug = slugText

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

// UpdateProject updates an existing project in the database.
// @Summary Update a project
// @Description Updates an existing project in the database.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param project body types.Project true "Project"
// @Success 204 "No content"
// @Failure 400 {string} string "Invalid ID or error parsing request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/projects/{id} [put]
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

	slugText := utils.CreateSlug(project.Title)
	project.Slug = slugText

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

// @Summary Delete a project
// @Description Deletes a project from the database based on the provided ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {string} string "Project deleted successfully"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Project not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/projects/{id} [delete]
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

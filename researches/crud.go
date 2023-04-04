package researches

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
	"gopkg.in/go-playground/validator.v9"
)

// @Summary Get all researches
// @Description Retrieves all researches
// @Tags researches
// @Accept json
// @Produce json
// @Success 200 {array} types.Research
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/researches [get]
func GetResearches(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("researches")

	filter := bson.M{}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error finding researches")
	}

	// Get the results from the cursor
	var results []types.Research
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	// Marshal the research struct to JSON format
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(fiber.StatusOK)
	return c.Send(jsonBytes)
}

// @Summary Get research by ID
// @Description Retrieves a research by its ID
// @Tags researches
// @Accept json
// @Produce json
// @Param id path string true "Research ID"
// @Success 200 {object} types.Research
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Research not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/researches/{id} [get]
func GetResearch(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("researches")

	// Get the research ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	var result types.Research

	// Find the research by ID
	err = collection.FindOne(
		context.TODO(),
		filter,
	).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).SendString("Research not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting research: " + err.Error())
	}

	// Marshal the research struct to JSON format
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(fiber.StatusOK)
	return c.Send(jsonBytes)
}

// @Summary Create research
// @Description Creates a new research
// @Tags researches
// @Accept json
// @Produce json
// @Param research body types.Research true "Research Object"
// @Success 201 {object} types.Research
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/researches [post]
func CreateResearch(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("researches")

	// Parse request body into research struct
	var research types.Research
	err := c.BodyParser(&research)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(research)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving created research: " + err.Error())
	}

	// Insert research document into MongoDB
	result, err := collection.InsertOne(context.TODO(), research)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error creating research: " + err.Error())
	}

	// Get the ID of the inserted research document
	objId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated research from MongoDB
	filter := bson.M{"_id": objId}
	var createdResearch types.Research
	err = collection.FindOne(context.TODO(), filter).Decode(&createdResearch)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated research: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusCreated).JSON(createdResearch)
}

// @Summary Update research by ID
// @Description Updates a research by its ID
// @Tags researches
// @Accept json
// @Produce json
// @Param id path string true "Research ID"
// @Param research body types.Research true "Research Object"
// @Success 200 {object} types.Research
// @Failure 400 {string} string "Invalid ID or Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/researches/{id} [put]
func UpdateResearch(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("researches")

	// Get the research ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Parse the request body into a research struct
	var research types.Research
	err = c.BodyParser(&research)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(research)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Validation error: " + err.Error())
	}

	// Update the research document in MongoDB
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": research}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating research: " + err.Error())
	}

	// Retrieve the updated research from MongoDB
	filter = bson.M{"_id": objId}
	var updatedResearch types.Research
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedResearch)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated research: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusOK).JSON(updatedResearch)
}

// @Summary Delete research by ID
// @Description Deletes a research document by its ID
// @Tags researches
// @Accept json
// @Produce json
// @Param id path string true "Research ID"
// @Success 200 {string} string "Research deleted successfully"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Research not found"
// @Failure 500 {string} string "Error deleting research: <error message>"
// @Router /v1/researches/{id} [delete]
func DeleteResearch(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("researches")

	// Get the research ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	// Delete research document from MongoDB
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error deleting research: " + err.Error())
	}

	// Check if any documents were deleted
	if result.DeletedCount == 0 {
		return c.Status(http.StatusNotFound).SendString("Research not found")
	}

	return c.SendString("Research deleted successfully")
}
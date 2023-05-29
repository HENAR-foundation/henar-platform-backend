package locations

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
	"gopkg.in/go-playground/validator.v9"
)

// @Summary Get all locations
// @Description Retrieves all locations
// @Tags locations
// @Accept json
// @Produce json
// @Success 200 {array} types.Location
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/locations [get]
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
func GetLocations(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("locations")

	filter := bson.M{}

	// Get the pagination options for the query
	findOptions, err := utils.GetPaginationOptions(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid pagination parameters")
	}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error finding locations")
	}

	// Get the results from the cursor
	var results []types.Location
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

// @Summary Get research by slug
// @Description Retrieves a research by its slug
// @Tags locations
// @Accept json
// @Produce json
// @Param slug path string true "Location slug"
// @Success 200 {object} types.Location
// @Failure 400 {string} string "Invalid slug"
// @Failure 404 {string} string "Location not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/locations/{slug} [get]
func GetLocation(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("locations")

	// Get the research ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	filter := bson.M{"_id": objId}

	var result types.Location

	// Find the research by slug
	err = collection.FindOne(
		context.TODO(),
		filter,
	).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).SendString("Location not found")
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
// @Tags locations
// @Accept json
// @Produce json
// @Param research body types.Location true "Location Object"
// @Success 201 {object} types.Location
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/locations [post]
func CreateLocation(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("locations")

	// Parse request body into research struct
	var research types.Location
	err := c.BodyParser(&research)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(research)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error retrieving created research: " + err.Error())
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
	var createdLocation types.Location
	err = collection.FindOne(context.TODO(), filter).Decode(&createdLocation)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated research: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusCreated).JSON(createdLocation)
}

// @Summary Update research by ID
// @Description Updates a research by its ID
// @Tags locations
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Param research body types.Location true "Location Object"
// @Success 200 {object} types.Location
// @Failure 400 {string} string "Invalid ID or Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/locations/{id} [patch]
func UpdateLocation(c *fiber.Ctx) error {
	if c.Locals("userRole") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	collection, _ := db.GetCollection("locations")

	// Get the research ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Parse the request body into a research struct
	var research types.Location
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
	var updatedLocation types.Location
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedLocation)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated research: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusOK).JSON(updatedLocation)
}

// @Summary Delete research by ID
// @Description Deletes a research document by its ID
// @Tags locations
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Success 200 {string} string "Location deleted successfully"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Location not found"
// @Failure 500 {string} string "Error deleting research: <error message>"
// @Router /v1/locations/{id} [delete]
func DeleteLocation(c *fiber.Ctx) error {
	if c.Locals("userRole") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	collection, _ := db.GetCollection("locations")

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
		return c.Status(http.StatusNotFound).SendString("Location not found")
	}

	return c.SendString("Location deleted successfully")
}

// @Summary Get Location Suggestions
// @Description Returns a list of suggested addresses based on a query string.
// @Tags locations
// @Accept json
// @Produce json
// @Success 200 {object} types.Suggestions
// @Failure 400 {string} string "Invalid request parameters."
// @Failure 500 {string} string "Internal server error."
// @Router /v1/locations/suggestions [get]
// @Param q query string false "Query string for location suggestions"
// @Param lanquage query string false "Language of response"
func GetLocationSuggestions(c *fiber.Ctx) error {
	url := "https://suggestions.dadata.ru/suggestions/api/4_1/rs/suggest/address"
	apiKey := "5a688e4bfd915586e67c7c0f42fcc08cda0d081b"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	language := c.Query("language")
	if language == "" {
		language = "en" // default language is English
	}

	// Set the query parameters
	q := req.URL.Query()
	q.Add("query", c.Query("q"))
	q.Add("language", language)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token "+apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	c.Set("Content-Type", "application/json; charset=utf-8")

	return c.Status(res.StatusCode).SendStream(res.Body)
}

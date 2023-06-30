package tags

import (
	"context"
	"encoding/json"
	"henar-backend/db"
	"henar-backend/sentry"
	"henar-backend/types"
	"henar-backend/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/go-playground/validator.v9"
)

// @Summary Get all tags
// @Description Retrieves a list of all tags in the database
// @Tags tags
// @Accept json
// @Produce json
// @Success 200 {array} types.Tag
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/tags [get]
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
func GetTags(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("tags")

	filter := bson.M{}

	// Get the pagination options for the query
	findOptions, err := utils.GetPaginationOptions(c)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid pagination parameters")
	}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error finding tags")
	}

	// Get the results from the cursor
	var results []types.Tag
	if err = cursor.All(context.TODO(), &results); err != nil {
		sentry.SentryHandler(err)
		panic(err)
	}

	// Marshal the tag struct to JSON format
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(fiber.StatusOK)
	return c.Send(jsonBytes)
}

// @Summary Get tag by ID
// @Description Retrieves a tag by its ID
// @Tags tags
// @Accept json
// @Produce json
// @Param id path string true "Tag ID"
// @Success 200 {object} types.Tag
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Tag not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/tags/{id} [get]
func GetTag(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("tags")

	// Get the tag ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	var result types.Tag

	// Find the tag by ID
	err = collection.FindOne(
		context.TODO(),
		filter,
	).Decode(&result)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).SendString("Tag not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting tag: " + err.Error())
	}

	// Marshal the tag struct to JSON format
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(fiber.StatusOK)
	return c.Send(jsonBytes)
}

// @Summary Create a new tag
// @Description Creates a new tag in the database
// @Tags tags
// @Accept json
// @Produce json
// @Param tag body types.Tag true "Tag object to create"
// @Success 201 {object} types.Tag "Created tag object"
// @Failure 400 {string} string "Error parsing request body or validation error"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/tags [post]
func CreateTag(c *fiber.Ctx) error {
	if c.Locals("userRole") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	collection, _ := db.GetCollection("tags")

	// Parse request body into tag struct
	var tag types.Tag
	err := c.BodyParser(&tag)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Error parsing request body: " + err.Error()})
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(tag)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Error retrieving created tag: " + err.Error()})
	}

	// Insert tag document into MongoDB
	result, err := collection.InsertOne(context.TODO(), tag)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error creating tag: " + err.Error()})
	}

	// Get the ID of the inserted tag document
	objId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated tag from MongoDB
	filter := bson.M{"_id": objId}
	var createdTag types.Tag
	err = collection.FindOne(context.TODO(), filter).Decode(&createdTag)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error retrieving updated tag: " + err.Error()})
	}

	// Marshal the tag struct to JSON format
	jsonBytes, err := json.Marshal(createdTag)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error encoding JSON: " + err.Error()})
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": jsonBytes})
}

// @Summary Update tag by ID
// @Description Updates a tag with the given ID in the database
// @Tags tags
// @Accept json
// @Produce json
// @Param id path string true "Tag ID"
// @Param tag body types.Tag true "Tag object to update"
// @Success 200 {object} types.Tag "Updated tag object"
// @Failure 400 {string} string "Invalid ID or error parsing request body or validation error"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/tags/{id} [patch]
func UpdateTag(c *fiber.Ctx) error {
	if c.Locals("userRole") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	collection, _ := db.GetCollection("tags")

	// Get the tag ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Parse the request body into a tag struct
	var tag types.Tag
	err = c.BodyParser(&tag)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(tag)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Validation error: " + err.Error())
	}

	// Update the tag document in MongoDB
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": tag}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating tag: " + err.Error())
	}

	// Retrieve the updated tag from MongoDB
	filter = bson.M{"_id": objId}
	var updatedTag types.Tag
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedTag)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated tag: " + err.Error())
	}

	// Marshal the updated tag struct to JSON format
	jsonBytes, err := json.Marshal(updatedTag)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error creating response: " + err.Error())
	}

	c.Set("Content-Type", "application/json")
	return c.Send(jsonBytes)
}

// @Summary Delete tag by ID
// @Description Deletes a tag with the given ID from the database
// @Tags tags
// @Param id path string true "Tag ID"
// @Success 200 {string} string "Tag deleted successfully"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Tag not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/tags/{id} [delete]
func DeleteTag(c *fiber.Ctx) error {
	if c.Locals("userRole") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	collection, _ := db.GetCollection("tags")

	// Get the tag ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	// Delete tag document from MongoDB
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error deleting tag: " + err.Error())
	}

	// Check if any documents were deleted
	if result.DeletedCount == 0 {
		return c.Status(http.StatusNotFound).SendString("Tag not found")
	}

	return c.SendString("Tag deleted successfully")
}

package statisticsCategories

import (
	"context"
	"encoding/json"
	"henar-backend/db"
	"henar-backend/sentry"
	"henar-backend/types"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/go-playground/validator.v9"
)

// @Summary Get all statistics
// @Description Retrieves all statistics
// @Tags statistics
// @Accept json
// @Produce json
// @Success 200 {array} types.Statistic
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/statistics [get]
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
func GetStatisticsCategories(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("statistics_categories")

	// Query the database and get the cursor
	filter := bson.M{}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error finding statistics categories")
	}

	// Get the results from the cursor
	var results []types.StatisticsCategory
	if err = cursor.All(context.TODO(), &results); err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error finding statistics categories")
	}

	// Marshal the statistic struct to JSON format
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

// @Summary Get statistic by ID
// @Description Retrieves a statistic by its ID
// @Tags statistics
// @Accept json
// @Produce json
// @Param id path string true "Statistic ID"
// @Success 200 {object} types.Statistic
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Statistic not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/statistics/{id} [get]
func GetStatisticsCategory(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("statistics_categories")

	// Get the statistic ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	var result types.StatisticsCategory

	// Find the statistic by ID
	err = collection.FindOne(
		context.TODO(),
		filter,
	).Decode(&result)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).SendString("Statistic category not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting statistics category: " + err.Error())
	}

	// Marshal the statistic struct to JSON format
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

// @Summary Create statistic
// @Description Creates a new statistic
// @Tags statistics
// @Accept json
// @Produce json
// @Param statistic body types.Statistic true "Statistic Object"
// @Success 201 {object} types.Statistic
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/statistics [post]
func CreateStatisticsCategory(c *fiber.Ctx) error {
	if c.Locals("userRole") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	collection, _ := db.GetCollection("statistics_categories")

	// Parse request body into statistic struct
	var statisticsCategory types.StatisticsCategory
	err := c.BodyParser(&statisticsCategory)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(statisticsCategory)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error retrieving created statistics category: " + err.Error())
	}

	// Insert statistic document into MongoDB
	result, err := collection.InsertOne(context.TODO(), statisticsCategory)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error creating statistics category: " + err.Error())
	}

	// Get the ID of the inserted statistic document
	objId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated statistic from MongoDB
	filter := bson.M{"_id": objId}
	var createdStatisticsCategory types.StatisticsCategory
	err = collection.FindOne(context.TODO(), filter).Decode(&createdStatisticsCategory)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated statistics category: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusCreated).JSON(createdStatisticsCategory)
}

// @Summary Update statistic by ID
// @Description Updates a statistic by its ID
// @Tags statistics
// @Accept json
// @Produce json
// @Param id path string true "Statistic ID"
// @Param statistic body types.Statistic true "Statistic Object"
// @Success 200 {object} types.Statistic
// @Failure 400 {string} string "Invalid ID or Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/statistics/{id} [patch]
func UpdateStatisticsCategory(c *fiber.Ctx) error {
	if c.Locals("userRole") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	collection, _ := db.GetCollection("statistics_categories")

	// Get the statistic ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Parse the request body into a statistic struct
	var statisticsCategory types.StatisticsCategory
	err = c.BodyParser(&statisticsCategory)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(statisticsCategory)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Validation error: " + err.Error())
	}

	// Update the statistic document in MongoDB
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": statisticsCategory}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating statistics category: " + err.Error())
	}

	// Retrieve the updated statistic from MongoDB
	filter = bson.M{"_id": objId}
	var updatedStatisticsCategory types.StatisticsCategory
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedStatisticsCategory)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated statistics category: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusOK).JSON(updatedStatisticsCategory)
}

// @Summary Delete statistic by ID
// @Description Deletes a statistic document by its ID
// @Tags statistics
// @Accept json
// @Produce json
// @Param id path string true "Statistic ID"
// @Success 200 {string} string "Statistic deleted successfully"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Statistic not found"
// @Failure 500 {string} string "Error deleting statistic: <error message>"
// @Router /v1/statistics/{id} [delete]
func DeleteStatisticsCategory(c *fiber.Ctx) error {
	if c.Locals("userRole") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	collection, _ := db.GetCollection("statistics_categories")

	// Get the statistic ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	// Delete statistic document from MongoDB
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error deleting statistics category: " + err.Error())
	}

	// Check if any documents were deleted
	if result.DeletedCount == 0 {
		return c.Status(http.StatusNotFound).SendString("Statistics category not found")
	}

	return c.SendString("Statistics category deleted successfully")
}

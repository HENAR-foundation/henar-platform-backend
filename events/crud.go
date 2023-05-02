package events

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

// @Summary Get all events
// @Description Retrieves all events
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {array} types.Event
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/events [get]
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param sort query string false "Comma-separated list of sort fields and directions, e.g. views,-applicants,tags"
// @Param language query string false "Language code for the title (default 'en')"
// @Param title query string false "Substring to match in the title"
// @Param tags query string false "Comma-separated list of tag IDs to filter by"
// @Param location query string false "Location ID to filter by"
func GetEvents(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("events")

	// Get the filter and options for the query
	findOptions, err := utils.GetPaginationOptions(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid pagination parameters")
	}

	filter, err := utils.GetFilter(c)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting filter: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString(errMsg)
	}

	sort := utils.GetSort(c)
	if len(sort) != 0 {
		findOptions.SetSort(sort)
	}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error finding events")
	}

	// Get the results from the cursor
	var results []types.Event
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	// Marshal the event struct to JSON format
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(fiber.StatusOK)
	return c.Send(jsonBytes)
}

// @Summary Get event by slug
// @Description Retrieves a event by its slug
// @Tags events
// @Accept json
// @Produce json
// @Param slug path string true "Event slug"
// @Success 200 {object} types.Event
// @Failure 400 {string} string "Invalid slug"
// @Failure 404 {string} string "Event not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/events/{slug} [get]
func GetEvent(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("events")

	slug := c.Params("slug")

	filter := bson.D{{Key: "slug", Value: slug}}

	var result types.Event

	// Find the event by slug
	err := collection.FindOne(
		context.TODO(),
		filter,
	).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).SendString("Event not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting event: " + err.Error())
	}

	// Marshal the event struct to JSON format
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(fiber.StatusOK)
	return c.Send(jsonBytes)
}

// @Summary Create event
// @Description Creates a new event
// @Tags events
// @Accept json
// @Produce json
// @Param event body types.Event true "Event Object"
// @Success 201 {object} types.Event
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/events [post]
func CreateEvent(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("events")

	// Parse request body into event struct
	var event types.Event
	err := c.BodyParser(&event)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	slugText := utils.CreateSlug(event.Title)
	event.Slug = slugText

	// Validate the required fields
	v := validator.New()
	err = v.Struct(event)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving created event: " + err.Error())
	}

	// Insert event document into MongoDB
	result, err := collection.InsertOne(context.TODO(), event)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error creating event: " + err.Error())
	}

	// Get the ID of the inserted event document
	objId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated event from MongoDB
	filter := bson.M{"_id": objId}
	var createdEvent types.Event
	err = collection.FindOne(context.TODO(), filter).Decode(&createdEvent)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated event: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusCreated).JSON(createdEvent)
}

// @Summary Update event by ID
// @Description Updates a event by its ID
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param event body types.Event true "Event Object"
// @Success 200 {object} types.Event
// @Failure 400 {string} string "Invalid ID or Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/events/{id} [patch]
func UpdateEvent(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("events")

	// Get the event ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Parse the request body into a event struct
	var event types.Event
	err = c.BodyParser(&event)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	slugText := utils.CreateSlug(event.Title)
	event.Slug = slugText

	// Validate the required fields
	v := validator.New()
	err = v.Struct(event)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Validation error: " + err.Error())
	}

	// Update the event document in MongoDB
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": event}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating event: " + err.Error())
	}

	// Retrieve the updated event from MongoDB
	filter = bson.M{"_id": objId}
	var updatedEvent types.Event
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedEvent)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated event: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusOK).JSON(updatedEvent)
}

// @Summary Delete event by ID
// @Description Deletes a event document by its ID
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {string} string "Event deleted successfully"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Event not found"
// @Failure 500 {string} string "Error deleting event: <error message>"
// @Router /v1/events/{id} [delete]
func DeleteEvent(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("events")

	// Get the event ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	// Delete event document from MongoDB
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error deleting event: " + err.Error())
	}

	// Check if any documents were deleted
	if result.DeletedCount == 0 {
		return c.Status(http.StatusNotFound).SendString("Event not found")
	}

	return c.SendString("Event deleted successfully")
}

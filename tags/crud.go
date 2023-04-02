package tags

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

func GetTags(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("tags")

	filter := bson.M{}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error finding tags")
	}

	// Get the results from the cursor
	var results []types.Tag
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	// Marshal the tag struct to JSON format
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(fiber.StatusOK)
	return c.Send(jsonBytes)
}

func GetTag(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("tags")

	// Get the tag ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
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
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).SendString("Tag not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting tag: " + err.Error())
	}

	// Marshal the tag struct to JSON format
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	c.Status(fiber.StatusOK)
	return c.Send(jsonBytes)
}

func CreateTag(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("tags")

	// Parse request body into tag struct
	var tag types.Tag
	err := c.BodyParser(&tag)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Error parsing request body: " + err.Error()})
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(tag)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error retrieving created tag: " + err.Error()})
	}

	// Insert tag document into MongoDB
	result, err := collection.InsertOne(context.TODO(), tag)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error creating tag: " + err.Error()})
	}

	// Get the ID of the inserted tag document
	objId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated tag from MongoDB
	filter := bson.M{"_id": objId}
	var createdTag types.Tag
	err = collection.FindOne(context.TODO(), filter).Decode(&createdTag)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error retrieving updated tag: " + err.Error()})
	}

	// Marshal the tag struct to JSON format
	jsonBytes, err := json.Marshal(createdTag)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error encoding JSON: " + err.Error()})
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusCreated).JSON(fiber.Map{"data": jsonBytes})
}

func UpdateTag(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("tags")

	// Get the tag ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Parse the request body into a tag struct
	var tag types.Tag
	err = c.BodyParser(&tag)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(tag)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Validation error: " + err.Error())
	}

	// Update the tag document in MongoDB
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": tag}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating tag: " + err.Error())
	}

	// Retrieve the updated tag from MongoDB
	filter = bson.M{"_id": objId}
	var updatedTag types.Tag
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedTag)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated tag: " + err.Error())
	}

	// Marshal the updated tag struct to JSON format
	jsonBytes, err := json.Marshal(updatedTag)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error creating response: " + err.Error())
	}

	c.Set("Content-Type", "application/json")
	return c.Send(jsonBytes)
}

func DeleteTag(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("tags")

	// Get the tag ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	filter := bson.D{{Key: "_id", Value: objId}}

	// Delete tag document from MongoDB
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error deleting tag: " + err.Error())
	}

	// Check if any documents were deleted
	if result.DeletedCount == 0 {
		return c.Status(http.StatusNotFound).SendString("Tag not found")
	}

	return c.SendString("Tag deleted successfully")
}

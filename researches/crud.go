package researches

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

// @Summary Get all researches
// @Description Retrieves all researches
// @Tags researches
// @Accept json
// @Produce json
// @Success 200 {array} types.Research
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/researches [get]
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param sort query string false "Comma-separated list of sort fields and directions, e.g. views,-applicants,tags"
// @Param language query string false "Language code for the title (default 'en')"
// @Param title query string false "Substring to match in the title"
// @Param tags query string false "Comma-separated list of tag IDs to filter by"
// @Param location query string false "Location ID to filter by"
func GetResearches(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("researches")

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
	// TODO: GET: User can't get pending | rejected entity

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error finding researches")
	}

	// Get the results from the cursor
	var results []types.Research
	if err = cursor.All(context.TODO(), &results); err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error finding reseaches")
	}
	if c.Locals("userRole") != "admin" {
		fieldsToUpdate := []string{"ModerationStatus", "ReasonOfReject"}
		utils.UpdateResultsForUserRole(results, fieldsToUpdate)
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
// @Tags researches
// @Accept json
// @Produce json
// @Param slug path string true "Research slug"
// @Success 200 {object} types.Research
// @Failure 400 {string} string "Invalid slug"
// @Failure 404 {string} string "Research not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/researches/{slug} [get]
func GetResearch(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("researches")

	slug := c.Params("slug")

	filter := bson.D{{Key: "slug", Value: slug}}

	var result types.Research

	// Find the research by slug
	err := collection.FindOne(
		context.TODO(),
		filter,
	).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).SendString("Research not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting research: " + err.Error())
	}

	if c.Locals("userRole") != "admin" &&
		c.Locals("user_id") != result.CreatedBy.Hex() {
		fieldsToUpdate := []string{"ModerationStatus", "ReasonOfReject"}
		utils.UpdateResultForUserRole(&result, fieldsToUpdate)
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
		return c.Status(http.StatusBadRequest).SendString("Error retrieving created research: " + err.Error())
	}

	// update fields
	userId, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}
	research.CreatedBy = userId
	pending := types.Pending
	research.ModerationStatus = &pending
	slugText := utils.CreateSlug(research.Title)
	research.Slug = slugText

	// Insert research document into MongoDB
	result, err := collection.InsertOne(context.TODO(), research)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error creating research: " + err.Error())
	}

	// Retrieve the updated research from MongoDB
	filter := bson.M{"_id": result.InsertedID.(primitive.ObjectID)}
	var createdResearch types.Research
	err = collection.FindOne(context.TODO(), filter).Decode(&createdResearch)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated research: " + err.Error())
	}

	// update user
	usersCollection, _ := db.GetCollection("users")
	userFilter := bson.M{"_id": userId}

	var user types.User
	err = usersCollection.FindOne(context.TODO(), userFilter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	if user.Researches == nil {
		user.Researches = make(map[primitive.ObjectID]bool)
	}
	user.Researches[createdResearch.ID] = true

	update := bson.M{"$set": user}
	_, err = usersCollection.UpdateOne(context.TODO(), userFilter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
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
// @Router /v1/researches/{id} [patch]
func UpdateResearch(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("researches")

	// Get the research ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Parse the request body into a updateBody struct
	var updateBody types.Research
	err = c.BodyParser(&updateBody)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(updateBody)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Validation error: " + err.Error())
	}

	// Find the Research document from MongoDB
	var result types.Research
	err = collection.FindOne(context.TODO(), bson.M{"_id": objId}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).SendString("Research not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting research: " + err.Error())
	}

	if c.Locals("userRole") != "admin" &&
		c.Locals("user_id") != result.CreatedBy.Hex() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	if c.Locals("userRole") != "admin" {
		// owner can't edit the following fields
		if updateBody.ModerationStatus != nil ||
			updateBody.ReasonOfReject != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Permission or ownership error",
			})
		}
	}

	slugText := utils.CreateSlug(updateBody.Title)
	updateBody.Slug = slugText

	// Update the research document in MongoDB
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": updateBody}
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
	researchId := c.Params("id")
	researchObjId, err := primitive.ObjectIDFromHex(researchId)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Find the research document from MongoDB
	var research types.Research
	err = collection.FindOne(context.TODO(), bson.M{"_id": researchObjId}).Decode(&research)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).SendString("Research not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting research: " + err.Error())
	}

	userId := c.Locals("user_id").(string)
	if c.Locals("userRole") != "admin" &&
		userId != research.CreatedBy.Hex() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	researchFilter := bson.D{{Key: "_id", Value: researchObjId}}

	// Delete research document from MongoDB
	result, err := collection.DeleteOne(context.TODO(), researchFilter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error deleting research: " + err.Error())
	}

	// Check if any documents were deleted
	if result.DeletedCount == 0 {
		return c.Status(http.StatusNotFound).SendString("Research not found")
	}

	// update user
	usersCollection, _ := db.GetCollection("users")
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}
	var user types.User
	userFilter := bson.M{"_id": userObjId}
	err = usersCollection.FindOne(context.TODO(), userFilter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	delete(user.Researches, researchObjId)

	update := bson.M{"$set": user}
	_, err = usersCollection.UpdateOne(context.TODO(), userFilter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	return c.SendString("Research deleted successfully")
}

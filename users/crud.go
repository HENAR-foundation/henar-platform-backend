package users

import (
	"context"
	"fmt"
	"henar-backend/db"
	"henar-backend/types"
	"henar-backend/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

// @Summary Create a new user
// @Description Creates a new user in the database
// @Tags users
// @Accept json
// @Produce json
// @Success 201 {object} types.User "The created user"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /users [post]
func CreateUser(c *fiber.Ctx) error {
	// Parse request body into user struct
	var uc types.UserCredentials
	err := c.BodyParser(&uc)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}
	fmt.Println(uc)

	// Validate the required fields
	v := validator.New()
	err = v.Struct(uc)
	if err != nil {
		return fmt.Errorf("error validating user: %w", err)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(uc.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Error hashing password: %w", err)
	}
	// var user types.User
	user := types.User{
		Email:          uc.Email,
		HashedPassword: hashedPassword,
		Role:           "specialist",
	}
	v.Struct(user)

	// Check if the email address is already in use
	collection, _ := db.GetCollection("users")
	filter := bson.M{"email": user.Email}
	var existingUser types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&existingUser)
	if err == nil {
		return fmt.Errorf("Email address already in use")
	}

	// Insert user document into MongoDB
	result, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		return fmt.Errorf("Error creating user: ", err)
	}

	// Get the ID of the inserted user document
	objId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated user from MongoDB
	filter = bson.M{"_id": objId}
	var createdUser types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&createdUser)
	if err != nil {
		return fmt.Errorf("Error retrieving created user: ", err)
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusCreated).JSON(createdUser)
}

// @Summary Update an existing user
// @Description Updates an existing user in the database
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body types.UserBody true "User details"
// @Success 200 {object} types.User
// @Failure 400 {string} string "Bad request"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /users/{id} [patch]
func UpdateUser(c *fiber.Ctx) error {
	// Parse request body into user struct
	var userBody types.UserBody
	err := c.BodyParser(&userBody)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(userBody)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error validating user: " + err.Error())
	}

	// Get the ID of the user to update
	userId := c.Params("id")

	// Convert the ID to a MongoDB ObjectID
	objId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing user ID: " + err.Error())
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userBody.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Error hashing password: %w", err)
	}
	var user types.User
	user.HashedPassword = hashedPassword

	// Update the user document in MongoDB
	collection, _ := db.GetCollection("users")
	filter := bson.M{"email": userBody.Email}
	var existingUser types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&existingUser)
	if existingUser.ID != objId {
		if err == nil {
			return fmt.Errorf("Email address already in use")
		}
	}

	filter = bson.M{"_id": objId}
	update := bson.M{
		"$set": bson.M{
			"email":          userBody.Email,
			"hashedPassword": hashedPassword,
			"avatar":         userBody.Avatar,
			"full_name":      userBody.FullName,
			"description":    userBody.Description,
			"contacts":       userBody.Contacts,
			"location":       userBody.Location,
			"role":           userBody.Role,
			"job":            userBody.Job,
			"tags":           userBody.Tags,
			// add other fields here as necessary
		},
	}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	// Retrieve the updated user from MongoDB
	var updatedUser types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedUser)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated user: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusOK).JSON(updatedUser)
}

// @Summary Get user by id
// @Description Retrieves a user by its id
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User "
// @Success 200 {object} types.User
// @Failure 400 {string} string "Invalid id"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/users/{id} [get]
func GetUser(c *fiber.Ctx) error {
	// Parse the user ID from the request parameters
	id := c.Params("id")

	// Convert the user ID string to an ObjectID
	objId, _ := primitive.ObjectIDFromHex(id)

	// Retrieve the user from MongoDB
	collection, _ := db.GetCollection("users")
	filter := bson.M{"_id": objId}
	var user types.User
	err := collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusOK).JSON(user)
}

// @Summary Get all users
// @Description Retrieves all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} types.User
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/users [get]
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param sort query string false "Comma-separated list of sort fields and directions, e.g. tags"
// @Param language query string false "Language code for the title (default 'en')"
// @Param full_name query string false "Substring to match in the full name"
// @Param job query string false "Substring to match in the job"
// @Param tags query string false "Comma-separated list of tag IDs to filter by"
// @Param location query string false "Location ID to filter by"
func GetUsers(c *fiber.Ctx) error {
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

	// Retrieve the list of users from MongoDB
	collection, _ := db.GetCollection("users")
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving users: " + err.Error())
	}
	defer cur.Close(context.Background())

	// Convert the list of users to a slice and set the response headers and body
	var users []types.User
	for cur.Next(context.Background()) {
		var user types.User
		err := cur.Decode(&user)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error decoding user: " + err.Error())
		}
		users = append(users, user)
	}
	return c.Status(http.StatusOK).JSON(users)
}

// @Summary Delete user by ID
// @Description Deletes a user document by its ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {string} string "User deleted successfully"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error deleting user: <error message>"
// @Router /v1/users/{id} [delete]
func DeleteUser(c *fiber.Ctx) error {
	// Get the ID of the user to delete
	userId := c.Params("id")

	// Convert the ID to a MongoDB ObjectID
	objId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing user ID: " + err.Error())
	}

	// Delete the user document in MongoDB
	collection, _ := db.GetCollection("users")
	filter := bson.M{"_id": objId}
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error deleting user: " + err.Error())
	}

	// Check if any user was deleted
	if result.DeletedCount == 0 {
		return c.Status(http.StatusNotFound).SendString("User not found")
	}

	// Set the response headers and write the response body
	return c.SendString("User deleted successfully")
}

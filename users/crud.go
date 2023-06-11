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
	var uc types.User
	err := c.BodyParser(&uc)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(uc)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error validating user: " + err.Error())
	}

	if uc.Password == nil {
		return c.Status(http.StatusBadRequest).SendString("Password is required")
	}

	// Hash the password
	Password, err := bcrypt.GenerateFromPassword(
		[]byte(*uc.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error hashing password: " + err.Error())
	}

	passwordString := string(Password)
	specialist := types.Specialist
	user := types.User{
		UserCredentials: types.UserCredentials{
			Email:    uc.Email,
			Password: &passwordString,
		},
		UserBody: types.UserBody{
			Role: &specialist,
			ContactsRequest: types.ContactsRequest{
				IncomingContactRequests:   make(map[primitive.ObjectID]string),
				OutgoingContactRequests:   make(map[primitive.ObjectID]string),
				ConfirmedContactsRequests: make(map[primitive.ObjectID]string),
				BlockedUsers:              make(map[primitive.ObjectID]string),
			},
			UserProjects: types.UserProjects{
				// TODO: delete this
				ProjectsApplications:  make(map[primitive.ObjectID]primitive.ObjectID),
				ConfirmedApplications: make(map[primitive.ObjectID]primitive.ObjectID),
				RejectedApplicants:    make(map[primitive.ObjectID]primitive.ObjectID),
				CreatedProjects:       make(map[primitive.ObjectID]bool),
			},
		},
	}
	// TODO: check user
	v.Struct(user)

	// Check if the email address is already in use
	collection, _ := db.GetCollection("users")
	filter := bson.M{"user_credentials.email": user.UserCredentials.Email}
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
	createdUser.Password = nil

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
	var updateBody types.User
	err := c.BodyParser(&updateBody)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(updateBody)
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

	// Update the user document in MongoDB
	collection, _ := db.GetCollection("users")
	filter := bson.M{"_id": objId}
	var existingUser types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&existingUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	if updateBody.Password != nil {
		Password, err := bcrypt.GenerateFromPassword([]byte(*updateBody.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error hashing password: " + err.Error())
		}
		passwordString := string(Password)
		updateBody.Password = &passwordString
	} else {
		updateBody.Password = existingUser.Password
	}

	// check unique email
	filter = bson.M{"user_credentials.email": updateBody.Email}
	var userByEmail types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&userByEmail)

	if userByEmail.ID != objId {
		if err == nil {
			return fmt.Errorf("Email address already in use")
		}
	}

	userRole := c.Locals("userRole")
	if userRole != "admin" &&
		c.Locals("user_id") != existingUser.ID.Hex() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	if userRole != "admin" {
		specialist := types.Specialist
		updateBody.Role = &specialist
	} else if updateBody.Role == nil {
		updateBody.Role = existingUser.Role
	}

	filter = bson.M{"_id": objId}
	// TODO: don't update all fields
	update := bson.M{"$set": updateBody}
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

	fieldsToUpdate := []string{"Password"}
	utils.UpdateResultForUserRole(&updatedUser, fieldsToUpdate)

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
	// TODO: dont send projects for author
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

	fieldsToUpdate := []string{"Password"}
	utils.UpdateResultForUserRole(&user, fieldsToUpdate)

	userRole := c.Locals("userRole")

	if userRole != "admin" {
		fieldsToUpdate := []string{"Role"}
		utils.UpdateResultForUserRole(&user, fieldsToUpdate)
	}

	userId := c.Locals("user_id")
	var userObjId primitive.ObjectID
	if userId != nil {
		userObjId, err = primitive.ObjectIDFromHex(userId.(string))
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString("Invalid ID")
		}
	}

	// fmt.Println((user.ConfirmedContactsRequests[userObjId] == "" ||
	// 	user.ConfirmedApplications[userObjId] == primitive.NilObjectID))
	// fmt.Println(user.ConfirmedApplications[userObjId] == primitive.NilObjectID)

	// hide fields for anon user
	if userId == nil {
		fieldsToUpdate := []string{
			"Contacts", "ContactsRequest", "UserProjects", "UserCredentials",
			"Location", "Language",
		}
		utils.UpdateResultForUserRole(&user, fieldsToUpdate)

		return c.Status(http.StatusOK).JSON(user)
	}

	// hide fields for not admin, not owner and not approved requester
	if userRole != "admin" &&
		userId != user.ID {
		if user.ConfirmedContactsRequests[userObjId] == "" &&
			user.ConfirmedApplications[userObjId] == primitive.NilObjectID {
			fieldsToUpdate := []string{
				"Contacts", "ContactsRequest", "UserProjects", "UserCredentials",
				"Location", "Language",
			}
			utils.UpdateResultForUserRole(&user, fieldsToUpdate)

			return c.Status(http.StatusOK).JSON(user)
		}
	}

	// check access to requests and projects
	// can I show these fields for confirmed user?
	// fmt.Println(user.ConfirmedContactsRequests[userObjId])
	// if userRole != "admin" ||
	// 	user.ConfirmedApplications[userObjId] ||
	// 	user.ConfirmedContactsRequests[userObjId] != "" ||
	// 	userId != user.ID {
	// 	fmt.Println(3, "admin or confirmed or user")
	// 	fmt.Println(userRole != "admin")
	// 	fmt.Println(userId != user.ID)

	// 	fieldsToUpdate := []string{"ContactsRequest", "UserProjects", "Contacts", "UserCredentials", "Location"}
	// 	utils.UpdateResultForUserRole(&user, fieldsToUpdate)
	// }

	// check access to contacts
	// if userRole != "admin" ||
	// 	user.ConfirmedContactsRequests[userObjId] != "" ||
	// 	userId != user.ID {
	// 	fmt.Println(4)
	// 	fieldsToUpdate := []string{"Contacts", "UserCredentials"}
	// 	utils.UpdateResultForUserRole(&user, fieldsToUpdate)
	// }

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

	userRole := c.Locals("userRole")

	if userRole != "admin" {
		filter["user_body.role"] = "specialist"
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
		user.Password = nil
		users = append(users, user)
	}

	if userRole != "admin" {
		fieldsToUpdate := []string{"Role", "Contacts", "ContactsRequest", "UserProjects", "UserCredentials", "Location"}
		utils.UpdateResultsForUserRole(users, fieldsToUpdate)
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
	if c.Locals("userRole") != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

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

// RequestContacts sends a contact request to another user.
// @Summary Send contact request
// @Description Sends a contact request to the specified user.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body types.RequestMessage true "Request body"
// @Success 200 {string} string "Contact request added successfully."
// @Failure 400 {string} string "Invalid project ID or user ID"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error connecting to database or updating user"
// @Router /users/request-contacts/{id} [post]
func RequestContacts(c *fiber.Ctx) error {
	var rm types.RequestMessage
	err := c.BodyParser(&rm)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(rm)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error validating user: " + err.Error())
	}

	collection, err := db.GetCollection("users")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}

	// Get the project ID from the URL path parameter
	approverId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid project ID")
	}

	requesterId, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// get requester
	filter := bson.M{"_id": requesterId}
	var requester types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&requester)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	// get approver
	filter = bson.M{"_id": approverId}
	var approver types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&approver)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	var msg string
	if approver.IncomingContactRequests[requesterId] != "" {
		delete(approver.IncomingContactRequests, requesterId)
		delete(requester.OutgoingContactRequests, approverId)
		msg = "Contact request deleted successfully."
	} else {
		if approver.IncomingContactRequests == nil {
			approver.IncomingContactRequests = make(map[primitive.ObjectID]string)
		}
		if requester.OutgoingContactRequests == nil {
			requester.OutgoingContactRequests = make(map[primitive.ObjectID]string)
		}

		approver.IncomingContactRequests[requesterId] = rm.Message
		requester.OutgoingContactRequests[approverId] = rm.Message
		msg = "Contact request added successfully."
	}

	// update approver
	filter = bson.M{"_id": approverId}
	update := bson.M{"$set": approver}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	// update requester
	filter = bson.M{"_id": requesterId}
	update = bson.M{"$set": requester}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.SendString(msg)
}

// ApproveContactsRequest approves a contact request from another user.
// @Summary Approve contact request
// @Description Approves a contact request from the specified user.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security ApiKeyAuth
// @Success 200 {string} string "Done"
// @Failure 400 {string} string "Invalid project ID or user ID"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error connecting to database or updating user"
// @Router /users/approve-contacts-request/{id} [get]
func ApproveContactsRequest(c *fiber.Ctx) error {
	collection, err := db.GetCollection("users")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}

	// Get the project ID from the URL path parameter
	requesterId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid project ID")
	}

	approverId, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// get approver
	filter := bson.M{"_id": approverId}
	var user types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	if user.IncomingContactRequests[requesterId] != "" {
		if user.ConfirmedContactsRequests == nil {
			user.ConfirmedContactsRequests = make(map[primitive.ObjectID]string)
		}
		user.ConfirmedContactsRequests[requesterId] = user.IncomingContactRequests[requesterId]
		delete(user.IncomingContactRequests, requesterId)
	}

	// update approver
	filter = bson.M{"_id": approverId}
	update := bson.M{"$set": user}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	return c.SendString("Done")
}

// RejectContactsRequest rejects a contact request from another user.
// @Summary Reject contact request
// @Description Rejects a contact request from the specified user.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security ApiKeyAuth
// @Success 200 {string} string "Done"
// @Failure 400 {string} string "Invalid project ID or user ID"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error connecting to database or updating user"
// @Router /users/reject-contacts-request/{id} [get]
func RejectContactsRequest(c *fiber.Ctx) error {
	collection, err := db.GetCollection("users")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}

	// Get the project ID from the URL path parameter
	requesterId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid project ID")
	}

	userId, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// get approver
	filter := bson.M{"_id": userId}
	var user types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	if user.IncomingContactRequests[requesterId] != "" {
		if user.BlockedUsers == nil {
			user.BlockedUsers = make(map[primitive.ObjectID]string)
		}

		user.BlockedUsers[requesterId] = user.IncomingContactRequests[requesterId]
		delete(user.IncomingContactRequests, requesterId)
	}

	// update approver
	filter = bson.M{"_id": userId}
	update := bson.M{"$set": user}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	return c.SendString("Done")
}

// ApproveProjectRequest approves a project request for the user.
// @Summary Approve project request
// @Description Approves a project request for the user.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {string} string "Request approved successfully."
// @Failure 400 {string} string "Invalid project ID or user ID"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error connecting to database or updating user"
// @Router /users/approve/{id} [get]
func ApproveProjectRequest(c *fiber.Ctx) error {
	userCollection, err := db.GetCollection("users")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}

	projectCollection, err := db.GetCollection("projects")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}

	requesterId := c.Params("id")
	requesterObjId, err := primitive.ObjectIDFromHex(requesterId)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid project ID")
	}

	approverId, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// get approver
	filter := bson.M{"_id": approverId}
	var user types.User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	// get project
	projectId := user.ProjectsApplications[requesterObjId]
	filter = bson.M{"_id": projectId}
	var project types.Project
	err = projectCollection.FindOne(context.TODO(), filter).Decode(&project)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated project: " + err.Error())
	}

	// TODO: when I cant find requester: Error retrieving updated project: mongo: no documents in result
	// TODO: validation of unexisted requster
	// if user.ProjectsApplications[requesterObjId] {

	// }
	// primitive.NilObjectID

	delete(user.ProjectsApplications, requesterObjId)
	if user.ConfirmedApplications == nil {
		user.ConfirmedApplications = make(map[primitive.ObjectID]primitive.ObjectID)
	}
	user.ConfirmedApplications[requesterObjId] = projectId
	delete(project.Applicants, requesterObjId)
	if project.SuccessfulApplicants == nil {
		project.SuccessfulApplicants = make(map[primitive.ObjectID]bool)
	}
	project.SuccessfulApplicants[requesterObjId] = true

	// update project
	projectFilter := bson.M{"_id": projectId}
	projectUpdate := bson.M{"$set": project}
	_, err = projectCollection.UpdateOne(context.TODO(), projectFilter, projectUpdate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	// update approver
	filter = bson.M{"_id": approverId}
	update := bson.M{"$set": user}
	_, err = userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	return c.SendString("Done")
}

// RejectProjectRequest rejects a project request for the user.
// @Summary Reject project request
// @Description Rejects a project request for the user.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {string} string "Request rejected successfully."
// @Failure 400 {string} string "Invalid project ID or user ID"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Error connecting to database or updating user"
// @Router /users/reject/{id} [get]
func RejectProjectRequest(c *fiber.Ctx) error {
	userCollection, err := db.GetCollection("users")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}
	projectCollection, err := db.GetCollection("projects")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}

	// Get the project ID from the URL path parameter
	requesterId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid project ID")
	}

	userId, err := primitive.ObjectIDFromHex(c.Locals("user_id").(string))
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// get approver
	filter := bson.M{"_id": userId}
	var user types.User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	// get project
	projectId := user.ProjectsApplications[requesterId]
	filter = bson.M{"_id": projectId}
	var project types.Project
	err = projectCollection.FindOne(context.TODO(), filter).Decode(&project)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated project: " + err.Error())
	}

	// TODO: handle one applicant to different project

	// if user.ProjectsApplications[incomingRequestUserId] {
	// 	delete(user.ProjectsApplications, incomingRequestUserId)
	// 	user.RejectedApplicants[incomingRequestUserId] = true
	// }
	delete(user.ProjectsApplications, requesterId)
	if user.RejectedApplicants == nil {
		user.RejectedApplicants = make(map[primitive.ObjectID]primitive.ObjectID)
	}
	user.RejectedApplicants[requesterId] = projectId

	delete(project.Applicants, requesterId)
	if project.RejectedApplicants == nil {
		project.RejectedApplicants = make(map[primitive.ObjectID]bool)
	}
	project.RejectedApplicants[requesterId] = true
	fmt.Println(project.Applicants)

	// update project
	projectFilter := bson.M{"_id": projectId}
	projectUpdate := bson.M{"$set": project}
	_, err = projectCollection.UpdateOne(context.TODO(), projectFilter, projectUpdate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	// update approver
	filter = bson.M{"_id": userId}
	update := bson.M{"$set": user}
	_, err = userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	return c.SendString("Done")
}

func UpdatePassword(c *fiber.Ctx) error {
	var requestBody types.PasswordUpdate
	err := c.BodyParser(&requestBody)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(requestBody)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("error validating user: " + err.Error())
	}

	userId := c.Locals("user_id").(string)
	objId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing user ID: " + err.Error())
	}

	collection, _ := db.GetCollection("users")
	filter := bson.M{"_id": objId}

	var user types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	if userId != user.ID.Hex() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	// Comparing the password with the hash
	err = bcrypt.CompareHashAndPassword(
		[]byte(*user.Password),
		[]byte(*requestBody.Password))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "wrong credentials",
		})
	}

	Password, err := bcrypt.GenerateFromPassword([]byte(*requestBody.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error hashing password: " + err.Error())
	}
	passwordString := string(Password)
	user.Password = &passwordString

	filter = bson.M{"_id": objId}
	update := bson.M{"$set": user}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	return c.SendString("Password successfully updated")
}

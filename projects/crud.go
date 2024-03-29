package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"henar-backend/db"
	"henar-backend/notifications"
	"henar-backend/sentry"
	"henar-backend/types"
	"henar-backend/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
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
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("Project not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error updating project: " + err.Error())
	}

	// Remove the fields if the user is not admin or author
	if c.Locals("userRole") != "admin" &&
		c.Locals("user_id") != result.CreatedBy.Hex() {
		fieldsToUpdate := []string{"ModerationStatus", "ReasonOfReject", "Applicants", "RejectApplicant"}
		utils.UpdateResultForUserRole(&result, fieldsToUpdate)
	}

	// Marshal the project struct to JSON format
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set("Content-Type", "application/json")
	c.Status(http.StatusOK)
	_, err = c.Write(jsonBytes)
	if err != nil {
		sentry.SentryHandler(err)
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
// @Param status query string false "Project statuses"
// @Param help query string false "How to help the project"
func GetProjects(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	// Get the filter and options for the query
	findOptions, err := utils.GetPaginationOptions(c)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid pagination parameters")
	}

	filter, err := utils.GetFilter(c)
	if err != nil {
		sentry.SentryHandler(err)
		errMsg := fmt.Sprintf("Error getting projects filter: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString(errMsg)
	}

	userRole := c.Locals("userRole")

	if userRole != "admin" {
		filter["moderation_status"] = "approved"
	}

	sort := utils.GetSort(c)
	if len(sort) != 0 {
		findOptions.SetSort(sort)
	}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error finding projects")
	}

	// Get the results from the cursor
	var results []types.Project
	if err := cursor.All(context.TODO(), &results); err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error querying database: " + err.Error())
	}

	// Remove the fields if the user is not admin
	if c.Locals("userRole") != "admin" {
		fieldsToUpdate := []string{"ModerationStatus", "ReasonOfReject", "Applicants", "RejectApplicant"}
		utils.UpdateResultsForUserRole(results, fieldsToUpdate)
	}

	// Marshal the result to JSON
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set("Content-Type", "application/json")
	c.Status(http.StatusOK)
	_, err = c.Write(jsonBytes)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error writing response: " + err.Error())
	}
	return nil
}

func GetSelfProjects(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	userId := c.Locals("user_id")

	if userId == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "not authorized",
		})
	}
	objId, _ := primitive.ObjectIDFromHex(userId.(string))

	// Get the filter and options for the query
	findOptions, err := utils.GetPaginationOptions(c)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid pagination parameters")
	}

	filter := bson.M{"created_by": objId}

	sort := utils.GetSort(c)
	if len(sort) != 0 {
		findOptions.SetSort(sort)
	}

	// Query the database and get the cursor
	cursor, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error finding projects")
	}

	// Get the results from the cursor
	var results []types.Project
	if err := cursor.All(context.TODO(), &results); err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error querying database: " + err.Error())
	}

	// Marshal the result to JSON
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set("Content-Type", "application/json")
	c.Status(http.StatusOK)
	_, err = c.Write(jsonBytes)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error writing response: " + err.Error())
	}
	return nil
}

func GetUserProjects(c *fiber.Ctx) error {
	projectsCollection, _ := db.GetCollection("projects")
	usersCollection, _ := db.GetCollection("users")

	userId := c.Locals("user_id")

	if userId == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "not authorized",
		})
	}

	id := c.Params("id")
	objId, _ := primitive.ObjectIDFromHex(id)

	filter := bson.M{"_id": objId}

	var user types.User
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}
	// Get the filter and options for the query
	findOptions, err := utils.GetPaginationOptions(c)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid pagination parameters")
	}

	userConfirmedProjectsIds := make([]primitive.ObjectID, 0, len(user.ConfirmedApplications))

	for _, _id := range user.ConfirmedApplications {
		userConfirmedProjectsIds = append(userConfirmedProjectsIds, _id)
	}
	fmt.Println(userConfirmedProjectsIds)

	filter = bson.M{"_id": bson.M{"$in": userConfirmedProjectsIds}}

	sort := utils.GetSort(c)
	if len(sort) != 0 {
		findOptions.SetSort(sort)
	}

	// Query the database and get the cursor
	cursor, err := projectsCollection.Find(context.TODO(), filter)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error finding projects")
	}

	// Get the results from the cursor
	var results []types.Project
	if err := cursor.All(context.TODO(), &results); err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error querying database: " + err.Error())
	}

	// Marshal the result to JSON
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error encoding JSON: " + err.Error())
	}

	// Set the response headers and write the response body
	c.Set("Content-Type", "application/json")
	c.Status(http.StatusOK)
	_, err = c.Write(jsonBytes)
	if err != nil {
		sentry.SentryHandler(err)
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
	projectsCollection, _ := db.GetCollection("projects")
	usersCollection, _ := db.GetCollection("users")

	// Parse request body into project struct
	var project types.Project
	err := c.BodyParser(&project)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(project)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error retrieving created project: " + err.Error())
	}

	userId := c.Locals("user_id").(string)
	userObjId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	userFilter := bson.M{"_id": userObjId}
	var user types.User
	err = usersCollection.FindOne(context.TODO(), userFilter).Decode(&user)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	project.CreatedBy = userObjId
	pending := types.Pending
	project.ModerationStatus = &pending

	slugText := utils.CreateSlug(project.Title)
	views := int64(0)
	project.Slug = &slugText
	project.Views = &views
	project.Applicants = make(map[primitive.ObjectID]bool)
	project.SuccessfulApplicants = make(map[primitive.ObjectID]bool)

	// Insert project document into MongoDB
	result, err := projectsCollection.InsertOne(context.TODO(), project)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error creating project: " + err.Error())
	}

	// Get the ID of the inserted project document
	projectObjId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated project from MongoDB
	filter := bson.M{"_id": projectObjId}
	var createdProject types.Project
	err = projectsCollection.FindOne(context.TODO(), filter).Decode(&createdProject)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated project: " + err.Error())
	}

	if user.CreatedProjects == nil {
		user.CreatedProjects = make(map[primitive.ObjectID]bool)
	}
	user.CreatedProjects[createdProject.ID] = true

	update := bson.M{"$set": user}
	_, err = usersCollection.UpdateOne(context.TODO(), userFilter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
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
// @Router /v1/projects/{id} [patch]
func UpdateProject(c *fiber.Ctx) error {
	collection, _ := db.GetCollection("projects")

	// Get the project ID from the URL path parameter
	id := c.Params("id")
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Parse the request body into a updateBody struct
	var updateBody types.Project
	err = c.BodyParser(&updateBody)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	var views int64

	if updateBody.Views == nil {
		views = int64(0)
	} else {
		views = *updateBody.Views
	}

	// Validate the required fields
	v := validator.New()
	v.RegisterValidation("enum", types.ValidateEnum)
	err = v.Struct(updateBody)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Validation error: " + err.Error())
	}

	// Find the project document from MongoDB
	var project types.Project
	err = collection.FindOne(context.TODO(), bson.M{"_id": objId}).Decode(&project)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("Project not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error finding project: " + err.Error())
	}
	if c.Locals("userRole") != "admin" &&
		c.Locals("user_id") != project.CreatedBy.Hex() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Permission or ownership error",
		})
	}

	if c.Locals("userRole") != "admin" {
		// owner can't edit the following fields
		if updateBody.ModerationStatus != nil ||
			updateBody.ReasonOfReject != nil ||
			updateBody.Views != nil ||
			updateBody.Slug != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Permission or ownership error",
			})
		}
		pending := types.Pending

		updateBody.ModerationStatus = &pending
	}

	slugText := utils.CreateSlug(updateBody.Title)
	updateBody.Slug = &slugText
	updateBody.Views = &views

	// Update the project document in MongoDB
	filter := bson.M{"_id": objId}
	update := bson.M{"$set": updateBody}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating project: " + err.Error())
	}

	// Retrieve the updated project from MongoDB
	filter = bson.M{"_id": objId}
	var updatedProject types.Project
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedProject)
	if err != nil {
		sentry.SentryHandler(err)
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
func DeleteProject(store *session.Store) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		projectsCollection, _ := db.GetCollection("projects")

		// Get the project ID from the URL path parameter
		id := c.Params("id")
		projectObjId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			sentry.SentryHandler(err)
			return c.Status(http.StatusBadRequest).SendString("Invalid ID")
		}

		// Find the project document from MongoDB
		var project types.Project
		err = projectsCollection.FindOne(context.TODO(), bson.M{"_id": projectObjId}).Decode(&project)
		if err != nil {
			sentry.SentryHandler(err)
			if err == mongo.ErrNoDocuments {
				return c.Status(http.StatusNotFound).SendString("Project not found")
			}
			return c.Status(http.StatusInternalServerError).SendString("Error finding project: " + err.Error())
		}

		// Check if the user has access to delete the project
		userId := c.Locals("user_id").(string)

		if c.Locals("userRole") != "admin" &&
			userId != project.CreatedBy.Hex() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Permission or ownership error",
			})
		}
		userObjId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			sentry.SentryHandler(err)
			return c.Status(http.StatusBadRequest).SendString("Invalid ID")
		}

		// update user projects list
		// TODO: delete for all applicants
		usersCollection, _ := db.GetCollection("users")
		userFilter := bson.M{"_id": userObjId}
		var user types.User
		err = usersCollection.FindOne(context.TODO(), userFilter).Decode(&user)
		if err != nil {
			sentry.SentryHandler(err)
			if err == mongo.ErrNoDocuments {
				return c.Status(http.StatusNotFound).SendString("User not found")
			}
			return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
		}
		delete(user.CreatedProjects, projectObjId)
		update := bson.M{"$set": user}
		_, err = usersCollection.UpdateOne(context.TODO(), userFilter, update)
		if err != nil {
			sentry.SentryHandler(err)
			return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
		}

		// Delete project document from MongoDB
		projectFilter := bson.D{{Key: "_id", Value: projectObjId}}
		result, err := projectsCollection.DeleteOne(context.TODO(), projectFilter)
		if err != nil {
			sentry.SentryHandler(err)
			return c.Status(http.StatusInternalServerError).SendString("Error deleting project: " + err.Error())
		}

		// Check if any documents were deleted
		if result.DeletedCount == 0 {
			return c.Status(http.StatusNotFound).SendString("Project not found")
		}

		return c.SendString("Project deleted successfully")
	}
}

// RespondToProject responds to a project by adding the current user as an applicant.
// @Summary Respond to a project
// @Description Adds the current user as an applicant to the specified project.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} types.Project
// @Failure 400 {string} string "Invalid ID or project ID"
// @Failure 500 {string} string "Error connecting to database or updating/retrieving project"
// @Router /projects/respond/{id} [get]
func RespondToProject(c *fiber.Ctx) error {
	collection, err := db.GetCollection("projects")
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}

	requsterId := c.Locals("user_id").(string)
	requesterObjId, err := primitive.ObjectIDFromHex(requsterId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Get the project ID from the URL path parameter
	projectId := c.Params("id")
	projectObjId, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid project ID")
	}

	// get project
	filter := bson.M{"_id": projectObjId}
	var project types.Project
	err = collection.FindOne(context.TODO(), filter).Decode(&project)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated project: " + err.Error())
	}

	if project.Applicants == nil {
		project.Applicants = make(map[primitive.ObjectID]bool)
	}
	project.Applicants[requesterObjId] = true

	// Update the project document in MongoDB
	filter = bson.M{"_id": projectObjId}
	update := bson.M{"$set": project}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating project: " + err.Error())
	}

	// update approver
	usersCollection, _ := db.GetCollection("users")

	approverId := project.CreatedBy

	approverFilter := bson.M{"_id": approverId}
	var approver types.User
	err = usersCollection.FindOne(context.TODO(), approverFilter).Decode(&approver)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	var requester types.User
	requesterFilter := bson.M{"_id": requesterObjId}
	err = usersCollection.FindOne(context.TODO(), requesterFilter).Decode(&requester)

	if approver.ProjectsApplications == nil {
		approver.ProjectsApplications = make(map[primitive.ObjectID]primitive.ObjectID)
	}
	approver.ProjectsApplications[requesterObjId] = projectObjId

	approverUpdate := bson.M{"$set": approver}
	_, err = usersCollection.UpdateOne(context.TODO(), approverFilter, approverUpdate)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	notificationBody := types.NotificationBody{
		PersonID:       requesterObjId,
		PersonFullName: requester.FirstName + " " + requester.LastName,
		Avatar:         requester.Avatar,
	}
	err = notifications.CreateNotification(types.ProjectRequest, approverId, notificationBody)

	if err != nil {
		sentry.SentryHandler(err)
		c.Status(http.StatusInternalServerError).SendString("Error creating notification:" + err.Error())
	}

	// Set the response headers and write the response body
	return c.SendString("Response sended successfully")
}

// TODO: respond/cancel delete applicants for public request

// CancelProjectApplication cancels the user's application for a project.
// @Summary Cancel project application
// @Description Cancels the user's application for the specified project.
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} types.Project
// @Failure 400 {string} string "Invalid ID or project ID"
// @Failure 500 {string} string "Error connecting to database or updating/retrieving project"
// @Router /projects/cancel/{id} [get]
func CancelProjectApplication(c *fiber.Ctx) error {
	collection, err := db.GetCollection("projects")
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}

	// TODO: error on update in mongo

	requsterId := c.Locals("user_id").(string)
	requesterObjId, err := primitive.ObjectIDFromHex(requsterId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Get the project ID from the URL path parameter
	projectId := c.Params("id")
	projectObjId, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid project ID")
	}

	// get project
	filter := bson.M{"_id": projectObjId}
	var project types.Project
	err = collection.FindOne(context.TODO(), filter).Decode(&project)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated project: " + err.Error())
	}

	delete(project.Applicants, requesterObjId)
	fmt.Println(project.Applicants)

	// Update the project document in MongoDB
	filter = bson.M{"_id": projectObjId}
	update := bson.M{"$set": bson.M{"applicants": project.Applicants}}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating project: " + err.Error())
	}

	// Retrieve the updated project from MongoDB
	filter = bson.M{"_id": projectObjId}
	var updatedProject types.Project
	err = collection.FindOne(context.TODO(), filter).Decode(&updatedProject)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated project: " + err.Error())
	}

	// update approver
	usersCollection, _ := db.GetCollection("users")

	approverId := updatedProject.CreatedBy

	approverFilter := bson.M{"_id": approverId}
	var approver types.User
	err = usersCollection.FindOne(context.TODO(), approverFilter).Decode(&approver)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	delete(approver.ProjectsApplications, requesterObjId)

	approverUpdate := bson.M{"$set": approver}
	_, err = usersCollection.UpdateOne(context.TODO(), approverFilter, approverUpdate)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.SendString("Response canceled successfully")
}

func ApproveApplicant(c *fiber.Ctx) error {
	collection, err := db.GetCollection("projects")
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}

	var ids map[string]string
	err = c.BodyParser(&ids)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	requsterId := c.Locals("user_id").(string)
	requesterObjId, err := primitive.ObjectIDFromHex(requsterId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Get the project ID from the URL path parameter
	projectId := ids["projectId"]

	projectObjId, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid project ID")
	}

	// get project
	filter := bson.M{"_id": projectObjId}
	var project types.Project
	err = collection.FindOne(context.TODO(), filter).Decode(&project)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated project: " + err.Error())
	}

	applicantId := ids["applicantId"]
	applicantObjId, err := primitive.ObjectIDFromHex(applicantId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid applicant ID")
	}

	if project.SuccessfulApplicants == nil {
		project.SuccessfulApplicants = make(map[primitive.ObjectID]bool)
	}
	project.SuccessfulApplicants[applicantObjId] = true

	projectApplicants := project.Applicants
	delete(projectApplicants, applicantObjId)
	project.Applicants = projectApplicants

	// Update the project document in MongoDB
	filter = bson.M{"_id": projectObjId}
	update := bson.M{"$set": project}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating project: " + err.Error())
	}

	// update applicant
	usersCollection, _ := db.GetCollection("users")

	applicantFilter := bson.M{"_id": applicantObjId}
	var applicant types.User
	err = usersCollection.FindOne(context.TODO(), applicantFilter).Decode(&applicant)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	if applicant.ConfirmedApplications == nil {
		applicant.ConfirmedApplications = make(map[primitive.ObjectID]primitive.ObjectID)
	}
	applicant.ConfirmedApplications[requesterObjId] = projectObjId

	applicantUpdate := bson.M{"$set": applicant}
	_, err = usersCollection.UpdateOne(context.TODO(), applicantFilter, applicantUpdate)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	notificationBody := types.NotificationBody{
		ProjectID:    *project.Slug,
		ProjectTitle: project.Title.En,
		Avatar:       applicant.Avatar,
	}
	err = notifications.CreateNotification(types.ApproveApplicant, applicantObjId, notificationBody)
	if err != nil {
		sentry.SentryHandler(err)
		c.Status(http.StatusInternalServerError).SendString("Error creating notification:" + err.Error())
	}

	// Set the response headers and write the response body
	return c.SendString("Response sended successfully")
}

func RejectApplicant(c *fiber.Ctx) error {
	collection, err := db.GetCollection("projects")
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error connecting to database: " + err.Error())
	}

	var ids map[string]string
	err = c.BodyParser(&ids)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	requsterId := c.Locals("user_id").(string)
	requesterObjId, err := primitive.ObjectIDFromHex(requsterId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid ID")
	}

	// Get the project ID from the URL path parameter
	projectId := ids["projectId"]

	projectObjId, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid project ID")
	}

	// get project
	filter := bson.M{"_id": projectObjId}
	var project types.Project
	err = collection.FindOne(context.TODO(), filter).Decode(&project)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving updated project: " + err.Error())
	}

	applicantId := ids["applicantId"]
	applicantObjId, err := primitive.ObjectIDFromHex(applicantId)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Invalid applicant ID")
	}

	if project.RejectedApplicants == nil {
		project.RejectedApplicants = make(map[primitive.ObjectID]bool)
	}
	project.RejectedApplicants[applicantObjId] = true

	projectApplicants := project.Applicants
	delete(projectApplicants, applicantObjId)
	project.Applicants = projectApplicants

	fmt.Println(project.Applicants)

	// Update the project document in MongoDB
	filter = bson.M{"_id": projectObjId}
	update := bson.M{"$set": project}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating project: " + err.Error())
	}

	// update applicant
	usersCollection, _ := db.GetCollection("users")

	applicantFilter := bson.M{"_id": applicantObjId}
	var applicant types.User
	err = usersCollection.FindOne(context.TODO(), applicantFilter).Decode(&applicant)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	if applicant.RejectedApplicants == nil {
		applicant.RejectedApplicants = make(map[primitive.ObjectID]primitive.ObjectID)
	}
	applicant.RejectedApplicants[requesterObjId] = projectObjId

	applicantUpdate := bson.M{"$set": applicant}
	_, err = usersCollection.UpdateOne(context.TODO(), applicantFilter, applicantUpdate)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	// Set the response headers and write the response body
	return c.SendString("Response sended successfully")
}

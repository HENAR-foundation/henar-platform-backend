package routes

import (
	"context"
	"fmt"
	"henar-backend/db"
	"henar-backend/sentry"
	"henar-backend/types"
	"henar-backend/utils"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

// TODO: update create user and sing up to db method
func SignUp(c *fiber.Ctx) error {
	var uc types.User
	err := c.BodyParser(&uc)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}
	if uc.Password == nil {
		return c.Status(http.StatusBadRequest).SendString("Password is required")
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(uc)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("error validating user: " + err.Error())
	}

	// Hash the password
	Password, err := bcrypt.GenerateFromPassword(
		[]byte(*uc.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		sentry.SentryHandler(err)
		return fmt.Errorf("Error hashing password: %w", err)
	}
	passwordString := string(Password)
	specialist := types.Specialist
	user := types.User{
		IsActivated: false,
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
				ProjectsApplications:  make(map[primitive.ObjectID]primitive.ObjectID),
				ConfirmedApplications: make(map[primitive.ObjectID]primitive.ObjectID),
				RejectedApplicants:    make(map[primitive.ObjectID]primitive.ObjectID),
				CreatedProjects:       make(map[primitive.ObjectID]bool),
			},
		},
	}
	v.Struct(user)

	// Check if the email address is already in use
	collection, _ := db.GetCollection("users")
	filter := bson.M{"user_credentials.email": user.UserCredentials.Email}
	var existingUser types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&existingUser)
	if err == nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Email address already in use")
	}

	// Insert user document into MongoDB
	result, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		sentry.SentryHandler(err)
		return fmt.Errorf("Error creating user: ", err)
	}

	// Get the ID of the inserted user document
	objId := result.InsertedID.(primitive.ObjectID)

	// Retrieve the updated user from MongoDB
	filter = bson.M{"_id": objId}
	var createdUser types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&createdUser)
	if err != nil {
		sentry.SentryHandler(err)
		return fmt.Errorf("Error retrieving created user: ", err)
	}

	// Set the response headers and write the response body
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"userId": createdUser.ID,
	})
}

func SignIn(c *fiber.Ctx) error {
	var uc types.UserCredentials

	err := c.BodyParser(&uc)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "kind an error: " + err.Error(),
		})
	}

	collection, _ := db.GetCollection("users")
	filter := bson.M{"user_credentials.email": uc.Email}
	var user types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&user)

	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "wrong credentials",
		})
	}

	// Comparing the password with the hash
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*uc.Password))
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "wrong credentials",
		})
	}

	sess, sessErr := store.Get(c)
	if sessErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "kind an error: " + err.Error(),
		})
	}

	sess.Set(AUTH_KEY, true)
	sess.Set(USER_ID, user.ID.Hex())
	sess.Set(USER_ROLE, string(*user.Role))

	sessErr = sess.Save()
	if sessErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "kind an error: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "logged in",
	})
}

func SignOut(c *fiber.Ctx) error {
	sess, err := store.Get(c)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "logged out (no session)",
		})
	}

	err = sess.Destroy()
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "kind an error: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "logged out",
	})
}

func Check(c *fiber.Ctx) error {
	userId := c.Locals("user_id")
	if userId == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "not authorized",
		})
	}
	objId, _ := primitive.ObjectIDFromHex(userId.(string))

	collection, _ := db.GetCollection("users")
	filter := bson.M{"_id": objId}
	var user types.User
	err := collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	fieldsToUpdate := []string{"Password"}
	utils.UpdateResultForUserRole(&user, fieldsToUpdate)

	return c.Status(http.StatusOK).JSON(user)
}

// ForgotPassword rejects a project request for the user.
// @Summary Forgot password
// @Description Sends a password reset email to the user with the specified email address.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body types.ForgotPassword true "Forgot password request body"
// @Success 200 {string} string "You will receive a reset email if a user with that email exists"
// @Failure 400 {string} string "Error parsing request body or passwords do not match"
// @Failure 500 {string} string "Error retrieving user or updating user"
// @Router /auth/forgot-password [post]
func ForgotPassword(c *fiber.Ctx) error {
	var requestBody types.ForgotPassword
	err := c.BodyParser(&requestBody)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	collection, _ := db.GetCollection("users")
	filter := bson.M{"user_credentials.email": requestBody.Email}
	var user types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.SendString("You will receive a reset email if user with that email exist")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	// Hash the password
	passwordResetToken, err := bcrypt.GenerateFromPassword(
		[]byte(*&user.Email),
		bcrypt.DefaultCost,
	)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error hashing password: " + err.Error())
	}

	// Comparing the hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Email), []byte(passwordResetToken))
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ERROR TOKEN",
		})
	}

	// passwordResetToken := "resetToken"

	// user.PasswordResetToken = passwordResetToken
	// user.PasswordResetAt = time.Now().Add(time.Minute * 15)

	// filter = bson.M{"_id": user.ID}
	// update := bson.M{"$set": user}
	// _, err = collection.UpdateOne(context.TODO(), filter, update)
	// if err != nil {
	// 	sentry.SentryHandler(err)
	// 	return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	// }

	return c.SendString("You will receive a reset email if user with that email exist")
}

// ResetPassword rejects a project request for the user.
// @Summary Reset password
// @Description Resets the password for the user using the provided reset token.
// @Tags auth
// @Accept json
// @Produce json
// @Param resettoken path string true "Reset token"
// @Param request body types.ResetPassword true "Reset password request body"
// @Success 200 {string} string "Password successfully updated"
// @Failure 400 {string} string "The reset token is invalid or has expired, or error parsing request body or passwords do not match"
// @Failure 500 {string} string "Error retrieving user or updating user"
// @Router /auth/reset-password/{token} [post]
func ResetPassword(c *fiber.Ctx) error {
	var payload types.ResetPassword
	err := c.BodyParser(&payload)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(payload)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("error validating user: " + err.Error())
	}
	if *payload.Password != *payload.PasswordConfirm {
		return c.Status(http.StatusBadRequest).SendString("Passwords do not match")
	}

	resetToken := c.Params("token")

	collection, _ := db.GetCollection("users")
	filter := bson.M{
		"password_reset_token": resetToken,
		"password_reset_at":    bson.M{"$gt": time.Now()},
	}
	var user types.User
	err = collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		sentry.SentryHandler(err)
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusBadRequest).SendString("The reset token is invalid or has expired")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	Password, err := bcrypt.GenerateFromPassword(
		[]byte(*payload.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		sentry.SentryHandler(err)
		return fmt.Errorf("Error hashing password: %w", err)
	}

	passwordString := string(Password)

	user.Password = &passwordString
	// TODO: change token flow
	// user.PasswordResetToken = ""

	filter = bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating user: " + err.Error())
	}

	return c.SendString("Password successfully updated")
}

package routes

import (
	"context"
	"fmt"
	"henar-backend/db"
	"henar-backend/internal/email"
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
	IsEmailVerified := false
	user := types.User{
		IsActivated:     false,
		IsEmailVerified: &IsEmailVerified,
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

	// Create verification data for the new user and insert it into db
	verificationData, err := CreateVerificationData(createdUser.ID, createdUser.Email, "mail_confirmation")
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	// Send email for email verification
	mailjetClient := email.Init()
	err = mailjetClient.SendConfirmationEmail(verificationData)
	if err != nil {
		sentry.SentryHandler(err)
		return err
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

	err = checkVerificationStatus(user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "email not verified",
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

// @Summary Initiate password reset
// @Description Send a password reset email to the user
// @Tags auth
// @Accept json
// @Produce json
// @Param email body string true "User email"
// @Success 200 {object} map[string]string "Password reset email sent"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /auth/forgot-password [post]
func ForgotPassword(c *fiber.Ctx) error {
	var requestBody types.ForgotPassword
	err := c.BodyParser(&requestBody)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	user, err := GetUserByEmail(requestBody.Email, c)
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	verificationData, err := CreateVerificationData(user.ID, user.Email, "pass_reset")
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	mailjetClient := email.Init()
	err = mailjetClient.SendPasswordResetEmail(verificationData)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Password reset email sent"})
}

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
// @Router /auth/reset-password/ [post]
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

	token := payload.Token

	verificationData, err := FindVerificationDataByCode(token)
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	// Check if the verification data is expired.
	if verificationData.ExpiresAt.Before(time.Now()) {
		return c.Status(http.StatusBadRequest).SendString("Token has expired")
	}

	// Check if the verification data is of the correct type.
	if verificationData.Type != types.PassReset {
		return c.Status(http.StatusBadRequest).SendString("Incorrect token type")
	}

	// Hash the new password.
	Password, err := bcrypt.GenerateFromPassword(
		[]byte(*payload.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		sentry.SentryHandler(err)
		return fmt.Errorf("error hashing password: %w", err)
	}

	passwordString := string(Password)

	// Retrieve the User instance for the user ID.
	user, err := GetUserByID(verificationData.User, c)
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	// Update the user's password.
	user.Password = &passwordString

	// Save the User instance.
	_, err = SaveUser(user, c)
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	// Mark the verification data as used and save
	if err := MarkVerificationDataAsUsed(verificationData, c); err != nil {
		sentry.SentryHandler(err)
		return err
	}

	return c.SendString("Password successfully updated")
}

// VerifyEmail verifies the provided code
// @Summary Verify email
// @Description verifies the provided code then set the User state to verified and deletes verification data
// @Tags auth
// @Accept json
// @Produce json
// @Param secret_code path string true "Secret code"
// @Success 200 {string} string "Email confirmed successfully"
// @Failure 400 {string} string "The secret code is invalid or has expired"
// @Failure 500 {string} string "Error verifying email"
// @Router /auth/verify-email/ [post]
func VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("secret_code")

	v := validator.New()
	err := v.Var(token, "required,hexadecimal")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Error validating code " + err.Error()})
	}

	verificationData, err := ValidateVerificationData(token, c)
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	err = UpdateUserVerificationStatus(verificationData.User, true, c)
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	err = UseToken(verificationData.ID, c)
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	return c.SendString("verification succeeded")
}

// ResendVerificationEmail resends verification email
// @Summary resend verification email
// @Description updates verification data and resends email with code
// @Tags auth
// @Accept json
// @Produce json
// @Param email path string true "email"
// @Success 200 {string} string "email resent successfully"
// @Failure 400 {string} string "email not found"
// @Failure 500 {string} string "Error resending email"
// @Router /auth/verify-email/{token, email} [post]
func ResendVerificationEmail(c *fiber.Ctx) error {
	var payload types.ResendVerificationEmail
	err := c.BodyParser(&payload)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	v := validator.New()
	err = v.Struct(payload)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("error validating: " + err.Error())
	}

	token := payload.Token
	userEmail := payload.Email

	if userEmail == "" {
		// if no email provided then searching it by token
		verificationData, err := FindVerificationDataByCode(token)
		if err != nil {
			sentry.SentryHandler(err)
			return err
		}
		userEmail = verificationData.Email
	}

	// Generate new code and update verification data
	updatedVerificationData, err := UpdateVerificationData(userEmail, c)
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	// Send email for email verification
	mailjetClient := email.Init()
	err = mailjetClient.SendConfirmationEmail(updatedVerificationData)
	if err != nil {
		sentry.SentryHandler(err)
		return err
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "email resend successfully", "email": updatedVerificationData.Email})
}

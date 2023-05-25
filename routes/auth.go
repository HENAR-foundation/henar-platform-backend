package routes

import (
	"context"
	"fmt"
	"henar-backend/db"
	"henar-backend/types"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

// TODO: update create user and sing up to db method
// TODO: create user return password, fix it
func SignUp(c *fiber.Ctx) error {
	var uc types.UserTest
	err := c.BodyParser(&uc)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	// Validate the required fields
	v := validator.New()
	err = v.Struct(uc)
	if err != nil {
		return fmt.Errorf("error validating user: %w", err)
	}

	// Hash the password
	Password, err := bcrypt.GenerateFromPassword(
		[]byte(uc.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("Error hashing password: %w", err)
	}

	specialist := types.Specialist
	user := types.UserTest{
		UserCredentials: types.UserCredentials{
			Email:    uc.Email,
			Password: string(Password),
		},
		UserBody: types.UserBody{
			Role: &specialist,
			ContactsRequest: types.ContactsRequest{
				IncomingContactRequests:   make(map[primitive.ObjectID]bool),
				OutgoingContactRequests:   make(map[primitive.ObjectID]bool),
				ConfirmedContactsRequests: make(map[primitive.ObjectID]bool),
				BlockedUsers:              make(map[primitive.ObjectID]bool),
			},
			UserProjects: types.UserProjects{
				ProjectsApplications:  make(map[primitive.ObjectID]bool),
				ConfirmedApplications: make(map[primitive.ObjectID]bool),
				RejectedApplicants:    make(map[primitive.ObjectID]bool),
				CreatedProjects:       make(map[primitive.ObjectID]bool),
			},
		},
	}
	v.Struct(user)

	// Check if the email address is already in use
	collection, _ := db.GetCollection("users")
	filter := bson.M{"user_credentials.email": user.UserCredentials.Email}
	var existingUser types.UserTest
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
	var createdUser types.UserTest
	err = collection.FindOne(context.TODO(), filter).Decode(&createdUser)
	if err != nil {
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "kind an error: " + err.Error(),
		})
	}

	collection, _ := db.GetCollection("users")
	filter := bson.M{"user_credentials.email": uc.Email}
	var user types.UserTest
	err = collection.FindOne(context.TODO(), filter).Decode(&user)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "wrong credentials",
		})
	}

	// Comparing the password with the hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(uc.Password))
	if err != nil {
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
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "logged out (no session)",
		})
	}

	err = sess.Destroy()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "kind an error: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "logged out",
	})
}

func Check(c *fiber.Ctx) error {
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized",
		})
	}

	auth := sess.Get(AUTH_KEY)

	if auth != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "authorized",
		})
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized",
		})
	}
}

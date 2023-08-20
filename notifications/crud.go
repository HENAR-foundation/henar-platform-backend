package notifications

import (
	"context"
	"fmt"
	"henar-backend/db"
	"henar-backend/sentry"
	"henar-backend/types"
	"henar-backend/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetNotifications(c *fiber.Ctx) error {
	userId := c.Locals("user_id")

	if userId == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "not authorized",
		})
	}
	objId, _ := primitive.ObjectIDFromHex(userId.(string))

	usersCollection, _ := db.GetCollection("users")
	usersFilter := bson.M{"_id": objId}
	var user types.User
	err := usersCollection.FindOne(context.TODO(), usersFilter).Decode(&user)
	if err != nil {
		sentry.SentryHandler(err)

		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusNotFound).SendString("User not found")
		}
		return c.Status(http.StatusInternalServerError).SendString("Error retrieving user: " + err.Error())
	}

	notificationsCollection, _ := db.GetCollection("notifications")
	notificationsFilter := bson.M{"_id": bson.M{"$in": user.Notifications}}

	if len(user.Notifications) == 0 {
		c.Status(http.StatusOK).JSON(nil)
		return nil
	}

	cursor, err := notificationsCollection.Find(context.TODO(), notificationsFilter)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error finding notifications" + err.Error())
	}

	var results []types.NotificationResponse
	if err := cursor.All(context.TODO(), &results); err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusInternalServerError).SendString("Error querying database: " + err.Error())
	}
	c.Status(http.StatusOK).JSON(results)

	return nil
}

func ReadNotifications(c *fiber.Ctx) error {
	var body types.NotificationAcceptiongRequestBody
	err := c.BodyParser(&body)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing request body: " + err.Error())
	}

	notificationsIds, err := utils.ConvertStringArrayToObjectIDArray(body.NotificationsIds)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error parsing IDs")
	}

	collection, _ := db.GetCollection("notifications")
	filter := bson.M{"_id": bson.M{"$in": notificationsIds}}
	update := bson.M{"$set": bson.M{
		"status": types.Read,
	}}

	updatedNotifications, err := collection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(http.StatusBadRequest).SendString("Error updating notifications")
	}

	fmt.Println(updatedNotifications)
	c.Status(http.StatusOK).JSON(updatedNotifications)

	return nil
}

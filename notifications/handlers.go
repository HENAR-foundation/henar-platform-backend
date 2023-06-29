package notifications

import (
	"context"
	"errors"
	"henar-backend/db"
	"henar-backend/types"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateNotification(notificationType types.NotificationType, userId primitive.ObjectID, body types.NotificationBody) error {
	usersCollection, _ := db.GetCollection("users")
	notificationsCollection, _ := db.GetCollection("notifications")

	result, err := notificationsCollection.InsertOne(context.TODO(), types.Notification{
		ID:        primitive.NewObjectID(),
		CreatedAt: time.Now(),
		Status:    types.New,
		Type:      notificationType,
		User:      userId,
		Body:      body,
	})
	if err != nil {
		return errors.New("failed to create notification")
	}

	notificationId, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return errors.New("failed to create notification id")
	}

	filter := bson.M{"_id": userId}
	var user types.User
	err = usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return errors.New("failed to find user")
	}

	user.Notifications = append(user.Notifications, notificationId)
	update := bson.M{"$set": user}
	_, err = usersCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return errors.New("failed to update user")
	}

	return nil
}

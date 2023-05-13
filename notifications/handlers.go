package notifications

import (
	"henar-backend/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateNotificationToAuthentificatedUser(notification types.NotificationType) error {
	return nil
}

func CreateNotificationToAnotherUser(notification types.NotificationType, userId primitive.ObjectID) error {
	return nil
}

package routes

import (
	"context"
	"errors"
	"henar-backend/db"
	"henar-backend/sentry"
	"henar-backend/types"
	"time"

	"henar-backend/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateVerificationData(userId primitive.ObjectID, email string, resendAttempts ...int) (types.VerificationData, error) {
	code, _ := utils.RandomHex(16)
	expirationTime := time.Now().Add(24 * time.Hour)

	verificationData := types.VerificationData{
		ID:        primitive.NewObjectID(),
		User:      userId,
		Email:     email,
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: expirationTime,
	}

	if len(resendAttempts) > 0 {
		verificationData.ResendAttempts = resendAttempts[0]
	} else {
		verificationData.ResendAttempts = 0
	}

	verificationDataCollection, _ := db.GetCollection("verificationData")

	_, err := verificationDataCollection.InsertOne(context.TODO(), verificationData)

	if err != nil {
		sentry.SentryHandler(err)
		return types.VerificationData{}, errors.New("failed to create verification data")
	}

	filter := bson.M{"_id": verificationData.ID}
	err = verificationDataCollection.FindOne(context.TODO(), filter).Decode(&verificationData)
	if err != nil {
		sentry.SentryHandler(err)
		return types.VerificationData{}, errors.New("failed to retrieve inserted verification data")
	}

	return verificationData, nil
}

// ValidateVerificationData fetches verification data from the db using the token and validates it
func ValidateVerificationData(token string, c *fiber.Ctx) (types.VerificationData, error) {
	verificationDataCollection, _ := db.GetCollection("verificationData")

	filter := bson.M{"code": token}
	var verificationData types.VerificationData

	err := verificationDataCollection.FindOne(context.TODO(), filter).Decode(&verificationData)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			sentry.SentryHandler(err)
			return types.VerificationData{}, errors.New("verification data not found")
		}
		sentry.SentryHandler(err)
		return types.VerificationData{}, errors.New("error verificating data")
	}

	if verificationData.UsedAt != nil {
		return verificationData, errors.New("token has already been used")
	}

	if verificationData.ExpiresAt.Before(time.Now()) {
		return verificationData, errors.New("verification code has expired")
	}

	if verificationData.Code != token {
		return verificationData, errors.New("invalid verification code")
	}

	return verificationData, nil
}

func UseToken(tokenID primitive.ObjectID, c *fiber.Ctx) error {
	verificationDataCollection, _ := db.GetCollection("verificationData")

	filter := bson.M{"_id": tokenID}
	update := bson.M{"$set": bson.M{"used_at": time.Now()}}

	result, err := verificationDataCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return errors.New("error updating token")
	}

	// Check if the document was deleted
	if result.ModifiedCount == 0 {
		// Token not found in the database
		return errors.New("verification code not found")
	}

	return nil
}

// updateUserVerificationStatus updates the user's verification status
func UpdateUserVerificationStatus(userID primitive.ObjectID, status bool, c *fiber.Ctx) error {
	collection, _ := db.GetCollection("users")

	// Update the user's verification status
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"is_email_verified": status}}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		sentry.SentryHandler(err)
		return errors.New("error updating verification status")
	}

	// Check if the document was updated
	if result.ModifiedCount == 0 {
		// User not found in the database
		return errors.New("user not found")
	}

	return nil
}

func checkVerificationStatus(user types.User) error {
	if user.IsEmailVerified == nil || *user.IsEmailVerified {
		return nil
	}

	return errors.New("email not verified")
}

func UpdateVerificationData(email string, c *fiber.Ctx) (types.VerificationData, error) {
	const MaxResendLimit = 5

	verificationDataCollection, _ := db.GetCollection("verificationData")

	filter := bson.M{"email": email}
	var verificationData types.VerificationData

	err := verificationDataCollection.FindOne(context.TODO(), filter).Decode(&verificationData)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			sentry.SentryHandler(err)
			return types.VerificationData{}, errors.New("email not found")
		}
		sentry.SentryHandler(err)
		return types.VerificationData{}, errors.New("error updating verification data")
	}

	// Check if the resend limit is reached
	if verificationData.ResendAttempts >= MaxResendLimit {
		return types.VerificationData{}, errors.New("resend limit exceeded")
	}

	// Generate a new verification code
	newCode, _ := utils.RandomHex(16)

	update := bson.M{
		"$set": bson.M{
			"code":            newCode,
			"resend_attempts": verificationData.ResendAttempts + 1,
			"used_at":         nil,
			"created_at":      time.Now(),
			"expires_at":      time.Now().Add(24 * time.Hour),
		},
	}

	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	if err := verificationDataCollection.FindOneAndUpdate(
		context.TODO(),
		filter,
		update,
		options,
	).Decode(&verificationData); err != nil {
		sentry.SentryHandler(err)
		return types.VerificationData{}, err
	}

	return verificationData, nil
}

func FindVerificationDataByCode(code string) (types.VerificationData, error) {
	verificationDataCollection, _ := db.GetCollection("verificationData")

	filter := bson.M{"code": code}
	var verificationData types.VerificationData

	err := verificationDataCollection.FindOne(context.TODO(), filter).Decode(&verificationData)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			sentry.SentryHandler(err)
			return types.VerificationData{}, errors.New("code not found")
		}
		sentry.SentryHandler(err)
		return types.VerificationData{}, errors.New("error finding verification data")
	}

	return verificationData, nil
}

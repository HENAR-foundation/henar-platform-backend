package utils

import (
	"fmt"
	"henar-backend/sentry"
	"henar-backend/types"
	"reflect"
	"strconv"
	"strings"

	"crypto/rand"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpdateResultForUserRole(result interface{}, fieldsToUpdate []string) {
	resultValue := reflect.ValueOf(result)

	if resultValue.Kind() != reflect.Ptr || resultValue.Elem().Kind() != reflect.Struct {
		// Handle error: result is not a pointer to a struct
		return
	}

	resultElem := resultValue.Elem()

	for _, field := range fieldsToUpdate {
		fieldValue := resultElem.FieldByName(field)

		if fieldValue.IsValid() && fieldValue.CanSet() {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
	}
}

func UpdateResultsForUserRole(results interface{}, fieldsToUpdate []string) {
	resultsValue := reflect.ValueOf(results)

	if resultsValue.Kind() != reflect.Slice {
		// TODO: Handle error: results is not a slice
		return
	}

	for i := 0; i < resultsValue.Len(); i++ {
		resultValue := resultsValue.Index(i)

		UpdateResultForUserRole(resultValue.Addr().Interface(), fieldsToUpdate)
	}
}

func GetSort(c *fiber.Ctx) primitive.D {
	sort := c.Query("sort")

	var sortOpts primitive.D
	switch sort {
	case "-views":
		sortOpts = bson.D{{Key: "views", Value: -1}}
	case "views":
		sortOpts = bson.D{{Key: "views", Value: 1}}
	case "-tags":
		sortOpts = bson.D{{Key: "tags", Value: -1}}
	case "tags":
		sortOpts = bson.D{{Key: "tags", Value: 1}}
	case "-applicants":
		sortOpts = bson.D{{Key: "applicants", Value: -1}}
	case "applicants":
		sortOpts = bson.D{{Key: "applicants", Value: 1}}
	default:
		// do nothing
	}

	return sortOpts
}

func GetFilter(c *fiber.Ctx) (bson.M, error) {
	filter := bson.M{}

	language := c.Query("language")
	if language == "" {
		language = "en" // default language is English
	}

	title := c.Query("title")
	if title != "" {
		filter["title."+language] = primitive.Regex{Pattern: title, Options: "i"}
	}

	name := c.Query("name")
	if name != "" {
		filter["full_name"] = primitive.Regex{Pattern: name, Options: "i"}
	}

	projectStatus := c.Query("status")
	if projectStatus != "" {
		filter["project_status"] = primitive.Regex{Pattern: projectStatus, Options: "i"}
	}

	// TODO: check filter for new array data type
	howToHelpTheProject := c.Query("help")
	if howToHelpTheProject != "" {
		filter["how_to_help_the_project"] = primitive.Regex{Pattern: howToHelpTheProject, Options: "i"}
	}

	job := c.Query("job")
	if job != "" {
		filter["job"] = primitive.Regex{Pattern: job, Options: "i"}
	}

	tags := c.Query("tags")
	if tags != "" {
		tagIDs := strings.Split(tags, ",")

		var tagObjectIDs []primitive.ObjectID
		for _, tagID := range tagIDs {
			objID, err := primitive.ObjectIDFromHex(tagID)
			if err != nil {
				sentry.SentryHandler(err)
				return nil, fmt.Errorf("invalid tag ID: %s", tagID)
			}
			tagObjectIDs = append(tagObjectIDs, objID)
		}

		filter["tags"] = bson.M{"$all": tagObjectIDs}
	}

	location := c.Query("location")
	if location != "" {
		objID, err := primitive.ObjectIDFromHex(location)
		if err != nil {
			sentry.SentryHandler(err)
			return nil, fmt.Errorf("invalid location ID: %s", location)
		}
		filter["location"] = objID
	}

	return filter, nil
}

func GetPaginationOptions(c *fiber.Ctx) (*options.FindOptions, error) {
	// Extract the limit parameter from the query string
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		sentry.SentryHandler(err)
		// Default to returning all documents if the limit parameter is not present or invalid
		limit = 0
	}

	if limit > 100 {
		// Limit exceeds maximum of 100
		return nil, fmt.Errorf("limit parameter exceeds maximum of 100")
	}

	// Extract the offset parameter from the query string
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		sentry.SentryHandler(err)
		// Default to zero offset if the offset parameter is not present or invalid
		offset = 0
	}

	// Create the options for the Find method
	findOptions := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset))

	return findOptions, nil
}

func CreateSlug(Title types.Translations) string {
	var title string
	if Title.En != "" {
		title = Title.En
	} else if Title.Ru != "" {
		title = Title.Ru
	} else if Title.Hy != "" {
		title = Title.Hy
	}

	slugText := slug.Make(title)
	return slugText
}

func RandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

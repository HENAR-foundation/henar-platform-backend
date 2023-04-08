package utils

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetPaginationOptions(c *fiber.Ctx) (*options.FindOptions, error) {
	// Extract the limit parameter from the query string
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
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
		// Default to zero offset if the offset parameter is not present or invalid
		offset = 0
	}

	// Create the options for the Find method
	findOptions := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset))

	return findOptions, nil
}

package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SessionMiddleware(c *fiber.Ctx) error {
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized",
		})
	}

	if sess.Get(AUTH_KEY) == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized",
		})
	}

	// Get the user ID from the session
	// userId := sess.Get(USER_ID).(string)
	// fmt.Println("user id", userId)
	// userId, ok := sess.Get("userId").(string)
	// if !ok || userId == "" {
	// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
	// 		"message": "not authorized",
	// 	})
	// }

	// // Store the user ID in the context
	// c.Locals("userId", userId)

	return c.Next()
}

func BaseMiddleware(c *fiber.Ctx) error {
	return c.Next()
}

package routes

import (
	"henar-backend/sentry"

	"github.com/gofiber/fiber/v2"
)

func SessionMiddleware(c *fiber.Ctx) error {
	sess, err := store.Get(c)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized (no session)",
		})
	}

	if sess.Get(AUTH_KEY) == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized (no user key)",
		})
	}

	return c.Next()
}

func AdminMiddleware(c *fiber.Ctx) error {
	sess, err := store.Get(c)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized",
		})
	}

	userRole := sess.Get("user_role")

	c.Locals("userRole", userRole)

	return c.Next()
}

func AuthorMiddleware(c *fiber.Ctx) error {
	sess, err := store.Get(c)
	if err != nil {
		sentry.SentryHandler(err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized",
		})
	}

	userId := sess.Get(USER_ID)

	c.Locals("user_id", userId)

	return c.Next()
}

func BaseMiddleware(c *fiber.Ctx) error {
	return c.Next()
}

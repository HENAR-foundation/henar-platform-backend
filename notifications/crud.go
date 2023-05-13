package notifications

import "github.com/gofiber/fiber/v2"

func GetNotifications(c *fiber.Ctx) error {
	c.Write([]byte("lalal"))

	return nil
}

func ReadNotifications(c *fiber.Ctx) error {
	return nil
}

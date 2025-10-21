package response

import "github.com/gofiber/fiber/v2"

// Success returns a standardized success response
func Success(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(data)
}

// Created returns a 201 created response
func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(data)
}

// NoContent returns a 204 no content response
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Message returns a simple message response
func Message(c *fiber.Ctx, message string) error {
	return c.JSON(fiber.Map{
		"message": message,
	})
}

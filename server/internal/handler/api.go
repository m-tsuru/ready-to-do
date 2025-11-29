package handler

import "github.com/gofiber/fiber/v2"

func Handler(app *fiber.App) {
	v1 := app.Group("/api/v1")
	v1.Get("/", pingHandler)
}

func pingHandler(c *fiber.Ctx) error {
	return c.SendString("pong")
}

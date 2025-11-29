package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/m-tsuru/ready-to-do/internal/handler"
)

func main() {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/api/v1")
	})

	handler.Handler(app)

	app.Listen(":80")
}

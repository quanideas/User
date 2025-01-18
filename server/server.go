package server

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

func RunServer() {
	env := os.Getenv("ENVIRONMENT")

	var app *fiber.App

	if env == "development" {
		app = fiber.New(fiber.Config{
			BodyLimit:               100 * 1024 * 1024, // 100 MB
			EnableTrustedProxyCheck: false,
		})
	} else {
		app = fiber.New(fiber.Config{
			BodyLimit:               100 * 1024 * 1024, // 100 MB
			EnableTrustedProxyCheck: true,
		})
	}

	SetupRoutes(app)

	port := fmt.Sprintf(":%v", os.Getenv("SERVER_IN_PORT"))
	app.Listen(port)
}

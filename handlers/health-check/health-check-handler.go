package healthcheck

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

func HealthCheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func ConnectionCheck(c *fiber.Ctx) error {
	healthCheck := struct {
		FileService    bool `json:"file_service"`
		ProjectService bool `json:"project_service"`
		TokenService   bool `json:"token_service"`
	}{}

	// Check File service
	host := os.Getenv("FILE_SERVICE_HOST")
	port := os.Getenv("FILE_SERVICE_PORT")
	api := "/health-check"
	url := fmt.Sprintf("%s:%s/%s", host, port, api)

	agent := fiber.Get(url)
	fileStatusCode, _, _ := agent.Bytes()
	healthCheck.FileService = fileStatusCode == fiber.StatusOK

	// Check Project service
	host = os.Getenv("PROJECT_SERVICE_HOST")
	port = os.Getenv("PROJECT_SERVICE_PORT")
	api = "/health-check"
	url = fmt.Sprintf("%s:%s/%s", host, port, api)

	agent = fiber.Get(url)
	projectStatusCode, _, _ := agent.Bytes()
	healthCheck.ProjectService = projectStatusCode == fiber.StatusOK

	// Check Token service
	host = os.Getenv("TOKEN_SERVICE_HOST")
	port = os.Getenv("TOKEN_SERVICE_PORT")
	api = "/health-check"
	url = fmt.Sprintf("%s:%s/%s", host, port, api)

	agent = fiber.Get(url)
	tokenStatusCode, _, _ := agent.Bytes()
	healthCheck.TokenService = tokenStatusCode == fiber.StatusOK

	c.JSON(healthCheck)
	return nil
}

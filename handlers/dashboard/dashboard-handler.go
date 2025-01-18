package dashboardhandlers

import (
	"user/common/constants"
	"user/common/helpers"
	"user/models/response"
	"user/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func GetDashboardStats(c *fiber.Ctx) error {
	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	isRoot := claims["is_root"].(bool)

	// Only allow root
	if !isRoot {
		helpers.BadRequest(c, "no permission", constants.ERR_COMMON_PERMISSION_NOT_ALLOWED)
		return nil
	}

	// Get stats from db
	var dashboardStats response.DashboardStatsResponse
	query := repositories.GetDashboardStats()
	if result := query.First(&dashboardStats); result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return user in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: dashboardStats,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

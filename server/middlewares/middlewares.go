package middlewares

import (
	"fmt"
	"os"
	"user/common/constants"
	"user/common/helpers"
	commonqueries "user/common/queries"
	"user/database"
	"user/models/entity"
	"user/models/request"
	"user/models/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func CatchPanic() fiber.Handler {
	// Return new handler
	return func(c *fiber.Ctx) (err error) { //nolint:nonamedreturns // Uses recover() to overwrite the error
		// Catch panics
		defer func() {
			if r := recover(); r != nil {
				helpers.InternalServerError(c, fmt.Sprintf("%v", r))
				log.Error(fmt.Sprintf("Panic: %v", r))
			}
		}()

		// Return err if exist, else move to next handler
		return c.Next()
	}
}

func ValidateJWT() fiber.Handler {
	// Return new handler
	return func(c *fiber.Ctx) (err error) { //nolint:nonamedreturns // Uses recover() to overwrite the error
		// Get token and refresh token from cookies
		token := c.Cookies("token")
		refreshToken := c.Cookies("refreshToken")

		// No token or refresh token, return unauthorized
		if token == "" || refreshToken == "" {
			// Clear cookies
			c.ClearCookie("token", "refreshToken")

			c.Status(fiber.StatusUnauthorized)
			c.JSON(response.ErrorResponse{
				ErrorCode: 401,
				Error:     "unauthorized",
			})
			return nil
		}

		// Get info of Token Validation microservice
		host := os.Getenv("TOKEN_SERVICE_HOST")
		port := os.Getenv("TOKEN_SERVICE_PORT")
		api := constants.TokenValidation
		url := fmt.Sprintf("%s:%s/%s", host, port, api)

		// Call token validation service to validate token
		agent := fiber.Post(url)
		agent.JSON(request.ValidationRequest{
			Token:        token,
			RefreshToken: refreshToken,
		})
		var data response.ValidationResponse
		errCode, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken)

		// If error then return bad request, or token is invalid then return unauthorize
		if err != nil {
			c.Status(fiber.StatusBadRequest)
			c.JSON(response.ErrorResponse{
				ErrorCode: errCode,
				Error:     err.Error(),
			})
			return nil
		} else if !data.IsValid {
			// Clear cookies
			c.ClearCookie("token", "refreshToken")

			c.Status(fiber.StatusUnauthorized)
			c.JSON(response.ErrorResponse{
				ErrorCode: 401,
				Error:     "unauthorized",
			})
			return nil
		}

		// Set new Token to cookies
		c.Cookie(&fiber.Cookie{
			Name:  "token",
			Value: data.Token,
		})

		// Set token in local scope (this request's scope) to be able to parse when needed
		var localToken *jwt.Token
		localToken, _ = jwt.Parse(data.Token, nil)
		c.Locals("user", localToken)

		// Move to next handler
		return c.Next()
	}
}

func ValidatePermimssion(permissionType, permissionLevel string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get username from token
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["username"].(string)
		isAdmin := claims["is_admin"].(bool)
		isRoot := claims["is_root"].(bool)

		if isRoot {
			return c.Next()
		}

		// If not admin, then check if this user has permission with exact level or the full level
		if !isAdmin {
			_, err := commonqueries.ValidatePermission(username, permissionType, permissionLevel, isRoot, isAdmin)
			if err != nil {
				helpers.BadRequest(c, "no permission\n"+err.Error(), constants.ERR_COMMON_PERMISSION_NOT_ALLOWED)
				return nil
			}
		}

		return c.Next()
	}
}

func CheckRoleBelongsToCompany(c *fiber.Ctx) error {
	// Parse ID to match entity model
	request := struct {
		ID     uuid.UUID `json:"id"`
		RoleID uuid.UUID `json:"role_id"`
	}{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Get roleID
	var roleID uuid.UUID
	if request.RoleID != uuid.Nil {
		roleID = request.RoleID
	} else {
		roleID = request.ID
	}

	var count int64
	result := database.DB.Db.Model(entity.Role{}).
		Where("id = ?", roleID).
		Where("company_id IN (?)", database.DB.Db.Model(entity.User{}).Select("company_id").Where("username = ?", username)).
		Count(&count)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	} else if count == 0 {
		helpers.BadRequest(c, "role not found", constants.ERR_ROLE_NOT_FOUND)
		return nil
	}

	return c.Next()
}

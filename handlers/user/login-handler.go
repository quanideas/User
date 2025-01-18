package userhandlers

import (
	"os"
	"strconv"
	"time"
	"user/common/constants"
	"user/common/helpers"
	"user/models/request"
	"user/models/response"
	sqlresponse "user/models/sql_response"
	"user/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Login returns JWT token if password matched hashed pwd on db,
// or 401 unauthorized otherwise.
// Params
// { username string
// password string }
func Login(c *fiber.Ctx) error {
	// Parse login information sent by request
	loginInfo := request.LoginRequest{}
	if err := c.BodyParser(&loginInfo); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Check if user exist on db
	var user sqlresponse.GetUserWithCompanyNameResponse
	result := repositories.GetUserWithCompanyName(loginInfo.Username).First(&user)

	// Not exist, throw error
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "user not found", constants.ERR_USER_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// User found, compare hashed and passed in password
	hashed := user.Password
	if !helpers.CompareHashedPassword(loginInfo.Password, hashed) {
		c.Status(fiber.StatusUnauthorized)
		c.JSON(response.ErrorResponse{
			ErrorCode: 401,
			Error:     "Username or password does not match",
		})
		return nil
	}

	// Password matched, create JWT Claims
	jwt, err := createJWTToken(loginInfo.Username, user.IsRoot, user.IsAdmin, user.CompanyID.String())
	if err != nil {
		helpers.InternalServerError(c, err.Error())
	}

	// Map to response and return
	body := response.LoginSuccessResponse{}
	body.UserGetByUsernameResponse = user.UserGetByUsernameResponse
	body.CompanyName = user.CompanyName
	body.Token = jwt

	c.Status(200)
	c.JSON(body)
	return nil
}

func createJWTToken(username string, is_root bool, is_admin bool, company_id string) (string, error) {
	expireTime, _ := strconv.Atoi(os.Getenv("JWT_EXPIRE_TIME_IN_HOURS"))
	secretKey := os.Getenv("JWT_SECRET_KEY")
	claims := jwt.MapClaims{
		"username":   username,
		"is_root":    is_root,
		"is_admin":   is_admin,
		"company_id": company_id,
		"exp":        time.Now().Add(time.Hour * time.Duration(expireTime)).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response
	jwt, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return jwt, nil
}

package permissionhandlers

import (
	"user/common/helpers"
	commonqueries "user/common/queries"
	"user/database"
	"user/models/entity"
	"user/models/request"
	"user/models/response"
	"user/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// GetByCompany get list of permissions. Anyone can get this
// Param is the list of sort and search columns
func GetAllPermissions(c *fiber.Ctx) error {
	// Parse search query model
	getAllRequest := request.GetAll{}
	if err := c.BodyParser(&getAllRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Init query
	query := repositories.GetAllPermissions()

	// Handle search and sort
	query, errCode, err := commonqueries.AddSearchAndSortGetAll(getAllRequest, query, entity.Permission{})
	if err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Get total number
	var count int64
	query.Count(&count)

	// Limit
	if getAllRequest.Count != 0 && getAllRequest.Page != 0 {
		query = query.Offset((getAllRequest.Page - 1) * getAllRequest.Count).Limit(getAllRequest.Count)
	}

	// Get
	var permissions []entity.Permission
	result := query.Find(&permissions)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return user in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: response.GetAll{
			List:  permissions,
			Total: count,
		},
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// ValidatePermission checks if user has permission in either Role or UserPermission.
// Automatically search for higher permission. e.g: if user has edit permission then
// user is assumed to have view permission, thus return Granted
// Params
// { ProjectID: projectID to search
// PermissionType: type of permission in constants. e.g: manage_permission
// PermissionLevel: level of permission. e.g: full, create, view, edit }
func ValidatePermission(c *fiber.Ctx) error {
	dataResponse := "Denied"

	// Parse validation model
	request := request.GetUserSpecificPermissionRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	isAdmin := claims["is_admin"].(bool)
	isRoot := claims["is_root"].(bool)

	// Check in cache
	var permissionCache entity.PermissionCache
	query := database.DB.Db.
		Where("username = ?", username).
		Where("permission_type = ?", request.PermissionType).
		Where("permission_level = ?", request.PermissionLevel)
	if request.ProjectID == nil {
		query = query.Where("project_id IS NULL")
	} else {
		query = query.Where("project_id = ?", request.ProjectID)
	}
	result := query.First(&permissionCache)
	if result.Error == nil { // Permission found
		// Set dataResponse corresponding to granted/denied found on cache
		if permissionCache.Granted {
			dataResponse = "Granted"
		}

		// Return permissions in request
		c.Status(200)
		c.JSON(response.BaseResponse{
			Data: dataResponse,
			Meta: struct{ Status int }{Status: 200},
		})
		return nil
	}

	// Validate permission in role and user permission
	if _, err := commonqueries.ValidatePermission(username,
		request.PermissionType,
		request.PermissionLevel,
		isRoot, isAdmin,
		request.ProjectID); err == nil {

		// Granted, pushed onto the cache
		dataResponse = "Granted"
		database.DB.Db.Create(&entity.PermissionCache{
			Username:        username,
			ProjectID:       request.ProjectID,
			PermissionType:  request.PermissionType,
			PermissionLevel: request.PermissionLevel,
			Granted:         true,
		})
	} else {
		// Denied, pushed denied onto the cache
		database.DB.Db.Create(&entity.PermissionCache{
			Username:        username,
			ProjectID:       request.ProjectID,
			PermissionType:  request.PermissionType,
			PermissionLevel: request.PermissionLevel,
			Granted:         false,
		})
	}

	// Return permissions in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: dataResponse,
		Meta: struct{ Status int }{Status: 200},
	})
	return nil
}

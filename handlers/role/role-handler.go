package rolehandlers

import (
	"fmt"
	"user/common/constants"
	"user/common/helpers"
	commonqueries "user/common/queries"
	"user/database"
	"user/models/entity"
	"user/models/request"
	"user/models/response"
	"user/repositories"

	"github.com/devfeel/mapper"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create creates a role corresponding to user's company.
// User has to be an admin or has manage role permission to be able to create new roles.
// Param takes in name of the role.
func Create(c *fiber.Ctx) error {
	// Parse role to match request model
	role := request.RoleRequest{}
	if err := c.BodyParser(&role); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Get user's company id
	var companyIDStr string
	var companyID uuid.UUID
	result := repositories.GetCompanyIDOfUser(username).First(&companyIDStr)
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "user not found", constants.ERR_USER_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}
	companyID, err := uuid.Parse(companyIDStr) // Parse string to uuid
	if err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Map to Role entity model
	roleEntity := entity.Role{
		BaseEntityModel: helpers.CreateMetaData(username),
		CompanyID:       companyID,
		Name:            role.Name,
	}

	// Call db to create
	result = database.DB.Db.Create(&roleEntity)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Map role to response model
	var roleResponse response.RoleResponse
	mapper.Mapper(&roleEntity, &roleResponse)
	roleResponse.BaseEntityModel = roleEntity.BaseEntityModel

	// Return created role
	c.Status(201)
	c.JSON(response.BaseResponse{
		Data: roleResponse,
		Meta: struct{ Status int }{Status: 201},
	})

	return nil
}

// Update updates info of a role. User needs to be an admin or has manage role permission.
// Param takes in ID of the role and name.
func UpdateInfo(c *fiber.Ctx) error {
	// Parse role to match entity model
	request := entity.Role{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Get role and also check if it belongs to user's company
	var role entity.Role
	result := database.DB.Db.Where("id = ?", request.ID).First(&role)
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "role not found", constants.ERR_ROLE_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Fill in metadata
	request.BaseEntityModel = helpers.EditMetaData(username, role.BaseEntityModel)
	request.CompanyID = role.CompanyID

	// Update
	result = database.DB.Db.Save(&request)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Map role to response model
	var roleResponse response.RoleResponse
	mapper.Mapper(&request, &roleResponse)
	roleResponse.BaseEntityModel = request.BaseEntityModel

	// Return user in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: roleResponse,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// GetAll returns all roles of user's company. User has to be an admin or has manage role
// permission to see all roles.
// Param takes in paging and sort/search.
func GetAll(c *fiber.Ctx) error {
	// Parse search query model
	getAllRequest := request.GetAll{}
	if err := c.BodyParser(&getAllRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Get user's company
	var companyID string
	result := repositories.GetCompanyIDOfUser(username).First(&companyID)
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "user not found", constants.ERR_USER_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Init query
	query := repositories.GetRoleListFromCompanyID(companyID)

	// Handle search and sort
	query, errCode, err := commonqueries.AddSearchAndSortGetAll(getAllRequest, query, entity.Role{})
	if err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Get total number
	var count int64
	query.Count(&count)

	// Limit
	query = query.Offset((getAllRequest.Page - 1) * getAllRequest.Count).Limit(getAllRequest.Count)

	// Get
	var roles []response.RoleResponse
	result = query.Find(&roles)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return user in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: response.GetAll{
			List:  roles,
			Total: count,
		},
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// GetByIDWithUserAndPermissions returns users and permissions of a role by using roleID.
// Param takes in the ID of the role.
func GetByIDWithUserAndPermissions(c *fiber.Ctx) error {
	// Parse role to match request model
	request := request.GetByIDRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get role
	var role entity.Role
	result := database.DB.Db.Where("roles.id = ?", request.ID).First(&role)
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "role not found", constants.ERR_ROLE_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Get users of role
	var users []response.RoleGetByIDUserResponse
	result = database.DB.Db.
		Model(entity.User{}).
		Where("username IN (?)", database.DB.Db.
			Model(entity.UserRoleMap{}).
			Select("user_username").
			Where("role_id = ?", request.ID)).
		Find(&users)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Get setting permissions of role
	var settingPermissions []response.GetByIDSettingPermissionResponse
	result = database.DB.Db.
		Model(entity.RoleSettingPermission{}).
		Where("role_id = ?", request.ID).
		Find(&settingPermissions)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Get project permissions of role
	var projectPermissions []response.GetByIDProjectPermissionResponse
	result = database.DB.Db.
		Model(entity.RoleProjectPermission{}).
		Where("role_id = ?", request.ID).
		Find(&projectPermissions)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Map to return model
	var roleResponse response.RoleGetByIDResponse
	m := mapper.NewMapper()
	m.SetEnableFieldIgnoreTag(true) // Ignore users, setting/project permissions
	m.Mapper(&role, &roleResponse)
	roleResponse.Users = users
	roleResponse.SettingPermissions = settingPermissions
	roleResponse.ProjectPermissions = projectPermissions

	// Return response model in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: roleResponse,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// Delete deletes a role from company along with all its permissions.
// Param takes in the ID of the role.
func Delete(c *fiber.Ctx) error {
	// Parse role to match request model
	request := request.DeleteByIDRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Call db to delete
	result := database.DB.Db.Where("id = ?", request.ID).Delete(entity.Role{})
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return created role
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: "Success",
		Meta: struct{ Status int }{Status: 200},
	})
	return nil
}

// AddUser adds a user to a role and inherit all permissions that role has.
// User making the request has to be an admin or has manage role permission.
// Param takes in ID of a role and ID of the user to be added to that role.
func AddUser(c *fiber.Ctx) error {
	// Parse role to match request model
	request := request.RoleAddUserRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	isRoot := claims["is_root"].(bool)
	isAdmin := claims["is_admin"].(bool)

	// Validate if user is in requester's company
	var count int64
	result := database.DB.Db.Model(entity.User{}).
		Where("username = ?", request.Username).
		Where("company_id IN (?)",
			database.DB.Db.Model(entity.User{}).Select("company_id").Where("username = ?", username)).
		Count(&count)
	if count == 0 {
		helpers.BadRequest(c, "user not found", constants.ERR_ROLE_USER_OR_ROLE_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Validate if requester has manage permission
	if errCode, err := commonqueries.ValidatePermission(username,
		constants.PERM_MANAGE_PERMISSION,
		constants.PERM_LEVEL_EDIT,
		isRoot,
		isAdmin); err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Create User-Role map entity
	userRoleMap := entity.UserRoleMap{
		RoleID:   request.RoleID,
		Username: request.Username,
	}

	// Call db to create
	result = database.DB.Db.Create(&userRoleMap)
	if result.Error == gorm.ErrForeignKeyViolated {
		helpers.BadRequest(c, "user or role not found", constants.ERR_ROLE_USER_OR_ROLE_NOT_FOUND)
		return nil
	} else if result.Error == gorm.ErrDuplicatedKey {
		helpers.BadRequest(c, "user already added to this role", constants.ERR_ROLE_USER_ALREADY_ADDED)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Remove all cache of the added user
	result = database.DB.Db.
		Where("username = ?", request.Username).
		Delete(entity.PermissionCache{})
	if result.Error != nil {
		helpers.InternalServerError(c, fmt.Sprintf("%s: %s", "error removing cache", result.Error.Error()))
		return nil
	}

	// Return success
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: "Success",
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// RemoveUser removes the user from a role.
// User making the request has to be an admin or has manage role permission.
// Param takes in the role ID and user ID
func RemoveUser(c *fiber.Ctx) error {
	// Parse role to match request model
	request := request.RoleAddUserRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	isRoot := claims["is_root"].(bool)
	isAdmin := claims["is_admin"].(bool)

	// Validate if user is in requester's company
	var count int64
	result := database.DB.Db.Model(entity.User{}).
		Where("username = ?", request.Username).
		Where("company_id IN (?)",
			database.DB.Db.Model(entity.User{}).Select("company_id").Where("username = ?", username)).
		Count(&count)
	if count == 0 {
		helpers.BadRequest(c, "user not found", constants.ERR_ROLE_USER_OR_ROLE_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Validate if requester has manage permission
	if errCode, err := commonqueries.ValidatePermission(username,
		constants.PERM_MANAGE_PERMISSION,
		constants.PERM_LEVEL_EDIT,
		isRoot,
		isAdmin); err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Call db to remove user from role
	result = database.DB.Db.
		Where("user_username = ?", request.Username).
		Where("role_id = ?", request.RoleID).
		Delete(entity.UserRoleMap{})
	if result.RowsAffected == 0 {
		helpers.BadRequest(c, "user or role not found", constants.ERR_ROLE_USER_OR_ROLE_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Remove all cache of the removed user
	result = database.DB.Db.
		Where("username = ?", request.Username).
		Delete(entity.PermissionCache{})
	if result.Error != nil {
		helpers.InternalServerError(c, fmt.Sprintf("%s: %s", "error removing cache", result.Error.Error()))
		return nil
	}

	// Return success
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: "Success",
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// AddPermission to add a new permission to a role that all users belong to that
// role will inherit.
// User making the request has to be an admin or has manage role permission.
// Param is the role ID to be added on, type of permission (setting/project), projectID of
// the project to be added (optional), and permission level (view/edit)
func AddPermission(c *fiber.Ctx) error {
	// Parse permission to match request model
	permissionRequest := request.RoleAddPermissionRequest{}
	if err := c.BodyParser(&permissionRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	isAdmin := claims["is_admin"].(bool)
	isRoot := claims["is_root"].(bool)

	// Validate user has manage permission
	if errCode, err := commonqueries.ValidatePermission(username,
		constants.PERM_MANAGE_PERMISSION,
		constants.PERM_LEVEL_EDIT,
		isRoot,
		isAdmin); err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Validate user has this permission
	if errCode, err := commonqueries.ValidatePermission(username,
		permissionRequest.PermissionType,
		permissionRequest.PermissionLevel,
		isRoot,
		isAdmin,
		permissionRequest.ProjectID); err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Check if setting or project permission
	var permission interface{}
	if len(permissionRequest.PermissionType) >= 7 && permissionRequest.PermissionType[:7] == "project" {
		// Create Role Project Permission entity
		m := mapper.NewMapper()
		m.SetEnabledTypeChecking(true)
		roleProjectPermission := entity.RoleProjectPermission{}
		m.Mapper(&permissionRequest, &roleProjectPermission)
		roleProjectPermission.ID = uuid.New()
		roleProjectPermission.RoleID = permissionRequest.RoleID
		roleProjectPermission.ProjectID = *permissionRequest.ProjectID
		permission = &roleProjectPermission
	} else {
		// Create Role Setting Permission entity
		roleSettingPermission := entity.RoleSettingPermission{}
		mapper.Mapper(&permissionRequest, &roleSettingPermission)
		roleSettingPermission.ID = uuid.New()
		roleSettingPermission.RoleID = permissionRequest.RoleID
		permission = &roleSettingPermission
	}

	// Call db to create
	result := database.DB.Db.Create(permission)
	if result.Error == gorm.ErrDuplicatedKey {
		helpers.BadRequest(c, "permission already added", constants.ERR_ROLE_PERMISSION_ALREADY_ADDED)
		return nil
	} else if result.Error == gorm.ErrForeignKeyViolated {
		helpers.BadRequest(c, "role or project not found", constants.ERR_ROLE_ROLE_OR_PROJECT_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return success
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: permission,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// RemovePermission removes a setting/project permission from a role
// User making the request has to be an admin or has manage role permission.
// Param
// { ID: id of permission }
func RemovePermission(c *fiber.Ctx) error {
	// Parse role to match request model
	request := request.DeleteByIDRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	isAdmin := claims["is_admin"].(bool)
	isRoot := claims["is_root"].(bool)

	// Validate user has manage permission
	if errCode, err := commonqueries.ValidatePermission(username,
		constants.PERM_MANAGE_PERMISSION,
		constants.PERM_LEVEL_EDIT,
		isRoot,
		isAdmin); err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Call db to find permission and validate if user has this permission
	var permission interface{}
	var roleID uuid.UUID
	var roleProjectPermission entity.RoleProjectPermission
	var roleSettingPermission entity.RoleSettingPermission
	result := database.DB.Db.Where("id = ?", request.ID).First(&roleProjectPermission)
	if result.Error == nil { // Project permission
		permission = &roleProjectPermission
		roleID = roleProjectPermission.RoleID

		// Validate user has this permission
		if errCode, err := commonqueries.ValidatePermission(username,
			roleProjectPermission.PermissionType,
			roleProjectPermission.PermissionLevel,
			isRoot,
			isAdmin,
			&roleProjectPermission.ProjectID); err != nil {
			helpers.BadRequest(c, err.Error(), errCode)
			return nil
		}
	} else { // Setting permission
		result = database.DB.Db.Where("id = ?", request.ID).First(&roleSettingPermission)
		if result.Error == nil {
			permission = &roleSettingPermission
			roleID = roleSettingPermission.RoleID

			// Validate user has this permission
			if errCode, err := commonqueries.ValidatePermission(username,
				roleSettingPermission.PermissionType,
				roleSettingPermission.PermissionLevel,
				isRoot,
				isAdmin); err != nil {
				helpers.BadRequest(c, err.Error(), errCode)
				return nil
			}
		} else {
			helpers.BadRequest(c, "permission not found", constants.ERR_COMMON_PERMISSION_NOT_FOUND)
			return nil
		}
	}

	// Check if Permission's Role belongs to this Company
	if err := validateRoleBelongToCompany(username, roleID); err != nil {
		helpers.BadRequest(c, err.Error(), constants.ERR_COMMON_PERMISSION_NOT_FOUND)
		return nil
	}

	// Call db to remove Permission from Role
	result = database.DB.Db.Delete(permission)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return success
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: "Success",
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

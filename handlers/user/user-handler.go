package userhandlers

import (
	"database/sql"
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

// Create is used to create a regular account or a root/admin account.
// Root accounts can only be created by root.
// Admin accounts can only be created by root or admin.
// Regular accounts can be created by root, admin, or accounts with permission.
// Admin/regular accounts will be added to a company based on CompanyID.
// Params
// { user User
// CompanyID uuid.UUID }
func Create(c *fiber.Ctx) error {
	// Parse user to match entity model
	userRequest := request.CreateUserRequest{}
	if err := c.BodyParser(&userRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	isRoot := claims["is_root"].(bool)
	isAdmin := claims["is_admin"].(bool)
	companyID := uuid.MustParse(claims["company_id"].(string))

	// Only allow when user is root user, admin of the company,
	// or has permission to create
	if isRoot { // No need to change anything

	} else if isAdmin { // Not allow to create root user if not root
		userRequest.IsRoot = false
		userRequest.User.CompanyID = companyID
	} else { // Regular user, already validated by middleware
		userRequest.IsAdmin = false
		userRequest.IsRoot = false
		userRequest.User.CompanyID = companyID
	}

	// Fill in metadata
	user := userRequest.User
	user.BaseEntityNoIDModel = helpers.CreateMetaDataNoID(username)
	// Encode password
	if password, err := helpers.HashPassword(user.Password); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	} else {
		user.Password = password
	}

	// If first user of the company, default to be admin
	var currentUsers int64
	result := database.DB.Db.Model(entity.User{}).Where("company_id = ?", user.CompanyID).Count(&currentUsers)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	} else if currentUsers == 0 {
		user.IsAdmin = true
	}

	// Get max users, if not root, then only allow to create if current < max
	// If root, then root can bypass max user limit
	var maxUsers sql.NullInt64
	result = database.DB.Db.Model(entity.Company{}).Where("id = ?", user.CompanyID).Select("max_users").First(&maxUsers)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}
	if !isRoot && maxUsers.Int64 != 0 && currentUsers >= maxUsers.Int64 {
		helpers.BadRequest(c, "maximum number of users reached, please contact site's admin", constants.ERR_USER_MAX_USER_IN_COMPANY)
		return nil
	}

	// Call db to create user
	result = database.DB.Db.Create(&user)
	if result.Error == gorm.ErrDuplicatedKey {
		helpers.BadRequest(c, "user already exists", constants.ERR_USER_MAX_USER_IN_COMPANY)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Update total number of users in this company
	database.DB.Db.Model(&entity.Company{}).
		Where("id = ?", user.CompanyID).
		Update("current_employees", currentUsers+1)

	// Map to return model
	userResponse := response.UserResponse{}
	mapper.Mapper(&user, &userResponse)
	userResponse.BaseEntityNoIDModel = user.BaseEntityNoIDModel

	// Return created user
	c.Status(201)
	c.JSON(response.BaseResponse{
		Data: userResponse,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// AddPermission adds a new project/setting permission to one user.
// Requester has to be admin or has manage_permission
// Params
// { Username: username of user to grand permission for.
// ProjectID: project's ID to grand permission for.
// PermissionType: type of permission in constants. e.g: manage_permission.
// PermissionLevel: level of permission. e.g: view/edit }
func AddPermission(c *fiber.Ctx) error {
	// Parse permission to match request model
	permissionRequest := request.UserAddPermissionRequest{}
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

	// Validate requester has manage permission
	if errCode, err := commonqueries.ValidatePermission(username,
		constants.PERM_MANAGE_PERMISSION,
		constants.PERM_LEVEL_EDIT,
		isRoot,
		isAdmin); err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Validate requester has this permission
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
		// Create User Project Permission entity
		m := mapper.NewMapper()
		m.SetEnabledTypeChecking(true)
		userProjectPermission := entity.UserProjectPermission{}
		m.Mapper(&permissionRequest, &userProjectPermission)
		userProjectPermission.ID = uuid.New()
		userProjectPermission.ProjectID = *permissionRequest.ProjectID
		permission = userProjectPermission
	} else {
		// Create User Setting Permission entity
		userSettingPermission := entity.UserSettingPermission{}
		mapper.Mapper(&permissionRequest, &userSettingPermission)
		userSettingPermission.ID = uuid.New()
		permission = &userSettingPermission
	}

	// Validate if user is in requester's company to make sure requester can't adding user of another company
	if err := validateUserIsInRequestersCompany(username, permissionRequest.Username); err != nil {
		helpers.BadRequest(c, err.Error(), constants.ERR_COMMON_PERMISSION_NOT_FOUND)
		return nil
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

	// Remove all cache of the requested user
	result = database.DB.Db.
		Where("username = ?", permissionRequest.Username).
		Delete(entity.PermissionCache{})
	if result.Error != nil {
		helpers.InternalServerError(c, fmt.Sprintf("%s: %s", "error removing cache", result.Error.Error()))
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

// AddPermission removes project/setting permission of one user.
// Requester has to be admin or has manage_permission
// Params
// { ID: id of permission }
func RemovePermission(c *fiber.Ctx) error {
	// Parse permission to match request model
	request := request.DeleteByIDRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get requester's username from token
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	isAdmin := claims["is_admin"].(bool)
	isRoot := claims["is_root"].(bool)

	// Validate requester has manage permission
	if errCode, err := commonqueries.ValidatePermission(username,
		constants.PERM_MANAGE_PERMISSION,
		constants.PERM_LEVEL_EDIT,
		isRoot,
		isAdmin); err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Call db to find permission and validate if requester has this permission
	var permission interface{}
	var permissionUsername string
	var userProjectPermission entity.UserProjectPermission
	var userSettingPermission entity.UserSettingPermission
	result := database.DB.Db.Where("id = ?", request.ID).First(&userProjectPermission)
	if result.Error == nil { // Permission found in User Project Permission
		permission = &userProjectPermission
		permissionUsername = userProjectPermission.Username

		// Validate requester has this permission
		if errCode, err := commonqueries.ValidatePermission(username,
			userProjectPermission.PermissionType,
			userProjectPermission.PermissionLevel,
			isRoot,
			isAdmin,
			&userProjectPermission.ProjectID); err != nil {
			helpers.BadRequest(c, err.Error(), errCode)
			return nil
		}
	} else { // Setting permission
		result = database.DB.Db.Where("id = ?", request.ID).First(&userSettingPermission)
		if result.Error == nil { // Permission found in User Setting Permission
			permission = &userSettingPermission
			permissionUsername = userSettingPermission.Username

			// Validate requester has this permission
			if errCode, err := commonqueries.ValidatePermission(username,
				userSettingPermission.PermissionType,
				userSettingPermission.PermissionLevel,
				isRoot,
				isAdmin); err != nil {
				helpers.BadRequest(c, err.Error(), errCode)
				return nil
			}
		} else { // Permission not found
			helpers.BadRequest(c, "permission not found", constants.ERR_COMMON_PERMISSION_NOT_FOUND)
			return nil
		}
	}

	// Validate if user is in requester's company to make sure requester can't remove user of another company
	if err := validateUserIsInRequestersCompany(username, permissionUsername); err != nil {
		helpers.BadRequest(c, err.Error(), constants.ERR_COMMON_PERMISSION_NOT_FOUND)
		return nil
	}

	// Call db to remove Permission from Role
	result = database.DB.Db.Delete(permission)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Remove all cache of the requested user
	result = database.DB.Db.
		Where("username = ?", permissionUsername).
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

// GetByID returns the requester's information based on
// the username in the JWT token. This is only used for getting
// self's information in order to view or update.
// No param required
func GetByUsername(c *fiber.Ctx) error {
	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Get user from db by username
	user := response.UserGetByUsernameResponse{}
	query := repositories.GetUserByUsername(username)
	result := query.First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "user not found", constants.ERR_USER_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return user in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: user,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// GetUserSpecificPermission returns the requester's permissions based on params.
// Automatically search for permission_type (if not empty string)
// Params
// { ProjectID: projectID to search
// PermissionType: type of permission in constants. e.g: manage_permission
// PermissionLevel: level of permission. e.g: full, create, view, edit }
func GetUserSpecificPermission(c *fiber.Ctx) error {
	var tableUserPermission *gorm.DB
	var tableRolePermission *gorm.DB

	// Parse search query model
	request := request.GetUserSpecificPermissionRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Check if permission is Project or Setting
	var permissions []response.UserGetSpecificPermissionResponse
	if len(request.PermissionType) >= 7 && request.PermissionType[:7] == "project" { // Project permission
		tableUserPermission = database.DB.Db.Model(&entity.UserProjectPermission{}).
			Select("project_id, permission_type, permission_level") // Select project_id
		tableRolePermission = database.DB.Db.Model(&entity.RoleProjectPermission{}).
			Select("project_id, permission_type, permission_level") // Select project_id
	} else { // Setting permission
		tableUserPermission = database.DB.Db.Model(&entity.UserSettingPermission{}).
			Select("permission_type, permission_level") // Don't select project_id
		tableRolePermission = database.DB.Db.Model(&entity.RoleSettingPermission{}).
			Select("permission_type, permission_level") // Don't select project_id
		request.ProjectID = nil
	}

	// Find permission in UserProjectPermission/UserSettingPermission
	tableUserPermission = tableUserPermission.Where("username = ?", username)

	// Find permission in RoleProjectPermission/RoleSettingPermission
	tableRolePermission = tableRolePermission.Where("role_id IN (?)",
		database.DB.Db.
			Model(&entity.UserRoleMap{}).
			Select("role_id").
			Where("user_username = ?", username))

	// Query by project_id
	if request.ProjectID != nil {
		projectID := *request.ProjectID
		tableUserPermission = tableUserPermission.Where("project_id = ?", projectID)
		tableRolePermission = tableRolePermission.Where("project_id = ?", projectID)
	}

	// Query by permission_type
	if request.PermissionType != "" {
		tableUserPermission = tableUserPermission.Where("permission_type = ?", request.PermissionType)
		tableRolePermission = tableRolePermission.Where("permission_type = ?", request.PermissionType)
	}

	// Query by permission_level
	if request.PermissionLevel != "" {
		tableUserPermission = tableUserPermission.Where("permission_level = ?", request.PermissionLevel)
		tableRolePermission = tableRolePermission.Where("permission_level = ?", request.PermissionLevel)
	}

	// Execute query
	if result := database.DB.Db.Raw("SELECT * FROM ((?) UNION (?)) AS p",
		tableUserPermission,
		tableRolePermission).
		Find(&permissions); result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return permissions in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: permissions,
		Meta: struct{ Status int }{Status: 200},
	})
	return nil
}

// TODO: Get list of user's permissions organized by project/setting/role/user
func GetPermissionByUsername(c *fiber.Ctx) error {
	// Parse search query model
	request := request.UserGetByUsernameRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Validate if Permission's is in this user's Company and get permissions of other Companies
	if err := validateUserIsInRequestersCompany(username, request.Username); err != nil {
		helpers.BadRequest(c, err.Error(), constants.ERR_USER_NOT_FOUND)
		return nil
	}

	// Get setting permissions
	var settingPermissions []response.GetByIDSettingPermissionResponse
	if result := database.DB.Db.
		Model(entity.UserSettingPermission{}).
		Where("username = ?", request.Username).
		Find(&settingPermissions); result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Get project permissions
	var projectPermissions []response.GetByIDProjectPermissionResponse
	if result := database.DB.Db.
		Model(entity.UserProjectPermission{}).
		Where("username = ?", request.Username).
		Find(&projectPermissions); result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Map to return model
	permissionResponse := response.UserGetPermissionByUsernameResponse{
		SettingPermissions: settingPermissions,
		ProjectPermissions: projectPermissions,
	}

	// Return user in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: permissionResponse,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// GetByCompany get list of users based on the requester's company
// Param is the list of sort and search columns
func GetAllByCompany(c *fiber.Ctx) error {
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
	isRoot := claims["is_root"].(bool)

	// Get user's company
	var companyID string
	if isRoot {
		companyID = getAllRequest.CompanyID
	} else {
		result := repositories.GetCompanyIDOfUser(username).First(&companyID)
		if result.Error == gorm.ErrRecordNotFound {
			helpers.BadRequest(c, "user not found", constants.ERR_USER_NOT_FOUND)
			return nil
		} else if result.Error != nil {
			helpers.InternalServerError(c, result.Error.Error())
			return nil
		}
	}

	// Init query
	query := repositories.GetUserListFromCompanyID(companyID)

	// Handle search and sort
	query, errCode, err := commonqueries.AddSearchAndSortGetAll(getAllRequest, query, entity.User{})
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
	var users []response.UserResponse
	result := query.Find(&users)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return user in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: response.GetAll{
			List:  users,
			Total: count,
		},
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// UpdateInfo updates user's information based on
// the username in the JWT token, excluding username & password,
// IsAdmin and IsRoot. Those cannot be changed using this API.
// Param is the full User entity model
func UpdateInfo(c *fiber.Ctx) error {
	// Parse user to match entity model
	userRequest := entity.User{}
	if err := c.BodyParser(&userRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Get user from db by username
	user := entity.User{}
	result := database.DB.Db.Where("username = ?", username).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "user not found", constants.ERR_USER_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Fill in metadata
	userRequest.BaseEntityNoIDModel = helpers.EditMetaDataNoID(username, user.BaseEntityNoIDModel)
	userRequest.Username = user.Username
	userRequest.Password = user.Password
	userRequest.CompanyID = user.CompanyID
	userRequest.IsAdmin = user.IsAdmin
	userRequest.IsRoot = user.IsRoot

	// Update
	result = database.DB.Db.Save(&userRequest)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return user in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: "Success",
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// ChangePassword updates user's password of the username in token.
// It checks for matching old password and validates password strength.
// Params
// { Password string
// OldPassword string }
func ChangePassword(c *fiber.Ctx) error {
	// Parse request model
	changePasswordRequest := request.ChangePasswordRequest{}
	if err := c.BodyParser(&changePasswordRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Get user from db
	user := entity.User{}
	result := database.DB.Db.Table("users").Where("username = ?",
		username).First(&user)

	// Not exist, throw error
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "user not found", constants.ERR_USER_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Compare hashed and old password
	if !helpers.CompareHashedPassword(changePasswordRequest.OldPassword, user.Password) {
		helpers.BadRequest(c,
			"Wrong password",
			constants.ERR_USER_WRONG_PASSWORD)
		return nil
	}

	// Old password matched, update new password
	if helpers.ValidatePassword(changePasswordRequest.Password) {
		// Hash new password
		hashedPassword, err := helpers.HashPassword(changePasswordRequest.Password)
		if err != nil {
			helpers.InternalServerError(c, result.Error.Error())
			return nil
		}

		// Update metadata
		user.Password = hashedPassword
		user.BaseEntityNoIDModel = helpers.EditMetaDataNoID(username, user.BaseEntityNoIDModel)

		// Update
		result := database.DB.Db.Updates(&user)
		if result.Error != nil {
			helpers.InternalServerError(c, result.Error.Error())
			return nil
		}

	} else { // Password too weak
		helpers.BadRequest(c,
			"Password too weak",
			constants.ERR_USER_WEAK_PASSWORD)
	}

	// Return user in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: "Success",
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// TODO: Root/Admin/User with manage_user & edit permission can change other users' password
func ResetPassword(c *fiber.Ctx) error {
	return nil
}

func Delete(c *fiber.Ctx) error {
	return nil
}

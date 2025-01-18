package commonqueries

import (
	"errors"
	"user/common/constants"
	"user/database"
	"user/models/entity"
	"user/repositories"

	"github.com/google/uuid"
)

func ValidateIsRootUser(username string) error {
	// Get user by username and root is true
	fetchedUser := entity.User{}
	result := database.DB.Db.Where("username = ? AND is_root = true",
		username).First(&fetchedUser)

	// Not exist or not root user, throw error
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func ValidateIsAdminUser(username string, companyID uuid.UUID) error {
	// Get user by username, companyID and is_admin is true
	var count int64
	result := database.DB.Db.
		Model(entity.User{}).
		Where("username = ?", username).
		Where("company_id = ?", companyID).
		Where("is_admin = true").
		Count(&count)

	// Not exist or not root user, throw error
	if result.Error != nil {
		return result.Error
	}

	if count == 0 {
		return errors.New("user is not admin")
	}

	return nil
}

// CheckPermission checks user's permission in RolePermission and UserPermission. If both counts are 0, then the
// user doesn't have the passed in permission
func ValidatePermission(username, permissionType, permissionLevel string, isRoot, isAdmin bool, projectIDs ...*uuid.UUID) (int, error) {
	var projectID uuid.UUID

	// Automatically allow if root
	if isRoot {
		return 0, nil
	}

	// Check if permission is project or setting
	var rolePermissionTable interface{}
	var userPermissionTable interface{}
	isProject := false
	if len(permissionType) >= 7 && permissionType[:7] == "project" {
		rolePermissionTable = entity.RoleProjectPermission{}
		userPermissionTable = entity.UserProjectPermission{}
		isProject = true
	} else {
		rolePermissionTable = entity.RoleSettingPermission{}
		userPermissionTable = entity.UserSettingPermission{}
	}

	// Validate permission is in permission list
	var permission entity.Permission
	query := database.DB.Db.Model(&entity.Permission{}).
		Where("permission_type = ?", permissionType).
		Where("permission_level = ?", permissionLevel)
	if isProject {
		query = query.Where("permission_group = ?", "project")
	} else {
		query = query.Where("permission_group = ?", "setting")
	}
	if result := query.First(&permission); result.Error != nil {
		return constants.ERR_COMMON_PERMISSION_NOT_FOUND, result.Error
	}

	// Get default permission levels with higher hierarchy than passed in level
	var higherLevels []string
	if result := database.DB.Db.
		Model(entity.Permission{}).
		Select("permission_level").
		Where("permission_type = ?", permissionType).
		Where("level_hierarchy < ?", permission.LevelHierarchy).
		Find(&higherLevels); result.Error != nil {
		return constants.ERR_COMMON_INTERNAL_SERVER_ERROR, result.Error
	}

	// If permission is project, check if project is in user's company
	// and check if user has view_all_projects permission
	if isProject {
		// Set projecdtID to query
		if len(projectID) != 0 {
			projectID = *projectIDs[0]
		} else {
			return constants.ERR_PROJECT_NOT_FOUND, errors.New("project not found")
		}

		// Check if project is in user's company
		if err := ValidateProjectBelongsToCompany(username, projectID); err != nil {
			return constants.ERR_COMMON_PROJECT_NOT_FOUND, errors.New("project not found")
		}

		// If admin then automatically allow
		if isAdmin {
			return 0, nil
		}

		// Check view_all_projects in role permissions
		var count int64
		query := repositories.ValidatePermissionFromRole(entity.RoleSettingPermission{},
			constants.PERM_VIEW_ALL_PROJECT, permissionLevel,
			username, higherLevels)
		if result := query.Count(&count); result.Error != nil {
			return constants.ERR_COMMON_INTERNAL_SERVER_ERROR, result.Error
		}
		if count != 0 {
			return 0, nil // Permission found in role, return nil as success
		}

		// Check view_all_projects in user permissions
		query = repositories.ValidatePermissionFromUser(entity.UserSettingPermission{},
			constants.PERM_VIEW_ALL_PROJECT, permissionLevel,
			username, higherLevels)
		if result := query.Count(&count); result.Error != nil {
			return constants.ERR_COMMON_INTERNAL_SERVER_ERROR, result.Error
		}
		if count != 0 {
			return 0, nil // Permission found in user permission, return nil as success
		}
	}

	// For a setting permission, if user is admin then automatically allow
	if !isProject && isAdmin {
		return 0, nil
	}

	// Find in RolePermission
	var count int64
	query = repositories.ValidatePermissionFromRole(rolePermissionTable,
		permissionType, permissionLevel,
		username, higherLevels)
	if isProject {
		query = query.Where("project_id = ?", projectID)
	}
	if result := query.Count(&count); result.Error != nil {
		return constants.ERR_COMMON_INTERNAL_SERVER_ERROR, result.Error
	}
	if count != 0 {
		return 0, nil // Permission found in role, return nil as success
	}

	// Permission not found in any of user's roles, try to find permission in UserPermission
	query = repositories.ValidatePermissionFromUser(userPermissionTable,
		permissionType, permissionLevel,
		username, higherLevels)
	if isProject {
		query = query.Where("project_id = ?", projectID)
	}
	if result := query.Count(&count); result.Error != nil {
		return constants.ERR_COMMON_INTERNAL_SERVER_ERROR, result.Error
	}
	if count != 0 {
		return 0, nil // Permission found in UserPermission, return nil as success
	}

	return constants.ERR_COMMON_PERMISSION_NOT_ALLOWED, errors.New("no permission")
}

// CheckProjectBelongsToCompany check the projectID belongs to this user's company
// Return nil if it does, error if it doesn't
func ValidateProjectBelongsToCompany(username string, projectID uuid.UUID) error {
	var count int64
	result := database.DB.Db.Model(entity.Company{}).
		Where("id IN (?)", database.DB.Db.Model(entity.User{}).
			Select("company_id").
			Where("username = ?", username)).
		Count(&count)

	if result.Error != nil {
		return result.Error
	}

	if count == 0 {
		return errors.New("project not found")
	}

	return nil
}

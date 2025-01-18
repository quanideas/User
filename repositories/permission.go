package repositories

import (
	"user/database"
	"user/models/entity"

	"gorm.io/gorm"
)

func GetAllPermissions() *gorm.DB {
	return database.DB.Db.
		Model(entity.Permission{})
}

func ValidatePermissionFromRole(model interface{}, permissionType, permissionLevel, requesterUsername string, higherLevels []string) *gorm.DB {
	return database.DB.Db.
		Model(model).
		Where("permission_type = ?", permissionType).
		Where("permission_level = ? OR permission_level IN (?)", permissionLevel, higherLevels).
		Where("role_id IN (?)",
			database.DB.Db.Model(entity.UserRoleMap{}).Select("role_id").Where("user_username = ?", requesterUsername))
}

func ValidatePermissionFromUser(model interface{}, permissionType, permissionLevel, requesterUsername string, higherLevels []string) *gorm.DB {
	return database.DB.Db.
		Model(model).
		Where("permission_type = ?", permissionType).
		Where("permission_level = ? OR permission_level IN (?)", permissionLevel, higherLevels).
		Where("username = ?", requesterUsername)
}

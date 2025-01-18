package repositories

import (
	"user/database"
	"user/models/entity"

	"gorm.io/gorm"
)

func GetRoleListFromCompanyID(companyID string) *gorm.DB {
	return database.DB.Db.
		Model(entity.Role{}).
		Where("company_id = ?", companyID)
}

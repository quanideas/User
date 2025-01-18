package repositories

import (
	"user/database"

	"gorm.io/gorm"
)

func GetCompanyIDOfUser(username string) *gorm.DB {
	return database.DB.Db.Raw("SELECT company_id FROM users WHERE username = ?", username)
}

func GetUserListFromCompanyID(companyID string) *gorm.DB {
	return database.DB.Db.
		Table("users").
		Where("company_id = ?", companyID)
}

func GetUserByUsername(username string) *gorm.DB {
	return database.DB.Db.
		Table("users").
		Where("username = ?", username)
}

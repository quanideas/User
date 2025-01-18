package repositories

import (
	"user/database"
	"user/models/entity"

	"gorm.io/gorm"
)

func GetDashboardStats() *gorm.DB {
	return database.DB.Db.Model(entity.User{}).Select("(SELECT COUNT(*) FROM users) AS user_count",
		"(SELECT COUNT(*) FROM companies) AS company_count",
		"(SELECT COUNT(*) FROM projects) AS project_count")
}

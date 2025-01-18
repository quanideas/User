package repositories

import (
	"user/database"

	"gorm.io/gorm"
)

func GetCompanyList() *gorm.DB {
	return database.DB.Db.
		Table("companies AS c").
		Select("id", "name", "city", "country", "is_disabled", "created_by", "created_time", "modified_by", "modified_time",
			"(SELECT count(*) FROM users AS u WHERE u.company_id = c.id) AS user_count",
			"(SELECT count(*) FROM projects AS p WHERE p.company_id = c.id) AS project_count")
}

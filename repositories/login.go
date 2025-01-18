package repositories

import (
	"user/database"

	"gorm.io/gorm"
)

func GetUserWithCompanyName(username string) *gorm.DB {
	return database.DB.Db.Raw("SELECT u.company_id, u.username, u.password, u.email, u.first_name, u.middle_name, u.last_name, u.language, u.is_root, u.is_admin, c.name AS name "+
		"FROM users u, companies c WHERE u.username = ? AND u.company_id = c.id", username)
}

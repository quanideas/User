package entity

import (
	"github.com/google/uuid"
)

const TableNameUser = "users"

// User mapped from table <users>
type User struct {
	BaseEntityNoIDModel
	Username   string    `gorm:"type:VARCHAR(30);column:username;primaryKey" json:"username"`
	CompanyID  uuid.UUID `gorm:"type:varchar(36);column:company_id" json:"company_id"`
	Password   string    `gorm:"type:NVARCHAR(100);column:password;not null" json:"password"`
	Email      *string   `gorm:"type:NVARCHAR(100);column:email;" json:"email"`
	FirstName  string    `gorm:"type:NVARCHAR(20);column:first_name;not null" json:"first_name"`
	MiddleName *string   `gorm:"type:NVARCHAR(50);column:middle_name" json:"middle_name"`
	LastName   string    `gorm:"type:NVARCHAR(20);column:last_name;not null" json:"last_name"`
	Language   *string   `gorm:"type:NVARCHAR(20);column:language" json:"language"`
	IsRoot     bool      `gorm:"type:BOOLEAN;column:is_root;not null" json:"is_root"`
	IsAdmin    bool      `gorm:"type:BOOLEAN;column:is_admin;not null" json:"is_admin"`

	UserSettingPermission []UserSettingPermission `gorm:"foreignKey:username;constraint:OnDelete:CASCADE;" json:"-"`
	UserProjectPermission []UserProjectPermission `gorm:"foreignKey:username;constraint:OnDelete:CASCADE;" json:"-"`
	Roles                 []Role                  `gorm:"many2many:user_role_maps;joinForeignKey:user_username;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName User's table name
func (*User) TableName() string {
	return TableNameUser
}

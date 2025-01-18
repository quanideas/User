package entity

import (
	"github.com/google/uuid"
)

const TableNameRole = "roles"
const TableNameUserRoleMap = "user_role_maps"

// Role mapped from table <roles>
type Role struct {
	BaseEntityModel
	CompanyID uuid.UUID `gorm:"type:varchar(36);column:company_id;not null" json:"company_id"`
	Name      string    `gorm:"type:NVARCHAR(100);column:name;not null" json:"name"`

	Users []User `gorm:"many2many:user_role_maps;joinForeignKey:role_id;constraint:OnDelete:CASCADE"`

	RoleSettingPermissions []RoleSettingPermission `gorm:"foreignKey:role_id;constraint:OnDelete:CASCADE;" json:"-"`
	RoleProjectPermissions []RoleProjectPermission `gorm:"foreignKey:role_id;constraint:OnDelete:CASCADE;" json:"-"`
}

// TableName Role's table name
func (*Role) TableName() string {
	return TableNameRole
}

// Role mapped from table <roles>
type UserRoleMap struct {
	RoleID   uuid.UUID `gorm:"type:varchar(36);column:role_id;not null;primaryKey" json:"role_id"`
	Username string    `gorm:"type:VARCHAR(30);column:user_username;not null;primaryKey" json:"username"`
}

// TableName Role's table name
func (*UserRoleMap) TableName() string {
	return TableNameUserRoleMap
}

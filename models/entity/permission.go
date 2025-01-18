package entity

import "github.com/google/uuid"

const TableNamePermission = "permissions"

const TableNameRoleSettingPermission = "role_setting_permissions"
const TableNameRoleProjectPermission = "role_project_permissions"

const TableNameUserSettingPermission = "user_setting_permissions"
const TableNameUserProjectPermission = "user_project_permissions"

const TableNamePermissionCache = "permission_caches"

// PermissionType mapped from table <permission_types>
type Permission struct {
	PermissionGroup string `gorm:"type:NVARCHAR(10);column:permission_group;not null;primaryKey" json:"permission_group"`
	PermissionType  string `gorm:"type:NVARCHAR(50);column:permission_type;not null;primaryKey" json:"permission_type"`
	PermissionLevel string `gorm:"type:NVARCHAR(20);column:permission_level;primaryKey" json:"permission_level"`
	LevelHierarchy  int    `gorm:"type:INT;column:level_hierarchy" json:"level_hierarchy"`
	Descriptions    string `gorm:"type:NVARCHAR(500);column:descriptions" json:"descriptions"`
}

// TableName PermissionType's table name
func (*Permission) TableName() string {
	return TableNamePermission
}

// RoleSettingPermission mapped from table <role_setting_permissions>
type RoleSettingPermission struct {
	// BaseEntityModel
	ID              uuid.UUID `gorm:"type:varchar(36);column:id;not null" json:"id"`
	RoleID          uuid.UUID `gorm:"type:varchar(36);column:role_id;not null;primaryKey" json:"role_id"`
	PermissionType  string    `gorm:"type:NVARCHAR(50);column:permission_type;not null;primaryKey" json:"permission_type"`
	PermissionLevel string    `gorm:"type:NVARCHAR(20);column:permission_level;primaryKey" json:"permission_level"`
}

// TableName RolePermission's table name
func (*RoleSettingPermission) TableName() string {
	return TableNameRoleSettingPermission
}

// RoleProjectPermission mapped from table <role_project_permissions>
type RoleProjectPermission struct {
	// BaseEntityModel
	ID              uuid.UUID `gorm:"type:varchar(36);column:id;not null" json:"id"`
	RoleID          uuid.UUID `gorm:"type:varchar(36);column:role_id;not null;primaryKey" json:"role_id"`
	ProjectID       uuid.UUID `gorm:"type:varchar(36);column:project_id;not null;primaryKey" json:"project_id"`
	PermissionType  string    `gorm:"type:NVARCHAR(50);column:permission_type;not null;primaryKey" json:"permission_type"`
	PermissionLevel string    `gorm:"type:NVARCHAR(20);column:permission_level;primaryKey" json:"permission_level"`
}

// TableName RolePermission's table name
func (*RoleProjectPermission) TableName() string {
	return TableNameRoleProjectPermission
}

// UserPermission mapped from table <user_permissions>
type UserSettingPermission struct {
	ID              uuid.UUID `gorm:"type:varchar(36);column:id;not null" json:"id"`
	Username        string    `gorm:"type:varchar(36);column:username;not null;primaryKey" json:"username"`
	PermissionType  string    `gorm:"type:NVARCHAR(50);column:permission_type;not null;primaryKey" json:"permission_type"`
	PermissionLevel string    `gorm:"type:NVARCHAR(20);column:permission_level;primaryKey" json:"permission_level"`
}

// TableName UserPermission's table name
func (*UserSettingPermission) TableName() string {
	return TableNameUserSettingPermission
}

// UserPermission mapped from table <user_permissions>
type UserProjectPermission struct {
	ID              uuid.UUID `gorm:"type:varchar(36);column:id;not null" json:"id"`
	Username        string    `gorm:"type:varchar(36);column:username;not null;primaryKey" json:"username"`
	ProjectID       uuid.UUID `gorm:"type:varchar(36);column:project_id;not null;primaryKey" json:"project_id"`
	PermissionType  string    `gorm:"type:NVARCHAR(50);column:permission_type;not null;primaryKey" json:"permission_type"`
	PermissionLevel string    `gorm:"type:NVARCHAR(20);column:permission_level;primaryKey" json:"permission_level"`
}

// TableName UserPermission's table name
func (*UserProjectPermission) TableName() string {
	return TableNameUserProjectPermission
}

type PermissionCache struct {
	Username        string     `gorm:"type:varchar(36);column:username;not null;primaryKey" json:"username"`
	ProjectID       *uuid.UUID `gorm:"type:varchar(36);column:project_id;primaryKey" json:"project_id"`
	PermissionType  string     `gorm:"type:NVARCHAR(50);column:permission_type;not null;primaryKey" json:"permission_type"`
	PermissionLevel string     `gorm:"type:NVARCHAR(20);column:permission_level;primaryKey" json:"permission_level"`
	Granted         bool       `gorm:"type:BOOLEAN;column:granted;not null" json:"granted"`
}

// TableName PermissionCache's table name
func (*PermissionCache) TableName() string {
	return TableNamePermissionCache
}

package response

import (
	"user/models/entity"

	"github.com/google/uuid"
)

type LoginSuccessResponse struct {
	UserGetByUsernameResponse
	CompanyName string `gorm:"type:NVARCHAR(200);column:name;not null" json:"CompanyName"`
	Token       string `json:"token"`
}

type UserResponse struct {
	entity.BaseEntityNoIDModel
	Username   string  `json:"username"`
	Email      *string `json:"email"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	LastName   string  `json:"last_name"`
	IsAdmin    bool    `json:"is_admin"`
}

type UserGetByUsernameResponse struct {
	Username   string  `json:"username"`
	Email      *string `json:"email"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	LastName   string  `json:"last_name"`
	Language   *string `json:"language"`
	IsRoot     bool    `json:"is_root"`
	IsAdmin    bool    `json:"is_admin"`
}

type UserGetSpecificPermissionResponse struct {
	ProjectID       *uuid.UUID `json:"project_id"`
	PermissionType  string     `json:"permission_type"`
	PermissionLevel string     `json:"permission_level"`
}

type UserGetPermissionByUsernameResponse struct {
	SettingPermissions []GetByIDSettingPermissionResponse `json:"SettingPermissions" mapper:"-"`
	ProjectPermissions []GetByIDProjectPermissionResponse `json:"ProjectPermissions" mapper:"-"`
}

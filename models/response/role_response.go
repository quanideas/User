package response

import (
	"user/models/entity"
)

type RoleResponse struct {
	entity.BaseEntityModel
	Name string `json:"name"`
}

type RoleGetByIDUserResponse struct {
	Username   string  `json:"username"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	LastName   string  `json:"last_name"`
}

type RoleGetByIDResponse struct {
	entity.BaseEntityModel
	Name               string                             `json:"name"`
	Users              []RoleGetByIDUserResponse          `json:"Users" mapper:"-"`
	SettingPermissions []GetByIDSettingPermissionResponse `json:"SettingPermissions" mapper:"-"`
	ProjectPermissions []GetByIDProjectPermissionResponse `json:"ProjectPermissions" mapper:"-"`
}

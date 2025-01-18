package response

import "github.com/google/uuid"

type BaseResponse struct {
	Meta struct {
		Status int
	}
	Data interface{}
}

type GetAll struct {
	List  interface{}
	Total int64
}

type GetByIDSettingPermissionResponse struct {
	ID              uuid.UUID `json:"id"`
	PermissionType  string    `json:"permission_type"`
	PermissionLevel string    `json:"permission_level"`
}

type GetByIDProjectPermissionResponse struct {
	ID              uuid.UUID `json:"id"`
	ProjectID       uuid.UUID `json:"project_id"`
	PermissionType  string    `json:"permission_type"`
	PermissionLevel string    `json:"permission_level"`
}

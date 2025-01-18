package request

import (
	"user/models/entity"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	entity.User
}

type ChangePasswordRequest struct {
	OldPassword string `json:"OldPassword"`
	Password    string `json:"Password"`
}

type GetUserSpecificPermissionRequest struct {
	ProjectID       *uuid.UUID `json:"project_id"`
	PermissionType  string     `json:"permission_type"`
	PermissionLevel string     `json:"permission_level"`
}

type UserAddPermissionRequest struct {
	Username        string     `json:"username"`
	ProjectID       *uuid.UUID `json:"project_id"`
	PermissionType  string     `json:"permission_type"`
	PermissionLevel string     `json:"permission_level"`
}

type UserGetByUsernameRequest struct {
	Username string `json:"username"`
}

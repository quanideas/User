package request

import "github.com/google/uuid"

type RoleRequest struct {
	Name string `json:"name"`
}

type RoleUpdateRequest struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type RoleAddUserRequest struct {
	RoleID   uuid.UUID `json:"role_id"`
	Username string    `json:"username"`
}

type RoleAddPermissionRequest struct {
	RoleID          uuid.UUID  `json:"role_id"`
	ProjectID       *uuid.UUID `json:"project_id"`
	PermissionType  string     `json:"permission_type"`
	PermissionLevel string     `json:"permission_level"`
}

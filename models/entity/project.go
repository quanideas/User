package entity

import (
	"github.com/google/uuid"
)

const TableNameProjectIteration = "project_iterations"
const TableNameProject = "projects"

// Project mapped from table <projects>
type Project struct {
	BaseEntityModel
	CompanyID  uuid.UUID `gorm:"type:varchar(36);column:company_id;not null" json:"company_id"`
	Name       string    `gorm:"type:NVARCHAR(200);column:name;not null" json:"name"`
	Location   *string   `gorm:"type:NVARCHAR(100);column:location" json:"location"`
	ShareLevel *string   `gorm:"type:NVARCHAR(10);column:share_level" json:"share_level"`
	ShareURL   *string   `gorm:"type:NVARCHAR(500);column:share_url" json:"share_url"`

	ProjectIterations      []ProjectIteration      `gorm:"foreignKey:project_id;constraint:OnDelete:CASCADE;" json:"-"`
	RoleProjectPermissions []RoleProjectPermission `gorm:"foreignKey:project_id;constraint:OnDelete:CASCADE;" json:"-"`
	UserProjectPermissions []UserProjectPermission `gorm:"foreignKey:project_id;constraint:OnDelete:CASCADE;" json:"-"`
}

// TableName Project's table name
func (*Project) TableName() string {
	return TableNameProject
}

// ProjectIteration mapped from table <project_iterations>
type ProjectIteration struct {
	BaseEntityModel
	ProjectID          uuid.UUID `gorm:"type:varchar(36);column:project_id;not null" json:"project_id"`
	Revision           *string   `gorm:"type:NVARCHAR(100);column:revision" json:"revision"`
	GeoJSONURL         *string   `gorm:"type:NVARCHAR(500);column:geojson_url" json:"geojson_url"`
	GeoJSONFileName    *string   `gorm:"type:NVARCHAR(100);column:geojson_file_name" json:"geojson_file_name"`
	OrthoPhotoURL      *string   `gorm:"type:NVARCHAR(500);column:ortho_photo_url" json:"ortho_photo_url"`
	OrthoPhotoFileName *string   `gorm:"type:NVARCHAR(100);column:ortho_photo_file_name" json:"ortho_photo_file_name"`
	Tile3DURL          *string   `gorm:"type:NVARCHAR(500);column:tile_3d_url" json:"tile_3d_url"`
	Tile3DFileName     *string   `gorm:"type:NVARCHAR(100);column:tile_3d_file_name" json:"tile_3d_file_name"`
}

// TableName ProjectIteration's table name
func (*ProjectIteration) TableName() string {
	return TableNameProjectIteration
}

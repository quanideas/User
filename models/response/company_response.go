package response

import "user/models/entity"

type CompanyResponse struct {
	entity.BaseEntityModel
	Name         string `json:"name"`
	City         string `json:"city"`
	Country      string `json:"country"`
	IsDisabled   bool   `json:"is_disabled"`
	UserCount    int64  `gorm:"column:user_count" json:"UserCount"`
	ProjectCount int64  `gorm:"column:project_count" json:"ProjectCount"`
}

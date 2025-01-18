package response

type DashboardStatsResponse struct {
	UserCount    int64 `gorm:"column:user_count" json:"UserCount"`
	CompanyCount int64 `gorm:"column:company_count" json:"CompanyCount"`
	ProjectCount int64 `gorm:"column:project_count" json:"ProjectCount"`
}

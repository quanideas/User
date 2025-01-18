package sqlresponse

import (
	"user/models/response"

	"github.com/google/uuid"
)

type GetUserWithCompanyNameResponse struct {
	response.UserGetByUsernameResponse
	Password    string    `gorm:"type:NVARCHAR(100);column:password;not null" json:"password"`
	CompanyName string    `gorm:"type:NVARCHAR(200);column:name;not null" json:"CompanyName"`
	CompanyID   uuid.UUID `gorm:"type:varchar(36);column:company_id" json:"company_id"`
}

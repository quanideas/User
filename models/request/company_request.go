package request

import (
	"user/models/entity"

	"github.com/google/uuid"
)

type CompanyEditRequest struct {
	entity.Company
	CompanyID uuid.UUID `json:"company_id"`
}

package rolehandlers

import (
	"errors"
	"user/database"
	"user/models/entity"

	"github.com/google/uuid"
)

// validateRoleBelongToCompany checks if role belongs to requester's company to prevent modifying permissions
// of other companies.
// Return nil if it does, error if it doesn't
func validateRoleBelongToCompany(requesterUsername string, roleID uuid.UUID) error {
	// Get requester's companyID from username
	var companyID string
	if result := database.DB.Db.
		Model(entity.User{}).
		Select("company_id").
		Where("username = ?", requesterUsername).
		First(&companyID); result.Error != nil {
		return result.Error
	}

	// Check if role belongs to requester's company
	var count int64
	result := database.DB.Db.
		Model(&entity.Role{}).
		Where("id = ?", roleID).
		Where("company_id = ?", companyID).
		Count(&count)

	if result.Error != nil {
		return result.Error
	} else if count == 0 { // Permission doesn't exist in the company
		return errors.New("permission not found")
	}

	return nil
}

package userhandlers

import (
	"errors"
	"user/database"
	"user/models/entity"
)

// validatePermissionIsInUserCompany checks if user of the permission belongs to requester user's company
// Return nil if it does, error if it doesn't
func validateUserIsInRequestersCompany(requesterUsername, permissionUsername string) error {
	// Get user's companyID from username
	var requesterCompanyID string
	if result := database.DB.Db.Model(entity.User{}).Select("company_id").Where("username = ?", requesterUsername).First(&requesterCompanyID); result.Error != nil {
		return result.Error
	}

	// Get user's companyID from username
	var permissionCompanyID string
	if result := database.DB.Db.Model(entity.User{}).Select("company_id").Where("username = ?", permissionUsername).First(&permissionCompanyID); result.Error != nil {
		return result.Error
	}

	if requesterCompanyID != permissionCompanyID {
		return errors.New("permission not found")
	}

	return nil
}

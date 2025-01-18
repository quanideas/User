package commonqueries

import (
	"errors"
	"strings"
	"user/common/constants"
	"user/common/helpers"
	"user/models/request"

	"gorm.io/gorm"
)

func AddSearchAndSortGetAll(getAllRequest request.GetAll, query *gorm.DB, model interface{}) (*gorm.DB, int, error) {
	// Search
	for i := 0; i < len(getAllRequest.Search); i++ {
		searchBy := getAllRequest.Search[i].By
		searchValue := getAllRequest.Search[i].Value

		// Validate field exists
		if isBool, err := helpers.ValidateFieldGetAllQuery(searchBy, model); err == nil {

			// Field is boolean, search by 1 and 0 instead of true false
			if isBool {
				if strings.ToLower(searchValue) == "true" || searchValue == "1" {
					query = query.Where(searchBy + " = 1")
				} else {
					query = query.Where(searchBy + " = 0")
				}
			} else {
				query = query.Where(searchBy+" LIKE ?", "%"+searchValue+"%")
			}
		} else { // Regular case, search by LIKE
			return nil, constants.ERR_COMMON_FIELD_NOT_FOUND, errors.New("field not found")
		}
	}

	// Sort
	sortedByModified := false
	for i := 0; i < len(getAllRequest.Sort); i++ {
		// Validate field exists
		if _, err := helpers.ValidateFieldGetAllQuery(getAllRequest.Sort[i].By, model); err == nil {
			query = query.Order(getAllRequest.Sort[i].By + " " + getAllRequest.Sort[i].Type)
		}

		// Check if request wants to sort by modified time, then don't add default sort by modified time below
		if getAllRequest.Sort[i].By == "modified_time" {
			sortedByModified = true
		}
	}
	if !sortedByModified {
		query = query.Order("modified_time desc") // If not sorted by modified time then always sort by modified time last
	}

	return query, 0, nil
}

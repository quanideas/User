package helpers

import (
	"errors"
	"reflect"
	"unicode"
	"user/common/constants"
	"user/models/request"
)

func ValidatePassword(password string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	hasMinLen = len(password) >= 6

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

func ValidateFieldGetAllQuery(fieldName string, object interface{}) (bool, error) {
	obj := reflect.ValueOf(object)

	// Iterate through each field
	for i := 0; i < obj.Type().NumField(); i++ {
		objField := obj.Type().Field(i)

		// Json name of field found (json name matches db field name)
		if jsonTag := objField.Tag.Get("json"); jsonTag != "" && jsonTag != "-" && jsonTag == fieldName {
			if objField.Type.Kind() == reflect.Bool {
				return true, nil
			} else {
				return false, nil
			}
		}
	}

	return false, errors.New("field not found")
}

func ValidateGetAllRequest(request request.GetAll) (int, error) {
	// Check for page limit
	if request.Count > 100 {
		return constants.ERR_COMMON_REQUEST_TOO_LARGE, errors.New("Request too large")
	}

	return 0, nil
}

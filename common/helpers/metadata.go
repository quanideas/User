package helpers

import (
	"time"
	"user/models/entity"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreateMetaData(username string) entity.BaseEntityModel {
	t := time.Now()
	return entity.BaseEntityModel{
		ID:           uuid.New(),
		CreatedBy:    username,
		CreatedTime:  t,
		ModifiedBy:   &username,
		ModifiedTime: &t,
	}
}

func CreateMetaDataNoID(username string) entity.BaseEntityNoIDModel {
	t := time.Now()
	return entity.BaseEntityNoIDModel{
		CreatedBy:    username,
		CreatedTime:  t,
		ModifiedBy:   &username,
		ModifiedTime: &t,
	}
}

func EditMetaData(username string, baseEntityFromDB entity.BaseEntityModel) entity.BaseEntityModel {
	t := time.Now()

	baseEntityFromDB.ModifiedBy = &username
	baseEntityFromDB.ModifiedTime = &t

	return baseEntityFromDB
}

func EditMetaDataNoID(modifierUsername string, baseEntityFromDB entity.BaseEntityNoIDModel) entity.BaseEntityNoIDModel {
	t := time.Now()

	baseEntityFromDB.ModifiedBy = &modifierUsername
	baseEntityFromDB.ModifiedTime = &t

	return baseEntityFromDB
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CompareHashedPassword(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

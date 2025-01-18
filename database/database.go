package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"user/common/helpers"
	"user/models/entity"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbInstance struct {
	Db *gorm.DB
}

var DB DbInstance

func ConnectDb() {
	// Connect to Postgres db via Gorm
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := gorm.Open(mysql.Open(connectionStr), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info),
		TranslateError: true})

	// Handle connection fail and ping
	if err != nil {
		log.Fatal("Failed to connect.\n", err)
	} else {
		log.Println("connected")
	}

	sqlDb, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get sql db.\n", err)
	}

	if err := sqlDb.Ping(); err != nil {
		log.Fatal("Failed to ping Postgres db.\n", err)
	}

	err = db.SetupJoinTable(&entity.Role{}, "Users", &entity.UserRoleMap{})
	err = db.SetupJoinTable(&entity.User{}, "Roles", &entity.UserRoleMap{})

	if err = db.AutoMigrate(
		&entity.Permission{},
		&entity.Company{},
		&entity.User{},
		&entity.Project{},
		&entity.Role{},
		&entity.RoleSettingPermission{},
		&entity.RoleProjectPermission{},
		&entity.UserSettingPermission{},
		&entity.UserProjectPermission{},
		&entity.ProjectIteration{},
		&entity.PermissionCache{}); err != nil {
		log.Fatal("failed migrating db.\n", err)
	}

	if db.Migrator().HasTable(&entity.User{}) {
		if err := db.First(&entity.User{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {

			if password, err := helpers.HashPassword(os.Getenv("DEFAULT_ROOT_PASSWORD")); err != nil {
				log.Fatal("error hashing password\n", err)
			} else {
				// Metadata
				defaultRootUsername := os.Getenv("DEFAULT_ROOT_USERNAME")
				currentTime := time.Now()

				// Create root company
				company := entity.Company{
					BaseEntityModel: entity.BaseEntityModel{
						ID:           uuid.New(),
						CreatedBy:    defaultRootUsername,
						CreatedTime:  currentTime,
						ModifiedBy:   &defaultRootUsername,
						ModifiedTime: &currentTime,
					},
					Name:         "Root",
					AddressLine1: "Root",
					City:         "Root",
					Country:      "Root",
					IsDisabled:   false,
				}
				result := db.Create(&company)
				if result.Error != nil {
					log.Fatal("failed migrating records.\n", err)
				}

				// Create root user
				user := entity.User{
					BaseEntityNoIDModel: entity.BaseEntityNoIDModel{
						CreatedBy:    defaultRootUsername,
						CreatedTime:  currentTime,
						ModifiedBy:   &defaultRootUsername,
						ModifiedTime: &currentTime,
					},
					CompanyID: company.ID,
					Username:  os.Getenv("DEFAULT_ROOT_USERNAME"),
					Password:  password,
					FirstName: "Root",
					LastName:  "Root",
					IsRoot:    false,
				}
				result = db.Create(&user)
				if result.Error != nil {
					log.Fatal("failed migrating records.\n", err)
				}
			}

		}
	}

	// Set global db instance to connected db
	DB = DbInstance{
		Db: db,
	}
}

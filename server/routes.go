package server

import (
	"os"
	"user/common/constants"
	companyhandlers "user/handlers/company"
	dashboardhandlers "user/handlers/dashboard"
	healthcheck "user/handlers/health-check"
	permissionhandlers "user/handlers/permission"
	rolehandlers "user/handlers/role"
	userhandlers "user/handlers/user"
	"user/server/middlewares"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupRoutes(app *fiber.App) {
	allowedDevOrigins := os.Getenv("ALLOWED_DEV_ORIGINS")
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

	// Apply CORS
	if os.Getenv("ENVIRONMENT") == "development" {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     allowedDevOrigins,
			AllowCredentials: true,
		}))
	} else {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     allowedOrigins,
			AllowCredentials: true,
		}))
	}

	// Recover from panics
	app.Use(middlewares.CatchPanic())

	// Unauthenticated
	app.Get("/health-check", healthcheck.HealthCheck)
	app.Get("/connection-check", healthcheck.ConnectionCheck)
	app.Post(constants.UserLogin, userhandlers.Login)

	// JWT Middleware
	app.Use(middlewares.ValidateJWT())

	// Authenticated
	// Company
	companyRoutes := app.Group("/company")
	companyRoutes.Post(constants.CompanyCreate, companyhandlers.Create)
	companyRoutes.Get(constants.CompanyGetByID, companyhandlers.GetByID)
	companyRoutes.Post(constants.CompanyGetByID, companyhandlers.GetByID)
	companyRoutes.Post(constants.CompanyGetAll, companyhandlers.GetAll)
	companyRoutes.Post(constants.CompanyUpdate, companyhandlers.Update)

	// User
	userRoutes := app.Group("/user")
	userRoutes.Post(constants.UserAddPermission,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_CREATE),
		userhandlers.AddPermission)
	userRoutes.Post(constants.UserRemovePermission,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_EDIT),
		userhandlers.RemovePermission)
	userRoutes.Post(constants.UserGetPermissionByUsername,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_VIEW),
		userhandlers.GetPermissionByUsername)
	userRoutes.Post(constants.UserCreate,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_USER, constants.PERM_LEVEL_CREATE),
		userhandlers.Create)
	userRoutes.Get(constants.UserGetByID, userhandlers.GetByUsername)
	userRoutes.Post(constants.UserGetByCompany,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_USER, constants.PERM_LEVEL_VIEW),
		userhandlers.GetAllByCompany)
	userRoutes.Post(constants.UserGetSpecificPermission, userhandlers.GetUserSpecificPermission)
	userRoutes.Post(constants.UserUpdate, userhandlers.UpdateInfo)

	// Role
	roleRoutes := app.Group("/role")
	roleRoutes.Post(constants.RoleCreate,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_CREATE),
		rolehandlers.Create)
	roleRoutes.Post(constants.RoleGetAll,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_VIEW),
		rolehandlers.GetAll)
	roleRoutes.Post(constants.RoleRemovePermission,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_EDIT),
		rolehandlers.RemovePermission)

	useValidateRoleBelongsToCompany := roleRoutes.Group("/", middlewares.CheckRoleBelongsToCompany)
	useValidateRoleBelongsToCompany.Post(constants.RoleGetByID,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_VIEW),
		rolehandlers.GetByIDWithUserAndPermissions)
	useValidateRoleBelongsToCompany.Post(constants.RoleUpdate,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_EDIT),
		rolehandlers.UpdateInfo)
	useValidateRoleBelongsToCompany.Post(constants.RoleDelete,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_EDIT),
		rolehandlers.Delete)
	useValidateRoleBelongsToCompany.Post(constants.RoleAddUser,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_EDIT),
		rolehandlers.AddUser)
	useValidateRoleBelongsToCompany.Post(constants.RoleRemoveUser,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_EDIT),
		rolehandlers.RemoveUser)
	useValidateRoleBelongsToCompany.Post(constants.RoleAddPermission,
		middlewares.ValidatePermimssion(constants.PERM_MANAGE_PERMISSION, constants.PERM_LEVEL_EDIT),
		rolehandlers.AddPermission)

	// Permission
	permissionRoutes := app.Group("/permission")
	permissionRoutes.Post(constants.PermissionGetAll, permissionhandlers.GetAllPermissions)
	permissionRoutes.Post(constants.PermissionValidate, permissionhandlers.ValidatePermission)

	// Dashboard
	dashboardRoutes := app.Group("/dashboard")
	dashboardRoutes.Get(constants.DashboardStats, dashboardhandlers.GetDashboardStats)
}

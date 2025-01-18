package companyhandlers

import (
	"user/common/constants"
	"user/common/helpers"
	commonqueries "user/common/queries"
	"user/database"
	"user/models/entity"
	"user/models/request"
	"user/models/response"
	"user/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Create returns a company record after it's created on db,
// and can only be used by root users.
// Params
// { company Company }
func Create(c *fiber.Ctx) error {
	// Parse company to match entity model
	company := entity.Company{}
	if err := c.BodyParser(&company); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}
	company.CurrentUsers = 0

	// Get username from token
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	isRoot := claims["is_root"].(bool)

	// Only allow when user is root user
	if !isRoot {
		helpers.BadRequest(c, "no permission to create company", constants.ERR_COMPANY_NO_PERM_TO_CREATE)
		return nil
	}

	// Fill in metadata
	company.BaseEntityModel = helpers.CreateMetaData(username)
	company.IsDisabled = false

	// Call db to create
	result := database.DB.Db.Create(&company)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return created company
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: company,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// GetByID returns the company's information based on
// the companyID in request if user is root.
// Returns requester's own company if user is not root.
// Params (optional, only used if Post by root)
// { CompanyID uuid.UUID }
func GetByID(c *fiber.Ctx) error {
	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	// username := claims["username"].(string)
	isRoot := claims["is_root"].(bool)
	companyID := claims["company_id"].(string)

	// If method is post and user is root, then get companyID from json body
	if c.Method() == "POST" {
		if isRoot {
			// Parse request
			idRequest := request.GetByIDRequest{}
			if err := c.BodyParser(&idRequest); err != nil {
				helpers.InternalServerError(c, err.Error())
				return err
			}

			companyID = idRequest.ID
		} else {
			helpers.BadRequest(c, "no permission", constants.ERR_COMMON_PERMISSION_NOT_ALLOWED)
			return nil
		}
	}

	// Check if company exists on db
	var fetchedCompany entity.Company
	if result := database.DB.Db.
		Model(&entity.Company{}).
		Where("id = ?", companyID).
		First(&fetchedCompany); result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "company not found", constants.ERR_COMPANY_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
	}

	// Return found company
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: fetchedCompany,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

func GetAll(c *fiber.Ctx) error {
	// Parse search query model
	getAllRequest := request.GetAll{}
	if err := c.BodyParser(&getAllRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	// username := claims["username"].(string)
	isRoot := claims["is_root"].(bool)

	// Only allow if root
	if !isRoot {
		helpers.BadRequest(c, "Not allowed", constants.ERR_COMPANY_GET_ALL_NOT_ALLOWED)
		return nil
	}

	// Init query
	query := repositories.GetCompanyList()

	// Handle search and sort
	query, errCode, err := commonqueries.AddSearchAndSortGetAll(getAllRequest, query, entity.Company{})
	if err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Get total number
	var count int64
	query.Count(&count)

	// Limit
	query = query.Offset((getAllRequest.Page - 1) * getAllRequest.Count).Limit(getAllRequest.Count)

	// Get
	var companies []response.CompanyResponse
	result := query.Find(&companies)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return user in request
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: response.GetAll{
			List:  companies,
			Total: count,
		},
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// Updates returns updated company's information if successful.
// Can only be used by root users or admins of requested company.
// Params
// { company Company }
func Update(c *fiber.Ctx) error {
	// Parse company to match entity model
	companyRequest := request.CompanyEditRequest{}
	if err := c.BodyParser(&companyRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	companyID := claims["company_id"].(string)
	isRoot := claims["is_root"].(bool)
	isAdmin := claims["is_admin"].(bool)

	// Only allow root or admin to update company
	if isRoot {
		companyID = companyRequest.CompanyID.String()
	} else if !isAdmin {
		helpers.BadRequest(c, "no permission to update company", constants.ERR_COMPANY_NO_PERM_TO_UPDATE)
		return nil
	}

	// Check if company exists on db
	var fetchedCompany entity.Company
	result := database.DB.Db.Model(&entity.Company{}).Where("id = ?", companyID).First(&fetchedCompany)
	if result.Error != nil {
		helpers.BadRequest(c, "company not found", constants.ERR_COMPANY_NOT_FOUND)
		return nil
	}

	// Map metadata to user input's company
	companyRequest.Company.BaseEntityModel = fetchedCompany.BaseEntityModel
	companyRequest.Company.CurrentUsers = fetchedCompany.CurrentUsers // Don't update number of users
	companyRequest.Company.BaseEntityModel = helpers.EditMetaData(username, fetchedCompany.BaseEntityModel)
	if !isRoot { // Only root can disable other companies and max users
		companyRequest.Company.IsDisabled = fetchedCompany.IsDisabled
		companyRequest.Company.MaxUsers = fetchedCompany.MaxUsers
	}

	// Update
	result = database.DB.Db.Save(&companyRequest.Company)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return edited company
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: companyRequest.Company,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

func Delete(c *fiber.Ctx) error {
	return nil
}

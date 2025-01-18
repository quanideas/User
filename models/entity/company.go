package entity

const TableNameCompany = "companies"

// Company mapped from table <companies>
type Company struct {
	BaseEntityModel
	Name         string  `gorm:"type:NVARCHAR(200);column:name;not null" json:"name"`
	AddressLine1 string  `gorm:"type:NVARCHAR(100);column:address_line_1;not null" json:"address_line_1"`
	AddressLine2 *string `gorm:"type:NVARCHAR(100);column:address_line_2" json:"address_line_2"`
	City         string  `gorm:"type:NVARCHAR(50);column:city;not null" json:"city"`
	State        *string `gorm:"type:NVARCHAR(50);column:state" json:"state"`
	Country      string  `gorm:"type:NVARCHAR(50);column:country;not null" json:"country"`
	ZipCode      *string `gorm:"type:NVARCHAR(10);column:zip_code" json:"zip_code"`
	IsDisabled   bool    `gorm:"type:BOOLEAN;column:is_disabled;not null" json:"is_disabled"`
	CurrentUsers int     `gorm:"type:INT;column:current_users" json:"current_users"`
	MaxUsers     int     `gorm:"type:INT;column:max_users" json:"max_users"`

	Roles    []Role    `gorm:"foreignKey:company_id;constraint:OnDelete:CASCADE;" json:"-"`
	Projects []Project `gorm:"foreignKey:company_id;constraint:OnDelete:CASCADE;" json:"-"`
	Users    []User    `gorm:"foreignKey:company_id;constraint:OnDelete:CASCADE;" json:"-"`
}

// TableName Company's table name
func (*Company) TableName() string {
	return TableNameCompany
}

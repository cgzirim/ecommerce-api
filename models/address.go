package models

// Address represents a user's address.
type Address struct {
	BaseModel
	FirstName     string `gorm:"varchar(255);not null" json:"first_name"`
	LastName      string `gorm:"varchar(255);not null" json:"last_name"`
	City          string `gorm:"varchar(50);not null" json:"city"`
	Country       string `gorm:"varchar(50);not null" json:"country"`
	ZipCode       string `gorm:"varchar(15);not null" json:"zip_code"`
	StreetAddress string `gorm:"varchar(255);not null" json:"street_address"`

	UserID uint `gorm:"not null" json:"user_id"`
}

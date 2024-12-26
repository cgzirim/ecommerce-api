package models

type Category struct {
	BaseModel
	Name        string `gorm:"not null"`
	Description string
	Products    []Product
}

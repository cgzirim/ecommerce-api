package models

// Product represents a product in the store.
type Product struct {
	BaseModel
	Name        string  `gorm:"varchar(255);not null" json:"name"`
	Category    string  `gorm:"varchar(255);not null" json:"category"`
	Description string  `gorm:"type:text" json:"description"`
	Price       float64 `gorm:"not null;check:price_gt_zero,price > 0" json:"price"`
	Stock       int     `gorm:"not null;check:stock_non_negative,stock >= 0" json:"stock"`
}

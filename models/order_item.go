package models

// OrderItem represents an item in an order.
type OrderItem struct {
	BaseModel
	Order     Order   `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"-"`
	OrderID   uint    `gorm:"not null" json:"order_id"`
	Product   Product `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT" json:"-"`
	ProductID uint    `gorm:"not null" json:"product_id"`
	Price     float64 `gorm:"not null;check:price_gt_zero,price > 0" json:"price"`
	Quantity  int     `gorm:"not null;check:quantity_gt_zero,quantity > 0" json:"quantity"`
}

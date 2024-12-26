package models

// Order represents an order placed by a user.
type Order struct {
	BaseModel
	UserID     uint        `gorm:"not null" json:"-"`
	User       User        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
	AddressID  uint        `gorm:"default:null" json:"-"`
	Address    Address     `gorm:"foreignKey:AddressID;constraint:OnDelete:SET NULL" json:"address"`
	Total      float64     `gorm:"not null" json:"total"`
	Status     string      `gorm:"default:'pending'" json:"status"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items"`
}

const (
	OrderStatusPending   = "pending"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
)

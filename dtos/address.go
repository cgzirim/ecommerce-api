package dtos

// CreateAddressRequest represents the expected request body for creating an address
type CreateAddressRequest struct {
	FirstName     string `json:"first_name" binding:"required"`
	LastName      string `json:"last_name" binding:"required"`
	City          string `json:"city" binding:"required"`
	Country       string `json:"country" binding:"required"`
	ZipCode       string `json:"zip_code" binding:"required"`
	StreetAddress string `json:"street_address" binding:"required"`
}

// AddressDetail represents the response body for a successful address creation
type AddressDetail struct {
	ID            uint   `json:"id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	City          string `json:"city"`
	Country       string `json:"country"`
	ZipCode       string `json:"zip_code"`
	StreetAddress string `json:"street_address"`
	UserID        uint   `json:"user_id"`
	CreatedAt     string `json:"created_at" example:"2024-12-26T01:59:44.840049+01:00"`
	UpdatedAt     string `json:"updated_at" example:"2024-12-26T01:59:44.840049+01:00"`
}

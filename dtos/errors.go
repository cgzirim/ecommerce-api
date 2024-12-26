package dtos

type ErrorResponse_400 struct {
	Error string `json:"error" example:"Validation failed"`
}

type ErrorResponse_401 struct {
	Error string `json:"error" example:"Unauthenticated, login is required"`
}

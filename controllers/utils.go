package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// handleValidationErrors is a helper function that handles validation errors
func handleValidationErrors(err error, c *gin.Context) {
	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); ok {
		errorMessages := make(map[string]string)

		re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
		for _, validationErr := range validationErrors {
			// Convert the field name to snake_case from camelCase
			snakeCase := re.ReplaceAllString(validationErr.Field(), `${1}_${2}`)
			field := strings.ToLower(snakeCase)

			switch validationErr.Tag() {
			case "gt":
				errorMessages[field] = fmt.Sprintf("Value must be greater than %s.", validationErr.Param())
			case "required":
				errorMessages[field] = "This field is required."
			case "min":
				errorMessages[field] = fmt.Sprintf("Value length must be greater than or equal to %s", validationErr.Param())
			default:
				errorMessages[field] = validationErr.Error()
			}
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": errorMessages})
		return
	}

	log.Printf("Non-validation error occurred: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred. Please try again."})
}

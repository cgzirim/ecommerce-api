package middleware

import (
	"errors"
	"log"
	"strings"

	"github.com/cgzirim/ecommerce-api/db"
	"github.com/cgzirim/ecommerce-api/models"
	"github.com/cgzirim/ecommerce-api/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// LoadAuthUserMiddleware adds the authenticated user's object to the request context.
func LoadAuthUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetAuthenticatedUser(c)
		if err != nil {
			log.Printf("Failed to get authenticated user: %v, route: %s\n", err, c.FullPath())
			c.Next()
			return
		}

		if user != nil {
			c.Set("user", *user)
		}

		c.Next()
	}
}

// GetAuthenticatedUser retrieves the authenticated user from the context
func GetAuthenticatedUser(c *gin.Context) (*models.User, error) {
	userID, err := GetUserIDFromJWT(c)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	return &user, nil
}

// GetUserIDFromJWT extracts the user ID from the JWT token in the Authorization header
func GetUserIDFromJWT(c *gin.Context) (uint, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, errors.New("authorization header is missing")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	/* This check is commented out because Swagger might fail to prepend "Bearer" to the token. */

	// if tokenString == authHeader {
	// 	return 0, errors.New("invalid authorization header format")
	// }

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		jwtSecret := []byte(utils.GetEnv("JWT_SECRET", "!2E"))
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if userID, ok := claims["userID"].(float64); ok {
			return uint(userID), nil
		}
	}

	return 0, errors.New("userID not found in token")
}

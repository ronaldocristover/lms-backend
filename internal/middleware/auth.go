package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ronaldocristover/lms-backend/pkg/apierror"
	"github.com/ronaldocristover/lms-backend/pkg/response"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, apierror.Unauthorized("Missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(c, apierror.Unauthorized("Invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &struct {
			UserID uuid.UUID `json:"user_id"`
			Email  string    `json:"email"`
			Role   string    `json:"role"`
			jwt.RegisteredClaims
		}{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			response.Error(c, apierror.Unauthorized("Invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*struct {
			UserID uuid.UUID `json:"user_id"`
			Email  string    `json:"email"`
			Role   string    `json:"role"`
			jwt.RegisteredClaims
		})
		if !ok {
			response.Error(c, apierror.Unauthorized("Invalid token claims"))
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

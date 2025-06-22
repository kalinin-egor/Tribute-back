package middleware

import (
	"net/http"
	"strings"
	"tribute-back/internal/infrastructure/auth"
	"tribute-back/internal/interfaces/api/dto"

	"github.com/gin-gonic/gin"
)

// TelegramAuthMiddleware validates the 'Authorization: TgAuth <initData>' header.
func TelegramAuthMiddleware(authService *auth.TelegramAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authorization header is required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "TgAuth" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authorization header format must be 'TgAuth <initData>'"})
			return
		}

		initData := parts[1]
		parsedData, err := authService.Validate(initData)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "Invalid authentication data: " + err.Error()})
			return
		}

		// Set the validated user ID in the context for handlers to use.
		c.Set("userID", parsedData.User.ID)
		c.Next()
	}
}

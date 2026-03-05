package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			fail(c, http.StatusUnauthorized, fmt.Errorf("missing or invalid authorization header"))
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return secret, nil
		})
		if err != nil || !token.Valid {
			fail(c, http.StatusUnauthorized, fmt.Errorf("invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fail(c, http.StatusUnauthorized, fmt.Errorf("invalid token claims"))
			c.Abort()
			return
		}

		userID, _ := claims["user_id"].(float64)
		username, _ := claims["username"].(string)
		c.Set("user_id", int64(userID))
		c.Set("username", username)
		c.Next()
	}
}

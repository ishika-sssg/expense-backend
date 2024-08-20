package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/config"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/models"
)

func RequireAuth(c *gin.Context) {

	// from cookies  way :
	// tokenString, err := c.Cookie("Authorization")

	// from bearer token :
	tokenString := c.GetHeader("Authorization")
	fmt.Println(tokenString)

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization Token is required"})
		c.Abort()
		return
	}

	// Trim the "Bearer " prefix
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validating
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the key for validation (replace with your own key)
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check the exp claim
		if exp, ok := claims["EXP"]; ok {
			if exp == nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has no expiration"})
				c.Abort()
				return
			} else if expFloat64, ok := exp.(float64); ok {
				expTime := time.Unix(int64(expFloat64), 0)
				if time.Now().After(expTime) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
					c.Abort()
					return
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expiration is not a valid number"})
				c.Abort()
				return
			}
		}
		// 	// find the user with token sub
		var user models.User
		// config.DB.First(&user, claims["sub"])
		config.DB.Where("ID=?", claims["sub"]).Find(&user)

		// attach to request
		c.Set("user", user)
		c.Set("user_id", user.ID)

		c.Next()

	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}

}

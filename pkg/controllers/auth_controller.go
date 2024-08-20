package controllers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/config"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/models"

	"time"

	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {
	// get username, email, password from body
	var body struct {
		Name     string
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// check if user already exists :
	var userFound models.User
	config.DB.Where("email=?", body.Email).Find(&userFound)

	if userFound.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{

			"success": false,
			"error":   true,
			"message": "Error: Email already exist",
			"status":  400,
		})
		return
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	// create user
	new_user := models.User{Name: body.Name, Email: body.Email, Password: string(hash)}
	fmt.Println(new_user)
	result := config.DB.Create(&new_user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"message": fmt.Sprintf("Error: %v", result.Error.Error()),
			"status":  400,
		})
		return

	}

	// send response

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"message": "User registered successfully",
		"status":  200,
	})

}

// login controller :
func Login(c *gin.Context) {

	//  get the email, password from req body'
	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"message": "Error occurred",
			"status":  400,
		})
		return
	}

	// look up requested user
	var loggedInUser models.User //this means variable "user" of type models.user
	config.DB.First(&loggedInUser, "email = ?", body.Email)

	if loggedInUser.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   true,
			"message": "Invalid email or password",
			"status":  401,
		})
		return
	}

	// compare sent in password with saved user pasword hash
	err := bcrypt.CompareHashAndPassword([]byte(loggedInUser.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"message": "Invalid email or password",
			"status":  400,
		})
		return
	}
	// generate a JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": loggedInUser.ID,
		"EXP": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   true,
			"message": "Failed to create token",
			"status":  400,
		})
		return
	}

	// send it back

	// setting token in cookies :
	// c.SetSameSite(http.SameSiteLaxMode)
	// c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	// user, _ := c.Get("user")

	// Set Authorization header with the token
	c.Header("Authorization", "Bearer "+tokenString)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"data": gin.H{
			"token": tokenString,
			"user":  loggedInUser,
		},
		"message": "Login successful",
		"status":  200,
	})
}

// middleware controlller :
func Validate(c *gin.Context) {
	// user, _ := c.Get("details")
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"error":   false,
		"data": gin.H{
			"userDetails": user,
		},
		"message": "User details fetched successfully",
		"status":  200,
	})
}

package models

import (
	"github.com/jinzhu/gorm"
)

// var db *gorm.DB

type User struct {
	gorm.Model
	Name     string `gorm:"column:user_name" json:"user_name"`
	Email    string `gorm:"column:email" json:"email"`
	Password string `gorm:"column:password" json:"password"`
}

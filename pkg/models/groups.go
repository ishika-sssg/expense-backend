package models

import (
	"github.com/jinzhu/gorm"
)

type Group struct {
	gorm.Model

	Group_name     string `gorm:"column:group_name" json:"group_name"`
	Description    string `gorm:"column:description" json:"description"`
	Category       string `gorm:"column:category" json:"category"`
	Group_admin_id int    `gorm:"not null" json:"group_admin_id"`         // Foreign key referencing User.ID
	Admin          User   `gorm:"foreignKey:Group_admin_id" json:"admin"` // Association
}

type GroupMembers struct {
	gorm.Model

	GroupId        int    `gorm:"not null"`
	Group          Group  `gorm:"foreignKey:GroupId"`
	MemberId       int    `gorm:"not null"`
	User           User   `gorm:"foreignKey:MemberId"`
	Member_email   string `gorm:"column:member_email" json:"member_email"`
	group_admin_id int    `gorm:"not null"`
}

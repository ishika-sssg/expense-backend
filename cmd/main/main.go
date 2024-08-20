package main

import (
	"log"

	"github.com/ishika-rg/expenseTrackerBackend/pkg/config"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/models"
	"github.com/ishika-rg/expenseTrackerBackend/pkg/routes"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	config.Connect()

	config.DB.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.GroupMembers{},
		&models.Expense{},
		&models.ExpenseShare{},
		&models.Transactions{},
		&models.Settlement{},
	)

	r := routes.SetRouter()

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}

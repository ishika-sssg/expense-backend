package config

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	// "gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func Connect() {
	//    connectfunction opens the connection to database :
	// dsn := "root:Pass@123@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
	// d, err := gorm.Open("mysql", "root:Pass@123@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local")
	dsn := `admin:ishika1234@tcp(database-ig.cz2y4ww0yeis.eu-north-1.rds.amazonaws.com:3306)/mydatabase?charset=utf8mb4&parseTime=True&loc=Local`

	var err error
	// DB, err = gorm.Open("mysql", "root:Pass@123@tcp(localhost:3306)/mydb?parseTime=true")
	DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
}

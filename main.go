package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"qor-admin-3/admin"
)

func main() {

	// Set up the database
	DB, _ := gorm.Open("sqlite3", ":memory:")

	r := gin.New()
	a := admin.New(DB, "", "secret")
	a.Bind(r)
	r.Run("127.0.0.1:8080")

}

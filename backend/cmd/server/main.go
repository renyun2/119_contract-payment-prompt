package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"contractpayment/internal/database"
	"contractpayment/internal/handlers"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	if strings.EqualFold(os.Getenv("GIN_MODE"), "release") {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	handlers.Register(r, db)
	log.Println("listening :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

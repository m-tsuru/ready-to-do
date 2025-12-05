package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/m-tsuru/ready-to-do/internal/database"
	"github.com/m-tsuru/ready-to-do/internal/handler"
	"gorm.io/gorm"
)

func main() {
	var dialector gorm.Dialector

	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "duckdb"
	}

	switch dbType {
	case "duckdb":
		dialector = database.NewDuckDBDialector("database.duckdb")
	default:
		panic("Unsupported DB_TYPE: " + dbType)
	}

	db, err := database.ConnectDatabase(dialector)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	err = database.Init(db)
	if err != nil {
		panic("Failed to initialize database: " + err.Error())
	}

	app := fiber.New()
	handler.Handler(app, db)

	app.Listen(":8080")
}

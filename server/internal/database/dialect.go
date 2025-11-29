package database

import (
	"github.com/alifiroozi80/duckdb"
	"gorm.io/gorm"
)

func ConnectDatabase(dialect gorm.Dialector) (*gorm.DB, error) {
	config := &gorm.Config{}

	db, err := gorm.Open(dialect, config)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}

func NewDuckDBDialector(databasePath string) gorm.Dialector {
	return duckdb.Open(databasePath)
}

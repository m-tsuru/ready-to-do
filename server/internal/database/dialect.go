package database

import (
	"database/sql"

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

func InitDuckDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	queries := []string{
		"CREATE SCHEMA IF NOT EXISTS main",
		"SET schema = 'main'",
	}

	for _, query := range queries {
		if err := sqlDB.QueryRow(query).Err(); err != nil && err != sql.ErrNoRows {
			// エラーを無視（既に存在する場合など）
			continue
		}
	}

	return nil
}

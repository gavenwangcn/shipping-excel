package db

import (
	"os"
	"path/filepath"

	"shipping-excel/backend/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(dataDir string) (*gorm.DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dataDir, "shipping.db")
	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := database.AutoMigrate(&model.Job{}, &model.OutputFile{}, &model.DataRow{}); err != nil {
		return nil, err
	}
	return database, nil
}

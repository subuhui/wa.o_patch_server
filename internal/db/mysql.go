package db

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection and runs auto-migration.
func InitDB(dbType, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch dbType {
	case "mysql":
		dialector = mysql.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate schema
	err = db.AutoMigrate(
		&User{},
		&App{},
		&Channel{},
		&Release{},
		&ReleaseArtifact{},
		&Patch{},
		&PatchArtifact{},
		&PatchEvent{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate database schema: %w", err)
	}

	// Seed default admin user if none exists
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		defaultUser := User{
			Email:                 "admin@subuhui.internal",
			Name:                  "Su Huihui",
			HasActiveSubscription: true,
		}
		if err := db.Create(&defaultUser).Error; err != nil {
			log.Printf("Warning: failed to seed default user: %v", err)
		} else {
			log.Println("Initialized default admin user: admin@subuhui.internal")
		}
	}

	return db, nil
}

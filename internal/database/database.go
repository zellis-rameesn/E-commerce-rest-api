package database

import (
	"fmt"
	"log"
	"time"

	"github.com/zellis-rameesn/go-ecommerce/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// DSN = Data Source Name
	// It is the connection string used to connect to the database.
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()

	if err != nil {
		return nil, fmt.Errorf("failed to configure db: %w", err)
	}

	// setting max connections of db for each api instance for horizontal scaling across multiple api instances
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxIdleTime(20 * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)

	stats := sqlDB.Stats()
	log.Printf("Max connections: %d", stats.MaxOpenConnections)
	log.Printf("Idle connections: %d", stats.Idle)
	return db, nil
}

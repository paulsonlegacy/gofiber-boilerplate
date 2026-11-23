// Package db provides GORM database initialization for PostgreSQL and SQLite.
// Use PostgreSQL for all production environments. SQLite is for local development only.
package db

import (
	"log"
	"os"
	"time"

	"boilerplate/internal/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var gormLogger = logger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags),
	logger.Config{
		SlowThreshold: time.Second,
		LogLevel:      logger.Error,
		Colorful:      true,
	},
)

// NewPostgresDB opens a PostgreSQL connection, configures the connection pool, and runs migrations.
func NewPostgresDB(cfg config.DBConfig) *gorm.DB {
	zap.S().Infow("Opening PostgreSQL connection", "dburl", cfg.DBURL)

	db, err := gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
		Logger:  gormLogger,
	})
	if err != nil {
		zap.L().Fatal("Failed to open PostgreSQL database", zap.Error(err))
	}

	if err := setupConnectionPool(db, cfg); err != nil {
		zap.L().Fatal("Failed to configure DB connection pool", zap.Error(err))
	}
	if err := AutoMigrateAll(db); err != nil {
		zap.L().Fatal("Database migration failed", zap.Error(err))
	}

	zap.L().Info("PostgreSQL database initialized")
	return db
}

// NewSQLiteDB opens a SQLite connection for local development and runs migrations.
// SQLite is limited to a single connection to avoid locking issues.
func NewSQLiteDB(path string) *gorm.DB {
	if path == "" {
		path = "database.db"
	}
	zap.S().Infow("Opening SQLite database", "path", path)

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
		Logger:  gormLogger,
	})
	if err != nil {
		zap.L().Fatal("Failed to open SQLite database", zap.Error(err))
	}

	// SQLite must use a single connection — concurrent writers cause SQLITE_BUSY errors
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	if err := AutoMigrateAll(db); err != nil {
		zap.L().Fatal("Database migration failed", zap.Error(err))
	}

	zap.L().Info("SQLite database initialized")
	return db
}

func setupConnectionPool(db *gorm.DB, cfg config.DBConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifeTime)
	sqlDB.SetConnMaxIdleTime(cfg.MaxIdleTime)
	return nil
}

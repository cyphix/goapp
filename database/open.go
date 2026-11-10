package database

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/cyphix/goapp/config"
)

// Open builds a DSN from cfg and opens the configured database.
// Callers only invoke this when cfg.Type is non-empty.
func Open(cfg *config.Database) (*gorm.DB, error) {
	if cfg.Type != "postgres" {
		return nil, fmt.Errorf("unsupported database type %q", cfg.Type)
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dsnValue(cfg.Host),
		dsnValue(cfg.User),
		dsnValue(cfg.Password),
		dsnValue(cfg.Name),
		cfg.Port,
		dsnValue(cfg.SSLMode),
	)
	if cfg.Schema != "" {
		dsn += " search_path=" + dsnValue(`"`+strings.ReplaceAll(cfg.Schema, `"`, `""`)+`"`)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("opening gorm connection: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("retrieving sql connection: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return db, nil
}

// dsnValue quotes a lib/pq keyword DSN value so values containing spaces,
// quotes or backslashes (e.g. passwords) stay a single token.
func dsnValue(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s) + "'"
}

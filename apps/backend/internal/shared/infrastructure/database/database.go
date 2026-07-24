// Package database provides the backend's Postgres connection via GORM.
package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connect opens a GORM connection to dsn. Callers decide whether a failure
// is fatal — Charli's audit trail degrades to log-only when there's no
// reachable database, it doesn't refuse to start.
func Connect(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
}

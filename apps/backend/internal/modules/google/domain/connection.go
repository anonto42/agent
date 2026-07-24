// Package domain models a browser installation's connection to Google
// (L4): the OAuth tokens needed to act on the user's behalf against Google
// APIs.
package domain

import "time"

// Connection is one device's OAuth connection to Google. Unlike an audit
// Entry, this is not append-only — tokens get refreshed in place, so
// UpdatedAt is meaningful here.
type Connection struct {
	ID           uint `gorm:"primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeviceID     string `gorm:"uniqueIndex"`
	AccessToken  string
	RefreshToken string
	TokenExpiry  time.Time
}

// Repository persists Google connections, one per device.
type Repository interface {
	Save(Connection) error
	FindByDevice(deviceID string) (*Connection, bool, error)
}

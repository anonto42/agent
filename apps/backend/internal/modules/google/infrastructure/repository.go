// Package infrastructure implements the google domain's repository via GORM.
package infrastructure

import (
	"gorm.io/gorm"

	"github.com/levelaxis/charli/backend/internal/modules/google/domain"
)

// GormRepository persists Google connections to Postgres via GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository builds a Repository backed by db, migrating the
// connections table if needed.
func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if err := db.AutoMigrate(&domain.Connection{}); err != nil {
		return nil, err
	}
	return &GormRepository{db: db}, nil
}

// Save creates or updates the connection for c.DeviceID (upsert on the
// unique device_id index — a device has at most one connection).
func (r *GormRepository) Save(c domain.Connection) error {
	var existing domain.Connection
	err := r.db.Where("device_id = ?", c.DeviceID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(&c).Error
	}
	if err != nil {
		return err
	}
	c.ID = existing.ID
	return r.db.Save(&c).Error
}

// FindByDevice looks up the connection for deviceID, if any.
func (r *GormRepository) FindByDevice(deviceID string) (*domain.Connection, bool, error) {
	var c domain.Connection
	err := r.db.Where("device_id = ?", deviceID).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &c, true, nil
}

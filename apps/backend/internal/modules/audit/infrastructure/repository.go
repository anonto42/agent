// Package infrastructure implements the audit domain's repository via GORM.
package infrastructure

import (
	"gorm.io/gorm"

	"github.com/levelaxis/charli/backend/internal/modules/audit/domain"
)

// GormRepository persists audit entries to Postgres via GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository builds a Repository backed by db, migrating the audit
// table if needed.
func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if err := db.AutoMigrate(&domain.Entry{}); err != nil {
		return nil, err
	}
	return &GormRepository{db: db}, nil
}

// Record inserts e.
func (r *GormRepository) Record(e domain.Entry) error {
	return r.db.Create(&e).Error
}

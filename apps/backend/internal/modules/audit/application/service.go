// Package application records audit entries.
package application

import (
	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/modules/audit/domain"
)

// Service records audit entries. A nil repo means no database is
// configured — Record becomes a no-op rather than an error, since the
// existing zap "audit: ..." log line at each call site remains the source
// of truth either way; this only adds queryability.
type Service struct {
	repo domain.Repository
	log  *zap.Logger
}

// NewService builds a Service. repo may be nil (see above).
func NewService(repo domain.Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Record persists e. A failure never propagates to the caller — a broken
// database must not take down a chat request — it's only logged.
func (s *Service) Record(e domain.Entry) {
	if s.repo == nil {
		return
	}
	if err := s.repo.Record(e); err != nil {
		s.log.Warn("audit: failed to persist entry", zap.Error(err))
	}
}

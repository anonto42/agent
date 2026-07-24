package application

import (
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/modules/audit/domain"
)

type fakeRepo struct {
	entries []domain.Entry
	err     error
}

func (f *fakeRepo) Record(e domain.Entry) error {
	if f.err != nil {
		return f.err
	}
	f.entries = append(f.entries, e)
	return nil
}

func TestRecordForwardsToRepo(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, zap.NewNop())

	svc.Record(domain.Entry{Session: "s1", Outcome: domain.OutcomeExecuted})

	if len(repo.entries) != 1 || repo.entries[0].Outcome != domain.OutcomeExecuted {
		t.Fatalf("expected the entry to reach the repo, got %+v", repo.entries)
	}
}

func TestRecordNoopsWithoutRepo(t *testing.T) {
	svc := NewService(nil, zap.NewNop())

	// Must not panic — this is the "no database configured" case.
	svc.Record(domain.Entry{Session: "s1", Outcome: domain.OutcomeExecuted})
}

func TestRecordSwallowsRepoErrors(t *testing.T) {
	repo := &fakeRepo{err: errors.New("db unavailable")}
	svc := NewService(repo, zap.NewNop())

	// Must not panic or propagate — a broken database must never take down
	// a chat request.
	svc.Record(domain.Entry{Session: "s1", Outcome: domain.OutcomeExecuted})
}

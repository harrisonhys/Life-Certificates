package repository

import (
	"context"
	"fmt"
	"time"

	"life-certificates/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FRIdentityRepository manages FR label mappings.
type FRIdentityRepository interface {
	Create(ctx context.Context, identity *domain.FRIdentity) error
	GetByLabel(ctx context.Context, label string) (*domain.FRIdentity, error)
	DeleteByParticipantID(ctx context.Context, participantID string) error
}

type frIdentityRepository struct {
	db *gorm.DB
}

// NewFRIdentityRepository creates a repository instance.
func NewFRIdentityRepository(db *gorm.DB) FRIdentityRepository {
	return &frIdentityRepository{db: db}
}

func (r *frIdentityRepository) Create(ctx context.Context, identity *domain.FRIdentity) error {
	if identity.CreatedAt.IsZero() {
		identity.CreatedAt = time.Now().UTC()
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(identity).Error; err != nil {
		return fmt.Errorf("create fr identity: %w", err)
	}
	return nil
}

func (r *frIdentityRepository) GetByLabel(ctx context.Context, label string) (*domain.FRIdentity, error) {
	var identity domain.FRIdentity
	if err := r.db.WithContext(ctx).First(&identity, "label = ?", label).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get fr identity by label: %w", err)
	}
	return &identity, nil
}

func (r *frIdentityRepository) DeleteByParticipantID(ctx context.Context, participantID string) error {
	if err := r.db.WithContext(ctx).Where("participant_id = ?", participantID).Delete(&domain.FRIdentity{}).Error; err != nil {
		return fmt.Errorf("delete fr identity: %w", err)
	}
	return nil
}

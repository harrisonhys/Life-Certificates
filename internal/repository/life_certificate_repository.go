package repository

import (
	"context"
	"fmt"

	"life-certificates/internal/domain"

	"gorm.io/gorm"
)

// LifeCertificateRepository exposes persistence for verification attempts.
type LifeCertificateRepository interface {
	Create(ctx context.Context, record *domain.LifeCertificate) error
	GetLatestByParticipant(ctx context.Context, participantID string) (*domain.LifeCertificate, error)
	DeleteByParticipant(ctx context.Context, participantID string) error
}

type lifeCertificateRepository struct {
	db *gorm.DB
}

// NewLifeCertificateRepository instantiates a gorm-backed implementation.
func NewLifeCertificateRepository(db *gorm.DB) LifeCertificateRepository {
	return &lifeCertificateRepository{db: db}
}

func (r *lifeCertificateRepository) Create(ctx context.Context, record *domain.LifeCertificate) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("create life certificate: %w", err)
	}
	return nil
}

func (r *lifeCertificateRepository) GetLatestByParticipant(ctx context.Context, participantID string) (*domain.LifeCertificate, error) {
	var record domain.LifeCertificate
	if err := r.db.WithContext(ctx).
		Where("participant_id = ?", participantID).
		Order("verified_at desc").
		First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest life certificate: %w", err)
	}
	return &record, nil
}

func (r *lifeCertificateRepository) DeleteByParticipant(ctx context.Context, participantID string) error {
	if err := r.db.WithContext(ctx).Where("participant_id = ?", participantID).Delete(&domain.LifeCertificate{}).Error; err != nil {
		return fmt.Errorf("delete life certificates: %w", err)
	}
	return nil
}

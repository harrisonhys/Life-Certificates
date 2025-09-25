package repository

import (
	"context"
	"fmt"

	"life-certificates/internal/domain"

	"gorm.io/gorm"
)

// ParticipantRepository defines persistence operations for participants.
type ParticipantRepository interface {
	Create(ctx context.Context, participant *domain.Participant) error
	GetByID(ctx context.Context, id string) (*domain.Participant, error)
	GetByNIK(ctx context.Context, nik string) (*domain.Participant, error)
	List(ctx context.Context) ([]domain.Participant, error)
	Update(ctx context.Context, participant *domain.Participant) error
	Delete(ctx context.Context, id string) error
}

type participantRepository struct {
	db *gorm.DB
}

// NewParticipantRepository creates a gorm-backed repository.
func NewParticipantRepository(db *gorm.DB) ParticipantRepository {
	return &participantRepository{db: db}
}

func (r *participantRepository) Create(ctx context.Context, participant *domain.Participant) error {
	if err := r.db.WithContext(ctx).Create(participant).Error; err != nil {
		return fmt.Errorf("create participant: %w", err)
	}
	return nil
}

func (r *participantRepository) GetByID(ctx context.Context, id string) (*domain.Participant, error) {
	var participant domain.Participant
	if err := r.db.WithContext(ctx).First(&participant, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get participant by id: %w", err)
	}
	return &participant, nil
}

func (r *participantRepository) GetByNIK(ctx context.Context, nik string) (*domain.Participant, error) {
	var participant domain.Participant
	if err := r.db.WithContext(ctx).First(&participant, "nik = ?", nik).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get participant by nik: %w", err)
	}
	return &participant, nil
}

func (r *participantRepository) List(ctx context.Context) ([]domain.Participant, error) {
	var participants []domain.Participant
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&participants).Error; err != nil {
		return nil, fmt.Errorf("list participants: %w", err)
	}
	return participants, nil
}

func (r *participantRepository) Update(ctx context.Context, participant *domain.Participant) error {
	if err := r.db.WithContext(ctx).Model(&domain.Participant{}).Where("id = ?", participant.ID).Updates(map[string]interface{}{
		"nik":        participant.NIK,
		"name":       participant.Name,
		"updated_at": participant.UpdatedAt,
	}).Error; err != nil {
		return fmt.Errorf("update participant: %w", err)
	}
	return nil
}

func (r *participantRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Participant{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete participant: %w", err)
	}
	return nil
}

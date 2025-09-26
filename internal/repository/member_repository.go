package repository

import (
	"context"
	"fmt"

	"life-certificates/internal/domain"

	"gorm.io/gorm"
)

// MemberRepository defines persistence operations for members.
type MemberRepository interface {
	Create(ctx context.Context, member *domain.Member) error
	GetByID(ctx context.Context, id string) (*domain.Member, error)
	GetByNIK(ctx context.Context, nik string) (*domain.Member, error)
	GetByNomorPeserta(ctx context.Context, nomorPeserta string) (*domain.Member, error)
	List(ctx context.Context) ([]domain.Member, error)
	Update(ctx context.Context, member *domain.Member) error
	Delete(ctx context.Context, id string) error
}

type memberRepository struct {
	db *gorm.DB
}

// NewMemberRepository creates a gorm-backed repository.
func NewMemberRepository(db *gorm.DB) MemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) Create(ctx context.Context, member *domain.Member) error {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("create member: %w", err)
	}
	return nil
}

func (r *memberRepository) GetByID(ctx context.Context, id string) (*domain.Member, error) {
	var member domain.Member
	if err := r.db.WithContext(ctx).First(&member, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get member by id: %w", err)
	}
	return &member, nil
}

func (r *memberRepository) GetByNIK(ctx context.Context, nik string) (*domain.Member, error) {
	var member domain.Member
	if err := r.db.WithContext(ctx).First(&member, "nik = ?", nik).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get member by nik: %w", err)
	}
	return &member, nil
}

func (r *memberRepository) GetByNomorPeserta(ctx context.Context, nomorPeserta string) (*domain.Member, error) {
	var member domain.Member
	if err := r.db.WithContext(ctx).First(&member, "nomor_peserta = ?", nomorPeserta).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get member by nomor peserta: %w", err)
	}
	return &member, nil
}

func (r *memberRepository) List(ctx context.Context) ([]domain.Member, error) {
	var members []domain.Member
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&members).Error; err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	return members, nil
}

func (r *memberRepository) Update(ctx context.Context, member *domain.Member) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Member{}).
		Where("id = ?", member.ID).
		Updates(map[string]interface{}{
			"nik":           member.NIK,
			"nomor_peserta": member.NomorPeserta,
			"birth_date":    member.BirthDate,
			"fullname":      member.FullName,
			"address":       member.Address,
			"city":          member.City,
			"province":      member.Province,
			"phone_number":  member.PhoneNumber,
			"email":         member.Email,
			"updated_at":    member.UpdatedAt,
		}).Error; err != nil {
		return fmt.Errorf("update member: %w", err)
	}
	return nil
}

func (r *memberRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Member{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete member: %w", err)
	}
	return nil
}

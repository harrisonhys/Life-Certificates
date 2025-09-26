package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"life-certificates/internal/domain"
	"life-certificates/internal/repository"
)

var (
	// ErrMemberNotFound indicates the requested member does not exist.
	ErrMemberNotFound = errors.New("member not found")
	// ErrMemberNIKExists signals that the requested NIK is already registered.
	ErrMemberNIKExists = errors.New("member with nik already exists")
	// ErrMemberNomorPesertaExists signals that the nomor peserta is already registered.
	ErrMemberNomorPesertaExists = errors.New("member with nomor peserta already exists")
)

// MemberService provides CRUD operations for members.
type MemberService struct {
	members repository.MemberRepository
}

// NewMemberService wires the required dependencies.
func NewMemberService(members repository.MemberRepository) *MemberService {
	return &MemberService{members: members}
}

// CreateMemberInput carries the payload required to create a member.
type CreateMemberInput struct {
	NIK          string `json:"nik"`
	NomorPeserta string `json:"nomor_peserta"`
	BirthDate    string `json:"birth_date"`
	FullName     string `json:"fullname"`
	Address      string `json:"address"`
	City         string `json:"city"`
	Province     string `json:"province"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
}

// UpdateMemberInput captures optional member fields for update operations.
type UpdateMemberInput struct {
	NIK          *string `json:"nik"`
	NomorPeserta *string `json:"nomor_peserta"`
	BirthDate    *string `json:"birth_date"`
	FullName     *string `json:"fullname"`
	Address      *string `json:"address"`
	City         *string `json:"city"`
	Province     *string `json:"province"`
	PhoneNumber  *string `json:"phone_number"`
	Email        *string `json:"email"`
}

// Create inserts a new member into the repository.
func (s *MemberService) Create(ctx context.Context, input CreateMemberInput) (*domain.Member, error) {
	nik := strings.TrimSpace(input.NIK)
	nomorPeserta := strings.TrimSpace(input.NomorPeserta)
	fullName := strings.TrimSpace(input.FullName)
	birthDateRaw := strings.TrimSpace(input.BirthDate)

	if nik == "" {
		return nil, fmt.Errorf("nik is required")
	}
	if nomorPeserta == "" {
		return nil, fmt.Errorf("nomor_peserta is required")
	}
	if fullName == "" {
		return nil, fmt.Errorf("fullname is required")
	}
	if birthDateRaw == "" {
		return nil, fmt.Errorf("birth_date is required")
	}

	birthDate, err := time.Parse("2006-01-02", birthDateRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid birth_date format, use YYYY-MM-DD")
	}

	existingByNIK, err := s.members.GetByNIK(ctx, nik)
	if err != nil {
		return nil, err
	}
	if existingByNIK != nil {
		return nil, ErrMemberNIKExists
	}

	existingByNomor, err := s.members.GetByNomorPeserta(ctx, nomorPeserta)
	if err != nil {
		return nil, err
	}
	if existingByNomor != nil {
		return nil, ErrMemberNomorPesertaExists
	}

	now := time.Now().UTC()
	member := &domain.Member{
		ID:           uuid.NewString(),
		NIK:          nik,
		NomorPeserta: nomorPeserta,
		BirthDate:    birthDate,
		FullName:     fullName,
		Address:      strings.TrimSpace(input.Address),
		City:         strings.TrimSpace(input.City),
		Province:     strings.TrimSpace(input.Province),
		PhoneNumber:  strings.TrimSpace(input.PhoneNumber),
		Email:        strings.TrimSpace(input.Email),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.members.Create(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// List returns all registered members ordered by creation date desc.
func (s *MemberService) List(ctx context.Context) ([]domain.Member, error) {
	return s.members.List(ctx)
}

// Get fetches a member by its identifier.
func (s *MemberService) Get(ctx context.Context, id string) (*domain.Member, error) {
	member, err := s.members.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrMemberNotFound
	}
	return member, nil
}

// Update applies modifications to an existing member.
func (s *MemberService) Update(ctx context.Context, id string, input UpdateMemberInput) (*domain.Member, error) {
	member, err := s.members.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrMemberNotFound
	}

	if input.NIK != nil {
		newNIK := strings.TrimSpace(*input.NIK)
		if newNIK == "" {
			return nil, fmt.Errorf("nik cannot be empty")
		}
		if newNIK != member.NIK {
			existing, err := s.members.GetByNIK(ctx, newNIK)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != member.ID {
				return nil, ErrMemberNIKExists
			}
		}
		member.NIK = newNIK
	}

	if input.NomorPeserta != nil {
		newNomor := strings.TrimSpace(*input.NomorPeserta)
		if newNomor == "" {
			return nil, fmt.Errorf("nomor_peserta cannot be empty")
		}
		if newNomor != member.NomorPeserta {
			existing, err := s.members.GetByNomorPeserta(ctx, newNomor)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != member.ID {
				return nil, ErrMemberNomorPesertaExists
			}
		}
		member.NomorPeserta = newNomor
	}

	if input.BirthDate != nil {
		birthDateRaw := strings.TrimSpace(*input.BirthDate)
		if birthDateRaw == "" {
			return nil, fmt.Errorf("birth_date cannot be empty")
		}
		birthDate, err := time.Parse("2006-01-02", birthDateRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid birth_date format, use YYYY-MM-DD")
		}
		member.BirthDate = birthDate
	}

	if input.FullName != nil {
		newFullName := strings.TrimSpace(*input.FullName)
		if newFullName == "" {
			return nil, fmt.Errorf("fullname cannot be empty")
		}
		member.FullName = newFullName
	}

	if input.Address != nil {
		member.Address = strings.TrimSpace(*input.Address)
	}
	if input.City != nil {
		member.City = strings.TrimSpace(*input.City)
	}
	if input.Province != nil {
		member.Province = strings.TrimSpace(*input.Province)
	}
	if input.PhoneNumber != nil {
		member.PhoneNumber = strings.TrimSpace(*input.PhoneNumber)
	}
	if input.Email != nil {
		member.Email = strings.TrimSpace(*input.Email)
	}

	member.UpdatedAt = time.Now().UTC()

	if err := s.members.Update(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// Delete removes a member from the repository.
func (s *MemberService) Delete(ctx context.Context, id string) error {
	member, err := s.members.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrMemberNotFound
	}

	return s.members.Delete(ctx, id)
}

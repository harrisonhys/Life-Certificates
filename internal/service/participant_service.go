package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"life-certificates/internal/domain"
	"life-certificates/internal/frcore"
	"life-certificates/internal/repository"
)

// Domain level errors used by handlers for precise status codes.
var (
	ErrParticipantExists   = errors.New("participant with nik already exists")
	ErrParticipantNotFound = errors.New("participant not found")
)

// ParticipantService provides registration operations.
type ParticipantService struct {
	participants repository.ParticipantRepository
	frIdentities repository.FRIdentityRepository
	frClient     frcore.Client
	certificates repository.LifeCertificateRepository
}

// RegisterInput contains the payload required to register a participant.
type RegisterInput struct {
	NIK       string
	Name      string
	Image     []byte
	ImageName string
}

// RegisterOutput returns identifiers produced during registration.
type RegisterOutput struct {
	ParticipantID string
	FRRef         string
	FRExternalRef string
}

// NewParticipantService wires dependencies for participant registration.
func NewParticipantService(participants repository.ParticipantRepository, frIdentities repository.FRIdentityRepository, certificates repository.LifeCertificateRepository, frClient frcore.Client) *ParticipantService {
	return &ParticipantService{
		participants: participants,
		frIdentities: frIdentities,
		frClient:     frClient,
		certificates: certificates,
	}
}

// Register registers a new participant and links them with FR Core.
func (s *ParticipantService) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if strings.TrimSpace(input.NIK) == "" {
		return nil, fmt.Errorf("nik is required")
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if len(input.Image) == 0 {
		return nil, fmt.Errorf("image is required")
	}

	existing, err := s.participants.GetByNIK(ctx, input.NIK)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrParticipantExists
	}

	participantID := uuid.NewString()
	imageName := input.ImageName
	if strings.TrimSpace(imageName) == "" {
		imageName = "registration.jpg"
	}

	frLabel := uuid.NewString()
	frExternalRef := participantID
	uploadResp, err := s.frClient.UploadFace(ctx, frcore.UploadRequest{
		Label:       frLabel,
		ExternalRef: frExternalRef,
		ImageName:   imageName,
		Image:       input.Image,
	})
	if err != nil {
		return nil, err
	}

	frRef := uploadResp.Label
	if strings.TrimSpace(frRef) == "" {
		frRef = uploadResp.ID
	}
	if strings.TrimSpace(frRef) == "" {
		frRef = frLabel
	}
	frExternal := uploadResp.ExternalRef
	if strings.TrimSpace(frExternal) == "" {
		frExternal = frExternalRef
	}

	now := time.Now().UTC()
	participant := &domain.Participant{
		ID:            participantID,
		NIK:           strings.TrimSpace(input.NIK),
		Name:          strings.TrimSpace(input.Name),
		FRLabel:       frRef,
		FRExternalRef: frExternal,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.participants.Create(ctx, participant); err != nil {
		return nil, err
	}

	if err := s.frIdentities.Create(ctx, &domain.FRIdentity{
		Label:         frRef,
		ParticipantID: participant.ID,
		ExternalRef:   frExternal,
	}); err != nil {
		return nil, err
	}

	return &RegisterOutput{ParticipantID: participant.ID, FRRef: participant.FRLabel, FRExternalRef: participant.FRExternalRef}, nil
}

// List returns all participants ordered by creation date desc.
func (s *ParticipantService) List(ctx context.Context) ([]domain.Participant, error) {
	return s.participants.List(ctx)
}

// Get returns a participant by ID.
func (s *ParticipantService) Get(ctx context.Context, id string) (*domain.Participant, error) {
	participant, err := s.participants.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, ErrParticipantNotFound
	}
	return participant, nil
}

// UpdateParticipantInput captures mutable participant fields.
type UpdateParticipantInput struct {
	NIK  string `json:"nik"`
	Name string `json:"name"`
}

// Update modifies participant metadata.
func (s *ParticipantService) Update(ctx context.Context, id string, input UpdateParticipantInput) (*domain.Participant, error) {
	participant, err := s.participants.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, ErrParticipantNotFound
	}

	newNIK := strings.TrimSpace(input.NIK)
	newName := strings.TrimSpace(input.Name)

	if newNIK == "" {
		newNIK = participant.NIK
	}
	if newName == "" {
		newName = participant.Name
	}

	if newNIK != participant.NIK {
		existing, err := s.participants.GetByNIK(ctx, newNIK)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != participant.ID {
			return nil, ErrParticipantExists
		}
	}

	participant.NIK = newNIK
	participant.Name = newName
	participant.UpdatedAt = time.Now().UTC()

	if err := s.participants.Update(ctx, participant); err != nil {
		return nil, err
	}

	return participant, nil
}

// Delete removes a participant and related records.
func (s *ParticipantService) Delete(ctx context.Context, id string) error {
	participant, err := s.participants.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if participant == nil {
		return ErrParticipantNotFound
	}

	if err := s.certificates.DeleteByParticipant(ctx, id); err != nil {
		return err
	}
	if err := s.frIdentities.DeleteByParticipantID(ctx, id); err != nil {
		return err
	}

	return s.participants.Delete(ctx, id)
}

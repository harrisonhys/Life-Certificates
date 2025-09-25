package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"life-certificates/internal/domain"
	"life-certificates/internal/frcore"
	"life-certificates/internal/liveness"
	"life-certificates/internal/repository"
)

// VerificationService coordinates life certificate verification flows.
type VerificationService struct {
	participants        repository.ParticipantRepository
	certificates        repository.LifeCertificateRepository
	frIdentities        repository.FRIdentityRepository
	frClient            frcore.Client
	livenessChecker     liveness.Checker
	distanceThreshold   float64
	similarityThreshold float64
}

// VerifyInput captures the payload for a verification attempt.
type VerifyInput struct {
	ParticipantID    string
	ImageBytes       []byte
	OriginalFilename string
}

// VerifyOutput contains persisted verification metadata.
type VerifyOutput struct {
	ParticipantID string
	Status        domain.LifeCertificateStatus
	Distance      *float64
	Similarity    *float64
	VerifiedAt    time.Time
}

// StatusOutput returns the latest verification record.
type StatusOutput struct {
	ParticipantID string
	Status        domain.LifeCertificateStatus
	Distance      *float64
	Similarity    *float64
	VerifiedAt    *time.Time
	SelfiePath    string
}

// NewVerificationService wires dependencies for verification flows.
func NewVerificationService(participants repository.ParticipantRepository, certificates repository.LifeCertificateRepository, frIdentities repository.FRIdentityRepository, frClient frcore.Client, checker liveness.Checker, distanceThreshold, similarityThreshold float64) *VerificationService {
	return &VerificationService{
		participants:        participants,
		certificates:        certificates,
		frIdentities:        frIdentities,
		frClient:            frClient,
		livenessChecker:     checker,
		distanceThreshold:   distanceThreshold,
		similarityThreshold: similarityThreshold,
	}
}

// Verify processes a life certificate submission from a participant.
func (s *VerificationService) Verify(ctx context.Context, input VerifyInput) (*VerifyOutput, error) {
	participantID := strings.TrimSpace(input.ParticipantID)
	if participantID == "" {
		return nil, fmt.Errorf("participant_id is required")
	}
	if len(input.ImageBytes) == 0 {
		return nil, fmt.Errorf("image payload is required")
	}

	participant, err := s.participants.GetByID(ctx, participantID)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, ErrParticipantNotFound
	}

	filename := input.OriginalFilename
	if filename == "" {
		filename = "verification.jpg"
	}

	now := time.Now().UTC()

	passed, reason, err := s.livenessChecker.Evaluate(ctx, input.ImageBytes)
	if err != nil {
		return nil, fmt.Errorf("liveness evaluation failed: %w", err)
	}

	if !passed {
		notes := reason
		record := &domain.LifeCertificate{
			ID:            uuid.NewString(),
			ParticipantID: participant.ID,
			SelfiePath:    "",
			Status:        domain.LifeCertificateStatusReview,
			VerifiedAt:    now,
			Notes:         &notes,
		}
		if err := s.certificates.Create(ctx, record); err != nil {
			return nil, err
		}
		return &VerifyOutput{
			ParticipantID: participant.ID,
			Status:        domain.LifeCertificateStatusReview,
			VerifiedAt:    now,
		}, nil
	}

	recognizeResp, err := s.frClient.Recognize(ctx, frcore.RecognizeRequest{
		ImageName: filename,
		Image:     input.ImageBytes,
	})
	if err != nil {
		return nil, err
	}

	status := domain.LifeCertificateStatusInvalid
	distanceOk := false
	if recognizeResp.Distance != nil {
		distanceOk = *recognizeResp.Distance <= s.distanceThreshold
	}
	similarityOk := recognizeResp.Similarity >= s.similarityThreshold

	matchLabel := false
	label := strings.TrimSpace(recognizeResp.Label)
	if label != "" {
		identity, err := s.frIdentities.GetByLabel(ctx, label)
		if err != nil {
			return nil, err
		}
		if identity != nil {
			matchLabel = identity.ParticipantID == participant.ID
		} else if similarityOk && (recognizeResp.Distance == nil || distanceOk) {
			// New alias detected with high confidence – associate label with participant for future lookups.
			_ = s.frIdentities.Create(ctx, &domain.FRIdentity{
				Label:         label,
				ParticipantID: participant.ID,
				ExternalRef:   participant.FRExternalRef,
			})
			matchLabel = true
		}
	}

	if matchLabel && (distanceOk || (!distanceOk && recognizeResp.Distance == nil && similarityOk)) {
		status = domain.LifeCertificateStatusValid
	}

	similarity := recognizeResp.Similarity
	record := &domain.LifeCertificate{
		ID:            uuid.NewString(),
		ParticipantID: participant.ID,
		SelfiePath:    "",
		Status:        status,
		Distance:      recognizeResp.Distance,
		Similarity:    &similarity,
		VerifiedAt:    now,
	}

	if err := s.certificates.Create(ctx, record); err != nil {
		return nil, err
	}

	return &VerifyOutput{
		ParticipantID: participant.ID,
		Status:        status,
		Distance:      recognizeResp.Distance,
		Similarity:    &similarity,
		VerifiedAt:    now,
	}, nil
}

// LatestStatus returns the most recent verification record for the participant.
func (s *VerificationService) LatestStatus(ctx context.Context, participantID string) (*StatusOutput, error) {
	participantID = strings.TrimSpace(participantID)
	if participantID == "" {
		return nil, fmt.Errorf("participant_id is required")
	}

	participant, err := s.participants.GetByID(ctx, participantID)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, ErrParticipantNotFound
	}

	record, err := s.certificates.GetLatestByParticipant(ctx, participantID)
	if err != nil {
		return nil, err
	}

	if record == nil {
		return &StatusOutput{ParticipantID: participantID}, nil
	}

	return &StatusOutput{
		ParticipantID: participantID,
		Status:        record.Status,
		Distance:      record.Distance,
		Similarity:    record.Similarity,
		VerifiedAt:    &record.VerifiedAt,
		SelfiePath:    record.SelfiePath,
	}, nil
}

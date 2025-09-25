package domain

import "time"

// LifeCertificateStatus captures the verification outcome stored per attempt.
type LifeCertificateStatus string

const (
	LifeCertificateStatusValid   LifeCertificateStatus = "VALID"
	LifeCertificateStatusInvalid LifeCertificateStatus = "INVALID"
	LifeCertificateStatusReview  LifeCertificateStatus = "REVIEW"
)

// Participant represents a pension participant tracked by the service.
type Participant struct {
	ID            string    `gorm:"type:char(36);primaryKey" json:"participant_id"`
	NIK           string    `gorm:"size:20;uniqueIndex" json:"nik"`
	Name          string    `gorm:"size:100" json:"name"`
	FRLabel       string    `gorm:"column:fr_label;size:64;uniqueIndex" json:"fr_label"`
	FRExternalRef string    `gorm:"column:fr_external_ref;size:64;uniqueIndex" json:"fr_external_ref"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// LifeCertificate represents a single verification attempt.
type LifeCertificate struct {
	ID            string                `gorm:"type:char(36);primaryKey" json:"id"`
	ParticipantID string                `gorm:"type:char(36);index" json:"participant_id"`
	SelfiePath    string                `gorm:"type:text" json:"selfie_path"`
	Status        LifeCertificateStatus `gorm:"type:varchar(16)" json:"status"`
	Distance      *float64              `json:"distance"`
	Similarity    *float64              `json:"similarity"`
	VerifiedAt    time.Time             `json:"verified_at"`
	Notes         *string               `json:"notes"`
}

// TableName overrides gorm pluralisation for consistency.
func (LifeCertificate) TableName() string {
	return "life_certificate"
}

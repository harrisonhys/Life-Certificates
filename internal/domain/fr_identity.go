package domain

import "time"

// FRIdentity maps FR Core labels to participants for verification.
type FRIdentity struct {
	Label         string    `gorm:"primaryKey;size:128" json:"label"`
	ParticipantID string    `gorm:"type:char(36);index" json:"participant_id"`
	ExternalRef   string    `gorm:"size:128" json:"external_ref"`
	CreatedAt     time.Time `json:"created_at"`
}
